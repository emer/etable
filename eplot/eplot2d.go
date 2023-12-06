// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

//go:generate goki generate

import (
	"fmt"
	"image"
	"log"
	"math"
	"path/filepath"
	"strings"

	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/etview"
	"goki.dev/girl/styles"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/ki/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
)

// Plot2D is a GoGi Widget that provides a 2D plot of selected columns of etable data
type Plot2D struct { //goki:add
	gi.Layout

	// the idxview of the table that we're plotting
	Table *etable.IdxView

	// the overall plot parameters
	Params PlotParams

	// the parameters for each column of the table
	Cols []*ColParams

	// the gonum plot that actually does the plotting -- always save the last one generated
	GPlot *plot.Plot

	// current svg file
	SVGFile gi.FileName

	// current csv data file
	DataFile gi.FileName

	// currently doing a plot
	InPlot bool `inactive:"+"`
}

func (pl *Plot2D) CopyFieldsFrom(frm any) {
	fr := frm.(*Plot2D)
	pl.Layout.CopyFieldsFrom(&fr.Layout)
	pl.Params.CopyFrom(&fr.Params)
	pl.SetTableView(fr.Table)
	mx := min(len(pl.Cols), len(fr.Cols))
	for i := 0; i < mx; i++ {
		pl.Cols[i].CopyFrom(fr.Cols[i])
	}
}

func (pl *Plot2D) OnInit() {
	pl.Params.Plot = pl
	pl.Params.Defaults()
	plot.DefaultFont = font.Font{Typeface: "Liberation", Variant: "Sans"}
	pl.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		// s.StdDialogVSpaceUnits
	})
	// plot
	// play.Lay = gi.LayoutHoriz
	// play.SetProp("max-width", -1)
	// play.SetProp("max-height", -1)
	// play.SetProp("spacing", gi.StdDialogVSpaceUnits)

	// cols
	// vl.Lay = gi.LayoutVert
	// vl.SetProp("spacing", 0)
	// vl.SetProp("vertical-align", gist.AlignTop)
	// vl.SetMinPrefHeight(units.NewEm(5)) // get separate scroll on cols
	// vl.SetStretchMaxHeight()
}

// SetTable sets the table to view and updates view
func (pl *Plot2D) SetTable(tab *etable.Table) {
	pl.Defaults()
	pl.Table = etable.NewIdxView(tab)
	// pl.Cols = nil // todo: is this necessary!?
	pl.ConfigPlot()
}

