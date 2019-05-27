// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsragg

import (
	"math"

	"github.com/emer/etable/agg"
	"github.com/emer/etable/etensor"
)

// Count returns the count of non-Null, non-NaN elements in given Tensor.
func Count(tsr etensor.Tensor) float64 {
	return tsr.Agg(0, agg.CountFunc)
}

// Sum returns the sum of non-Null, non-NaN elements in given Tensor.
func Sum(tsr etensor.Tensor) float64 {
	return tsr.Agg(0, agg.SumFunc)
}

// Prod returns the product of non-Null, non-NaN elements in given Tensor.
func Prod(tsr etensor.Tensor) float64 {
	return tsr.Agg(0, agg.ProdFunc)
}

// Max returns the maximum of non-Null, non-NaN elements in given Tensor.
func Max(tsr etensor.Tensor) float64 {
	return tsr.Agg(-math.MaxFloat64, agg.MaxFunc)
}

// Min returns the minimum of non-Null, non-NaN elements in given Tensor.
func Min(tsr etensor.Tensor) float64 {
	return tsr.Agg(math.MaxFloat64, agg.MinFunc)
}

// Mean returns the mean of non-Null, non-NaN elements in given Tensor.
func Mean(tsr etensor.Tensor) float64 {
	cnt := Count(tsr)
	if cnt == 0 {
		return 0
	}
	return Sum(tsr) / cnt
}

// Var returns the sample variance of non-Null, non-NaN elements in given Tensor.
func Var(tsr etensor.Tensor) float64 {
	cnt := Count(tsr)
	if cnt < 2 {
		return 0
	}
	mean := Sum(tsr) / cnt
	vr := tsr.Agg(0, func(idx int, val float64, agg float64) float64 {
		dv := val - mean
		return agg + dv*dv
	})
	return vr / (cnt - 1)
}

// Std returns the sample standard deviation of non-Null, non-NaN elements in given Tensor.
func Std(tsr etensor.Tensor) float64 {
	return math.Sqrt(Var(tsr))
}

// Sem returns the sample standard error of the mean of non-Null, non-NaN elements in given Tensor.
func Sem(tsr etensor.Tensor) float64 {
	cnt := Count(tsr)
	if cnt < 2 {
		return 0
	}
	return Std(tsr) / math.Sqrt(cnt)
}

// VarPop returns the population variance of non-Null, non-NaN elements in given Tensor.
func VarPop(tsr etensor.Tensor) float64 {
	cnt := Count(tsr)
	if cnt < 2 {
		return 0
	}
	mean := Sum(tsr) / cnt
	vr := tsr.Agg(0, func(idx int, val float64, agg float64) float64 {
		dv := val - mean
		return agg + dv*dv
	})
	return vr / cnt
}

// StdPop returns the population standard deviation of non-Null, non-NaN elements in given Tensor.
func StdPop(tsr etensor.Tensor) float64 {
	return math.Sqrt(VarPop(tsr))
}

// SemPop returns the population standard error of the mean of non-Null, non-NaN elements in given Tensor.
func SemPop(tsr etensor.Tensor) float64 {
	cnt := Count(tsr)
	if cnt < 2 {
		return 0
	}
	return StdPop(tsr) / math.Sqrt(cnt)
}

// SumSq returns the sum-of-squares of non-Null, non-NaN elements in given Tensor.
func SumSq(tsr etensor.Tensor) float64 {
	return tsr.Agg(0, agg.SumSqFunc)
}
