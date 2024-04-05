// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package split

import (
	"fmt"

	"github.com/emer/etable/v2/agg"
	"github.com/emer/etable/v2/etable"
)

// AggIndex performs aggregation using given standard aggregation function across
// all splits, and returns the SplitAgg container of the results, which are also
// stored in the Splits.  Column is specified by index.
func AggIndex(spl *etable.Splits, colIndex int, aggTyp agg.Aggs) *etable.SplitAgg {
	ag := spl.AddAgg(agg.AggsName(aggTyp), colIndex)
	for _, sp := range spl.Splits {
		agv := agg.AggIndex(sp, colIndex, aggTyp)
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
	return AggIndex(spl, dt.ColIndex(colNm), aggTyp)
}

// AggTry performs aggregation using given standard aggregation function across
// all splits, and returns the SplitAgg container of the results, which are also
// stored in the Splits.  Column is specified by name -- returns error for bad column name.
func AggTry(spl *etable.Splits, colNm string, aggTyp agg.Aggs) (*etable.SplitAgg, error) {
	dt := spl.Table()
	if dt == nil {
		return nil, fmt.Errorf("split.AggTry: No splits to aggregate over")
	}
	colIndex, err := dt.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return AggIndex(spl, colIndex, aggTyp), nil
}

// AggAllNumericCols performs aggregation using given standard aggregation function across
// all splits, for all number-valued columns in the table.
func AggAllNumericCols(spl *etable.Splits, aggTyp agg.Aggs) {
	dt := spl.Table()
	for ci, cl := range dt.Cols {
		if !cl.DataType().IsNumeric() {
			continue
		}
		AggIndex(spl, ci, aggTyp)
	}
}

///////////////////////////////////////////////////
//   Desc

// DescIndex performs aggregation using standard aggregation functions across
// all splits, and stores results in the Splits.  Column is specified by index.
func DescIndex(spl *etable.Splits, colIndex int) {
	dt := spl.Table()
	if dt == nil {
		return
	}
	col := dt.Cols[colIndex]
	allAggs := agg.DescAggs
	if col.NumDims() > 1 { // nd cannot do qiles
		allAggs = agg.DescAggsND
	}
	for _, ag := range allAggs {
		AggIndex(spl, colIndex, ag)
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
	DescIndex(spl, dt.ColIndex(colNm))
}

// DescTry performs aggregation using standard aggregation functions across
// all splits, and stores results in the Splits.
// Column is specified by name -- returns error for bad column name.
func DescTry(spl *etable.Splits, colNm string) error {
	dt := spl.Table()
	if dt == nil {
		return fmt.Errorf("split.DescTry: No splits to aggregate over")
	}
	colIndex, err := dt.ColIndexTry(colNm)
	if err != nil {
		return err
	}
	DescIndex(spl, colIndex)
	return nil
}
