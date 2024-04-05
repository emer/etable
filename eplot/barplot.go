// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"

	"cogentcore.org/core/colors"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/minmax"
	"github.com/emer/etable/v2/split"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

// bar plot is on integer positions, with different Y values and / or
// legend values interleaved

// GenPlotBar generates a Bar plot, setting GPlot variable
func (pl *Plot2D) GenPlotBar() {
	plt := plot.New() // note: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()
	// TODO(kai): better bar plot styling
	plt.BackgroundColor = colors.Scheme.Surface

	if pl.Params.BarWidth > 1 {
		pl.Params.BarWidth = .8
	}

	// process xaxis first
	xi, xview, _, err := pl.PlotXAxis(plt, pl.Table)
	if err != nil {
		return
	}
	xp := pl.Cols[xi]

	var lsplit *etable.Splits
	nleg := 1
	if pl.Params.LegendCol != "" {
		_, err = pl.Table.Table.ColIndexTry(pl.Params.LegendCol)
		if err != nil {
			log.Println("eplot.LegendCol: " + err.Error())
		} else {
			xview.SortColNames([]string{pl.Params.LegendCol, xp.Col}, etable.Ascending) // make it fit!
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

	stride := nys * nleg
	if stride > 1 {
		stride += 1 // extra gap
	}

	yoff := 0
	yidx := 0
	maxx := 0 // max number of x values
	for _, cp := range pl.Cols {
		if !cp.On || cp == xp {
			continue
		}
		if cp.IsString {
			continue
		}
		start := yoff
		for li := 0; li < nleg; li++ {
			lview := xview
			leg := ""
			if lsplit != nil && len(lsplit.Values) > li {
				leg = lsplit.Values[li][0]
				lview = lsplit.Splits[li]
			}
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
				xy, _ := NewTableXYName(lview, xi, xp.TensorIndex, cp.Col, idx, cp.Range)
				if xy == nil {
					continue
				}
				maxx = max(maxx, lview.Len())
				if firstXY == nil {
					firstXY = xy
				}
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
				ec := -1
				if cp.ErrCol != "" {
					ec = pl.Table.Table.ColIndex(cp.ErrCol)
				}
				var bar *ErrBarChart
				if ec >= 0 {
					exy, _ := NewTableXY(lview, ec, 0, ec, 0, minmax.Range64{})
					bar, err = NewErrBarChart(xy, exy)
					if err != nil {
						log.Println(err)
						continue
					}
				} else {
					bar, err = NewErrBarChart(xy, nil)
					if err != nil {
						log.Println(err)
						continue
					}
				}
				bar.Color = clr
				bar.Stride = float64(stride)
				bar.Start = float64(start)
				bar.Width = pl.Params.BarWidth
				plt.Add(bar)
				plt.Legend.Add(lbl, bar)
				start++
			}
		}
		yidx++
		yoff += nleg
	}
	mid := (stride - 1) / 2
	if stride > 1 {
		mid = (stride - 2) / 2
	}
	if firstXY != nil && len(strCols) > 0 {
		firstXY.Table = xview
		n := xview.Len()
		for _, cp := range strCols {
			xy, _ := NewTableXYName(xview, xi, xp.TensorIndex, cp.Col, cp.TensorIndex, firstXY.YRange)
			xy.LblCol = xy.YCol
			xy.YCol = firstXY.YCol
			xy.YIndex = firstXY.YIndex

			xyl := plotter.XYLabels{}
			xyl.XYs = make(plotter.XYs, n)
			xyl.Labels = make([]string, n)

			for i := range xview.Indexes {
				y := firstXY.Value(i)
				x := float64(mid + (i%maxx)*stride)
				xyl.XYs[i] = plotter.XY{x, y}
				xyl.Labels[i] = xy.Label(i)
			}
			lbls, _ := plotter.NewLabels(xyl)
			if lbls != nil {
				plt.Add(lbls)
			}
		}
	}

	netn := pl.Table.Len() * stride
	xc := pl.Table.Table.Cols[xi]
	vals := make([]string, netn)
	for i, dx := range pl.Table.Indexes {
		pi := mid + i*stride
		if pi < netn && dx < xc.Len() {
			vals[pi] = xc.StringVal1D(dx)
		}
	}
	plt.NominalX(vals...)

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
