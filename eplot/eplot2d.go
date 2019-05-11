// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"

	"github.com/emer/etable/etable"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin/key"
	"github.com/goki/gi/svg"
	"github.com/goki/gide/gide"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Plot2D is a GoGi Widget that provides a 2D plot of selected columns of etable data
type Plot2D struct {
	gi.Layout
	Table    *etable.Table `desc:"the table that we're plotting"`
	Params   PlotParams    `desc:"the overall plot parameters"`
	Cols     []*ColParams  `desc:"the parameters for each column of the table"`
	GPlot    *plot.Plot    `desc:"the gonum plot that actually does the plotting -- always save the last one generated"`
	SVGFile  gi.FileName   `desc:"current svg file"`
	DataFile gi.FileName   `desc:"current csv data file"`
}

var KiT_Plot2D = kit.Types.AddType(&Plot2D{}, Plot2DProps)

func (pl *Plot2D) Defaults() {
	pl.Params.Defaults()
}

// SetTable sets the table to view and updates view
func (pl *Plot2D) SetTable(tab *etable.Table) {
	pl.Defaults()
	if pl.Table != tab {
		pl.Table = tab
		pl.Cols = nil
	}
	pl.Config()
}

// ColParamsTry returns the current column parameters by name (to access by index, just use Cols directly)
// Try version returns error message if not found.
func (pl *Plot2D) ColParamsTry(colNm string) (*ColParams, error) {
	for _, cp := range pl.Cols {
		if cp.Col == colNm {
			return cp, nil
		}
	}
	return nil, fmt.Errorf("eplot plot: %v column named: %v not found", pl.Name, colNm)
}

// ColParams returns the current column parameters by name (to access by index, just use Cols directly)
// returns nil if not found
func (pl *Plot2D) ColParams(colNm string) *ColParams {
	cp, _ := pl.ColParamsTry(colNm)
	return cp
}

// SetColParams sets main parameters for one column
func (pl *Plot2D) SetColParams(colNm string, on bool, fixMin bool, min float64, fixMax bool, max float64) {
	cp, err := pl.ColParamsTry(colNm)
	if err != nil {
		log.Println(err)
		return
	}
	cp.On = on
	cp.Range.FixMin = fixMin
	if fixMin {
		cp.Range.Min = min
	}
	cp.Range.FixMax = fixMax
	if fixMax {
		cp.Range.Max = max
	}
}

// SaveSVG saves the plot to an svg -- first updates to ensure that plot is current
func (pl *Plot2D) SaveSVG(fname gi.FileName) {
	pl.Update()
	sv := pl.SVGPlot()
	SaveSVGView(string(fname), pl.GPlot, sv, 2)
	pl.SVGFile = fname
}

// SaveData saves the Table data to a csv (comma-separated values) file
func (pl *Plot2D) SaveData(fname gi.FileName) {
	pl.Table.SaveCSV(fname, ',', true)
	pl.DataFile = fname
}

// YLabel returns the Y-axis label
func (pl *Plot2D) YLabel() string {
	if pl.Params.YAxisLabel != "" {
		return pl.Params.YAxisLabel
	}
	for _, cp := range pl.Cols {
		if cp.On {
			return cp.Label()
		}
	}
	return "Y"
}

// XLabel returns the X-axis label
func (pl *Plot2D) XLabel() string {
	if pl.Params.XAxisLabel != "" {
		return pl.Params.XAxisLabel
	}
	if pl.Params.XAxisCol != "" {
		cp := pl.ColParams(pl.Params.XAxisCol)
		if cp != nil {
			return cp.Label()
		}
		return pl.Params.XAxisCol
	}
	return "X"
}

// Update updates the display based on current state of table
func (pl *Plot2D) Update() {
	if !pl.IsVisible() || pl.Table == nil {
		return
	}
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.NumCols() {
		pl.Config()
	}
	pl.GenPlot() // just need to redraw svg, not full thing.
}

// GenPlot generates the plot
func (pl *Plot2D) GenPlot() {
	plt, _ := plot.New() // todo: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()
	plt.BackgroundColor = nil

	for _, cp := range pl.Cols {
		cp.Update()
		if cp.Col == pl.Params.XAxisCol {
			if cp.Range.FixMin {
				plt.X.Min = math.Min(plt.X.Min, cp.Range.Min)
			}
			if cp.Range.FixMax {
				plt.X.Max = math.Max(plt.X.Max, cp.Range.Max)
			}
		}
		if !cp.On {
			continue
		}
		if cp.Range.FixMin {
			plt.Y.Min = math.Min(plt.Y.Min, cp.Range.Min)
		}
		if cp.Range.FixMax {
			plt.Y.Max = math.Max(plt.Y.Max, cp.Range.Max)
		}
		xy, _ := NewTableXYNames(pl.Table, pl.Params.XAxisCol, cp.Col)
		l, _ := plotter.NewLine(xy)
		l.LineStyle.Width = vg.Points(pl.Params.LineWidth)
		l.LineStyle.Color = cp.Color
		plt.Add(l)
		plt.Legend.Add(cp.Label(), l)
	}
	plt.Legend.Top = true

	pl.GPlot = plt
	sv := pl.SVGPlot()
	PlotViewSVG(plt, sv, pl.Params.Scale)
}

