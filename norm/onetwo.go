// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"math"

	"github.com/chewxy/math32"
)

///////////////////////////////////////////
//  L1

// L132 computes the sum of absolute values (L1 Norm).
// Skips NaN's
func L132(a []float32) float32 {
	ss := float32(0)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		ss += math32.Abs(av)
	}
	return ss
}

// L164 computes the sum of absolute values (L1 Norm).
// Skips NaN's
func L164(a []float64) float64 {
	ss := float64(0)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		ss += math.Abs(av)
	}
	return ss
}

///////////////////////////////////////////
//  SumSquares

// SumSquares32 computes the sum-of-squares of vector.
// Skips NaN's
func SumSquares32(a []float32) float32 {
	ss := float32(0)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		ss += av * av
	}
	return ss
}

// SumSquares64 computes the sum-of-squares of vector.
// Skips NaN's
func SumSquares64(a []float64) float64 {
	ss := float64(0)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		ss += av * av
	}
	return ss
}

///////////////////////////////////////////
//  L2

// L232 computes the square-root of sum-of-squares of vector, i.e., the L2 norm.
// Skips NaN's
func L232(a []float32) float32 {
	ss := SumSquares32(a)
	return math32.Sqrt(ss)
}

// L264 computes the square-root of sum-of-squares of vector, i.e., the L2 norm.
// Skips NaN's
func L264(a []float64) float64 {
	ss := SumSquares64(a)
	return math.Sqrt(ss)
}
