// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"math"

	"goki.dev/etable/v2/etensor"
)

// ClosestRow32 returns the closest fit between probe pattern and patterns in
// an etensor.Float32 where the outer-most dimension is assumed to be a row
// (e.g., as a column in an etable), using the given metric function,
// *which must have the Increasing property* -- i.e., larger = further.
// returns the row and metric value for that row.
// Col cell sizes must match size of probe (panics if not).
func ClosestRow32(probe *etensor.Float32, col *etensor.Float32, mfun Func32) (int, float32) {
	rows := col.Dim(0)
	csz := col.Len() / rows
	if csz != probe.Len() {
		panic("metric.ClosestRow32: probe size != cell size of tensor column!\n")
	}
	ci := -1
	minv := float32(math.MaxFloat32)
	for ri := 0; ri < rows; ri++ {
		st := ri * csz
		rvals := col.Values[st : st+csz]
		v := mfun(probe.Values, rvals)
		if v < minv {
			ci = ri
			minv = v
		}
	}
	return ci, minv
}

// ClosestRow64 returns the closest fit between probe pattern and patterns in
// an etensor.Tensor where the outer-most dimension is assumed to be a row
// (e.g., as a column in an etable), using the given metric function,
// *which must have the Increasing property* -- i.e., larger = further.
// returns the row and metric value for that row.
// Col cell sizes must match size of probe (panics if not).
// Optimized for etensor.Float64 but works for any tensor.
func ClosestRow64(probe etensor.Tensor, col etensor.Tensor, mfun Func64) (int, float64) {
	rows := col.Dim(0)
	csz := col.Len() / rows
	if csz != probe.Len() {
		panic("metric.ClosestRow64: probe size != cell size of tensor column!\n")
	}
	ci := -1
	minv := math.MaxFloat64
	fp, pok := probe.(*etensor.Float64)
	fc, cok := col.(*etensor.Float64)
	if pok && cok {
		for ri := 0; ri < rows; ri++ {
			st := ri * csz
			rvals := fc.Values[st : st+csz]
			v := mfun(fp.Values, rvals)
			if v < minv {
				ci = ri
				minv = v
			}
		}
	} else if cok {
		var fpv []float64
		probe.Floats(&fpv)
		for ri := 0; ri < rows; ri++ {
			st := ri * csz
			rvals := fc.Values[st : st+csz]
			v := mfun(fpv, rvals)
			if v < minv {
				ci = ri
				minv = v
			}
		}
	} else {
		var fpv, fcv []float64
		probe.Floats(&fpv)
		col.Floats(&fcv)
		for ri := 0; ri < rows; ri++ {
			st := ri * csz
			rvals := fcv[st : st+csz]
			v := mfun(fpv, rvals)
			if v < minv {
				ci = ri
				minv = v
			}
		}
	}
	return ci, minv
}

// ClosestRow32Py returns the closest fit between probe pattern and patterns in
// an etensor.Float32 where the outer-most dimension is assumed to be a row
// (e.g., as a column in an etable), using the given metric function,
// *which must have the Increasing property* -- i.e., larger = further.
// returns the row and metric value for that row.
// Col cell sizes must match size of probe (panics if not).
// Py version is for Python, returns a slice with row, cor, takes std metric
func ClosestRow32Py(probe *etensor.Float32, col *etensor.Float32, std StdMetrics) []float32 {
	row, cor := ClosestRow32(probe, col, StdFunc32(std))
	return []float32{float32(row), cor}
}

// ClosestRow64Py returns the closest fit between probe pattern and patterns in
// an etensor.Tensor where the outer-most dimension is assumed to be a row
// (e.g., as a column in an etable), using the given metric function,
// *which must have the Increasing property* -- i.e., larger = further.
// returns the row and metric value for that row.
// Col cell sizes must match size of probe (panics if not).
// Optimized for etensor.Float64 but works for any tensor.
// Py version is for Python, returns a slice with row, cor, takes std metric
func ClosestRow64Py(probe etensor.Tensor, col etensor.Tensor, std StdMetrics) []float64 {
	row, cor := ClosestRow64(probe, col, StdFunc64(std))
	return []float64{float64(row), cor}
}
