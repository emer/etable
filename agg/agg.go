// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"math"

	"github.com/goki/etable/v2/etable"
)

// Every standard Agg method in this file follows one of these signatures:

// IdxViewAggFuncIdx is an aggregation function operating on IdxView, taking a column index arg
type IdxViewAggFuncIdx func(ix *etable.IdxView, colIdx int) []float64

// IdxViewAggFunc is an aggregation function operating on IdxView, taking a column name arg
type IdxViewAggFunc func(ix *etable.IdxView, colNm string) []float64

// IdxViewAggFuncTry is an aggregation function operating on IdxView, taking a column name arg,
// returning an error message
type IdxViewAggFuncTry func(ix *etable.IdxView, colIdx int) ([]float64, error)

///////////////////////////////////////////////////
//   Count

// CountIdx returns the count of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, CountFunc)
}

// Count returns the count of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Count(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return CountIdx(ix, colIdx)
}

// CountTry returns the count of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return CountIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Sum

// SumIdx returns the sum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, SumFunc)
}

// Sum returns the sum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sum(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SumIdx(ix, colIdx)
}

// SumTry returns the sum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Prod

// ProdIdx returns the product of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, 1, ProdFunc)
}

// Prod returns the product of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Prod(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return ProdIdx(ix, colIdx)
}

// ProdTry returns the product of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return ProdIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Max

// MaxIdx returns the maximum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, -math.MaxFloat64, MaxFunc)
}

// Max returns the maximum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Max(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MaxIdx(ix, colIdx)
}

// MaxTry returns the maximum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MaxIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Min

// MinIdx returns the minimum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, math.MaxFloat64, MinFunc)
}

// Min returns the minimum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Min(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MinIdx(ix, colIdx)
}

// MinTry returns the minimum of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MinIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Mean

// MeanIdx returns the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanIdx(ix *etable.IdxView, colIdx int) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumIdx(ix, colIdx)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	return mean
}

// Mean returns the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Mean(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MeanIdx(ix, colIdx)
}

// MeanTry returns the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MeanIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Var

// VarIdx returns the sample variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarIdx(ix *etable.IdxView, colIdx int) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumIdx(ix, colIdx)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	col := ix.Table.Cols[colIdx]
	_, csz := col.RowCellSize()
	vr := ix.AggCol(colIdx, 0, func(idx int, val float64, agg float64) float64 {
		cidx := idx % csz
		dv := val - mean[cidx]
		return agg + dv*dv
	})
	for i := range vr {
		if cnt[i] > 1 {
			vr[i] /= (cnt[i] - 1)
		}
	}
	return vr
}

// Var returns the sample variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Var(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return VarIdx(ix, colIdx)
}

// VarTry returns the sample variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Std

// StdIdx returns the sample std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdIdx(ix *etable.IdxView, colIdx int) []float64 {
	std := VarIdx(ix, colIdx)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// Std returns the sample std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Std(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return StdIdx(ix, colIdx)
}

// StdTry returns the sample std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Sem

// SemIdx returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemIdx(ix *etable.IdxView, colIdx int) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	sem := StdIdx(ix, colIdx)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// Sem returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sem(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SemIdx(ix, colIdx)
}

// SemTry returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   VarPop

// VarPopIdx returns the population variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopIdx(ix *etable.IdxView, colIdx int) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumIdx(ix, colIdx)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	col := ix.Table.Cols[colIdx]
	_, csz := col.RowCellSize()
	vr := ix.AggCol(colIdx, 0, func(idx int, val float64, agg float64) float64 {
		cidx := idx % csz
		dv := val - mean[cidx]
		return agg + dv*dv
	})
	for i := range vr {
		if cnt[i] > 0 {
			vr[i] /= cnt[i]
		}
	}
	return vr
}

// VarPop returns the population variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPop(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return VarPopIdx(ix, colIdx)
}

// VarPopTry returns the population variance of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarPopIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   StdPop

// StdPopIdx returns the population std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopIdx(ix *etable.IdxView, colIdx int) []float64 {
	std := VarPopIdx(ix, colIdx)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// StdPop returns the population std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPop(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return StdPopIdx(ix, colIdx)
}

// StdPopTry returns the population std deviation of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdPopIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   SemPop

// SemPopIdx returns the population standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopIdx(ix *etable.IdxView, colIdx int) []float64 {
	cnt := CountIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	sem := StdPopIdx(ix, colIdx)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// SemPop returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPop(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SemPopIdx(ix, colIdx)
}

// SemPopTry returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemPopIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   SumSq

// SumSqIdx returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqIdx(ix *etable.IdxView, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, SumSqFunc)
}

// SumSq returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSq(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SumSqIdx(ix, colIdx)
}

// SumSqTry returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumSqIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Median

// MedianIdx returns the median of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MedianIdx(ix *etable.IdxView, colIdx int) []float64 {
	return QuantilesIdx(ix, colIdx, []float64{.5})
}

// Median returns the median of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Median(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MedianIdx(ix, colIdx)
}

// MedianTry returns the median of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MedianTry(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MedianIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Q1

// Q1Idx returns the first quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1Idx(ix *etable.IdxView, colIdx int) []float64 {
	return QuantilesIdx(ix, colIdx, []float64{.25})
}

// Q1 returns the first quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return Q1Idx(ix, colIdx)
}

// Q1Try returns the first quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1Try(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return Q1Idx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Q3

// Q3Idx returns the third quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3Idx(ix *etable.IdxView, colIdx int) []float64 {
	return QuantilesIdx(ix, colIdx, []float64{.75})
}

// Q3 returns the third quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3(ix *etable.IdxView, colNm string) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return Q3Idx(ix, colIdx)
}

// Q3Try returns the third quartile of non-Null, non-NaN elements in given
// IdxView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3Try(ix *etable.IdxView, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return Q3Idx(ix, colIdx), nil
}