// SetTableView sets the idxview of table to view and updates view
func (pl *Plot2D) SetTableView(tab *etable.IdxView) {
	pl.Defaults()
	pl.Table = tab
	pl.Cols = nil
	pl.ConfigPlot()
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

// SaveAll saves the current plot to a png, svg, and the data to a tsv -- full save
// Any extension is removed and appropriate extensions are added
func (pl *Plot2D) SaveAll(fname gi.FileName) {
	fn := string(fname)
	fn = strings.TrimSuffix(fn, filepath.Ext(fn))
	pl.SaveCSV(gi.FileName(fn+".tsv"), etable.Tab)
	pl.SavePNG(gi.FileName(fn + ".png"))
	pl.SaveSVG(gi.FileName(fn + ".svg"))
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
	pl.GenPlot()
}

// GenPlot generates the plot and renders it to SVG
// It surrounds operation with InPlot true / false to prevent multiple updates
func (pl *Plot2D) GenPlot() {
	if !pl.IsVisible() { // need this to make things render better on tab opening etc
		return
	}
	if pl.InPlot {
		fmt.Printf("error: in plot already\n")
		return
	}
	pl.InPlot = true
	sv := pl.SVGPlot()
	if pl.Table == nil || pl.Table.Table == nil || pl.Table.Table.Rows == 0 || pl.Table.Len() == 0 {
		sv.DeleteChildren(ki.DestroyKids)
		pl.InPlot = false
		return
	}
	lsti := pl.Table.Idxs[pl.Table.Len()-1]
	if lsti >= pl.Table.Table.Rows { // out of date
		pl.Table.Sequential()
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
		xbreaks = append(xbreaks, xview.Len())
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

// ConfigPlot configures the overall view widget
func (pl *Plot2D) ConfigPlot() {
	pl.Params.FmMeta(pl.Table.Table)
	if !pl.HasChildren() {
		pl.AddDefaultTopAppBar()
		gi.NewFrame(pl, "cols")
		gi.NewSVG(pl, "plot")
	}
	updt := pl.UpdateStart()
	defer pl.UpdateEndLayout(updt)

	pl.ColsConfig()
	pl.PlotConfig()
}

func (pl *Plot2D) ColsLay() *gi.Frame {
	return pl.ChildByName("cols", 0).(*gi.Frame)
}

func (pl *Plot2D) SVGPlot() *gi.SVG {
	return pl.ChildByName("plot", 1).(*gi.SVG)
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
		cp.FmMetaMap(pl.Table.Table.MetaData)
		pl.Cols[ci] = cp
		clri += inc
	}
}

// ColsFmMetaMap updates all the column settings from given meta map
func (pl *Plot2D) ColsFmMetaMap(meta map[string]string) {
	for _, cp := range pl.Cols {
		cp.FmMetaMap(meta)
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
	updt := pl.UpdateStart()
	defer pl.UpdateEnd(updt)
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
	updt := pl.UpdateStart()
	defer pl.UpdateEnd(updt)

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
	pl.ColsListUpdate()
	vl.DeleteChildren(true)
	if len(pl.Cols) == 0 {
		return
	}
	sc := gi.NewLayout(vl, "sel-cols")
	sw := gi.NewSwitch(sc, "on").SetTooltip("Toggle off all columns").
		OnChanged(func(e events.Event) {
			sw.SetChecked(false)
			pl.SetAllCols(false)
		})
	bt := gi.NewButton(sc, "col").SetText("Select Cols").
		SetTooltip("click to select columns based on column name").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, pl.SetColsByName)
		})
	gi.NewSeparator(vl, "sep").SetHoriz(true)

	for _, cp := range pl.Cols {
		cp.Plot = pl
		cl := gi.NewLayout(vl, cn.Col.Name)
		cl.Style(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(1, 0)
		})
		sw := gi.NewSwitch(cl, "on").SetTooltip("toggle plot on").
			OnChanged(func(e events.Event) {
				cp.On = sw.IsChecked()
				pl.Update()
			})
		sw.SetChecked(cp.On)
		bt := gi.NewButton(cl, "col").SetText(cp.Col).
			OnClick(func(e events.Event) {
			})
		bt.SetMenu(func(m *gi.Scene) {
			gi.NewButton(m, "set-x").SetText("Set X Axis").OnClick(func(e events.Event) {
				pl.Params.XAxisCol = cp.Col
				pl.Update()
			})
			gi.NewButton(m, "set-legend").SetText("Set Legend").OnClick(func(e events.Event) {
				pl.Params.LegendCol = cpp.Col
				pl.Update()
			})
			gi.NewButton(m, "edit").SetText("Edit").OnClick(func(e events.Event) {
				d := gi.NewBody().AddTitle("Col Params")
				NewStructView(d).SetStruct(cp)
				d.NewFullDialog(tv).Run()
			})
		})
	}
}

// PlotConfig configures the PlotView
func (pl *Plot2D) PlotConfig() {
	sv := pl.SVGPlot()
	sv.InitScale()

	sv.Fill = true
	sv.SetProp("background-color", &gi.Prefs.Colors.Background)
	sv.SetStretchMax()
}

