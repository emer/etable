// Copyright (c) 2019, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

//go:generate goki generate

import (
	"fmt"
	"strings"

	"goki.dev/etable/v2/etable"
)

// Aggs is a list of different standard aggregation functions, which can be used
// to choose an aggregation function
type Aggs int32 //enums:enum

const (
	// Count of number of elements
	AggCount Aggs = iota

	// Sum of elements
	AggSum

	// Product of elements
	AggProd

	// Min minimum value
	AggMin

	// Max maximum value
	AggMax

	// Mean mean value
	AggMean

	// Var sample variance (squared diffs from mean, divided by n-1)
	AggVar

	// Std sample standard deviation (sqrt of Var)
	AggStd

	// Sem sample standard error of the mean (Std divided by sqrt(n))
	AggSem

	// VarPop population variance (squared diffs from mean, divided by n)
	AggVarPop

	// StdPop population standard deviation (sqrt of VarPop)
	AggStdPop

	// SemPop population standard error of the mean (StdPop divided by sqrt(n))
	AggSemPop

	// Median middle value in sorted ordering
	AggMedian

	// Q1 first quartile = 25%ile value = .25 quantile value
	AggQ1

	// Q3 third quartile = 75%ile value = .75 quantile value
	AggQ3

	// SumSq sum of squares
	AggSumSq
)

// AggsName returns the name of the Aggs varaible without the Agg prefix..
func AggsName(ag Aggs) string {
	return strings.TrimPrefix(ag.String(), "Agg")
}

// AggIdx returns aggregate according to given agg type applied
// to all non-Null, non-NaN elements in given IdxView indexed view of
// an etable.Table, for given column index.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq, 25%, 1Q, Median, 50%, 2Q, 75%, 3Q (case insensitive)
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func AggIdx(ix *etable.IdxView, colIdx int, ag Aggs) []float64 {
	switch ag {
	case AggCount:
		return CountIdx(ix, colIdx)
	case AggSum:
		return SumIdx(ix, colIdx)
	case AggProd:
		return ProdIdx(ix, colIdx)
	case AggMin:
		return MinIdx(ix, colIdx)
	case AggMax:
		return MaxIdx(ix, colIdx)
	case AggMean:
		return MeanIdx(ix, colIdx)
	case AggVar:
		return VarIdx(ix, colIdx)
	case AggStd:
		return StdIdx(ix, colIdx)
	case AggSem:
		return SemIdx(ix, colIdx)
	case AggVarPop:
		return VarPopIdx(ix, colIdx)
	case AggStdPop:
		return StdPopIdx(ix, colIdx)
	case AggSemPop:
		return SemPopIdx(ix, colIdx)
	case AggQ1:
		return Q1Idx(ix, colIdx)
	case AggMedian:
		return MedianIdx(ix, colIdx)
	case AggQ3:
		return Q3Idx(ix, colIdx)
	case AggSumSq:
		return SumSqIdx(ix, colIdx)
	}
	return nil
}

// Agg returns aggregate according to given agg type applied
// to all non-Null, non-NaN elements in given IdxView indexed view of
// an etable.Table, for given column name.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq (case insensitive)
// If name not found, nil is returned -- use Try version for error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func Agg(ix *etable.IdxView, colNm string, ag Aggs) []float64 {
	colIdx := ix.Table.ColIdx(colNm)
	if colIdx == -1 {
		return nil
	}
	return AggIdx(ix, colIdx, ag)
}

// AggTry returns aggregate according to given agg type applied
// to all non-Null, non-NaN elements in given IdxView indexed view of
// an etable.Table, for given column name.
// valid names are: Count, Sum, Var, Std, Sem, VarPop, StdPop, SemPop,
// Min, Max, SumSq (case insensitive)
// If col name not found, returns error message.
// Return value is size of each column cell -- 1 for scalar 1D columns
// and N for higher-dimensional columns.
func AggTry(ix *etable.IdxView, colNm string, ag Aggs) ([]float64, error) {
	colIdx, err := ix.Table.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	rv := AggIdx(ix, colIdx, ag)
	if rv == nil {
		return nil, fmt.Errorf("etable agg.AggTry: agg type: %v not recognized", ag)
	}
	return rv, nil
}
