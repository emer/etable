// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package histogram

import (
	"cogentcore.org/core/math32"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
)

// F64 generates a histogram of counts of values within given
// number of bins and min / max range.  hist vals is sized to nBins.
// if value is < min or > max it is ignored.
func F64(hist *[]float64, vals []float64, nBins int, min, max float64) {
	if cap(*hist) >= nBins {
		*hist = (*hist)[0:nBins]
	} else {
		*hist = make([]float64, nBins)
	}
	h := *hist
	// 0.1.2.3 = 3-0 = 4 bins
	inc := (max - min) / float64(nBins)
	for i := 0; i < nBins; i++ {
		h[i] = 0
	}
	for _, v := range vals {
		if v < min || v > max {
			continue
		}
		bin := int((v - min) / inc)
		if bin >= nBins {
			bin = nBins - 1
		}
		h[bin] += 1
	}
}

// F64Table generates an etable with a histogram of counts of values within given
// number of bins and min / max range. The table has columns: Value, Count
// if value is < min or > max it is ignored.
// The Value column represents the min value for each bin, with the max being
// the value of the next bin, or the max if at the end.
func F64Table(dt *etable.Table, vals []float64, nBins int, min, max float64) {
	sch := etable.Schema{
		{"Value", etensor.FLOAT64, nil, nil},
		{"Count", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, nBins)
	F64(&dt.Cols[1].(*etensor.Float64).Values, vals, nBins, min, max)
	inc := (max - min) / float64(nBins)
	vls := dt.Cols[0].(*etensor.Float64).Values
	for i := 0; i < nBins; i++ {
		vls[i] = math32.Truncate64(min+float64(i)*inc, 4)
	}
}

//////////////////////////////////////////////////////
// float32

// F32 generates a histogram of counts of values within given
// number of bins and min / max range.  hist vals is sized to nBins.
// if value is < min or > max it is ignored.
func F32(hist *[]float32, vals []float32, nBins int, min, max float32) {
	if cap(*hist) >= nBins {
		*hist = (*hist)[0:nBins]
	} else {
		*hist = make([]float32, nBins)
	}
	h := *hist
	// 0.1.2.3 = 3-0 = 4 bins
	inc := (max - min) / float32(nBins)
	for i := 0; i < nBins; i++ {
		h[i] = 0
	}
	for _, v := range vals {
		if v < min || v > max {
			continue
		}
		bin := int((v - min) / inc)
		if bin >= nBins {
			bin = nBins - 1
		}
		h[bin] += 1
	}
}

// F32Table generates an etable with a histogram of counts of values within given
// number of bins and min / max range. The table has columns: Value, Count
// if value is < min or > max it is ignored.
// The Value column represents the min value for each bin, with the max being
// the value of the next bin, or the max if at the end.
func F32Table(dt *etable.Table, vals []float32, nBins int, min, max float32) {
	sch := etable.Schema{
		{"Value", etensor.FLOAT32, nil, nil},
		{"Count", etensor.FLOAT32, nil, nil},
	}
	dt.SetFromSchema(sch, nBins)
	F32(&dt.Cols[1].(*etensor.Float32).Values, vals, nBins, min, max)
	inc := (max - min) / float32(nBins)
	vls := dt.Cols[0].(*etensor.Float32).Values
	for i := 0; i < nBins; i++ {
		vls[i] = math32.Truncate(min+float32(i)*inc, 4)
	}
}
