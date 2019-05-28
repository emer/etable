// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"fmt"
	"math"

	"github.com/emer/etable/etable"
)

// QuantilesIdx returns the given quantile(s) of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func QuantilesIdx(ix *etable.IdxView, colIdx int, qs []float64) []float64 {
	nq := len(qs)
	if nq == 0 {
		return nil
	}
	col := ix.Table.Cols[colIdx]
	if col.NumDims() > 1 { // only valid for 1D
		return nil
	}
	rvs := make([]float64, nq)
	six := ix.Clone()                                 // leave original indexes intact
	six.Filter(func(et *etable.Table, row int) bool { // get rid of nulls in this column
		if col.IsNull1D(row) {
			return false
		}
		return true
	})
	six.SortCol(colIdx, true)
	sz := len(six.Idxs) - 1 // length of our own index list
	fsz := float64(sz)
	for i, q := range qs {
		val := 0.0
		qi := q * fsz
		lwi := math.Floor(qi)
		lwii := int(lwi)
		if lwii >= sz {
			val = col.FloatVal1D(six.Idxs[sz])
		} else if lwii < 0 {
			val = col.FloatVal1D(six.Idxs[0])
		} else {
			phi := qi - lwi
			lwv := col.FloatVal1D(six.Idxs[lwii])
			hiv := col.FloatVal1D(six.Idxs[lwii+1])
			val = (1-phi)*lwv + phi*hiv
		}
		rvs[i] = val
	}
	return rvs
}

// Quantiles returns the given quantile(s) of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func Quantiles(ix *etable.IdxView, colNm string, qs []float64) []float64 {
	colIdx := ix.Table.ColIdxByName(colNm)
	if colIdx == -1 {
		return nil
	}
	return QuantilesIdx(ix, colIdx, qs)
}

// QuantilesTry returns the given quantile(s) of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name
// If name not found, error message is returned.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func QuantilesTry(ix *etable.IdxView, colNm string, qs []float64) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	rv := QuantilesIdx(ix, colIdx, qs)
	if rv == nil {
		return nil, fmt.Errorf("etable agg.QuantilesTry: either qs: %v empty or column: %v not 1D", qs, colNm)
	}
	return rv, nil
}
