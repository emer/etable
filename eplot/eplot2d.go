// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package eplot provides an interactive, graphical plotting utility for etable data.
*/
package eplot

import (
	"fmt"

	"github.com/emer/etable/etable"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/svg"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Plot2D is a GoGi Widget that provides a 2D plot of selected columns of etable data
type Plot2D struct {
	gi.Layout
	Table  *etable.Table `desc:"the table that we're plotting"`
	Params PlotParams    `desc:"the overall plot parameters"`
	Cols   []*ColParams  `desc:"the parameters for each column of the table"`
	GPlot  *plot.Plot    `desc:"the gonum plot that actually does the plotting -- always save the last one generated"`
}

var KiT_Plot2D = kit.Types.AddType(&Plot2D{}, Plot2DProps)

func (pl *Plot2D) Defaults() {
	pl.Params.Defaults()
}

// SetTable sets the table to view and updates view
func (pl *Plot2D) SetTable(tab *etable.Table) {
	pl.Defaults()
	pl.Table = tab
	pl.Config()
}

// // SetVar sets the variable to view and updates the display
// func (pl *Plot2D) SetVar(vr string) {
// 	pl.Var = vr
// 	pl.Update("")
// }

// func (pl *Plot2D) HasLayers() bool {
// 	if pl.Net == nil || pl.Net.NLayers() == 0 {
// 		return false
// 	}
// 	return true
// }

// Update updates the display based on current state of table
func (pl *Plot2D) Update() {
	if !pl.IsVisible() || pl.Table == nil {
		return
	}
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.NumCols() {
		pl.Config()
	}

	plt, _ := plot.New() // todo: not clear how to re-use
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = "Epoch"
	plt.Y.Label.Text = "Y"

	for _, cp := range pl.Cols {
		if !cp.On {
			continue
		}
		cp.Update()
		xy, _ := NewTableXYNames(pl.Table, pl.Params.XAxisCol, cp.Column)
		l, _ := plotter.NewLine(xy)
		l.LineStyle.Width = vg.Points(pl.Params.LineWidth)
		l.LineStyle.Color = cp.Color
		plt.Add(l)
		plt.Legend.Add(cp.Label(), l)
	}
	plt.Legend.Top = true

	pl.GPlot = plt
	sv := pl.Plot()
	PlotViewSVG(plt, sv, 5, 5, 2) // todo: compute height etc
	pl.UpdateSig()
}

// Config configures the overall view widget
func (pl *Plot2D) Config() {
	pl.Lay = gi.LayoutVert
	pl.Defaults()
	// pl.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_ToolBar, "tbar")
	config.Add(gi.KiT_Layout, "plot")
	mods, updt := pl.ConfigChildren(config, false)
	if !mods {
		updt = pl.UpdateStart()
	}

	play := pl.PlotLay()
	play.Lay = gi.LayoutHoriz
	play.SetProp("max-width", -1)
	play.SetProp("max-height", -1)

	vncfg := kit.TypeAndNameList{}
	vncfg.Add(gi.KiT_Frame, "cols")
	vncfg.Add(svg.KiT_Editor, "plot")
	play.ConfigChildren(vncfg, false) // won't do update b/c of above updt

	pl.ColsConfig()
	pl.PlotConfig()
	pl.ToolbarConfig()

	pl.UpdateEnd(updt)
}

func (pl *Plot2D) Toolbar() *gi.ToolBar {
	return pl.ChildByName("tbar", 0).(*gi.ToolBar)
}

func (pl *Plot2D) PlotLay() *gi.Layout {
	return pl.ChildByName("plot", 1).(*gi.Layout)
}

func (pl *Plot2D) Plot() *svg.Editor {
	return pl.PlotLay().ChildByName("plot", 1).(*svg.Editor)
}

func (pl *Plot2D) ColsLay() *gi.Frame {
	return pl.PlotLay().ChildByName("cols", 0).(*gi.Frame)
}

// ColsListUpdate updates the list of columns
func (pl *Plot2D) ColsListUpdate() {
	if pl.Table == nil {
		pl.Cols = nil
		return
	}
	npc := len(PlotColorNames)
	nc := pl.Table.NumCols()
	pl.Cols = make([]*ColParams, nc)
	for ci := range pl.Table.Cols {
		cn := pl.Table.ColNames[ci]
		cp := &ColParams{Column: cn, ColorName: PlotColorNames[ci%npc]}
		cp.Defaults()
		pl.Cols[ci] = cp
	}
}

// ColsConfig configures the column gui buttons
func (pl *Plot2D) ColsConfig() {
	vl := pl.ColsLay()
	vl.SetReRenderAnchor()
	vl.Lay = gi.LayoutVert
	vl.SetProp("spacing", 0)
	vl.SetProp("vertical-align", gi.AlignTop)
	pl.ColsListUpdate()
	if len(pl.Cols) == 0 {
		vl.DeleteChildren(true)
		return
	}
	config := kit.TypeAndNameList{}
	for _, cn := range pl.Cols {
		config.Add(gi.KiT_Layout, cn.Column)
	}
	mods, updt := vl.ConfigChildren(config, false)
	if !mods {
		updt = vl.UpdateStart()
	}
	clcfg := kit.TypeAndNameList{}
	clcfg.Add(gi.KiT_CheckBox, "on")
	clcfg.Add(gi.KiT_Action, "col")

	for i, cli := range *vl.Children() {
		cp := pl.Cols[i]
		cl := cli.(*gi.Layout)
		cl.Lay = gi.LayoutHoriz
		cl.ConfigChildren(clcfg, false)
		cl.SetProp("margin", 0)
		cl.SetProp("max-width", -1)
		cb := cl.Child(0).(*gi.CheckBox)
		cb.SetChecked(cp.On)
		cb.SetProp("idx", i)
		cb.ButtonSig.Connect(pl.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonToggled) {
				pll := recv.Embed(KiT_Plot2D).(*Plot2D)
				cbb := send.(*gi.CheckBox)
				idx := cb.Prop("idx").(int)
				cpp := pll.Cols[idx]
				cpp.On = cbb.IsChecked()
				pll.Update()
			}
		})

		ca := cl.Child(1).(*gi.Action)
		ca.SetText(cp.Column)
		ca.Data = i
		ca.ActionSig.Connect(pl.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			pll := recv.Embed(KiT_Plot2D).(*Plot2D)
			caa := send.(*gi.Action)
			idx := caa.Data.(int)
			cpp := pll.Cols[idx]
			giv.StructViewDialog(pl.Viewport, cpp, giv.DlgOpts{Title: "ColParams"}, nil, nil)
		})
	}
	vl.UpdateEnd(updt)
}

