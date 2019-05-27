// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import "math"

// These are standard AggFunc functions that can operate on etensor.Tensor or etable.Table

// CountFunc is an AggFunc that computes number of elements (non-Null, non-NaN)
// Use 0 as initial value.
func CountFunc(idx int, val float64, agg float64) float64 {
	return agg + 1
}

// SumFunc is an AggFunc that computes a sum aggregate.
// use 0 as initial value.
func SumFunc(idx int, val float64, agg float64) float64 {
	return agg + val
}

// Prodfunc is an AggFunc that computes a product aggregate.
// use 1 as initial value.
func ProdFunc(idx int, val float64, agg float64) float64 {
	return agg * val
}

// MaxFunc is an AggFunc that computes a max aggregate.
// use -math.MaxFloat64 for initial agg value.
func MaxFunc(idx int, val float64, agg float64) float64 {
	return math.Max(agg, val)
}

// MinFunc is an AggFunc that computes a min aggregate.
// use math.MaxFloat64 for initial agg value.
func MinFunc(idx int, val float64, agg float64) float64 {
	return math.Min(agg, val)
}

// SumSqFunc is an AggFunc that computes a sum-of-squares aggregate.
// use 0 as initial value.
func SumSqFunc(idx int, val float64, agg float64) float64 {
	return agg + val*val
}
