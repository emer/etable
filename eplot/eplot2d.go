// Copyright (c) 2019, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

//go:generate goki generate

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"path/filepath"
	"strings"

	"goki.dev/colors"
	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/etview"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/mat32/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
)

// Plot2D is a GoGi Widget that provides a 2D plot of selected columns of etable data
type Plot2D struct { //gti:add
	gi.Layout

	// the idxview of the table that we're plotting
	Table *etable.IdxView `set:"-"`

	// the overall plot parameters
	Params PlotParams

	// the parameters for each column of the table
	Cols []*ColParams `set:"-"`

	// the gonum plot that actually does the plotting -- always save the last one generated
	Plot *plot.Plot `set:"-" edit:"-" json:"-" xml:"-"`

	// ConfigPlotFunc is a function to call to configure [Plot2D.Plot], the gonum plot that
	// actually does the plotting. It is called after [Plot] is generated, and properties
	// of [Plot] can be modified in it. Properties of [Plot] should not be modified outside
	// of this function, as doing so will have no effect.
	ConfigPlotFunc func() `json:"-" xml:"-"`

	// current svg file
	SVGFile gi.FileName

	// current csv data file
	DataFile gi.FileName

	// currently doing a plot
	InPlot bool `set:"-" edit:"-" json:"-" xml:"-"`
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
	plot.DefaultFont = font.Font{Typeface: "Roboto", Variant: "Sans"}
	pl.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Grow.Set(1, 1)
	})
}

// SetTable sets the table to view and updates view
func (pl *Plot2D) SetTable(tab *etable.Table) *Plot2D {
	pl.Table = etable.NewIdxView(tab)
	pl.DeleteCols()
	pl.Update()
	return pl
}

// SetTableView sets the idxview of table to view and updates view
func (pl *Plot2D) SetTableView(tab *etable.IdxView) *Plot2D {
	pl.Table = tab
	pl.DeleteCols()
	pl.Update()
	return pl
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
func (pl *Plot2D) SaveSVG(fname gi.FileName) { //gti:add
	pl.Update()
	sv := pl.SVGPlot()
	SaveSVGView(string(fname), pl.Plot, sv, 2)
	pl.SVGFile = fname
}

// SavePNG saves the current plot to a png, capturing current render
func (pl *Plot2D) SavePNG(fname gi.FileName) { //gti:add
	sv := pl.SVGPlot()
	sv.SavePNG(fname)
}

// SaveCSV saves the Table data to a csv (comma-separated values) file with headers (any delim)
func (pl *Plot2D) SaveCSV(fname gi.FileName, delim etable.Delims) { //gti:add
	pl.Table.SaveCSV(fname, delim, etable.Headers)
	pl.DataFile = fname
}

// SaveAll saves the current plot to a png, svg, and the data to a tsv -- full save
// Any extension is removed and appropriate extensions are added
func (pl *Plot2D) SaveAll(fname gi.FileName) { //gti:add
	fn := string(fname)
	fn = strings.TrimSuffix(fn, filepath.Ext(fn))
	pl.SaveCSV(gi.FileName(fn+".tsv"), etable.Tab)
	pl.SavePNG(gi.FileName(fn + ".png"))
	pl.SaveSVG(gi.FileName(fn + ".svg"))
}

// OpenCSV opens the Table data from a csv (comma-separated values) file (or any delim)
func (pl *Plot2D) OpenCSV(fname gi.FileName, delim etable.Delims) { //gti:add
	pl.Table.Table.OpenCSV(fname, delim)
	pl.DataFile = fname
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

// GoUpdatePlot updates the display based on current IdxView into table.
// this version can be called from go routines.
func (pl *Plot2D) GoUpdatePlot() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.Table.Table == nil || pl.InPlot {
		return
	}
	updt := pl.Sc.UpdateStartAsync()
	if !updt {
		pl.Sc.UpdateEndAsyncRender(updt)
		return
	}
	pl.Table.Sequential()
	pl.GenPlot()
	pl.Sc.UpdateEndAsyncRender(updt)
}

// UpdatePlot updates the display based on current IdxView into table.
// This version can only be called within main goroutine for
// window eventloop -- use GoUpdateUplot for other-goroutine updates.
func (pl *Plot2D) UpdatePlot() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.Table.Table == nil || pl.InPlot {
		return
	}
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.Table.NumCols() {
		pl.Update()
	}
	pl.Table.Sequential()
	pl.GenPlot()
}

