// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import "github.com/emer/etable/etable"

// IfFunc is used for the *If aggregators -- counted if it returns true
type IfFunc func(idx int, val float64) bool

///////////////////////////////////////////////////
//   CountIf

// CountIfIdx returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIfIdx(ix *etable.IdxView, colIdx int, iffun IfFunc) []float64 {
	return ix.AggCol(colIdx, 0, func(idx int, val float64, agg float64) float64 {
		if iffun(idx, val) {
			return agg + 1
		}
		return agg
	})
}

// CountIf returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIf(ix *etable.IdxView, colNm string, iffun IfFunc) []float64 {
	colIdx := ix.Table.ColIdxByName(colNm)
	if colIdx == -1 {
		return nil
	}
	return CountIfIdx(ix, colIdx, iffun)
}

// CountIfTry returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIfTry(ix *etable.IdxView, colNm string, iffun IfFunc) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	return CountIfIdx(ix, colIdx, iffun), nil
}

///////////////////////////////////////////////////
//   PropIf

// PropIfIdx returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIfIdx(ix *etable.IdxView, colIdx int, iffun IfFunc) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	pif := CountIfIdx(ix, colIdx, iffun)
	for i := range pif {
		if cnt[i] > 0 {
			pif[i] /= cnt[i]
		}
	}
	return pif
}

// PropIf returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIf(ix *etable.IdxView, colNm string, iffun IfFunc) []float64 {
	colIdx := ix.Table.ColIdxByName(colNm)
	if colIdx == -1 {
		return nil
	}
	return PropIfIdx(ix, colIdx, iffun)
}

// PropIfTry returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIfTry(ix *etable.IdxView, colNm string, iffun IfFunc) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	return PropIfIdx(ix, colIdx, iffun), nil
}

///////////////////////////////////////////////////
//   PctIf

// PctIfIdx returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIfIdx(ix *etable.IdxView, colIdx int, iffun IfFunc) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	pif := CountIfIdx(ix, colIdx, iffun)
	for i := range pif {
		if cnt[i] > 0 {
			pif[i] /= cnt[i]
		}
	}
	return pif
}

// PctIf returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIf(ix *etable.IdxView, colNm string, iffun IfFunc) []float64 {
	colIdx := ix.Table.ColIdxByName(colNm)
	if colIdx == -1 {
		return nil
	}
	return PctIfIdx(ix, colIdx, iffun)
}

// PctIfTry returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IdxView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIfTry(ix *etable.IdxView, colNm string, iffun IfFunc) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	return PctIfIdx(ix, colIdx, iffun), nil
}
