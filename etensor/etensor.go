// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"gonum.org/v1/gonum/mat"
)

//go:generate tmpl -i -data=numeric.tmpldata numeric.gen.go.tmpl

// Tensor is the general interface for n-dimensional tensors.
//
// Tensor is automatically a gonum/mat.Matrix, implementing the Dims(), At(), and T() methods
// which automatically operate on the inner-most two dimensions, assuming default row-major
// layout. Error messages will be logged if applied to a Tensor with less than 2 dimensions,
// and care should be taken when using with > 2 dimensions (e.g., will only affect the first
// 2D subspace within a higher-dimensional space -- typically you'll want to call SubSpace
// to get a 2D subspace of the higher-dimensional Tensor (SubSpace is not part of interface
// as it returns the specific type, but is defined for all Tensor types).
//
type Tensor interface {
	mat.Matrix

	// Len returns the number of elements in the tensor (product of shape dimensions).
	Len() int

	// DataType returns the type of data, using arrow.DataType (ID() is the arrow.Type enum value)
	DataType() Type

	// ShapeObj returns a pointer to the shape object that fully parameterizes the tensor shape
	ShapeObj() *Shape

	// Shapes returns the size in each dimension of the tensor. (Shape is the full Shape struct)
	Shapes() []int

	// Strides returns the number of elements to step in each dimension when traversing the tensor.
	Strides() []int

	// Shape64 returns the size in each dimension using int64 (arrow compatbile)
	Shape64() []int64

	// Strides64 returns the strides in each dimension using int64 (arrow compatbile)
	Strides64() []int64

	// NumDims returns the number of dimensions of the tensor.
	NumDims() int

	// Dim returns the size of the given dimension
	Dim(i int) int

	// DimNames returns the string slice of dimension names
	DimNames() []string

	// DimName returns the name of the i-th dimension.
	DimName(i int) string

	IsContiguous() bool

	// IsRowMajor returns true if shape is row-major organized:
	// first dimension is the row or outer-most storage dimension.
	// Values *along a row* are contiguous, as you increment along
	// the minor, inner-most (column) dimension.
	// Importantly: ColMajor and RowMajor both have the *same*
	// actual memory storage arrangement, with values along a row
	// (across columns) contiguous in memory -- the only difference
	// is in the order of the indexes used to access this memory.
	IsRowMajor() bool

	// IsColMajor returns true if shape is column-major organized:
	// first dimension is column, i.e., inner-most storage dimension.
	// Values *along a row* are contiguous, as you increment along
	// the major inner-most (column) dimension.
	// Importantly: ColMajor and RowMajor both have the *same*
	// actual memory storage arrangement, with values along a row
	// (across columns) contiguous in memory -- the only difference
	// is in the order of the indexes used to access this memory.
	IsColMajor() bool

	// RowCellSize returns the size of the outer-most Row shape dimension,
	// and the size of all the remaining inner dimensions (the "cell" size)
	// e.g., for Tensors that are columns in a data table.
	// Only valid for RowMajor organization.
	RowCellSize() (rows, cells int)

	// Offset returns the flat 1D array / slice index into an element
	// at the given n-dimensional index.
	// No checking is done on the length or size of the index values
	// relative to the shape of the tensor.
	Offset(i []int) int

	// IsNull returns true if the given index has been flagged as a Null
	// (undefined, not present) value
	IsNull(i []int) bool

	// IsNull1D returns true if the given 1-dimensional index has been flagged as a Null
	// (undefined, not present) value
	IsNull1D(i int) bool

	// SetNull sets whether given index has a null value or not.
	// All values are assumed valid (non-Null) until marked otherwise, and calling
	// this method creates a Null bitslice map if one has not already been set yet.
	SetNull(i []int, nul bool)

	// SetNull1D sets whether given 1-dimensional index has a null value or not.
	// All values are assumed valid (non-Null) until marked otherwise, and calling
	// this method creates a Null bitslice map if one has not already been set yet.
	SetNull1D(i int, nul bool)

	// Generic accessor routines support Float (float64) or String, either full dimensional or 1D

	// FloatVal returns the value of given index as a float64
	FloatVal(i []int) float64

	// SetFloat sets the value of given index as a float64
	SetFloat(i []int, val float64)

	// StringVal returns the value of given index as a string
	StringVal(i []int) string

	// SetString sets the value of given index as a string
	SetString(i []int, val string)

	// FloatVal1D returns the value of given 1-dimensional index (0-Len()-1) as a float64
	FloatVal1D(i int) float64

	// SetFloat1D sets the value of given 1-dimensional index (0-Len()-1) as a float64
	SetFloat1D(i int, val float64)

	// Floats1D returns a flat []float64 slice of all elements in the tensor
	// For Float64 tensor type, this directly returns its underlying Values
	// which are writable as well -- for all others this is a new slice (read only).
	// This can be used for all of the gonum/floats methods for basic math, gonum/stats, etc
	Floats1D() []float64

	// StringVal1D returns the value of given 1-dimensional index (0-Len()-1) as a string
	StringVal1D(i int) string

	// SetString1D sets the value of given 1-dimensional index (0-Len()-1) as a string
	SetString1D(i int, val string)

	// SubSpace returns a new tensor as a subspace of the current one, incorporating the
	// given number of dimensions (0 < subdim < NumDims of this tensor). Only valid for
	// row or column major layouts.
	// * subdim are the inner, contiguous dimensions (i.e., the last in RowMajor
	//   and the first in ColMajor).
	// * offs are offsets for the outer dimensions (len = NDims - subdim)
	//   for the subspace to return.
	// The new tensor points to the values of the this tensor (i.e., modifications
	// will affect both), as its Values slice is a view onto the original (which
	// is why only inner-most contiguous supsaces are supported).
	// Use Clone() method to separate the two.
	SubSpace(subdim int, offs []int) Tensor

	// SubSpaceTry is SubSpace but returns an error message if the subdim and offs
	// do not match the tensor Shape.
	SubSpaceTry(subdim int, offs []int) (Tensor, error)

	// Range returns the min, max (and associated indexes, -1 = no values) for the tensor.
	// This is needed for display and is thus in the core api in optimized form
	// Other math operations can be done using gonum/floats package.
	Range() (min, max float64, minIdx, maxIdx int)

	// AggFunc applies given aggregation function to each element in the tensor, using float64
	// conversions of the values.  init is the initial value for the agg variable.  returns final
	// aggregate value
	AggFunc(ini float64, fun func(val float64, agg float64) float64) float64

	// EvalFunc applies given function to each element in the tensor, using float64
	// conversions of the values, and puts the results into given float64 slice, which is
	// ensured to be of the proper length
	EvalFunc(fun func(val float64) float64, res *[]float64)

	// SetFunc applies given function to each element in the tensor, using float64
	// conversions of the values, and writes the results back into the same tensor values
	SetFunc(fun func(val float64) float64)

	// SetZeros is simple convenience function initialize all values to 0
	SetZeros()

	// Clone clones this tensor, creating a duplicate copy of itself with its
	// own separate memory representation of all the values, and returns
	// that as a Tensor (which can be converted into the known type as needed).
	Clone() Tensor

	// CopyFrom copies all avail values from other tensor into this tensor, with an
	// optimized implementation if the other tensor is of the same type, and
	// otherwise it goes through appropriate standard type.
	CopyFrom(from Tensor)

	// CopyCellsFrom copies given range of values from other tensor into this tensor,
	// using flat 1D indexes: to = starting index in this Tensor to start copying into,
	// start = starting index on from Tensor to start copying from, and n = number of
	// values to copy.  Uses an optimized implementation if the other tensor is
	// of the same type, and otherwise it goes through appropriate standard type.
	CopyCellsFrom(from Tensor, to, start, n int)

	// SetShape sets the shape parameters of the tensor, and resizes backing storage appropriately.
	// existing RowMajor or ColMajor stride preference will be used if strides is nil, and
	// existing names will be preserved if nil
	SetShape(shape, strides []int, names []string)

	// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
	// Does nothing for other stride layouts
	AddRows(n int)

	// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
	// Does nothing for other stride layouts
	SetNumRows(rows int)
}

// Check for interface implementation
var _ Tensor = (*Float32)(nil)
