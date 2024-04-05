// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import "github.com/emer/etable/v2/etable"

// IfFunc is used for the *If aggregators -- counted if it returns true
type IfFunc func(idx int, val float64) bool

///////////////////////////////////////////////////
//   CountIf

// CountIfIndex returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIfIndex(ix *etable.IndexView, colIndex int, iffun IfFunc) []float64 {
	return ix.AggCol(colIndex, 0, func(idx int, val float64, agg float64) float64 {
		if iffun(idx, val) {
			return agg + 1
		}
		return agg
	})
}

// CountIf returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIf(ix *etable.IndexView, colNm string, iffun IfFunc) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return CountIfIndex(ix, colIndex, iffun)
}

// CountIfTry returns the count of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIfTry(ix *etable.IndexView, colNm string, iffun IfFunc) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return CountIfIndex(ix, colIndex, iffun), nil
}

///////////////////////////////////////////////////
//   PropIf

// PropIfIndex returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIfIndex(ix *etable.IndexView, colIndex int, iffun IfFunc) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	pif := CountIfIndex(ix, colIndex, iffun)
	for i := range pif {
		if cnt[i] > 0 {
			pif[i] /= cnt[i]
		}
	}
	return pif
}

// PropIf returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIf(ix *etable.IndexView, colNm string, iffun IfFunc) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return PropIfIndex(ix, colIndex, iffun)
}

// PropIfTry returns the proportion (0-1) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PropIfTry(ix *etable.IndexView, colNm string, iffun IfFunc) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return PropIfIndex(ix, colIndex, iffun), nil
}

///////////////////////////////////////////////////
//   PctIf

// PctIfIndex returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIfIndex(ix *etable.IndexView, colIndex int, iffun IfFunc) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	pif := CountIfIndex(ix, colIndex, iffun)
	for i := range pif {
		if cnt[i] > 0 {
			pif[i] /= cnt[i]
		}
	}
	return pif
}

// PctIf returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIf(ix *etable.IndexView, colNm string, iffun IfFunc) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return PctIfIndex(ix, colIndex, iffun)
}

// PctIfTry returns the percentage (0-100) of true return values for given IfFunc on
// non-Null, non-NaN elements in given IndexView indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func PctIfTry(ix *etable.IndexView, colNm string, iffun IfFunc) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return PctIfIndex(ix, colIndex, iffun), nil
}
