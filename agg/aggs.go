// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"fmt"
	"math"
	"strings"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

///////////////////////////////////////////////////
//   Desc

// DescAll returns a table of standard descriptive aggregate stats for
// all numeric columns in given table, operating over all non-Null, non-NaN elements
// in each column.
func DescAll(ix *etable.IdxTable) *etable.Table {
	st := ix.Table
	nonqNms := []string{"Count", "Mean", "Std", "Sem", "Min", "Max"} // everything else done wth quantiles
	statNms := []string{"Count", "Mean", "Std", "Sem", "Min", "Max", "25%", "50%", "75%"}
	nStat := len(statNms)
	sc := etable.Schema{
		{"Stat", etensor.STRING, nil, nil},
	}
	for ci := range st.Cols {
		col := st.Cols[ci]
		if col.DataType() == etensor.STRING {
			continue
		}
		sc = append(sc, etable.Column{st.ColNames[ci], etensor.FLOAT64, col.Shapes()[1:], col.DimNames()[1:]})
	}
	dt := etable.New(sc, nStat)
	dtnm := dt.Cols[0]
	dtci := 1
	qs := []float64{.25, .5, .75}
	sq := len(nonqNms)
	for ci := range st.Cols {
		col := st.Cols[ci]
		if col.DataType() == etensor.STRING {
			continue
		}
		_, csz := col.RowCellSize()
		dtst := dt.Cols[dtci]
		for i, snm := range nonqNms {
			stat := StatByIdx(snm, ix, ci)
			si := i * csz
			for j := 0; j < csz; j++ {
				dtst.SetFloat1D(si+j, stat[j])
			}
			if dtci == 1 {
				dtnm.SetString1D(i, snm)
			}
		}
		if col.NumDims() == 1 {
			qvs := QuantilesByIdx(ix, ci, qs)
			for i, qv := range qvs {
				dtst.SetFloat1D(sq+i, qv)
				dtnm.SetString1D(sq+i, statNms[sq+i])
			}
		}
		dtci++
	}
	return dt
}

// DescByIdx returns a table of standard descriptive aggregate stats
// of non-Null, non-NaN elements in given IdxTable indexed view of an
// etable.Table, for given column index.
func DescByIdx(ix *etable.IdxTable, colIdx int) *etable.Table {
	st := ix.Table
	col := st.Cols[colIdx]
	nonqNms := []string{"Count", "Mean", "Std", "Sem"} // everything else done wth quantiles
	statNms := []string{"Count", "Mean", "Std", "Sem", "Min", "25%", "50%", "75%", "Max"}
	nStat := len(statNms)
	if col.NumDims() > 1 { // nd cannot do qiles
		nonqNms = append(nonqNms, []string{"Min", "Max"}...)
		statNms = nonqNms
		nStat += 2
	}
	sc := etable.Schema{
		{"Stat", etensor.STRING, nil, nil},
		{st.ColNames[colIdx], etensor.FLOAT64, col.Shapes()[1:], col.DimNames()[1:]},
	}
	dt := etable.New(sc, nStat)
	dtnm := dt.Cols[0]
	dtst := dt.Cols[1]
	_, csz := col.RowCellSize()
	for i, snm := range nonqNms {
		stat := StatByIdx(snm, ix, colIdx)
		si := i * csz
		for j := 0; j < csz; j++ {
			dtst.SetFloat1D(si+j, stat[j])
		}
		dtnm.SetString1D(i, snm)
	}
	if col.NumDims() == 1 {
		qs := []float64{0, .25, .5, .75, 1}
		qvs := QuantilesByIdx(ix, colIdx, qs)
		sq := len(nonqNms)
		for i, qv := range qvs {
			dtst.SetFloat1D(sq+i, qv)
			dtnm.SetString1D(sq+i, statNms[sq+i])
		}
	}
	return dt
}

// Desc returns a table of standard descriptive aggregate stats
// of non-Null, non-NaN elements in given IdxTable indexed view of an
// etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
func Desc(ix *etable.IdxTable, colNm string) *etable.Table {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return DescByIdx(ix, colIdx)
}

