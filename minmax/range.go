// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minmax

// Range32 represents a range of values for plotting, where the min or max can optionally be fixed
type Range32 struct {

	// Min and Max range values
	F32 `desc:"Min and Max range values"`

	// fix the minimum end of the range
	FixMin bool `desc:"fix the minimum end of the range"`

	// fix the maximum end of the range
	FixMax bool `desc:"fix the maximum end of the range"`
}

// SetMin sets a fixed min value
func (rr *Range32) SetMin(min float32) {
	rr.FixMin = true
	rr.Min = min
}

// SetMax sets a fixed max value
func (rr *Range32) SetMax(max float32) {
	rr.FixMax = true
	rr.Max = max
}

// Range returns Max - Min
func (rr *Range32) Range() float32 {
	return rr.Max - rr.Min
}

///////////////////////////////////////////////////////////////////////
//  64

// Range64 represents a range of values for plotting, where the min or max can optionally be fixed
type Range64 struct {

	// Min and Max range values
	F64 `desc:"Min and Max range values"`

	// fix the minimum end of the range
	FixMin bool `desc:"fix the minimum end of the range"`

	// fix the maximum end of the range
	FixMax bool `desc:"fix the maximum end of the range"`
}

// SetMin sets a fixed min value
func (rr *Range64) SetMin(min float64) {
	rr.FixMin = true
	rr.Min = min
}

// SetMax sets a fixed max value
func (rr *Range64) SetMax(max float64) {
	rr.FixMax = true
	rr.Max = max
}

// Range returns Max - Min
func (rr *Range64) Range() float64 {
	return rr.Max - rr.Min
}
