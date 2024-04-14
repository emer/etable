// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/core"
	"cogentcore.org/core/errors"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etview"
)

//go:embed *.tsv
var tsv embed.FS

func main() {
	pats := etable.NewTable("pats")
	pats.SetMetaData("name", "TrainPats")
	pats.SetMetaData("desc", "Training patterns")
	// todo: meta data for grid size
	errors.Log(pats.OpenFS(tsv, "random_5x5_25.tsv", etable.Tab))

	b := core.NewBody("grids")

	tv := core.NewTabs(b)

	// nt := tv.NewTab("First")
	nt := tv.NewTab("Patterns")
	etv := etview.NewTableView(nt).SetTable(pats)
	b.AddAppBar(etv.ConfigToolbar)

	b.RunMainWindow()
}