// GenPlot generates the plot and renders it to SVG
// It surrounds operation with InPlot true / false to prevent multiple updates
func (pl *Plot2D) GenPlot() {
	if !pl.IsVisible() { // need this to make things render better on tab opening etc
		return
	}
	if pl.InPlot {
		slog.Error("eplot: in plot already")
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
	pl.Plot = nil
	switch pl.Params.Type {
	case XY:
		pl.GenPlotXY()
	case Bar:
		pl.GenPlotBar()
	}
	if pl.Plot != nil {
		PlotViewSVG(pl.Plot, sv, pl.Params.Scale)
	} else {
		sv.SVG.DeleteAll()
		// slog.Error("eplot: no plot generated from gonum plot")
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

func (pl *Plot2D) ConfigWidget() {
	if pl.Table != nil {
		pl.ConfigPlot()
	}
}

// ConfigPlot configures the overall view widget
func (pl *Plot2D) ConfigPlot() {
	pl.Params.FmMeta(pl.Table.Table)
	if !pl.HasChildren() {
		fr := gi.NewFrame(pl, "cols")
		fr.Style(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Grow.Set(0, 1)
			s.Overflow.Y = styles.OverflowAuto
			s.BackgroundColor.SetSolid(colors.Scheme.SurfaceContainerLow)
		})
		pt := gi.NewSVG(pl, "plot")
		pt.Style(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})

	}
	updt := pl.UpdateStart()
	defer pl.UpdateEndLayout(updt)

	pl.ColsConfig()
	pl.PlotConfig()
}

// DeleteCols deletes any existing cols, to ensure an update to new table
func (pl *Plot2D) DeleteCols() {
	pl.Cols = nil
	if pl.HasChildren() {
		vl := pl.ColsLay()
		vl.DeleteChildren(ki.DestroyKids)
	}
}

func (pl *Plot2D) ColsLay() *gi.Frame {
	return pl.ChildByName("cols", 0).(*gi.Frame)
}

func (pl *Plot2D) SVGPlot() *gi.SVG {
	return pl.ChildByName("plot", 1).(*gi.SVG)
}

const PlotColsHeaderN = 2

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
	pl.Cols = make([]*ColParams, nc)
	clri := 0
	for ci := range dt.Cols {
		cn := dt.ColNames[ci]
		inc := 1
		if cn == pl.Params.XAxisCol { // re-use xaxis color
			inc = 0
		}
		cp := &ColParams{Col: cn, Color: colors.BinarySpacedAccentVariant(clri)}
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
		if i < PlotColsHeaderN {
			continue
		}
		ci := i - PlotColsHeaderN
		cp := pl.Cols[ci]
		cl := cli.(*gi.Layout)
		sw := cl.Child(0).(*gi.Switch)
		if sw.StateIs(states.Checked) != cp.On {
			sw.SetChecked(cp.On)
			sw.SetNeedsRender(true)
		}
	}
}

// SetAllCols turns all Cols on or off (except X axis)
func (pl *Plot2D) SetAllCols(on bool) {
	updt := pl.UpdateStart()
	defer pl.UpdateEndRender(updt)
	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		if i < PlotColsHeaderN {
			continue
		}
		ci := i - PlotColsHeaderN
		cp := pl.Cols[ci]
		if cp.Col == pl.Params.XAxisCol {
			continue
		}
		cp.On = on
		cl := cli.(*gi.Layout)
		sw := cl.Child(0).(*gi.Switch)
		sw.SetChecked(cp.On)
		sw.SetNeedsRender(true)
	}
	pl.UpdatePlot()
}

// SetColsByName turns cols On or Off if their name contains given string
func (pl *Plot2D) SetColsByName(nameContains string, on bool) { //gti:add
	updt := pl.UpdateStart()
	defer pl.UpdateEndRender(updt)

	vl := pl.ColsLay()
	for i, cli := range *vl.Children() {
		if i < PlotColsHeaderN {
			continue
		}
		ci := i - PlotColsHeaderN
		cp := pl.Cols[ci]
		if cp.Col == pl.Params.XAxisCol {
			continue
		}
		if !strings.Contains(cp.Col, nameContains) {
			continue
		}
		cp.On = on
		cl := cli.(*gi.Layout)
		sw := cl.Child(0).(*gi.Switch)
		sw.SetChecked(cp.On)
		sw.SetNeedsRender(true)
	}
	pl.UpdatePlot()
}

