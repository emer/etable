// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"goki.dev/etable/v2/etensor"
)

///////////////////////////////////////////
//  DivNorm

// TensorDivNorm32 does divisive normalization by given norm function
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0.
func TensorDivNorm32(tsr *etensor.Float32, ndim int, nfunc Func32) {
	if ndim == 0 {
		DivNorm32(tsr.Values, nfunc)
	}
	if ndim >= tsr.NumDims() {
		panic("norm.TensorSubNorm32: number of dims must be < NumDims()")
	}
	sln := 1
	ln := tsr.Len()
	for i := 0; i < ndim; i++ {
		sln *= tsr.Dim(i)
	}
	dln := ln / sln
	for sl := 0; sl < sln; sl++ {
		st := sl * dln
		vl := tsr.Values[st : st+dln]
		DivNorm32(vl, nfunc)
	}
}

// TensorDivNorm64 does divisive normalization by given norm function
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0.
func TensorDivNorm64(tsr *etensor.Float64, ndim int, nfunc Func64) {
	if ndim == 0 {
		DivNorm64(tsr.Values, nfunc)
	}
	if ndim >= tsr.NumDims() {
		panic("norm.TensorSubNorm64: number of dims must be < NumDims()")
	}
	sln := 1
	ln := tsr.Len()
	for i := 0; i < ndim; i++ {
		sln *= tsr.Dim(i)
	}
	dln := ln / sln
	for sl := 0; sl < sln; sl++ {
		st := sl * dln
		vl := tsr.Values[st : st+dln]
		DivNorm64(vl, nfunc)
	}
}

///////////////////////////////////////////
//  SubNorm

// TensorSubNorm32 does subtractive normalization by given norm function
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0 (panics).
func TensorSubNorm32(tsr *etensor.Float32, ndim int, nfunc Func32) {
	if ndim == 0 {
		SubNorm32(tsr.Values, nfunc)
	}
	if ndim >= tsr.NumDims() {
		panic("norm.TensorSubNorm32: number of dims must be < NumDims()")
	}
	sln := 1
	ln := tsr.Len()
	for i := 0; i < ndim; i++ {
		sln *= tsr.Dim(i)
	}
	dln := ln / sln
	for sl := 0; sl < sln; sl++ {
		st := sl * dln
		vl := tsr.Values[st : st+dln]
		SubNorm32(vl, nfunc)
	}
}

// TensorSubNorm64 does subtractive normalization by given norm function
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0.
func TensorSubNorm64(tsr *etensor.Float64, ndim int, nfunc Func64) {
	if ndim == 0 {
		SubNorm64(tsr.Values, nfunc)
	}
	if ndim >= tsr.NumDims() {
		panic("norm.TensorSubNorm64: number of dims must be < NumDims()")
	}
	sln := 1
	ln := tsr.Len()
	for i := 0; i < ndim; i++ {
		sln *= tsr.Dim(i)
	}
	dln := ln / sln
	for sl := 0; sl < sln; sl++ {
		st := sl * dln
		vl := tsr.Values[st : st+dln]
		SubNorm64(vl, nfunc)
	}
}

///////////////////////////////////////////
//  ZScore

// TensorZScore32 subtracts the mean and divides by the standard deviation
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0 (panics).
func TensorZScore32(tsr *etensor.Float32, ndim int) {
	TensorSubNorm32(tsr, ndim, Mean32)
	TensorDivNorm32(tsr, ndim, Std32)
}

// TensorZScore64 subtracts the mean and divides by the standard deviation
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0 (panics).
func TensorZScore64(tsr *etensor.Float64, ndim int) {
	TensorSubNorm64(tsr, ndim, Mean64)
	TensorDivNorm64(tsr, ndim, Std64)
}

///////////////////////////////////////////
//  Unit

// TensorUnit32 subtracts the min and divides by the max, so that values are in 0-1 unit range
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0 (panics).
func TensorUnit32(tsr *etensor.Float32, ndim int) {
	TensorSubNorm32(tsr, ndim, Min32)
	TensorDivNorm32(tsr, ndim, Max32)
}

// TensorUnit64 subtracts the min and divides by the max, so that values are in 0-1 unit range
// computed on the first ndim dims of the tensor, where 0 = all values,
// 1 = norm each of the sub-dimensions under the first outer-most dimension etc.
// ndim must be < NumDims() if not 0 (panics).
func TensorUnit64(tsr *etensor.Float64, ndim int) {
	TensorSubNorm64(tsr, ndim, Min64)
	TensorDivNorm64(tsr, ndim, Max64)
}
