// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"fmt"
	"log"
	"math"

	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// GenPlotBar generates a Bar plot, setting GPlot variable
func (pl *Plot2D) GenPlotBar() {
	plt, _ := plot.New() // todo: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()
	plt.BackgroundColor = nil

	// process xaxis first
	xi, _, err := pl.PlotXAxis(plt)
	if err != nil {
		return
	}
	xp := pl.Cols[xi]

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
		nys++
		if cp.Range.FixMin {
			plt.Y.Min = math.Min(plt.Y.Min, cp.Range.Min)
		}
		if cp.Range.FixMax {
			plt.Y.Max = math.Max(plt.Y.Max, cp.Range.Max)
		}
	}

	stRow := 0
	edRow := pl.Table.Rows
	offset := -0.5 * float64(nys-1) * float64(pl.Params.BarWidth)

	for _, cp := range pl.Cols {
		if !cp.On || cp == xp {
			continue
		}
		if cp.IsString {
			continue
		}

		nidx := 1
		stidx := cp.TensorIdx
		if cp.TensorIdx < 0 { // do all
			yc := pl.Table.ColByName(cp.Col)
			_, sz := yc.RowCellSize()
			nidx = sz
			stidx = 0
		}
		for ii := 0; ii < nidx; ii++ {
			idx := stidx + ii
			xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, idx)
			if firstXY == nil {
				firstXY = xy
			}
			bar, err := plotter.NewBarChart(xy, vg.Points(pl.Params.BarWidth))
			if err != nil {
				log.Println(err)
				continue
			}
			lbl := cp.Label()
			clr := cp.Color
			if nidx > 1 {
				clr, _ = gi.ColorFromString(PlotColorNames[idx%len(PlotColorNames)], nil)
				lbl = fmt.Sprintf("%s_%02d", lbl, idx)
			}
			bar.Color = clr
			bar.Offset = vg.Points(offset)
			offset += pl.Params.BarWidth
			plt.Add(bar)
			plt.Legend.Add(lbl, bar)
			if cp.ErrCol != "" {
				ec := pl.Table.ColIdx(cp.ErrCol)
				if ec >= 0 {
					xy.ErrCol = ec
					eb, _ := plotter.NewYErrorBars(xy)
					eb.LineStyle.Color = clr
					plt.Add(eb)
				}
			}
		}
	}
	if firstXY != nil && len(strCols) > 0 {
		for _, cp := range strCols {
			xy, _ := NewTableXYName(pl.Table, stRow, edRow, xi, xp.TensorIdx, cp.Col, cp.TensorIdx)
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
	xc := pl.Table.Cols[xi]
	if xc.DataType() == etensor.STRING {
		xcs := xc.(*etensor.String)
		plt.NominalX(xcs.Values...)
	}

	plt.Legend.Top = true
	plt.X.Tick.Label.Rotation = math.Pi * (pl.Params.XAxisRot / 180)
	if pl.Params.XAxisRot > 10 {
		plt.X.Tick.Label.YAlign = draw.YCenter
		plt.X.Tick.Label.XAlign = draw.XRight
	}
	pl.GPlot = plt
}
