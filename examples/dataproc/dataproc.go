// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etview"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"gonum.org/v1/plot"
)

// this is the stub main for gogi that calls our actual mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

// Planets is raw data
var Planets *etable.Table

// PlanetsDesc are descriptive stats of all (non-Null) data
var PlanetsDesc *etable.Table

// PlanetsNNDesc are descriptive stats of planets where entire row is non-null
var PlanetsNNDesc *etable.Table

// AnalyzePlanets analyzes planets.csv data following some of the examples
// given here, using pandas:
// 	https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html
func AnalyzePlanets() {
	Planets = etable.NewTable("planets")
	Planets.OpenCSV("./planets.csv", ',')

	PlanetsIdx := etable.NewIdxTable(Planets) // full original data

	NonNull := etable.NewIdxTable(Planets)
	NonNull.Filter(etable.FilterNull) // filter out all rows with Null values

	PlanetsDesc = agg.DescAll(PlanetsIdx) // individually excludes Null values in each col, but not row-wise
	PlanetsNNDesc = agg.DescAll(NonNull)  // standard descriptive stats for row-wise non-nulls
}

func OpenGUI() {
	width := 1600
	height := 1200

	gi.SetAppName("dataproc")
	gi.SetAppAbout(`This demonstrates data processing using etable.Table. See <a href="https://github.com/emer/etable">etable on GitHub</a>.</p>`)

	plot.DefaultFont = "Helvetica"

	win := gi.NewWindow2D("ra25", "Leabra Random Associator", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tv := gi.AddNewTabView(mfr, "tv")

	plv := tv.AddNewTab(etview.KiT_TableView, "Planets Data").(*etview.TableView)
	plv.SetTable(Planets, nil)

	plnndscv := tv.AddNewTab(etview.KiT_TableView, "Planets Non-Null Rows Desc").(*etview.TableView)
	plnndscv.SetTable(PlanetsNNDesc, nil)

	pldscv := tv.AddNewTab(etview.KiT_TableView, "Planets All Desc").(*etview.TableView)
	pldscv.SetTable(PlanetsDesc, nil)

	tv.SelectTabIndex(0)

	vp.UpdateEndNoSig(updt)
	win.StartEventLoop()
}

func mainrun() {
	AnalyzePlanets()
	OpenGUI()
}
