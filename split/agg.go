// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package split

import (
	"fmt"

	"github.com/goki/etable/v2/agg"
	"github.com/goki/etable/v2/etable"
)

// AggIdx performs aggregation using given standard aggregation function across
// all splits, and returns the SplitAgg container of the results, which are also
// stored in the Splits.  Column is specified by index.
func AggIdx(spl *etable.Splits, colIdx int, aggTyp agg.Aggs) *etable.SplitAgg {
	ag := spl.AddAgg(agg.AggsName(aggTyp), colIdx)
	for _, sp := range spl.Splits {
		agv := agg.AggIdx(sp, colIdx, aggTyp)
		ag.Aggs = append(ag.Aggs, agv)
	}
	return ag
}

// Agg performs aggregation using given standard aggregation function across
// all splits, and returns the SplitAgg container of the results, which are also
// stored in the Splits.  Column is specified by name -- see Try for error msg version.
func Agg(spl *etable.Splits, colNm string, aggTyp agg.Aggs) *etable.SplitAgg {
	dt := spl.Table()
	if dt == nil {
		return nil
	}
	return AggIdx(spl, dt.ColIdx(colNm), aggTyp)
}

// AggTry performs aggregation using given standard aggregation function across
// all splits, and returns the SplitAgg container of the results, which are also
// stored in the Splits.  Column is specified by name -- returns error for bad column name.
func AggTry(spl *etable.Splits, colNm string, aggTyp agg.Aggs) (*etable.SplitAgg, error) {
	dt := spl.Table()
	if dt == nil {
		return nil, fmt.Errorf("split.AggTry: No splits to aggregate over")
	}
	colIdx, err := dt.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return AggIdx(spl, colIdx, aggTyp), nil
}

// AggAllNumericCols performs aggregation using given standard aggregation function across
// all splits, for all number-valued columns in the table.
func AggAllNumericCols(spl *etable.Splits, aggTyp agg.Aggs) {
	dt := spl.Table()
	for ci, cl := range dt.Cols {
		if !cl.DataType().IsNumeric() {
			continue
		}
		AggIdx(spl, ci, aggTyp)
	}
}

///////////////////////////////////////////////////
//   Desc

// DescIdx performs aggregation using standard aggregation functions across
// all splits, and stores results in the Splits.  Column is specified by index.
func DescIdx(spl *etable.Splits, colIdx int) {
	dt := spl.Table()
	if dt == nil {
		return
	}
	col := dt.Cols[colIdx]
	allAggs := agg.DescAggs
	if col.NumDims() > 1 { // nd cannot do qiles
		allAggs = agg.DescAggsND
	}
	for _, ag := range allAggs {
		AggIdx(spl, colIdx, ag)
	}
}

// Desc performs aggregation using standard aggregation functions across
// all splits, and stores results in the Splits.
// Column is specified by name -- see Try for error msg version.
func Desc(spl *etable.Splits, colNm string) {
	dt := spl.Table()
	if dt == nil {
		return
	}
	DescIdx(spl, dt.ColIdx(colNm))
}

// DescTry performs aggregation using standard aggregation functions across
// all splits, and stores results in the Splits.
// Column is specified by name -- returns error for bad column name.
func DescTry(spl *etable.Splits, colNm string) error {
	dt := spl.Table()
	if dt == nil {
		return fmt.Errorf("split.DescTry: No splits to aggregate over")
	}
	colIdx, err := dt.ColIdxTry(colNm)
	if err != nil {
		return err
	}
	DescIdx(spl, colIdx)
	return nil
}
