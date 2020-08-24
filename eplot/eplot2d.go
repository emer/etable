// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/svg"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"gonum.org/v1/plot"
)

// Plot2D is a GoGi Widget that provides a 2D plot of selected columns of etable data
type Plot2D struct {
	gi.Layout
	Table    *etable.IdxView `desc:"the idxview of the table that we're plotting"`
	Params   PlotParams      `desc:"the overall plot parameters"`
	Cols     []*ColParams    `desc:"the parameters for each column of the table"`
	GPlot    *plot.Plot      `desc:"the gonum plot that actually does the plotting -- always save the last one generated"`
	SVGFile  gi.FileName     `desc:"current svg file"`
	DataFile gi.FileName     `desc:"current csv data file"`
	InPlot   bool            `inactive:"+" desc:"currently doing a plot"`
}

var KiT_Plot2D = kit.Types.AddType(&Plot2D{}, Plot2DProps)

// AddNewPlot2D adds a new Plot2D to given parent node, with given name.
func AddNewPlot2D(parent ki.Ki, name string) *Plot2D {
	return parent.AddNewChild(KiT_Plot2D, name).(*Plot2D)
}

func (pl *Plot2D) CopyFieldsFrom(frm interface{}) {
	fr := frm.(*Plot2D)
	pl.Layout.CopyFieldsFrom(&fr.Layout)
	pl.Params.CopyFrom(&fr.Params)
	pl.SetTableView(fr.Table)
	mx := ints.MinInt(len(pl.Cols), len(fr.Cols))
	for i := 0; i < mx; i++ {
		pl.Cols[i].CopyFrom(fr.Cols[i])
	}
}

func (pl *Plot2D) Defaults() {
	pl.Params.Plot = pl
	pl.Params.Defaults()
	plot.DefaultFont = "Helvetica"
}

// SetTable sets the table to view and updates view
func (pl *Plot2D) SetTable(tab *etable.Table) {
	pl.Defaults()
	pl.Table = etable.NewIdxView(tab)
	pl.Cols = nil
	pl.Config()
}

// SetTableView sets the idxview of table to view and updates view
func (pl *Plot2D) SetTableView(tab *etable.IdxView) {
	pl.Defaults()
	pl.Table = tab
	pl.Cols = nil
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
	return nil, fmt.Errorf("eplot plot: %v column named: %v not found", pl.Nm, colNm)
}

// ColParams returns the current column parameters by name (to access by index, just use Cols directly)
// returns nil if not found
func (pl *Plot2D) ColParams(colNm string) *ColParams {
	cp, _ := pl.ColParamsTry(colNm)
	return cp
}

// use these for SetColParams args
const (
	On       bool = true
	Off           = false
	FixMin        = true
	FloatMin      = false
	FixMax        = true
	FloatMax      = false
)

