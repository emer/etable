// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import "github.com/goki/ki/kit"

// Func32 is a distance / similarity metric operating on slices of float32 numbers
type Func32 func(a, b []float32) float32

// Func64 is a distance / similarity metric operating on slices of float64 numbers
type Func64 func(a, b []float64) float64

// StdMetrics is a logical type -- the subset supported by etable.
// This is copied directly from arrow.Type
// They can be expressed as either a primitive physical type
// (bytes or bits of some fixed size), a nested type consisting of other data types,
// or another data type (e.g. a timestamp encoded as an int64)
type StdMetrics int

const (
	Euclidean StdMetrics = iota
	SumSquares
	Covariance
	Correlation
	Cosine
	InvCosine
	InvCorrelation
	InnerProduct
	Abs
	Hamming
	CrossEntropy

	StdMetricsN
)

//go:generate stringer -type=StdMetrics

var KiT_StdMetrics = kit.Enums.AddEnum(StdMetricsN, false, nil)

func (ev StdMetrics) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *StdMetrics) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// StdMetricFunc32 returns a standard metric function as specified
func StdMetricFunc32(std StdMetrics) Func32 {
	switch std {
	case Euclidean:
		return Euclidean32
	case SumSquares:
		return SumSquares32
	case Covariance:
		return Covariance32
	case Correlation:
		return Correlation32
	case Cosine:
		return Cosine32
	case InvCorrelation:
		return InvCorrelation32
	case InvCosine:
		return InvCosine32
	case InnerProduct:
		return InnerProduct32
	case Abs:
		return Abs32
	case Hamming:
		return Hamming32
	case CrossEntropy:
		return CrossEntropy32
	}
	return nil
}

// StdMetricFunc64 returns a standard metric function as specified
func StdMetricFunc64(std StdMetrics) Func64 {
	switch std {
	case Euclidean:
		return Euclidean64
	case SumSquares:
		return SumSquares64
	case Covariance:
		return Covariance64
	case Correlation:
		return Correlation64
	case Cosine:
		return Cosine64
	case InvCorrelation:
		return InvCorrelation64
	case InvCosine:
		return InvCosine64
	case InnerProduct:
		return InnerProduct64
	case Abs:
		return Abs64
	case Hamming:
		return Hamming64
	case CrossEntropy:
		return CrossEntropy64
	}
	return nil
}
