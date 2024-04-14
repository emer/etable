// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/core"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etable"
)

//go:embed *.tsv
var tsv embed.FS

func main() {
	b := core.NewBody("plot")

	epc := etable.NewTable("epc")
	epc.OpenFS(tsv, "ra25epoch.tsv", etable.Tab)

	pl := eplot.NewPlot2D(b)
	pl.SetTable(epc)
	pl.Params.Title = "RA25 Epoch Train"
	pl.Params.XAxisCol = "Epoch"
	pl.ColParams("UnitErr").On = true

	b.AddAppBar(pl.ConfigToolbar)

	b.RunMainWindow()
}