// Desc returns a table of standard descriptive aggregate stats
// of non-Null, non-NaN elements in given IdxTable indexed view of an
// etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func DescTry(ix *etable.IdxTable, colNm string) (*etable.Table, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return DescByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Stat

// StatByIdx returns statistic according to given stat name applied
// to all non-Null, non-NaN elements in given IdxTable indexed view of
// an etable.Table, for given column index.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq, 25%, 1Q, Median, 50%, 2Q, 75%, 3Q (case insensitive)
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StatByIdx(statNm string, ix *etable.IdxTable, colIdx int) []float64 {
	statNm = strings.ToLower(statNm)
	switch statNm {
	case "count":
		return CountByIdx(ix, colIdx)
	case "sum":
		return SumByIdx(ix, colIdx)
	case "prod":
		return ProdByIdx(ix, colIdx)
	case "min", "0%":
		return MinByIdx(ix, colIdx)
	case "max", "100%":
		return MaxByIdx(ix, colIdx)
	case "mean":
		return MeanByIdx(ix, colIdx)
	case "var":
		return VarByIdx(ix, colIdx)
	case "std":
		return StdByIdx(ix, colIdx)
	case "sem":
		return SemByIdx(ix, colIdx)
	case "varpop", "var-pop", "var_pop":
		return VarPopByIdx(ix, colIdx)
	case "stdpop", "std-pop", "std_pop":
		return StdPopByIdx(ix, colIdx)
	case "sempop", "sem-pop", "sem_pop":
		return SemPopByIdx(ix, colIdx)
	case "sumsq":
		return SumSqByIdx(ix, colIdx)
	case "25%", "1q":
		return QuantilesByIdx(ix, colIdx, []float64{.25})
	case "median", "50%", "2q":
		return QuantilesByIdx(ix, colIdx, []float64{.5})
	case "75%", "3q":
		return QuantilesByIdx(ix, colIdx, []float64{.75})
	}
	return nil
}

// StatByIdx returns statistic according to given stat name applied
// to all non-Null, non-NaN elements in given IdxTable indexed view of
// an etable.Table, for given column name.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq (case insensitive)
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Stat(statNm string, ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return StatByIdx(statNm, ix, colIdx)
}

// StatByIdx returns statistic according to given stat name applied
// to all non-Null, non-NaN elements in given IdxTable indexed view of
// an etable.Table, for given column name.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq (case insensitive)
// If stat name not recognized, or name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StatTry(statNm string, ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	rv := StatByIdx(statNm, ix, colIdx)
	if rv == nil {
		return nil, fmt.Errorf("etable agg.StatTry: stat name: %v not recognized", statNm)
	}
	return rv, nil
}

///////////////////////////////////////////////////
//   Count

// CountByIdx returns the count of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, CountFunc)
}

// Count returns the count of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Count(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return CountByIdx(ix, colIdx)
}

// CountTry returns the count of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func CountTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return CountByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Sum

// SumByIdx returns the sum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, SumFunc)
}

// Sum returns the sum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sum(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SumByIdx(ix, colIdx)
}

// SumTry returns the sum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Prod

// ProdByIdx returns the product of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, 1, ProdFunc)
}

// Prod returns the product of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Prod(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return ProdByIdx(ix, colIdx)
}

// ProdTry returns the product of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func ProdTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return ProdByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Max

// MaxByIdx returns the maximum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, -math.MaxFloat64, MaxFunc)
}

// Max returns the maximum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Max(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MaxByIdx(ix, colIdx)
}

// MaxTry returns the maximum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MaxTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MaxByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Min

// MinByIdx returns the minimum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, math.MaxFloat64, MinFunc)
}

// Min returns the minimum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Min(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MinByIdx(ix, colIdx)
}

// MinTry returns the minimum of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MinTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MinByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Mean

// MeanByIdx returns the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	cnt := CountByIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumByIdx(ix, colIdx)
	for i := range mean {
		if cnt[i] > 0 {
			mean[i] /= cnt[i]
		}
	}
	return mean
}

// Mean returns the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Mean(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return MeanByIdx(ix, colIdx)
}

