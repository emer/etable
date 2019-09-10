// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"math"

	"github.com/chewxy/math32"
)

///////////////////////////////////////////
//  N

// N32 computes the number of non-NaN vector values.
// Skips NaN's
func N32(a []float32) float32 {
	n := 0
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		n++
	}
	return float32(n)
}

// N64 computes the number of non-NaN vector values.
// Skips NaN's
func N64(a []float64) float64 {
	n := 0
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		n++
	}
	return float64(n)
}

///////////////////////////////////////////
//  Sum

// Sum32 computes the sum of vector values.
// Skips NaN's
func Sum32(a []float32) float32 {
	s := float32(0)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		s += av
	}
	return s
}

// Sum64 computes the sum of vector values.
// Skips NaN's
func Sum64(a []float64) float64 {
	s := float64(0)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		s += av
	}
	return s
}

///////////////////////////////////////////
//  Mean

// Mean32 computes the mean of the vector (sum / N).
// Skips NaN's
func Mean32(a []float32) float32 {
	s := float32(0)
	n := 0
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		s += av
		n++
	}
	if n > 0 {
		s /= float32(n)
	}
	return s
}

// Mean64 computes the mean of the vector (sum / N).
// Skips NaN's
func Mean64(a []float64) float64 {
	s := float64(0)
	n := 0
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		s += av
		n++
	}
	if n > 0 {
		s /= float64(n)
	}
	return s
}

///////////////////////////////////////////
//  Var

// Var32 returns the sample variance of non-NaN elements.
func Var32(a []float32) float32 {
	mean := Mean32(a)
	n := 0
	s := float32(0)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		dv := av - mean
		s += dv * dv
		n++
	}
	if n > 1 {
		s /= float32(n - 1)
	}
	return s
}

// Var64 returns the sample variance of non-NaN elements.
func Var64(a []float64) float64 {
	mean := Mean64(a)
	n := 0
	s := float64(0)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		dv := av - mean
		s += dv * dv
		n++
	}
	if n > 1 {
		s /= float64(n - 1)
	}
	return s
}

///////////////////////////////////////////
//  Std

// Std32 returns the sample standard deviation of non-NaN elements in vector.
func Std32(a []float32) float32 {
	return math32.Sqrt(Var32(a))
}

// Std64 returns the sample standard deviation of non-NaN elements in vector.
func Std64(a []float64) float64 {
	return math.Sqrt(Var64(a))
}

///////////////////////////////////////////
//  Max

// Max32 computes the max over vector values.
// Skips NaN's
func Max32(a []float32) float32 {
	m := float32(-math.MaxFloat32)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		m = math32.Max(m, av)
	}
	return m
}

// Max64 computes the max over vector values.
// Skips NaN's
func Max64(a []float64) float64 {
	m := float64(-math.MaxFloat64)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		m = math.Max(m, av)
	}
	return m
}

///////////////////////////////////////////
//  MaxAbs

// MaxAbs32 computes the max of absolute value over vector values.
// Skips NaN's
func MaxAbs32(a []float32) float32 {
	m := float32(0)
	for _, av := range a {
		if math32.IsNaN(av) {
			continue
		}
		m = math32.Max(m, math32.Abs(av))
	}
	return m
}

// MaxAbs64 computes the max over vector values.
// Skips NaN's
func MaxAbs64(a []float64) float64 {
	m := float64(0)
	for _, av := range a {
		if math.IsNaN(av) {
			continue
		}
		m = math.Max(m, math.Abs(av))
	}
	return m
}