// ColsConfig configures the column gui buttons
func (pl *Plot2D) ColsConfig() {
	vl := pl.ColsLay()
	pl.ColsListUpdate()
	if len(vl.Kids) == len(pl.Cols)+PlotColsHeaderN {
		pl.ColsUpdate()
		return
	}
	vl.DeleteChildren(true)
	if len(pl.Cols) == 0 {
		return
	}
	sc := gi.NewLayout(vl, "sel-cols")
	sw := gi.NewSwitch(sc, "on").SetTooltip("Toggle off all columns")
	sw.OnChange(func(e events.Event) {
		sw.SetChecked(false)
		pl.SetAllCols(false)
	})
	gi.NewButton(sc, "col").SetText("Select Cols").SetType(gi.ButtonAction).
		SetTooltip("click to select columns based on column name").
		OnClick(func(e events.Event) {
			giv.CallFunc(pl, pl.SetColsByName)
		})
	gi.NewSeparator(vl, "sep")

	for _, cp := range pl.Cols {
		cp := cp
		cp.Plot = pl
		cl := gi.NewLayout(vl, cp.Col)
		cl.Style(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(0, 0)
		})
		sw := gi.NewSwitch(cl, "on").SetType(gi.SwitchCheckbox).SetTooltip("toggle plot on")
		sw.OnChange(func(e events.Event) {
			cp.On = sw.StateIs(states.Checked)
			pl.UpdatePlot()
		})
		sw.SetState(cp.On, states.Checked)
		bt := gi.NewButton(cl, "col").SetText(cp.Col).SetType(gi.ButtonAction)
		bt.SetMenu(func(m *gi.Scene) {
			gi.NewButton(m, "set-x").SetText("Set X Axis").OnClick(func(e events.Event) {
				pl.Params.XAxisCol = cp.Col
				pl.UpdatePlot()
			})
			gi.NewButton(m, "set-legend").SetText("Set Legend").OnClick(func(e events.Event) {
				pl.Params.LegendCol = cp.Col
				pl.UpdatePlot()
			})
			gi.NewButton(m, "edit").SetText("Edit").OnClick(func(e events.Event) {
				d := gi.NewBody().AddTitle("Col Params")
				giv.NewStructView(d).SetStruct(cp)
				d.NewFullDialog(pl).Run()
				// todo: add update on ok
			})
		})
	}
}

// PlotConfig configures the PlotView
func (pl *Plot2D) PlotConfig() {
	sv := pl.SVGPlot()
	sv.SVG.Fill = true
	sv.SVG.Norm = true
	sv.SVG.Scale = 1
	sv.SVG.Translate = mat32.Vec2{}
	sv.SetReadOnly(true)
}

func (pl *Plot2D) ConfigToolbar(tb *gi.Toolbar) {
	if pl.Table == nil || pl.Table.Table == nil {
		return
	}
	gi.NewButton(tb).SetIcon(icons.PanTool).
		SetTooltip("toggle the ability to zoom and pan the view").OnClick(func(e events.Event) {
		sv := pl.SVGPlot()
		sv.SetReadOnly(!sv.IsReadOnly())
		sv.ApplyStyleUpdate()
	})
	gi.NewButton(tb).SetIcon(icons.ArrowForward).
		SetTooltip("turn on select mode for selecting SVG elements").
		OnClick(func(e events.Event) {
			fmt.Println("this will select select mode")
		})
	gi.NewSeparator(tb)
	gi.NewButton(tb).SetText("Update").SetIcon(icons.Update).
		SetTooltip("update fully redraws display, reflecting any new settings etc").
		OnClick(func(e events.Event) {
			pl.ConfigPlot()
			pl.UpdatePlot()
		})
	gi.NewButton(tb).SetText("Config").SetIcon(icons.Settings).
		SetTooltip("set parameters that control display (font size etc)").
		OnClick(func(e events.Event) {
			d := gi.NewBody().AddTitle(pl.Nm + " Params")
			giv.NewStructView(d).SetStruct(&pl.Params)
			d.NewFullDialog(pl).Run()
		})
	gi.NewButton(tb).SetText("Table").SetIcon(icons.Edit).
		SetTooltip("open a TableView window of the data").
		OnClick(func(e events.Event) {
			d := gi.NewBody().AddTitle(pl.Nm + " Data")
			etv := etview.NewTableView(d).SetTable(pl.Table.Table)
			d.AddAppBar(etv.ConfigToolbar)
			d.NewFullDialog(pl).Run()
		})
	gi.NewSeparator(tb)

	gi.NewButton(tb).SetText("Save").SetIcon(icons.Save).SetMenu(func(m *gi.Scene) {
		giv.NewFuncButton(m, pl.SaveSVG).SetIcon(icons.Save)
		giv.NewFuncButton(m, pl.SavePNG).SetIcon(icons.Save)
		giv.NewFuncButton(m, pl.SaveCSV).SetIcon(icons.Save)
		gi.NewSeparator(m)
		giv.NewFuncButton(m, pl.SaveAll).SetIcon(icons.Save)
	})
	giv.NewFuncButton(tb, pl.OpenCSV).SetIcon(icons.Open)
	gi.NewSeparator(tb)
	giv.NewFuncButton(tb, pl.Table.FilterColName).SetText("Filter").SetIcon(icons.FilterAlt)
	giv.NewFuncButton(tb, pl.Table.Sequential).SetText("Unfilter").SetIcon(icons.FilterAltOff)
}

// NewSubPlot returns a Plot2D with its own separate Toolbar,
// suitable for a tab or other element that is not the main plot.
func NewSubPlot(par gi.Widget, name ...string) *Plot2D {
	fr := gi.NewFrame(par, name...)
	tb := gi.NewToolbar(fr, "tbar")
	pl := NewPlot2D(fr, "plot")
	fr.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})
	tb.ToolbarFuncs.Add(pl.ConfigToolbar)
	return pl
}