// MeanTry returns the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func MeanTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return MeanByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Var

// VarByIdx returns the sample variance of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	cnt := CountByIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumByIdx(ix, colIdx)
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
// IdxTable indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Var(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return VarByIdx(ix, colIdx)
}

// VarTry returns the sample variance of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// Sample variance is normalized by 1/(n-1) -- see VarPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Std

// StdByIdx returns the sample std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	std := VarByIdx(ix, colIdx)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// Std returns the sample std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Std(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return StdByIdx(ix, colIdx)
}

// StdTry returns the sample std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// Sample std deviation is normalized by 1/(n-1) -- see StdPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Sem

// SemByIdx returns the sample standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	cnt := CountByIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	sem := StdByIdx(ix, colIdx)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// Sem returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Sem(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SemByIdx(ix, colIdx)
}

// SemTry returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// Sample sem is normalized by 1/(n-1) -- see SemPop version for 1/n normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   VarPop

// VarPopByIdx returns the population variance of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	cnt := CountByIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	mean := SumByIdx(ix, colIdx)
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
// IdxTable indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPop(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return VarPopByIdx(ix, colIdx)
}

// VarPopTry returns the population variance of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// population variance is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func VarPopTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return VarPopByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   StdPop

// StdPopByIdx returns the population std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	std := VarPopByIdx(ix, colIdx)
	for i := range std {
		std[i] = math.Sqrt(std[i])
	}
	return std
}

// StdPop returns the population std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPop(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return StdPopByIdx(ix, colIdx)
}

// StdPopTry returns the population std deviation of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// population std dev is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func StdPopTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return StdPopByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   SemPop

// SemPopByIdx returns the population standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	cnt := CountByIdx(ix, colIdx)
	if cnt == nil {
		return nil
	}
	sem := StdPopByIdx(ix, colIdx)
	for i := range sem {
		if cnt[i] > 0 {
			sem[i] /= math.Sqrt(cnt[i])
		}
	}
	return sem
}

// SemPop returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPop(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SemPopByIdx(ix, colIdx)
}

// SemPopTry returns the standard error of the mean of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// population sem is normalized by 1/n -- see Var version for 1/(n-1) sample normalization.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SemPopTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SemPopByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   SumSq

// SumSqByIdx returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqByIdx(ix *etable.IdxTable, colIdx int) []float64 {
	return ix.AggCol(colIdx, 0, SumSqFunc)
}

// SumSq returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSq(ix *etable.IdxTable, colNm string) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return SumSqByIdx(ix, colIdx)
}

// SumSqTry returns the sum-of-squares of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func SumSqTry(ix *etable.IdxTable, colNm string) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return SumSqByIdx(ix, colIdx), nil
}

///////////////////////////////////////////////////
//   Quantiles

// QuantilesByIdx returns the given quantile(s) of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column index.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func QuantilesByIdx(ix *etable.IdxTable, colIdx int, qs []float64) []float64 {
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
// IdxTable indexed view of an etable.Table, for given column name.
// If name not found, nil is returned -- use Try version for error message.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func Quantiles(ix *etable.IdxTable, colNm string, qs []float64) []float64 {
	colIdx := ix.Table.ColByNameIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return QuantilesByIdx(ix, colIdx, qs)
}

// QuantilesTry returns the given quantile(s) of non-Null, non-NaN elements in given
// IdxTable indexed view of an etable.Table, for given column name
// If name not found, error message is returned.
// Column must be a 1d Column -- returns nil for n-dimensional columns.
// qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.  Uses linear interpolation.
// Because this requires a sort, it is more efficient to get as many quantiles
// as needed in one pass.
func QuantilesTry(ix *etable.IdxTable, colNm string, qs []float64) ([]float64, error) {
	colIdx, err := ix.Table.ColByNameIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	rv := QuantilesByIdx(ix, colIdx, qs)
	if rv == nil {
		return nil, fmt.Errorf("etable agg.QuantilesTry: either qs: %v empty or column: %v not 1D", qs, colNm)
	}
	return rv, nil
}
