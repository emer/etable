// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/svg"
	"github.com/goki/ki/ints"
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
	InPlot   bool          `desc:"currently doing a plot"`
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
	pl.SetTable(fr.Table)
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

// SaveCSV saves the Table data to a csv (comma-separated values) file with headers
func (pl *Plot2D) SaveCSV(fname gi.FileName) {
	pl.Table.SaveCSV(fname, etable.Comma, true)
	pl.DataFile = fname
}

// OpenCSV opens the Table data from a csv (comma-separated values) file (or any delim)
func (pl *Plot2D) OpenCSV(fname gi.FileName, delim rune) {
	pl.Table.OpenCSV(fname, delim)
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

// GoUpdate updates the display based on current state of table.
// This version must be used when called from another goroutine
// does proper blocking to synchronize with updating in the main
// goroutine.
func (pl *Plot2D) GoUpdate() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.InPlot {
		return
	}
	if pl.Viewport.IsUpdatingNode() { // already updating -- don't add to it
		return
	}

	pl.Viewport.BlockUpdates()
	plupdt := false
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.NumCols() {
		plupdt = pl.UpdateStart()
		pl.Config()
	}
	sv := pl.SVGPlot()
	updt := sv.UpdateStart()
	pl.GenPlot()
	pl.Viewport.UnblockUpdates()
	sv.UpdateEnd(updt)
	pl.UpdateEnd(plupdt)
}

// Update updates the display based on current state of table.
// This version can only be called within main goroutine for
// window eventloop -- use GoUpdate for other-goroutine updates.
func (pl *Plot2D) Update() {
	if pl == nil || pl.This() == nil {
		return
	}
	if !pl.IsVisible() || pl.Table == nil || pl.InPlot {
		return
	}
	if len(pl.Kids) != 2 || len(pl.Cols) != pl.Table.NumCols() {
		pl.Config()
	}
	if pl.Viewport.IsUpdatingNode() { // already updating -- don't add to it
		return
	}
	pl.GenPlot()
}

// GenPlot generates the plot
// if blockUpdts is true, then block any other updates on the parent Viewport
// while we're plotting, because this involves destroying and rebuilding the
// tree, and is often called from another goroutine.  if called as part of the
// normal render process (e.g., in Style2D, then do NOT block updates!)
func (pl *Plot2D) GenPlot() {
	if pl.InPlot {
		fmt.Printf("error: in plot already\n")
		return
	}
	pl.InPlot = true
	plt, _ := plot.New() // todo: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()
	plt.BackgroundColor = nil

	// process xaxis first
	xi, xbreaks, err := pl.PlotXAxis(plt)
	if err != nil {
		return
	}
	xp := pl.Cols[xi]

	var firstXY *TableXY
	var strCols []*ColParams

	for _, cp := range pl.Cols {
		cp.UpdateVals()
		if !cp.On {
			continue
		}
		if cp.IsString {
			strCols = append(strCols, cp)
			continue
		}
		if cp.Range.FixMin {
			plt.Y.Min = math.Min(plt.Y.Min, cp.Range.Min)
		}
		if cp.Range.FixMax {
			plt.Y.Max = math.Max(plt.Y.Max, cp.Range.Max)
		}
	}

	if xbreaks != nil {
		stRow := 0
		for bi, edRow := range xbreaks {
			firstXY = nil
			for _, cp := range pl.Cols {
				if !cp.On || cp == xp {
					continue
				}
				if cp.IsString {
					continue
				}
				xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, cp.TensorIdx)
				if firstXY == nil {
					firstXY = xy
				}
				var pts *plotter.Scatter
				var lns *plotter.Line
				if pl.Params.Lines && pl.Params.Points {
					lns, pts, _ = plotter.NewLinePoints(xy)
				} else if pl.Params.Points {
					pts, _ = plotter.NewScatter(xy)
				} else {
					lns, _ = plotter.NewLine(xy)
				}
				if lns != nil {
					lns.LineStyle.Width = vg.Points(pl.Params.LineWidth)
					lns.LineStyle.Color = cp.Color
					plt.Add(lns)
					if bi == 0 {
						plt.Legend.Add(cp.Label(), lns)
					}
				}
				if pts != nil {
					pts.GlyphStyle.Color = cp.Color
					pts.GlyphStyle.Radius = vg.Points(pl.Params.PointSize)
					plt.Add(pts)
					if lns == nil && bi == 0 {
						plt.Legend.Add(cp.Label(), pts)
					}
				}
			}
			if firstXY != nil && len(strCols) > 0 {
				for _, cp := range strCols {
					xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, cp.TensorIdx)
					xy.LblCol = xy.YCol
					xy.YCol = firstXY.YCol
					lbls, _ := plotter.NewLabels(xy)
					plt.Add(lbls)
				}
			}
			stRow = edRow
		}
	} else {
		stRow := 0
		edRow := pl.Table.Rows
		for _, cp := range pl.Cols {
			if !cp.On || cp == xp {
				continue
			}
			if cp.IsString {
				continue
			}
			if cp.Range.FixMin {
				plt.Y.Min = math.Min(plt.Y.Min, cp.Range.Min)
			}
			if cp.Range.FixMax {
				plt.Y.Max = math.Max(plt.Y.Max, cp.Range.Max)
			}

			xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, cp.TensorIdx)
			if firstXY == nil {
				firstXY = xy
			}
			var pts *plotter.Scatter
			var lns *plotter.Line
			if pl.Params.Lines && pl.Params.Points {
				lns, pts, _ = plotter.NewLinePoints(xy)
			} else if pl.Params.Points {
				pts, _ = plotter.NewScatter(xy)
			} else {
				lns, _ = plotter.NewLine(xy)
			}
			if lns != nil {
				lns.LineStyle.Width = vg.Points(pl.Params.LineWidth)
				lns.LineStyle.Color = cp.Color
				plt.Add(lns)
				plt.Legend.Add(cp.Label(), lns)
			}
			if pts != nil {
				pts.GlyphStyle.Color = cp.Color
				pts.GlyphStyle.Radius = vg.Points(pl.Params.PointSize)
				plt.Add(pts)
				if lns == nil {
					plt.Legend.Add(cp.Label(), pts)
				}
			}
		}
		if firstXY != nil && len(strCols) > 0 {
			for _, cp := range strCols {
				xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, cp.TensorIdx)
				xy.LblCol = xy.YCol
				xy.YCol = firstXY.YCol
				lbls, _ := plotter.NewLabels(xy)
				plt.Add(lbls)
			}
		}
	}

	plt.Legend.Top = true

	pl.GPlot = plt
	sv := pl.SVGPlot()
	PlotViewSVG(plt, sv, pl.Params.Scale)
	pl.InPlot = false
}

