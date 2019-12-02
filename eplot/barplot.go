// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"log"
	"math"

	"github.com/emer/etable/etensor"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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

	stRow := 0
	edRow := pl.Table.Rows
	nys := 0
	for _, cp := range pl.Cols {
		if !cp.On || cp == xp {
			continue
		}
		if cp.IsString {
			continue
		}
		nys++
	}
	offset := -0.5 * float64(nys) * float64(pl.Params.BarWidth)

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
		bar, err := plotter.NewBarChart(xy, vg.Points(pl.Params.BarWidth))
		if err != nil {
			log.Println(err)
			return
		}
		bar.Color = cp.Color
		bar.Offset = vg.Points(offset)
		offset += pl.Params.BarWidth
		plt.Add(bar)
		plt.Legend.Add(cp.Label(), bar)
		if cp.ErrCol != "" {
			ec := pl.Table.ColIdx(cp.ErrCol)
			if ec >= 0 {
				xy.ErrCol = ec
				eb, _ := plotter.NewYErrorBars(xy)
				plt.Add(eb)
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

	// Use string labels for X axis if X is a string
	xc := pl.Table.Cols[xi]
	if xc.DataType() == etensor.STRING {
		xcs := xc.(*etensor.String)
		plt.NominalX(xcs.Values...)
	}

	plt.Legend.Top = true
	pl.GPlot = plt
}