// Config configures the overall view widget
func (pl *Plot2D) Config() {
	pl.Lay = gi.LayoutVert
	pl.Defaults()
	pl.SetProp("spacing", gi.StdDialogVSpaceUnits)
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
	play.SetProp("spacing", gi.StdDialogVSpaceUnits)

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

func (pl *Plot2D) SVGPlot() *svg.Editor {
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
	nc := pl.Table.NumCols()
	if nc == len(pl.Cols) {
		return
	}
	npc := len(PlotColorNames)
	pl.Cols = make([]*ColParams, nc)
	clri := 0
	for ci := range pl.Table.Cols {
		cn := pl.Table.ColNames[ci]
		inc := 1
		if cn == pl.Params.XAxisCol { // re-use xaxis color
			inc = 0
		}
		cp := &ColParams{Col: cn, ColorName: PlotColorNames[clri%npc]}
		cp.Defaults()
		pl.Cols[ci] = cp
		clri += inc
	}
}

// ColsUpdate updates the display toggles for all the cols
func (pl *Plot2D) ColsUpdate() {
	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		cp := pl.Cols[i]
		cl := cli.(*gi.Layout)
		cb := cl.Child(0).(*gi.CheckBox)
		cb.SetChecked(cp.On)
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
		config.Add(gi.KiT_Layout, cn.Col)
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
		ca.SetText(cp.Col)
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
	sv := pl.SVGPlot()
	sv.InitScale()

	sv.Fill = true
	sv.SetProp("background-color", "white")
	sv.SetStretchMaxWidth()
	sv.SetStretchMaxHeight()
}

func (pl *Plot2D) ToolbarConfig() {
	tbar := pl.Toolbar()
	if len(tbar.Kids) != 0 {
		return
	}

	tbar.AddAction(gi.ActOpts{Icon: "pan", Tooltip: "return to default pan / orbit mode where mouse drags move camera around (Shift = pan, Alt = pan target)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select pan mode\n")
		})
	tbar.AddAction(gi.ActOpts{Icon: "arrow", Tooltip: "turn on select mode for selecting units and layers with mouse clicks"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select select mode\n")
		})
	tbar.AddSeparator("ctrl")
	tbar.AddAction(gi.ActOpts{Label: "Update", Icon: "update", Tooltip: "update fully redraws display, reflecting any new settings etc"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			pl.Config()
			pl.Update()
		})
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(pl.Viewport, &pl.Params, giv.DlgOpts{Title: pl.Nm + " Params"}, nil, nil)
		})
	tbar.AddSeparator("ctrl")
	tbar.AddAction(gi.ActOpts{Label: "Save SVG...", Icon: "file-save", Tooltip: "save plot to an .svg file that can be further enhanced using a drawing editor or directly included in publications etc"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveSVG", pl.Viewport)
		})
	tbar.AddAction(gi.ActOpts{Label: "Save Data...", Icon: "file-save", Tooltip: "save table data to a csv comma-separated-values file"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveData", pl.Viewport)
		})

}

func (pl *Plot2D) Style2D() {
	pl.Layout.Style2D()
	pl.GenPlot()
	pl.ColsUpdate()
}

var Plot2DProps = ki.Props{
	"max-width":  -1,
	"max-height": -1,
	// "width":      units.NewEm(5), // this gives the entire plot the scrollbars
	// "height":     units.NewEm(5),
	"ToolBar": ki.PropSlice{
		{"Update", ki.Props{
			"shortcut": "Command+U",
			"desc":     "update graph plot",
			"icon":     "update",
		}},
		// {"ViewFile", ki.Props{
		// 	"label": "Open...",
		// 	"icon":  "file-open",
		// 	"desc":  "open a file in current active text view",
		// 	"shortcut-func": giv.ShortcutFunc(func(gei interface{}, act *gi.Action) key.Chord {
		// 		return key.Chord(gide.ChordForFun(gide.KeyFunFileOpen).String())
		// 	}),
		// 	"Args": ki.PropSlice{
		// 		{"File Name", ki.Props{
		// 			"default-field": "ActiveFilename",
		// 		}},
		// 	},
		// }},
		{"SaveSVG", ki.Props{
			"label": "Save SVG...",
			"desc":  "save plot to an SVG file",
			"icon":  "file-save",
			"shortcut-func": giv.ShortcutFunc(func(gei interface{}, act *gi.Action) key.Chord {
				return key.Chord(gide.ChordForFun(gide.KeyFunBufSaveAs).String())
			}),
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"default-field": "SVGFile",
					"ext":           ".svg",
				}},
			},
		}},
		{"SaveData", ki.Props{
			"label": "Save Data...",
			"icon":  "file-save",
			"desc":  "save table data to a csv comma-separated-values file",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"default-field": "DataFile",
					"ext":           ".csv",
				}},
			},
		}},
	},
}

// these are the plot color names to use in order for successive lines -- feel free to choose your own!
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "yellow4", "SkyBlue", "pink"}
