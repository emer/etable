// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"

	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etview"
	"github.com/emer/etable/split"
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

// GpMethodOrbit shows the median of orbital period as a function of method
var GpMethodOrbit *etable.Table

// GpMethodYear shows all stats of year described by orbit
var GpMethodYear *etable.Table

// GpMethodDecade shows number of planets found in each decade by given method
var GpMethodDecade *etable.Table

// GpDecade shows number of planets found in each decade
var GpDecade *etable.Table

// AnalyzePlanets analyzes planets.csv data following some of the examples
// given here, using pandas:
// 	https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html
func AnalyzePlanets() {
	Planets = etable.NewTable("planets")
	Planets.OpenCSV("./planets.csv", ',')

	PlanetsAll := etable.NewIdxView(Planets) // full original data

	NonNull := etable.NewIdxView(Planets)
	NonNull.Filter(etable.FilterNull) // filter out all rows with Null values

	PlanetsDesc = agg.DescAll(PlanetsAll) // individually excludes Null values in each col, but not row-wise
	PlanetsNNDesc = agg.DescAll(NonNull)  // standard descriptive stats for row-wise non-nulls

	byMethod := split.GroupBy(PlanetsAll, []string{"method"})
	split.Agg(byMethod, "orbital_period", agg.AggMedian)
	GpMethodOrbit = byMethod.AggsToTable(false) // false = include agg name in column

	byMethod.DeleteAggs()
	split.Desc(byMethod, "year") // full desc stats of year

	byMethod.Filter(func(idx int) bool {
		ag := byMethod.AggByColName("year:Std")
		return ag.Aggs[idx][0] > 0 // exclude results with 0 std
	})

	GpMethodYear = byMethod.AggsToTable(false) // false = include agg name in column

	byMethodDecade := split.GroupByFunc(PlanetsAll, func(row int) []string {
		meth := Planets.CellStringByName("method", row)
		yr := Planets.CellFloatByName("year", row)
		decade := math.Floor(yr/10) * 10
		return []string{meth, fmt.Sprintf("%gs", decade)}
	})
	byMethodDecade.SetLevels("method", "decade")

	split.Agg(byMethodDecade, "number", agg.AggSum)

	// uncomment this to switch to decade first, then method
	// byMethodDecade.ReorderLevels([]int{1, 0})
	// byMethodDecade.SortLevels()

	decadeOnly, _ := byMethodDecade.ExtractLevels([]int{1})
	split.Agg(decadeOnly, "number", agg.AggSum)
	GpDecade = decadeOnly.AggsToTable(false)

	GpMethodDecade = byMethodDecade.AggsToTable(false) // here to ensure that decadeOnly didn't mess up..

	// todo: need unstack -- should be specific to the splits data because we already have the cols and
	// groups etc -- the ExtractLevels method provides key starting point.

	// todo: pivot table -- neeeds unstack function.

	// todo: could have a generic unstack-like method that takes a column for the data to turn into columns
	// and another that has the data to put in the cells.
}

func OpenGUI() {
	width := 1600
	height := 1200

	gi.SetAppName("dataproc")
	gi.SetAppAbout(`This demonstrates data processing using etable.Table. See <a href="https://github.com/emer/etable">etable on GitHub</a>.</p>`)

	plot.DefaultFont = "Helvetica"

	win := gi.NewWindow2D("dataproc", "eTable Data Processing Demo", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tv := gi.AddNewTabView(mfr, "tv")
	tv.Viewport = vp

	tv.AddNewTab(etview.KiT_TableView, "Planets Data").(*etview.TableView).SetTable(Planets, nil)
	tv.AddNewTab(etview.KiT_TableView, "Non-Null Rows Desc").(*etview.TableView).SetTable(PlanetsNNDesc, nil)
	tv.AddNewTab(etview.KiT_TableView, "All Desc").(*etview.TableView).SetTable(PlanetsDesc, nil)
	tv.AddNewTab(etview.KiT_TableView, "By Method Orbit").(*etview.TableView).SetTable(GpMethodOrbit, nil)
	tv.AddNewTab(etview.KiT_TableView, "By Method Year").(*etview.TableView).SetTable(GpMethodYear, nil)
	tv.AddNewTab(etview.KiT_TableView, "By Method Decade").(*etview.TableView).SetTable(GpMethodDecade, nil)
	tv.AddNewTab(etview.KiT_TableView, "By Decade").(*etview.TableView).SetTable(GpDecade, nil)

	tv.SelectTabIndex(0)

	vp.UpdateEndNoSig(updt)
	win.StartEventLoop()
}

func mainrun() {
	AnalyzePlanets()
	OpenGUI()
}