// PlotConfig configures the PlotView
func (pl *Plot2D) PlotConfig() {
	sv := pl.Plot()
	sv.InitScale()
	sv.Fill = true
	sv.SetProp("background-color", "white")
	// sv.SetProp("width", units.NewValue(float32(width/2), units.Px))
	// sv.SetProp("height", units.NewValue(float32(height-100), units.Px))
	sv.SetStretchMaxWidth()
	sv.SetStretchMaxHeight()
}

func (pl *Plot2D) ToolbarConfig() {
	tbar := pl.Toolbar()
	if len(tbar.Kids) != 0 {
		return
	}

	//	todo: add save button!

	tbar.AddAction(gi.ActOpts{Icon: "pan", Tooltip: "return to default pan / orbit mode where mouse drags move camera around (Shift = pan, Alt = pan target)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select pan mode\n")
		})
	tbar.AddAction(gi.ActOpts{Icon: "arrow", Tooltip: "turn on select mode for selecting units and layers with mouse clicks"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select select mode\n")
		})
	tbar.AddSeparator("ctrl")
	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update", Tooltip: "fully redraw display"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			pl.Config()
			pl.Update()
		})
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(pl.Viewport, &pl.Params, giv.DlgOpts{Title: pl.Nm + " Params"}, nil, nil)
		})
	// todo: colorbar
}

var Plot2DProps = ki.Props{
	"max-width":  -1,
	"max-height": -1,
}

// these are the plot color names to use in order for successive lines -- feel free to choose your own!
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "yellow4", "SkyBlue", "pink"}