func (pl *Plot2D) PlotTopAppBar(tb *gi.TopAppBar) {
	if pl.Table == nil || pl.Table.Table == nil {
		return
	}
	gi.NewButton(tb).SetIcon(icons.Pan).
		SetTooltip("return to default pan / orbit mode where mouse drags move camera around (Shift = pan, Alt = pan target)").OnClick(func(e events.Event) {
			fmt.Printf("this will select pan mode\n")
		})
	gi.NewButton(tb).SetIcon(icons.Arrow).
		SetTooltip(turn on select mode for selecting units and layers with mouse clicks").
		OnClick(func(e events.Event) {
			fmt.Printf("this will select select mode\n")
		})
	gi.NewSeparator(tb)
	gi.NewButton(tb).SetLabel("Update").SetIcon(icons.Update).
		SetTooltip("update fully redraws display, reflecting any new settings etc").
		OnClick(func(e events.Event) {
			pl.ConfigPlot()
			pl.Update()
		})
	gi.NewButton(tb).SetLabel("Config...").SetIcon(icons.Gear).
		SetTooltip("set parameters that control display (font size etc)").
		OnClick(func(e events.Event) {
			d := gi.NewBody().AddTitle(pl.Nm + " Params")
			NewStructView(d).SetStruct(&pl.Params)
			d.NewFullDialog(tv).Run()
		})
	gi.NewButton(tb).SetLabel("Table...").SetIcon(icons.Edit).
		SetTooltip("open a TableView window of the data").
		OnClick(func(e events.Event) {
			etview.TableViewDialog(pl.ViewportSafe(), pl.Table.Table, giv.DlgOpts{Title: pl.Nm + " Data"}, nil, nil)
		})
	gi.NewSeparator(tb)
	
	gi.NewButton(tb).SetLabel("Save...").SetIcon(icons.Save).SetMenu(func(m *gi.Scene) {
	gi.NewButton(m).SetText("Save SVG...").SetIcon(icons.Save).
		SetTooltip("save plot to an .svg file that can be further enhanced using a drawing editor or directly included in publications etc").OnClick(func(e events.Event) {
			giv.CallFunc(pl, pl.SaveSVG)
		})
	gi.NewButton(m).SetText("Save PNG...").SetIcon(icons.Save).
		SetTooltip("save plot to a .png file, capturing the exact bits you currently see as the render").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, "SavePNG")
		})
	gi.NewButton(m).SetText("Save CSV...").SetIcon(icons.Save).
		SetTooltip("Save CSV-formatted data (or any delimiter) -- header outputs emergent-style header data").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, "SaveCSV")
		})
	gi.NewSeparator(m)
	gi.NewButton(m).SetText("Save All...").SetIcon(icons.Save).
		SetTooltip("Save a .png, .svg, and .tsv of the data").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, "SaveAll")
		})
	})
	gi.NewButton(tb).SetLabel("Open CSV...").SetIcon(icons.Open).
	 SetTooltip("Open CSV-formatted data -- also recognizes emergent-style headers").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, "OpenCSV")
		})
	gi.NewSeparator(tb)
	gi.NewButton(tb).SetLabel("Filter...").SetIcon(icons.Search).
		SetTooltip("filter rows of data being plotted by values in given column name, using string representation, with exclude, contains and ignore case options").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl.Table, "FilterColName")
		})
	gi.NewButton(tb).SetLabel("Unfilter").SetIcon(icons.Search).
		SetTooltip("plot all rows in the table (undo any filtering )".
		OnClick(func(e events.Event) {
			giv.CallFuncd(pl.Table, "Sequential")
		})
}

// func (pl *Plot2D) Style2D() {
// 	pl.Layout.Style2D()
// 	pl.ToolbarConfig() // safe
// 	if !pl.IsConfiged() && pl.Table != nil && pl.Table.Table != nil {
// 		pl.Config()
// 	}
// 	if pl.IsConfiged() {
// 		pl.ColsUpdate()
// 	}
// }

// func (pl *Plot2D) Layout2D(parBBox image.Rectangle, iter int) bool {
// 	redo := pl.Layout.Layout2D(parBBox, iter)
// 	mvp := pl.ViewportSafe()
// 	if pl.IsConfiged() && !pl.InPlot && mvp != nil { // note: for tabs, not full re-rend
// 		pl.GenPlot() // only if visible; this is recursive
// 	}
// 	return redo
// }

/*
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
		{"SaveAll", ki.Props{
			"label": "Save All...",
			"desc":  "save plot to SVG, PNG, and TSV files",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"default-field": "SVGFile",
					"ext":           ".svg",
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
*/

// these are the plot color names to use in order for successive lines -- feel free to choose your own!
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "goldenrod", "SkyBlue", "pink"}
