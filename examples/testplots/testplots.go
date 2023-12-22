// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"goki.dev/colors"
	"goki.dev/etable/v2/eplot"
	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etensor"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/gimain"
)

func main() { gimain.Run(app) }

func app() {
	b := gi.NewAppBody("testplots")
	b.App().About = `This runs various testing data plots using etable.Table. See <a href="https://goki.dev/etable/v2">etable on GitHub</a>.</p>`

	tv := gi.NewTabs(b)

	PlotColorSpread(tv)

	b.NewWindow().Run().Wait()
}

func PlotColorSpread(tv *gi.Tabs) {
	label := "Color Spread"
	dt := etable.NewTable(label)
	dt.SetMetaData("name", label)
	dt.SetMetaData("read-only", "true")

	sch := etable.Schema{
		{"Idx", etensor.INT, nil, nil},
		{"Collapse", etensor.INT, nil, nil},
		{"Val", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, 0)

	mx := 100
	dt.SetNumRows(mx)

	for i := 0; i < mx; i++ {
		val := colors.BinarySpacedNumber(i)
		dt.SetCellFloat("Idx", i, float64(i))      // select this to see the timecourse
		dt.SetCellFloat("Collapse", i, float64(0)) // select this to collapse all points on top
		dt.SetCellFloat("Val", i, float64(val))
	}

	pl := eplot.NewSubPlot(tv.NewTab(label))
	pl.SetTable(dt)
	pl.Params.XAxisCol = "Idx"
	pl.Params.Lines = false
	pl.Params.Points = true
	pl.ColParams("Val").On = true
}
