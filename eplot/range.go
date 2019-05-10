// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import "github.com/emer/etable/minmax"

// Range represents a range of values for plotting, where the min or max can optionally be fixed
type Range struct {
	minmax.F64 `desc:"Min and Max range values"`
	FixMin     bool `desc:"fix the minimum end of the range"`
	FixMax     bool `desc:"fix the maximum end of the range"`
}

// SetMin sets a fixed min value
func (rr *Range) SetMin(min float64) {
	rr.FixMin = true
	rr.Min = min
}

// SetMax sets a fixed max value
func (rr *Range) SetMax(max float64) {
	rr.FixMax = true
	rr.Max = max
}

// Range returns Max - Min
func (rr *Range) Range() float64 {
	return rr.Max - rr.Min
}