// SetColParams sets main parameters for one column
func (pl *Plot2D) SetColParams(colNm string, on bool, fixMin bool, min float64, fixMax bool, max float64) *ColParams {
	cp, err := pl.ColParamsTry(colNm)
	if err != nil {
		log.Println(err)
		return nil
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
	return cp
}

// SaveSVG saves the plot to an svg -- first updates to ensure that plot is current
func (pl *Plot2D) SaveSVG(fname gi.FileName) {
	pl.Update()
	sv := pl.SVGPlot()
	SaveSVGView(string(fname), pl.GPlot, sv, 2)
	pl.SVGFile = fname
}

// SavePNG saves the current plot to a png, capturing current render
func (pl *Plot2D) SavePNG(fname gi.FileName) {
	sv := pl.SVGPlot()
	sv.SavePNG(string(fname))
}

// SaveCSV saves the Table data to a csv (comma-separated values) file with headers (any delim)
func (pl *Plot2D) SaveCSV(fname gi.FileName, delim etable.Delims) {
	pl.Table.SaveCSV(fname, delim, etable.Headers)
	pl.DataFile = fname
}

// OpenCSV opens the Table data from a csv (comma-separated values) file (or any delim)
func (pl *Plot2D) OpenCSV(fname gi.FileName, delim etable.Delims) {
	pl.Table.Table.OpenCSV(fname, delim)
	pl.DataFile = fname
	pl.Config()
	pl.Update()
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

// GoUpdate updates the display based on current state of table.
// This version must be used when called from another goroutine
// does proper blocking to synchronize with updating in the main
// goroutine.
func (pl *Plot2D) GoUpdate() {
	if pl == nil || pl.This() == nil {
		return
	}
	if pl.Table == nil || pl.Table.Table == nil {
		return
	}
	pl.Table.Sequential()
	pl.GoUpdatePlot()
}

// GoUpdatePlot updates the display based on current IdxView into table.
// This version must be used when called from another goroutine
// does proper blocking to synchronize with updating in the main
// goroutine.
func (pl *Plot2D) GoUpdatePlot() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.Table.Table == nil || pl.InPlot {
		return
	}
	mvp := pl.ViewportSafe()
	if mvp.IsUpdatingNode() { // already updating -- don't add to it
		return
	}

	mvp.BlockUpdates()
	plupdt := false
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.Table.NumCols() {
		plupdt = pl.UpdateStart()
		pl.Config()
	}
	sv := pl.SVGPlot()
	updt := sv.UpdateStart()
	pl.GenPlot()
	mvp.UnblockUpdates()
	sv.UpdateEnd(updt)
	pl.UpdateEnd(plupdt)
}

// Update updates the display based on current state of table.
// Calls Sequential method on etable.IdxView to view entire current table.
// This version can only be called within main goroutine for
// window eventloop -- use GoUpdate for other-goroutine updates.
func (pl *Plot2D) Update() {
	if pl == nil || pl.This() == nil {
		return
	}
	if pl.Table == nil || pl.Table.Table == nil {
		return
	}
	pl.Table.Sequential()
	pl.UpdatePlot()
}

// UpdatePlot updates the display based on current IdxView into table.
// This version can only be called within main goroutine for
// window eventloop -- use GoUpdate for other-goroutine updates.
func (pl *Plot2D) UpdatePlot() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.Table.Table == nil || pl.InPlot {
		return
	}
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.Table.NumCols() {
		pl.Config()
	}
	if pl.ViewportSafe().IsUpdatingNode() { // already updating -- don't add to it
		return
	}
	pl.GenPlot()
}

// GenPlot generates the plot and renders it to SVG
// It surrounds operation with InPlot true / false to prevent multiple updates
func (pl *Plot2D) GenPlot() {
	if pl.InPlot {
		fmt.Printf("error: in plot already\n")
		return
	}
	pl.InPlot = true
	sv := pl.SVGPlot()
	if pl.Table == nil || pl.Table.Table == nil || pl.Table.Len() == 0 {
		sv.DeleteChildren(ki.DestroyKids)
		pl.InPlot = false
		return
	}
	pl.GPlot = nil
	switch pl.Params.Type {
	case XY:
		pl.GenPlotXY()
	case Bar:
		pl.GenPlotBar()
	}
	if pl.GPlot != nil {
		PlotViewSVG(pl.GPlot, sv, pl.Params.Scale)
	}
	pl.InPlot = false
}

