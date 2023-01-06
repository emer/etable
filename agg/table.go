// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package agg

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
)

// MeanTables returns an etable.Table with the mean values across all float
// columns of the input tables, which must have the same columns but not
// necessarily the same number of rows.
func MeanTables(dts []*etable.Table) *etable.Table {
	nt := len(dts)
	if nt == 0 {
		return nil
	}
	maxRows := 0
	var maxdt *etable.Table
	for _, dt := range dts {
		if dt.Rows > maxRows {
			maxRows = dt.Rows
			maxdt = dt
		}
	}
	if maxRows == 0 {
		return nil
	}
	ot := maxdt.Clone()

	// N samples per row
	rns := make([]int, maxRows)
	for _, dt := range dts {
		dnr := dt.Rows
		mx := ints.MinInt(dnr, maxRows)
		for ri := 0; ri < mx; ri++ {
			rns[ri]++
		}
	}
	for ci, cl := range ot.Cols {
		if cl.DataType() != etensor.FLOAT32 && cl.DataType() != etensor.FLOAT64 {
			continue
		}
		_, cells := cl.RowCellSize()
		for di, dt := range dts {
			if di == 0 {
				continue
			}
			dc := dt.Cols[ci]
			dnr := dt.Rows
			mx := ints.MinInt(dnr, maxRows)
			for ri := 0; ri < mx; ri++ {
				si := ri * cells
				for j := 0; j < cells; j++ {
					ci := si + j
					cv := cl.FloatVal1D(ci)
					cv += dc.FloatVal1D(ci)
					cl.SetFloat1D(ci, cv)
				}
			}
		}
		for ri := 0; ri < maxRows; ri++ {
			si := ri * cells
			for j := 0; j < cells; j++ {
				ci := si + j
				cv := cl.FloatVal1D(ci)
				if rns[ri] > 0 {
					cv /= float64(rns[ri])
					cl.SetFloat1D(ci, cv)
				}
			}
		}
	}
	return ot
}
