// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"goki.dev/etable/v2/eplot"
	"goki.dev/etable/v2/etable"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/gimain"
)

func main() { gimain.Run(app) }

func app() {
	gi.SetAppName("plot")
	gi.SetAppAbout(`This demonstrates data plotting using etable.Table. See <a href="https://goki.dev/etable/v2">etable on GitHub</a>.</p>`)

	b := gi.NewBody()

	epc := etable.NewTable("epc")
	epc.OpenCSV("ra25epoch.tsv", etable.Tab)

	pl := eplot.NewPlot2D(b)
	pl.SetTable(epc)
	pl.Params.XAxisCol = "Epoch"
	pl.ColParams("UnitErr").On = true
	pl.ColsUpdate()
	pl.Update()

	b.AddTopBar(func(pw gi.Widget) {
		tb := b.DefaultTopAppBar(pw)
		pl.PlotTopAppBar(tb)
	})

	b.NewWindow().Run().Wait()
}
