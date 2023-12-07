// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etview"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/gimain"
	"goki.dev/grr"
)

func main() { gimain.Run(app) }

func app() {
	gi.SetAppName("dataproc")
	gi.SetAppAbout(`This demonstrates tensor grid and related functionality in etable.Table. See <a href="https://goki.dev/etable/v2">etable on GitHub</a>.</p>`)

	pats := etable.NewTable("pats")
	pats.SetMetaData("name", "TrainPats")
	pats.SetMetaData("desc", "Training patterns")
	// todo: meta data for grid size
	grr.Log(pats.OpenCSV("random_5x5_25.tsv", etable.Tab))

	b := gi.NewBody()

	tv := gi.NewTabs(b)

	nt := tv.NewTab("First")
	nt = tv.NewTab("Patterns")
	etview.NewTableView(nt).SetTable(pats)

	b.NewWindow().Run().Wait()
}
