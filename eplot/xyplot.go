// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log/slog"
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/grr"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/split"
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

	plt.BackgroundColor = colors.Scheme.Surface

	clr := colors.Scheme.OnSurface

	plt.Title.TextStyle.Color = clr
	plt.Legend.TextStyle.Color = clr
	plt.X.Color = clr
	plt.Y.Color = clr
	plt.X.Label.TextStyle.Color = clr
	plt.Y.Label.TextStyle.Color = clr
	plt.X.Tick.Color = clr
	plt.Y.Tick.Color = clr
	plt.X.Tick.Label.Color = clr
	plt.Y.Tick.Label.Color = clr

	// process xaxis first
	xi, xview, xbreaks, err := pl.PlotXAxis(plt, pl.Table)
	if err != nil {
		return
	}
	xp := pl.Cols[xi]

	var lsplit *etable.Splits
	nleg := 1
	if pl.Params.LegendCol != "" {
		_, err = pl.Table.Table.ColIndexTry(pl.Params.LegendCol)
		if err != nil {
			slog.Error("eplot.LegendCol", "err", err.Error())
		} else {
			grr.Log(xview.SortStableColNames([]string{pl.Params.LegendCol, xp.Col}, etable.Ascending))
			lsplit = split.GroupBy(xview, []string{pl.Params.LegendCol})
			nleg = max(lsplit.Len(), 1)
		}
	}

	var firstXY *TableXY
	var strCols []*ColParams
	nys := 0
	for _, cp := range pl.Cols {
		if !cp.On {
			continue
		}
		if cp.IsString {
			strCols = append(strCols, cp)
			continue
		}
		if cp.TensorIndex < 0 {
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
				stidx := cp.TensorIndex
				if cp.TensorIndex < 0 { // do all
					yc := pl.Table.Table.ColByName(cp.Col)
					_, sz := yc.RowCellSize()
					nidx = sz
					stidx = 0
				}
				for ii := 0; ii < nidx; ii++ {
					idx := stidx + ii
					tix := lview.Clone()
					tix.Indexes = tix.Indexes[stRow:edRow]
					xy, _ := NewTableXYName(tix, xi, xp.TensorIndex, cp.Col, idx, cp.Range)
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
						clr = colors.Spaced(cidx)
					}
					if nidx > 1 {
						clr = colors.Spaced(idx)
						lbl = fmt.Sprintf("%s_%02d", lbl, idx)
					}
					if cp.Lines.Or(pl.Params.Lines) && cp.Points.Or(pl.Params.Points) {
						lns, pts, _ = plotter.NewLinePoints(xy)
					} else if cp.Points.Or(pl.Params.Points) {
						pts, _ = plotter.NewScatter(xy)
					} else {
						lns, _ = plotter.NewLine(xy)
					}
					if lns != nil {
						lns.LineStyle.Width = vg.Points(cp.LineWidth.Or(pl.Params.LineWidth))
						lns.LineStyle.Color = clr
						plt.Add(lns)
						if bi == 0 {
							plt.Legend.Add(lbl, lns)
						}
					}
					if pts != nil {
						pts.GlyphStyle.Color = clr
						pts.GlyphStyle.Radius = vg.Points(cp.PointSize.Or(pl.Params.PointSize))
						pts.GlyphStyle.Shape = cp.PointShape.Or(pl.Params.PointShape).Glyph()
						plt.Add(pts)
						if lns == nil && bi == 0 {
							plt.Legend.Add(lbl, pts)
						}
					}
					if cp.ErrCol != "" {
						ec := pl.Table.Table.ColIndex(cp.ErrCol)
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
			xy, _ := NewTableXYName(xview, xi, xp.TensorIndex, cp.Col, cp.TensorIndex, firstXY.YRange)
			xy.LblCol = xy.YCol
			xy.YCol = firstXY.YCol
			xy.YIndex = firstXY.YIndex
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
		for i, dx := range pl.Table.Indexes {
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
	pl.Plot = plt
	if pl.ConfigPlotFunc != nil {
		pl.ConfigPlotFunc()
	}
}
