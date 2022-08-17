// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"math"

	"github.com/goki/mat32"
)

// Abs32 applies the Abs function to each element in given slice
func Abs32(a []float32) {
	for i, av := range a {
		if mat32.IsNaN(av) {
			continue
		}
		a[i] = mat32.Abs(av)
	}
}

// Abs64 applies the Abs function to each element in given slice
func Abs64(a []float64) {
	for i, av := range a {
		if math.IsNaN(av) {
			continue
		}
		a[i] = math.Abs(av)
	}
}
