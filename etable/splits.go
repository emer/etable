// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import "github.com/emer/etable/etensor"

// Splits are lists of indexed views into a given Table, that represent a particular
// way of splitting up the data, e.g., whenever a given column value changes.
// Each split can be given a name, which can be used for accessing the split
// and for labeling the results.
type Splits struct {
	Splts   []*IdxView    `desc:"the list of index views for each split"`
	Names   []string      `desc:"the name associated with each split -- if non-nil, same length as Splits"`
	Aggs    [][][]float64 `desc:"aggregate data, outer index is same len as number of splits, next inner index is number of aggregation variables, and final index is number of values per.."`
	AggVars []string      `desc:"name of aggregation variables, same length as inner index of Aggs"`
}

// NewSplitsColGroups returns a new Splits set based on the groups of values
// across the given set of column indexes, using given indexed view into
// a given table.
func NewSplitsCols(ix *IdxView, colIdxs []int) *Splits {
	srt := ix.Clone()
	srt.SortCols(colIdxs, true)
	spl := &Splits{}
	lstVals := ""
	var curIx *IdxView
	for _, rw := range srt.Idxs {
		curVals := ""
		for ci := range colIdxs {
			cl := ix.Table.Cols[ci]
			cv := cl.StringVal1D(rw)
			curVals += cv + ":" // todo: maybe need better delim?
		}
		if curVals != lstVals || curIx == nil {
			curIx = &IdxView{}
			curIx.Table = ix.Table
			curIx.AddIndex(rw) // use raw row indexes
			spl.Splts = append(spl.Splts, curIx)
			spl.Names = append(spl.Names, curVals)
			lstVals = curVals
		} else {
			curIx.AddIndex(rw)
		}
	}
	return spl
}

// todo: define standard agg functions in agg package!

// AddAgg adds an aggregation variable over splits, operating on given column index in the table,
// using given aggregation function.
func (spl *Splits) AddAgg(aggNm string, colIdx int, ini float64, fun etensor.AggFunc) {
	spl.AggVars = append(spl.AggVars, aggNm)
	nspl := len(spl.Splts)
	if len(spl.Aggs) != nspl {
		spl.Aggs = make([][][]float64, nspl)
	}
	for i, spix := range spl.Splts {
		ag := spix.AggCol(colIdx, ini, fun)
		spl.Aggs[i] = append(spl.Aggs[i], ag)
	}
}

// AggsToTable returns a Table containing aggregate data in Splits
func (spl *Splits) AggsToTable() *Table {
	return &Table{}
}