// PlotXAxis processes the XAxis and returns its index and any breaks to insert
// based on negative X axis traversals or NaN values.  xbreaks always ends in last row.
func (pl *Plot2D) PlotXAxis(plt *plot.Plot, ixvw *etable.IdxView) (xi int, xview *etable.IdxView, xbreaks []int, err error) {
	xi, err = ixvw.Table.ColIdxTry(pl.Params.XAxisCol)
	if err != nil {
		log.Println("eplot.PlotXAxis: " + err.Error())
		return
	}
	xview = ixvw
	xc := ixvw.Table.Cols[xi]
	xp := pl.Cols[xi]
	sz := 1
	lim := false
	if xp.Range.FixMin {
		lim = true
		plt.X.Min = math.Min(plt.X.Min, xp.Range.Min)
	}
	if xp.Range.FixMax {
		lim = true
		plt.X.Max = math.Max(plt.X.Max, xp.Range.Max)
	}
	if xc.NumDims() > 1 {
		sz = xc.Len() / xc.Dim(0)
		if xp.TensorIdx > sz || xp.TensorIdx < 0 {
			log.Printf("eplot.PlotXAxis: TensorIdx invalid -- reset to 0")
			xp.TensorIdx = 0
		}
	}
	if lim {
		xview = ixvw.Clone()
		xview.Filter(func(et *etable.Table, row int) bool {
			if !ixvw.Table.IsValidRow(row) { // sometimes it seems to get out of whack
				return false
			}
			var xv float64
			if xc.NumDims() > 1 {
				xv = xc.FloatValRowCell(row, xp.TensorIdx)
			} else {
				xv = xc.FloatVal1D(row)
			}
			if xp.Range.FixMin && xv < xp.Range.Min {
				return false
			}
			if xp.Range.FixMax && xv > xp.Range.Max {
				return false
			}
			return true
		})
	}
	if pl.Params.NegXDraw {
		return
	}
	lastx := -math.MaxFloat64
	for row := 0; row < xview.Len(); row++ {
		trow := xview.Idxs[row] // true table row
		var xv float64
		if xc.NumDims() > 1 {
			xv = xc.FloatValRowCell(trow, xp.TensorIdx)
		} else {
			xv = xc.FloatVal1D(trow)
		}
		if xv < lastx {
			xbreaks = append(xbreaks, row)
		}
		lastx = xv
	}
	xbreaks = append(xbreaks, xview.Len())
	return
}

// Config configures the overall view widget
func (pl *Plot2D) Config() {
	pl.Lay = gi.LayoutVert
	pl.Defaults()
	pl.Params.FmMeta(pl.Table.Table)
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

// IsConfiged returns true if widget is fully configured
func (pl *Plot2D) IsConfiged() bool {
	if len(pl.Kids) == 0 {
		return false
	}
	ppl := pl.PlotLay()
	if len(ppl.Kids) == 0 {
		return false
	}
	return true
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

const NColsHeader = 2

// ColsListUpdate updates the list of columns
func (pl *Plot2D) ColsListUpdate() {
	if pl.Table == nil || pl.Table.Table == nil {
		pl.Cols = nil
		return
	}
	dt := pl.Table.Table
	nc := dt.NumCols()
	if nc == len(pl.Cols) {
		return
	}
	npc := len(PlotColorNames)
	pl.Cols = make([]*ColParams, nc)
	clri := 0
	for ci := range dt.Cols {
		cn := dt.ColNames[ci]
		inc := 1
		if cn == pl.Params.XAxisCol { // re-use xaxis color
			inc = 0
		}
		cp := &ColParams{Col: cn, ColorName: gi.ColorName(PlotColorNames[clri%npc])}
		cp.Defaults()
		tcol := dt.Cols[ci]
		if tcol.DataType() == etensor.STRING {
			cp.IsString = true
		} else {
			cp.IsString = false
		}
		pl.Cols[ci] = cp
		clri += inc
	}
}

// ColsUpdate updates the display toggles for all the cols
func (pl *Plot2D) ColsUpdate() {
	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		if i < NColsHeader {
			continue
		}
		ci := i - NColsHeader
		cp := pl.Cols[ci]
		cl := cli.(*gi.Layout)
		cb := cl.Child(0).(*gi.CheckBox)
		cb.SetChecked(cp.On)
	}
}

// SetAllCols turns all Cols on or off (except X axis)
func (pl *Plot2D) SetAllCols(on bool) {
	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		if i < NColsHeader {
			continue
		}
		ci := i - NColsHeader
		cp := pl.Cols[ci]
		if cp.Col == pl.Params.XAxisCol {
			continue
		}
		cp.On = on
		cl := cli.(*gi.Layout)
		cb := cl.Child(0).(*gi.CheckBox)
		cb.SetChecked(cp.On)
	}
	pl.Update()
}