// PlotXAxis processes the XAxis and returns its index and any breaks to insert
// based on negative X axis traversals or NaN values
func (pl *Plot2D) PlotXAxis(plt *plot.Plot) (xi int, xbreaks []int, err error) {
	xi, err = pl.Table.ColIdxTry(pl.Params.XAxisCol)
	if err != nil {
		log.Println("eplot.PlotXAxis: " + err.Error())
		return
	}
	xc := pl.Table.Cols[xi]
	xp := pl.Cols[xi]
	sz := 1
	if xp.Range.FixMin {
		plt.X.Min = math.Min(plt.X.Min, xp.Range.Min)
	}
	if xp.Range.FixMax {
		plt.X.Max = math.Max(plt.X.Max, xp.Range.Max)
	}
	if xc.NumDims() > 1 {
		sz = xc.Len() / xc.Dim(0)
		if xp.TensorIdx > sz || xp.TensorIdx < 0 {
			log.Printf("eplot.PlotXAxis: TensorIdx invalid -- reset to 0")
			xp.TensorIdx = 0
		}
	}
	if pl.Params.NegXDraw {
		return
	}
	lastx := -math.MaxFloat64
	for row := 0; row < pl.Table.Rows; row++ {
		var xv float64
		if xc.NumDims() > 1 {
			off := row*sz + xp.TensorIdx
			xv = xc.FloatVal1D(off)
		} else {
			xv = xc.FloatVal1D(row)
		}
		if xv < lastx {
			xbreaks = append(xbreaks, row)
		}
		lastx = xv
	}
	if xbreaks != nil {
		xbreaks = append(xbreaks, pl.Table.Rows)
	}
	return
}

// Config configures the overall view widget
func (pl *Plot2D) Config() {
	pl.Lay = gi.LayoutVert
	pl.Defaults()
	pl.Params.FmMeta(pl.Table)
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
		cp := &ColParams{Col: cn, ColorName: gi.ColorName(PlotColorNames[clri%npc])}
		cp.Defaults()
		tcol := pl.Table.Cols[ci]
		if _, ok := tcol.(*etensor.String); ok {
			cp.IsString = true
		}
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
		cp.Plot = pl
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
	sv.SetStretchMax()
}

func (pl *Plot2D) ToolbarConfig() {
	if pl.Table == nil {
		return
	}
	tbar := pl.Toolbar()
	if len(tbar.Kids) != 0 || pl.Viewport == nil {
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
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(pl.Viewport, &pl.Params, giv.DlgOpts{Title: pl.Nm + " Params"}, nil, nil)
		})
	tbar.AddSeparator("ctrl")
	tbar.AddAction(gi.ActOpts{Label: "Save SVG...", Icon: "file-save", Tooltip: "save plot to an .svg file that can be further enhanced using a drawing editor or directly included in publications etc"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveSVG", pl.Viewport)
		})
	tbar.AddAction(gi.ActOpts{Label: "Open CSV...", Icon: "file-open", Tooltip: "open CSV-formatted file"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "OpenCSV", pl.Viewport)
		})
	tbar.AddAction(gi.ActOpts{Label: "Save CSV...", Icon: "file-save", Tooltip: "save table data to a csv comma-separated-values file with headers"}, pl.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(pl, "SaveCSV", pl.Viewport)
		})

}

func (pl *Plot2D) Style2D() {
	pl.Layout.Style2D()
	pl.ToolbarConfig() // safe
	if !pl.IsConfiged() {
		return
	}
	if !pl.InPlot && pl.Viewport != nil && pl.Viewport.IsDoingFullRender() {
		pl.GenPlot() // this is recursive
	}
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
		{"OpenCSV", ki.Props{
			"label": "Open CSV File...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted data (or any delimeter -- default is tab (9), comma = 44) -- also recognizes emergent-style headers",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{}},
				{"Delimiter", ki.Props{
					"default": '\t',
					"desc":    "can use any single-character rune here -- default is tab (9) b/c otherwise hard to type, comma = 44",
				}},
			},
		}},
		{"SaveCSV", ki.Props{
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
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "yellow", "SkyBlue", "pink"}
