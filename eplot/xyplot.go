// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gist"
	"github.com/goki/ki/ints"
	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/split"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// GenPlotXY generates an XY (lines, points) plot, setting GPlot variable
func (pl *Plot2D) GenPlotXY() {
	plt := plot.New() // todo: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()

	plt.Title.TextStyle.Color = gi.Prefs.Colors.Font
	plt.Legend.TextStyle.Color = gi.Prefs.Colors.Font
	plt.X.Color = gi.Prefs.Colors.Font
	plt.Y.Color = gi.Prefs.Colors.Font
	plt.X.Label.TextStyle.Color = gi.Prefs.Colors.Font
	plt.Y.Label.TextStyle.Color = gi.Prefs.Colors.Font
	plt.X.Tick.Color = gi.Prefs.Colors.Font
	plt.Y.Tick.Color = gi.Prefs.Colors.Font
	plt.X.Tick.Label.Color = gi.Prefs.Colors.Font
	plt.Y.Tick.Label.Color = gi.Prefs.Colors.Font

	plt.BackgroundColor = nil

	// process xaxis first
	xi, xview, xbreaks, err := pl.PlotXAxis(plt, pl.Table)
	if err != nil {
		return
	}
	xp := pl.Cols[xi]

	var lsplit *etable.Splits
	nleg := 1
	if pl.Params.LegendCol != "" {
		_, err = pl.Table.Table.ColIdxTry(pl.Params.LegendCol)
		if err != nil {
			log.Println("eplot.LegendCol: " + err.Error())
		} else {
			err = xview.SortStableColNames([]string{pl.Params.LegendCol, xp.Col}, etable.Ascending)
			if err != nil {
				log.Println(err)
			}
			lsplit = split.GroupBy(xview, []string{pl.Params.LegendCol})
			nleg = ints.MaxInt(lsplit.Len(), 1)
		}
	}

	var firstXY *TableXY
	var strCols []*ColParams
	nys := 0
	for _, cp := range pl.Cols {
		cp.UpdateVals()
		if !cp.On {
			continue
		}
		if cp.IsString {
			strCols = append(strCols, cp)
			continue
		}
		if cp.TensorIdx < 0 {
			yc := pl.Table.Table.ColByName(cp.Col)
			_, sz := yc.RowCellSize()
			nys += sz
		} else {
			nys++
		}
		if cp.Range.FixMin {
			plt.Y.Min = math.Min(plt.Y.Min, cp.Range.Min)
		}
		if cp.Range.FixMax {
			plt.Y.Max = math.Max(plt.Y.Max, cp.Range.Max)
		}
	}

	if nys == 0 {
		return
	}

	firstXY = nil
	yidx := 0
	for _, cp := range pl.Cols {
		if !cp.On || cp == xp {
			continue
		}
		if cp.IsString {
			continue
		}
		for li := 0; li < nleg; li++ {
			lview := xview
			leg := ""
			if lsplit != nil && len(lsplit.Values) > li {
				leg = lsplit.Values[li][0]
				lview = lsplit.Splits[li]
				_, _, xbreaks, _ = pl.PlotXAxis(plt, lview)
			}
			stRow := 0
			for bi, edRow := range xbreaks {
				nidx := 1
				stidx := cp.TensorIdx
				if cp.TensorIdx < 0 { // do all
					yc := pl.Table.Table.ColByName(cp.Col)
					_, sz := yc.RowCellSize()
					nidx = sz
					stidx = 0
				}
				for ii := 0; ii < nidx; ii++ {
					idx := stidx + ii
					tix := lview.Clone()
					tix.Idxs = tix.Idxs[stRow:edRow]
					xy, _ := NewTableXYName(tix, xi, xp.TensorIdx, cp.Col, idx, cp.Range)
					if xy == nil {
						continue
					}
					if firstXY == nil {
						firstXY = xy
					}
					var pts *plotter.Scatter
					var lns *plotter.Line
					lbl := cp.Label()
					clr := cp.Color
					if leg != "" {
						lbl = leg + " " + lbl
					}
					if nleg > 1 {
						cidx := yidx*nleg + li
						clr, _ = gist.ColorFromString(PlotColorNames[cidx%len(PlotColorNames)], nil)
					}
					if nidx > 1 {
						clr, _ = gist.ColorFromString(PlotColorNames[idx%len(PlotColorNames)], nil)
						lbl = fmt.Sprintf("%s_%02d", lbl, idx)
					}
					if pl.Params.Lines && pl.Params.Points {
						lns, pts, _ = plotter.NewLinePoints(xy)
					} else if pl.Params.Points {
						pts, _ = plotter.NewScatter(xy)
					} else {
						lns, _ = plotter.NewLine(xy)
					}
					if lns != nil {
						lns.LineStyle.Width = vg.Points(pl.Params.LineWidth)
						lns.LineStyle.Color = clr
						plt.Add(lns)
						if bi == 0 {
							plt.Legend.Add(lbl, lns)
						}
					}
					if pts != nil {
						pts.GlyphStyle.Color = clr
						pts.GlyphStyle.Radius = vg.Points(pl.Params.PointSize)
						plt.Add(pts)
						if lns == nil && bi == 0 {
							plt.Legend.Add(lbl, pts)
						}
					}
					if cp.ErrCol != "" {
						ec := pl.Table.Table.ColIdx(cp.ErrCol)
						if ec >= 0 {
							xy.ErrCol = ec
							eb, _ := plotter.NewYErrorBars(xy)
							eb.LineStyle.Color = clr
							plt.Add(eb)
						}
					}
				}
				stRow = edRow
			}
		}
		yidx++
	}
	if firstXY != nil && len(strCols) > 0 {
		for _, cp := range strCols {
			xy, _ := NewTableXYName(xview, xi, xp.TensorIdx, cp.Col, cp.TensorIdx, firstXY.YRange)
			xy.LblCol = xy.YCol
			xy.YCol = firstXY.YCol
			xy.YIdx = firstXY.YIdx
			lbls, _ := plotter.NewLabels(xy)
			if lbls != nil {
				plt.Add(lbls)
			}
		}
	}

	// Use string labels for X axis if X is a string
	xc := pl.Table.Table.Cols[xi]
	if xc.DataType() == etensor.STRING {
		xcs := xc.(*etensor.String)
		vals := make([]string, pl.Table.Len())
		for i, dx := range pl.Table.Idxs {
			vals[i] = xcs.Values[dx]
		}
		plt.NominalX(vals...)
	}

	plt.Legend.Top = true
	plt.X.Tick.Label.Rotation = math.Pi * (pl.Params.XAxisRot / 180)
	if pl.Params.XAxisRot > 10 {
		plt.X.Tick.Label.YAlign = draw.YCenter
		plt.X.Tick.Label.XAlign = draw.XRight
	}
	pl.GPlot = plt
}