// SetColsByName turns cols On or Off if their name contains given string
func (pl *Plot2D) SetColsByName(nameContains string, on bool) {
	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		if i < NColsHeader {
			continue
		}
		ci := i - NColsHeader
		cp := pl.Cols[ci]
		if cp.Col == pl.Params.XAxisCol {
			continue
		}
		if !strings.Contains(cp.Col, nameContains) {
			continue
		}
		cp.On = on
		cl := cli.(*gi.Layout)
		cb := cl.Child(0).(*gi.CheckBox)
		cb.SetChecked(cp.On)
	}
	pl.Update()
}

// ColsConfig configures the column gui buttons
func (pl *Plot2D) ColsConfig() {
	vl := pl.ColsLay()
	vl.SetReRenderAnchor()
	vl.Lay = gi.LayoutVert
	vl.SetProp("spacing", 0)
	vl.SetProp("vertical-align", gi.AlignTop)
	vl.SetMinPrefHeight(units.NewEm(5)) // get separate scroll on cols
	vl.SetStretchMaxHeight()
	pl.ColsListUpdate()
	if len(pl.Cols) == 0 {
		vl.DeleteChildren(true)
		return
	}
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_Layout, "sel-cols")
	config.Add(gi.KiT_Separator, "sep")
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
		if i == 1 {
			sp := cli.(*gi.Separator)
			sp.Horiz = true
			continue
		}
		cl := cli.(*gi.Layout)
		cl.Lay = gi.LayoutHoriz
		cl.ConfigChildren(clcfg, false)
		cl.SetProp("margin", 0)
		cl.SetProp("max-width", -1)
		cb := cl.Child(0).(*gi.CheckBox)
		ca := cl.Child(1).(*gi.Action)
		if i == 0 {
			cb.SetChecked(false)
			cb.Tooltip = "click to turn all columns on or off"
			cb.ButtonSig.Connect(pl.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				if sig == int64(gi.ButtonToggled) {
					pll := recv.Embed(KiT_Plot2D).(*Plot2D)
					cbb := send.(*gi.CheckBox)
					pll.SetAllCols(cbb.IsChecked())
				}
			})
			ca.SetText("Select Cols")
			ca.Tooltip = "click to select columns based on column name"
			ca.ActionSig.Connect(pl.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				pll := recv.Embed(KiT_Plot2D).(*Plot2D)
				giv.CallMethod(pll, "SetColsByName", pll.ViewportSafe())
			})
		} else {
			ci := i - NColsHeader
			cp := pl.Cols[ci]
			cp.Plot = pl

			cb.SetChecked(cp.On)
			cb.SetProp("idx", ci)
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
			ca.SetText(cp.Col)
			ca.Data = ci
			ca.ActionSig.Connect(pl.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				pll := recv.Embed(KiT_Plot2D).(*Plot2D)
				caa := send.(*gi.Action)
				idx := caa.Data.(int)
				cpp := pll.Cols[idx]
				giv.StructViewDialog(pl.ViewportSafe(), cpp, giv.DlgOpts{Title: "ColParams"}, nil, nil)
			})
		}
	}
	vl.UpdateEnd(updt)
}

// PlotConfig configures the PlotView
func (pl *Plot2D) PlotConfig() {
	sv := pl.SVGPlot()
	sv.InitScale()

	sv.Fill = true
	sv.SetProp("background-color", &gi.Prefs.Colors.Background)
	sv.SetStretchMax()
}

