// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cogentcore.org/core/gi"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etable"
)

func main() {
	b := gi.NewAppBody("plot")
	b.App().About = `This demonstrates data plotting using etable.Table. See <a href="https://github.com/emer/etable/v2">etable on GitHub</a>.</p>`

	epc := etable.NewTable("epc")
	epc.OpenCSV("ra25epoch.tsv", etable.Tab)

	pl := eplot.NewPlot2D(b)
	pl.SetTable(epc)
	pl.Params.Title = "RA25 Epoch Train"
	pl.Params.XAxisCol = "Epoch"
	pl.ColParams("UnitErr").On = true

	b.AddAppBar(pl.ConfigToolbar)

	b.NewWindow().Run().Wait()
}
