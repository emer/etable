#!/usr/local/bin/pyleabra -i

# Copyright (c) 2020, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# this tests transferring data between python and etable data.
# we're using the pyleabra gopy executable, built in emer/leabra/python

from leabra import go, etable, efile, split, etensor, etview, rand, erand, patgen, gi, giv, pygiv, mat32

import pyet

import io, sys, getopt
import numpy as np
import pandas as pd
import torch
import torch.utils.data as data_utils

# this will become Sim later.. 
TheSim = 1

def TestCB(recv, send, sig, data):
    TheSim.Test()
    TheSim.UpdateClassView()
    TheSim.vp.SetNeedsFullRender()

class Sim(pygiv.ClassViewObj):
    """
    Sim encapsulates the entire simulation model, and we define all the
    functionality as methods on this struct.  This structure keeps all relevant
    state information organized and available without having to pass everything around
    as arguments to methods, and provides the core GUI interface (note the view tags
    for the fields which provide hints to how things should be displayed).
    """

    def __init__(self):
        super(Sim, self).__init__()
        self.Pats = etable.Table()
        self.SetTags("Pats", 'view:"no-inline" desc:"test patterns"')

        self.PatsTable = 0
        self.SetTags("PatsTable", 'view:"-" desc:"view"')

        # statistics: note use float64 as that is best for etable.Table
        self.Win = 0
        self.SetTags("Win", 'view:"-" desc:"main GUI window"')
        self.ToolBar = 0
        self.SetTags("ToolBar", 'view:"-" desc:"the master toolbar"')
        self.vp  = 0
        self.SetTags("vp", 'view:"-" desc:"viewport"')
        
    def Config(ss):
        ss.ConfigPats()
        
    def ConfigPats(ss):
        dt = ss.Pats
        sch = etable.Schema(
            [etable.Column("Name", etensor.STRING, go.nil, go.nil),
            etable.Column("Input", etensor.FLOAT32, go.Slice_int([4, 5]), go.Slice_string(["Y", "X"])),
            etable.Column("Output", etensor.FLOAT32, go.Slice_int([4, 5]), go.Slice_string(["Y", "X"]))]
        )
        dt.SetFromSchema(sch, 3)
        patgen.PermutedBinaryRows(dt.Cols[1], 6, 1, 0)
        patgen.PermutedBinaryRows(dt.Cols[2], 6, 1, 0)
        cn = etensor.String(dt.Cols[0])
        cn.Values.copy(["any", "baker", "cheese"])

    def Numpy(ss):
        """
        test conversions to / from numpy
        """
        dt = ss.Pats
        
        etf = etensor.Float32(dt.Cols[1])
        npf = pyet.etensor_to_numpy(etf)
        print(npf)
        ctf = pyet.numpy_to_etensor(npf)
        print(ctf)
        
        etu32 = etensor.NewUint32(go.Slice_int([3,4,5]), go.nil, go.nil)
        sz = etf.Len()
        for i in range(sz):
            etu32.Values[i] = int(etf.Values[i])
        print(etu32)
        npu32 = pyet.etensor_to_numpy(etu32)
        print(npu32)
        ctu32 = pyet.numpy_to_etensor(npu32)
        print(ctu32)
        
        ets = etensor.String(dt.Cols[0])
        nps = pyet.etensor_to_numpy(ets)
        print(nps)
        cts = pyet.numpy_to_etensor(nps)
        print(cts)
        
        ets = etensor.String(dt.Cols[0])
        nps = pyet.etensor_to_numpy(ets)
        print(nps)
        cts = pyet.numpy_to_etensor(nps)
        print(cts)
        
        etb = etensor.NewBits(go.Slice_int([3,4,5]), go.nil, go.nil)
        sz = etb.Len()
        for i in range(sz):
            etb.Set1D(i, erand.BoolProb(.2, -1))
        print(etb)
        npb = pyet.etensor_to_numpy(etb)
        print(npb)
        ctb = pyet.numpy_to_etensor(npb)
        print(ctb)
        
    def Torch(ss):
        """
        test conversions to torch
        """
        dt = ss.Pats
        pdt = pyet.etable_to_py(dt)
        print(pdt)
        ttd = pyet.etable_to_torch(pdt)
        print(ttd)

    def Test(ss):
        ss.Numpy()
        ss.Torch()
        
    def ConfigGui(ss):
        """
        ConfigGui configures the GoGi gui interface for this simulation,
        """
        width = 1600
        height = 1200

        gi.SetAppName("epyarrow")
        gi.SetAppAbout('testing of using arrow to port data between Go and Python. See <a href="https://github.com/emer/etable/blob/master/examples/pyarrow/README.md">README.md on GitHub</a>.</p>')

        win = gi.NewMainWindow("epyarrow", "ePy Arrow", width, height)
        ss.Win = win

        vp = win.WinViewport2D()
        ss.vp = vp
        updt = vp.UpdateStart()

        mfr = win.SetMainFrame()

        tbar = gi.AddNewToolBar(mfr, "tbar")
        tbar.SetStretchMaxWidth()
        ss.ToolBar = tbar

        split = gi.AddNewSplitView(mfr, "split")
        split.Dim = mat32.X
        split.SetStretchMax()

        cv = ss.NewClassView("sv")
        cv.AddFrame(split)
        cv.Config()

        tv = gi.AddNewTabView(split, "tv")

        tabv = etview.TableView()
        tv.AddTab(tabv, "Pats")
        tabv.SetTable(ss.Pats, go.nil)
        ss.PatsTable = tabv

        split.SetSplitsList(go.Slice_float32([.2, .8]))
        recv = win.This()

        tbar.AddAction(gi.ActOpts(Label="Test", Icon="update", Tooltip="run the test."), recv, TestCB)

        # main menu
        appnm = gi.AppName()
        mmen = win.MainMenu
        mmen.ConfigMenus(go.Slice_string([appnm, "File", "Edit", "Window"]))

        amen = gi.Action(win.MainMenu.ChildByName(appnm, 0))
        amen.Menu.AddAppMenu(win)

        emen = gi.Action(win.MainMenu.ChildByName("Edit", 1))
        emen.Menu.AddCopyCutPaste(win)

        win.MainMenuUpdated()
        vp.UpdateEndNoSig(updt)
        win.GoStartEventLoop()

# TheSim is the overall state for this simulation
TheSim = Sim()
 
def main(argv):
    TheSim.Config()
    TheSim.Test()
    TheSim.ConfigGui()
    
main(sys.argv[1:])

        