func (pl *Plot2D) ToolbarConfig() {
	if pl.Table == nil || pl.Table.Table == nil {
		return
	}
	tbar := pl.Toolbar()
	if len(tbar.Kids) != 0 || pl.ViewportSafe() == nil {
		return
	}

	tbar.SetStretchMaxWidth()
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
	tbar.AddAction(gi.ActOpts{Label: "Config...", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(pl.ViewportSafe(), &pl.Params, giv.DlgOpts{Title: pl.Nm + " Params"}, nil, nil)
		})
	tbar.AddSeparator("file")
	tbar.AddAction(gi.ActOpts{Label: "Save SVG...", Icon: "file-save", Tooltip: "save plot to an .svg file that can be further enhanced using a drawing editor or directly included in publications etc"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveSVG", pl.ViewportSafe())
		})
	tbar.AddAction(gi.ActOpts{Label: "Save PNG...", Icon: "file-save", Tooltip: "save plot to a .png file, capturing the exact bits you currently see as the render"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SavePNG", pl.ViewportSafe())
		})
	tbar.AddSeparator("img")
	tbar.AddAction(gi.ActOpts{Label: "Open CSV...", Icon: "file-open", Tooltip: "Open CSV-formatted data -- also recognizes emergent-style headers"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "OpenCSV", pl.ViewportSafe())
		})
	tbar.AddAction(gi.ActOpts{Label: "Save CSV...", Icon: "file-save", Tooltip: "Save CSV-formatted data (or any delimiter) -- header outputs emergent-style header data"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveCSV", pl.ViewportSafe())
		})
	tbar.AddSeparator("filt")
	tbar.AddAction(gi.ActOpts{Label: "Filter...", Icon: "search", Tooltip: "filter rows of data being plotted by values in given column name, using string representation, with exclude, contains and ignore case options"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl.Table, "FilterColName", pl.ViewportSafe())
		})
	tbar.AddAction(gi.ActOpts{Label: "Unfilter", Icon: "search", Tooltip: "plot all rows in the table (undo any filtering )"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl.Table, "Sequential", pl.ViewportSafe())
		})

}

func (pl *Plot2D) Style2D() {
	pl.Layout.Style2D()
	pl.ToolbarConfig() // safe
	if !pl.IsConfiged() {
		return
	}
	mvp := pl.ViewportSafe()
	if !pl.InPlot && mvp != nil && mvp.IsDoingFullRender() {
		pl.GenPlot() // this is recursive
	}
	pl.ColsUpdate()
}

var Plot2DProps = ki.Props{
	"max-width":  -1,
	"max-height": -1,
	"ToolBar": ki.PropSlice{
		{"Update", ki.Props{
			"shortcut": "Command+U",
			"desc":     "update graph plot",
			"icon":     "update",
		}},
		{"SaveSVG", ki.Props{
			"label": "Save SVG...",
			"desc":  "save plot to an SVG file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"default-field": "SVGFile",
					"ext":           ".svg",
				}},
			},
		}},
		{"SavePNG", ki.Props{
			"label": "Save PNG...",
			"desc":  "save current render of plot to PNG file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".png",
				}},
			},
		}},
		{"OpenCSV", ki.Props{
			"label": "Open CSV File...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted data (or any delimeter) -- also recognizes emergent-style headers",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".tsv,.csv",
				}},
				{"Delimiter", ki.Props{
					"default": etable.Tab,
					"desc":    "delimiter between columns",
				}},
			},
		}},
		{"SaveCSV", ki.Props{
			"label": "Save Data...",
			"icon":  "file-save",
			"desc":  "Save CSV-formatted data (or any delimiter) -- header outputs emergent-style header data (recommended)",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"default-field": "DataFile",
					"ext":           ".tsv,.csv",
				}},
				{"Delimiter", ki.Props{
					"default": etable.Tab,
					"desc":    "delimiter between columns",
				}},
			},
		}},
	},
	"CallMethods": ki.PropSlice{
		{"SetColsByName", ki.Props{
			"desc": "Turn columns containing given string On or Off",
			"Args": ki.PropSlice{
				{"Name Contains", ki.Props{}},
				{"On", ki.Props{
					"default": true,
				}},
			},
		}},
	},
}

// these are the plot color names to use in order for successive lines -- feel free to choose your own!
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "yellow", "SkyBlue", "pink"}
