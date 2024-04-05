// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"math"

	"github.com/emer/etable/v2/etable"
)

// Every standard Agg method in this file follows one of these signatures:

// IndexViewAggFuncIndex is an aggregation function operating on IndexView, taking a column index arg
type IndexViewAggFuncIndex func(ix *etable.IndexView, colIndex int) []float64

// IndexViewAggFunc is an aggregation function operating on IndexView, taking a column name arg
type IndexViewAggFunc func(ix *etable.IndexView, colNm string) []float64

// IndexViewAggFuncTry is an aggregation function operating on IndexView, taking a column name arg,
// returning an error message
type IndexViewAggFuncTry func(ix *etable.IndexView, colIndex int) ([]float64, error)

///////////////////////////////////////////////////
//   Count

// CountIndex returns the count of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, 0, CountFunc)
}

// Count returns the count of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Count(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return CountIndex(ix, colIndex)
}

// CountTry returns the count of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return CountIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Sum

// SumIndex returns the sum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, 0, SumFunc)
}

// Sum returns the sum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sum(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return SumIndex(ix, colIndex)
}

// SumTry returns the sum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Prod

// ProdIndex returns the product of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, 1, ProdFunc)
}

// Prod returns the product of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Prod(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return ProdIndex(ix, colIndex)
}

// ProdTry returns the product of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return ProdIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Max

// MaxIndex returns the maximum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, -math.MaxFloat64, MaxFunc)
}

// Max returns the maximum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Max(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return MaxIndex(ix, colIndex)
}

// MaxTry returns the maximum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return MaxIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Min

// MinIndex returns the minimum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, math.MaxFloat64, MinFunc)
}

// Min returns the minimum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Min(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return MinIndex(ix, colIndex)
}

// MinTry returns the minimum of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return MinIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Mean

// MeanIndex returns the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanIndex(ix *etable.IndexView, colIndex int) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	mean := SumIndex(ix, colIndex)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	return mean
}

// Mean returns the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Mean(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return MeanIndex(ix, colIndex)
}

// MeanTry returns the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return MeanIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Var

// VarIndex returns the sample variance of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarIndex(ix *etable.IndexView, colIndex int) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	mean := SumIndex(ix, colIndex)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	col := ix.Table.Cols[colIndex]
	_, csz := col.RowCellSize()
	vr := ix.AggCol(colIndex, 0, func(idx int, val float64, agg float64) float64 {
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
// IndexView indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Var(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return VarIndex(ix, colIndex)
}

// VarTry returns the sample variance of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Std

// StdIndex returns the sample std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdIndex(ix *etable.IndexView, colIndex int) []float64 {
	std := VarIndex(ix, colIndex)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// Std returns the sample std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Std(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return StdIndex(ix, colIndex)
}

// StdTry returns the sample std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Sem

// SemIndex returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemIndex(ix *etable.IndexView, colIndex int) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	sem := StdIndex(ix, colIndex)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// Sem returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sem(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return SemIndex(ix, colIndex)
}

// SemTry returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   VarPop

// VarPopIndex returns the population variance of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopIndex(ix *etable.IndexView, colIndex int) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	mean := SumIndex(ix, colIndex)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	col := ix.Table.Cols[colIndex]
	_, csz := col.RowCellSize()
	vr := ix.AggCol(colIndex, 0, func(idx int, val float64, agg float64) float64 {
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
// IndexView indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPop(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return VarPopIndex(ix, colIndex)
}

// VarPopTry returns the population variance of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarPopIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   StdPop

// StdPopIndex returns the population std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopIndex(ix *etable.IndexView, colIndex int) []float64 {
	std := VarPopIndex(ix, colIndex)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// StdPop returns the population std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPop(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return StdPopIndex(ix, colIndex)
}

// StdPopTry returns the population std deviation of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdPopIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   SemPop

// SemPopIndex returns the population standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopIndex(ix *etable.IndexView, colIndex int) []float64 {
	cnt := CountIndex(ix, colIndex)
	if cnt == nil {
		return nil
	}
	sem := StdPopIndex(ix, colIndex)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// SemPop returns the standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPop(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return SemPopIndex(ix, colIndex)
}

// SemPopTry returns the standard error of the mean of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemPopIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   SumSq

// SumSqIndex returns the sum-of-squares of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqIndex(ix *etable.IndexView, colIndex int) []float64 {
	return ix.AggCol(colIndex, 0, SumSqFunc)
}

// SumSq returns the sum-of-squares of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSq(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return SumSqIndex(ix, colIndex)
}

// SumSqTry returns the sum-of-squares of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumSqIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Median

// MedianIndex returns the median of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MedianIndex(ix *etable.IndexView, colIndex int) []float64 {
	return QuantilesIndex(ix, colIndex, []float64{.5})
}

// Median returns the median of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Median(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return MedianIndex(ix, colIndex)
}

// MedianTry returns the median of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MedianTry(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return MedianIndex(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Q1

// Q1Index returns the first quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1Index(ix *etable.IndexView, colIndex int) []float64 {
	return QuantilesIndex(ix, colIndex, []float64{.25})
}

// Q1 returns the first quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return Q1Index(ix, colIndex)
}

// Q1Try returns the first quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q1Try(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return Q1Index(ix, colIndex), nil
}

///////////////////////////////////////////////////
//   Q3

// Q3Index returns the third quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3Index(ix *etable.IndexView, colIndex int) []float64 {
	return QuantilesIndex(ix, colIndex, []float64{.75})
}

// Q3 returns the third quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3(ix *etable.IndexView, colNm string) []float64 {
	colIndex := ix.Table.ColIndex(colNm)
	if colIndex == -1 {
		return nil
	}
	return Q3Index(ix, colIndex)
}

// Q3Try returns the third quartile of non-Null, non-NaN elements in given
// IndexView indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Q3Try(ix *etable.IndexView, colNm string) ([]float64, error) {
	colIndex, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return Q3Index(ix, colIndex), nil
}
