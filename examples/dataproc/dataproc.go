// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etview"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"gonum.org/v1/plot"
)

// this is the stub main for gogi that calls our actual mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

func mainrun() {
	width := 1600
	height := 1200

	gi.SetAppName("dataproc")
	gi.SetAppAbout(`This demonstrates data processing using etable.Table. See <a href="https://github.com/emer/etable">etable on GitHub</a>.</p>`)

	plot.DefaultFont = "Helvetica"

	planets := etable.NewTable("planets")
	planets.OpenCSV("./planets.csv", ',')

	win := gi.NewWindow2D("ra25", "Leabra Random Associator", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tv := gi.AddNewTabView(mfr, "tv")

	plv := tv.AddNewTab(etview.KiT_TableView, "TableView").(*etview.TableView)
	plv.SetTable(planets, nil)

	vp.UpdateEndNoSig(updt)
	win.StartEventLoop()
}
