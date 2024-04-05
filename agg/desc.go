// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
)

// DescAggs are all the standard aggregates
var DescAggs = []Aggs{AggCount, AggMean, AggStd, AggSem, AggMin, AggQ1, AggMedian, AggQ3, AggMax}

// DescAggsND are all the standard aggregates for n-dimensional (n > 1) data -- cannot do quantiles
var DescAggsND = []Aggs{AggCount, AggMean, AggStd, AggSem, AggMin, AggMax}

// DescAll returns a table of standard descriptive aggregates for
// all numeric columns in given table, operating over all non-Null, non-NaN elements
// in each column.
func DescAll(ix *etable.IndexView) *etable.Table {
	st := ix.Table
	nonQs := []Aggs{AggCount, AggMean, AggStd, AggSem, AggMin, AggMax} // everything else done wth quantiles
	allAggs := []Aggs{AggCount, AggMean, AggStd, AggSem, AggMin, AggMax, AggQ1, AggMedian, AggQ3}
	nAgg := len(allAggs)
	sc := etable.Schema{
		{"Agg", etensor.STRING, nil, nil},
	}
	for ci := range st.Cols {
		col := st.Cols[ci]
		if col.DataType() == etensor.STRING {
			continue
		}
		sc = append(sc, etable.Column{st.ColNames[ci], etensor.FLOAT64, col.Shapes()[1:], col.DimNames()[1:]})
	}
	dt := etable.New(sc, nAgg)
	dtnm := dt.Cols[0]
	dtci := 1
	qs := []float64{.25, .5, .75}
	sq := len(nonQs)
	for ci := range st.Cols {
		col := st.Cols[ci]
		if col.DataType() == etensor.STRING {
			continue
		}
		_, csz := col.RowCellSize()
		dtst := dt.Cols[dtci]
		for i, agtyp := range nonQs {
			ag := AggIndex(ix, ci, agtyp)
			si := i * csz
			for j := 0; j < csz; j++ {
				dtst.SetFloat1D(si+j, ag[j])
			}
			if dtci == 1 {
				dtnm.SetString1D(i, AggsName(agtyp))
			}
		}
		if col.NumDims() == 1 {
			qvs := QuantilesIndex(ix, ci, qs)
			for i, qv := range qvs {
				dtst.SetFloat1D(sq+i, qv)
				dtnm.SetString1D(sq+i, AggsName(allAggs[sq+i]))
			}
		}
		dtci++
	}
	return dt
}

// DescIndex returns a table of standard descriptive aggregates
// of non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column index.
func DescIndex(ix *etable.IndexView, colIndex int) *etable.Table {
	st := ix.Table
	col := st.Cols[colIndex]
	nonQs := []Aggs{AggCount, AggMean, AggStd, AggSem} // everything else done wth quantiles
	allAggs := []Aggs{AggCount, AggMean, AggStd, AggSem, AggMin, AggQ1, AggMedian, AggQ3, AggMax}
	nAgg := len(allAggs)
	if col.NumDims() > 1 { // nd cannot do qiles
		nonQs = append(nonQs, []Aggs{AggMin, AggMax}...)
		allAggs = nonQs
		nAgg += 2
	}
	sc := etable.Schema{
		{"Agg", etensor.STRING, nil, nil},
		{st.ColNames[colIndex], etensor.FLOAT64, col.Shapes()[1:], col.DimNames()[1:]},
	}
	dt := etable.New(sc, nAgg)
	dtnm := dt.Cols[0]
	dtst := dt.Cols[1]
	_, csz := col.RowCellSize()
	for i, agtyp := range nonQs {
		ag := AggIndex(ix, colIndex, agtyp)
		si := i * csz
		for j := 0; j < csz; j++ {
			dtst.SetFloat1D(si+j, ag[j])
		}
		dtnm.SetString1D(i, AggsName(agtyp))
	}
	if col.NumDims() == 1 {
		qs := []float64{0, .25, .5, .75, 1}
		qvs := QuantilesIndex(ix, colIndex, qs)
		sq := len(nonQs)
		for i, qv := range qvs {
			dtst.SetFloat1D(sq+i, qv)
			dtnm.SetString1D(sq+i, AggsName(allAggs[sq+i]))
		}
	}
	return dt
}

// Desc returns a table of standard descriptive aggregates
// of non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
func Desc(ix *etable.IndexView, colNm string) *etable.Table {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return DescIndex(ix, colIndex)
}

// Desc returns a table of standard descriptive aggregate aggs
// of non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func DescTry(ix *etable.IndexView, colNm string) (*etable.Table, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return DescIndex(ix, colIndex), nil
}
