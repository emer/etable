// Copyright (c) 2021, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convolve

import (
	"github.com/goki/etable/v2/etable"
	"github.com/goki/etable/v2/etensor"
)

// SmoothTable returns a cloned table with each of the floating-point
// columns in the source table smoothed over rows.
// khalf is the half-width of the Gaussian smoothing kernel,
// where larger values produce more smoothing.  A sigma of .5
// is used for the kernel.
func SmoothTable(src *etable.Table, khalf int) *etable.Table {
	k64 := GaussianKernel64(khalf, .5)
	k32 := GaussianKernel32(khalf, .5)
	dest := src.Clone()
	for ci, sci := range src.Cols {
		switch sci.DataType() {
		case etensor.FLOAT32:
			sc := sci.(*etensor.Float32)
			dc := dest.Cols[ci].(*etensor.Float32)
			Slice32(&dc.Values, sc.Values, k32)
		case etensor.FLOAT64:
			sc := sci.(*etensor.Float64)
			dc := dest.Cols[ci].(*etensor.Float64)
			Slice64(&dc.Values, sc.Values, k64)
		}
	}
	return dest
}
