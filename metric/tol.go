// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"math"

	"goki.dev/mat32/v2"
)

///////////////////////////////////////////
//  Tolerance

// Tolerance32 sets a = b for any element where |a-b| <= tol.
// This can be called prior to any metric function.
func Tolerance32(a, b []float32, tol float32) {
	if len(a) != len(b) {
		panic("metric: slice lengths do not match")
	}
	for i, av := range a {
		bv := b[i]
		if mat32.IsNaN(av) || mat32.IsNaN(bv) {
			continue
		}
		if mat32.Abs(av-bv) <= tol {
			a[i] = bv
		}
	}
}

// Tolerance64 sets a = b for any element where |a-b| <= tol.
// This can be called prior to any metric function.
func Tolerance64(a, b []float64, tol float64) {
	if len(a) != len(b) {
		panic("metric: slice lengths do not match")
	}
	for i, av := range a {
		bv := b[i]
		if math.IsNaN(av) || math.IsNaN(bv) {
			continue
		}
		if math.Abs(av-bv) <= tol {
			a[i] = bv
		}
	}
}
