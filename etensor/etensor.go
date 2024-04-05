// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"gonum.org/v1/gonum/mat"
)

//go:generate tmpl -i -data=numeric.tmpldata numeric.gen.go.tmpl

// AggFunc is an aggregation function that incrementally updates agg value
// from each element in the tensor in turn -- returns new agg value that
// will be passed into next item as agg
type AggFunc func(idx int, val float64, agg float64) float64

// EvalFunc is an evaluation function that computes a function on each
// element value, returning the computed value
type EvalFunc func(idx int, val float64) float64

// Tensor is the general interface for n-dimensional tensors.
//
// Tensor is automatically a gonum/mat.Matrix, implementing the Dims(), At(), T(), and Symmetric() methods
// which automatically operate on the inner-most two dimensions, assuming default row-major
// layout. Error messages will be logged if applied to a Tensor with less than 2 dimensions,
// and care should be taken when using with > 2 dimensions (e.g., will only affect the first
// 2D subspace within a higher-dimensional space -- typically you'll want to call SubSpace
// to get a 2D subspace of the higher-dimensional Tensor (SubSpace is not part of interface
// as it returns the specific type, but is defined for all Tensor types).
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
	FloatValue(i []int) float64

	// SetFloat sets the value of given index as a float64
	SetFloat(i []int, val float64)

	// StringVal returns the value of given index as a string
	StringValue(i []int) string

	// SetString sets the value of given index as a string
	SetString(i []int, val string)

	// FloatValue1D returns the value of given 1-dimensional index (0-Len()-1) as a float64
	FloatValue1D(i int) float64

	// SetFloat1D sets the value of given 1-dimensional index (0-Len()-1) as a float64
	SetFloat1D(i int, val float64)

	// FloatValueRowCell returns the value at given row and cell, where row is outer-most dim,
	// and cell is 1D index into remaining inner dims -- for etable.Table columns
	FloatValueRowCell(row, cell int) float64

	// SetFloatRowCell sets the value at given row and cell, where row is outer-most dim,
	// and cell is 1D index into remaining inner dims -- for etable.Table columns
	SetFloatRowCell(row, cell int, val float64)

	// Floats sets []float64 slice of all elements in the tensor
	// (length is ensured to be sufficient).
	// This can be used for all of the gonum/floats methods
	// for basic math, gonum/stats, etc.
	Floats(flt *[]float64)

	// SetFloats sets tensor values from a []float64 slice (copies values).
	SetFloats(vals []float64)

	// StringValue1D returns the value of given 1-dimensional index (0-Len()-1) as a string
	StringValue1D(i int) string

	// SetString1D sets the value of given 1-dimensional index (0-Len()-1) as a string
	SetString1D(i int, val string)

	// StringValueRowCell returns the value at given row and cell, where row is outer-most dim,
	// and cell is 1D index into remaining inner dims -- for etable.Table columns
	StringValueRowCell(row, cell int) string

	// SetStringRowCell sets the value at given row and cell, where row is outer-most dim,
	// and cell is 1D index into remaining inner dims -- for etable.Table columns
	SetStringRowCell(row, cell int, val string)

	// SubSpace returns a new tensor with innermost subspace at given
	// offset(s) in outermost dimension(s) (len(offs) < NumDims).
	// Only valid for row or column major layouts.
	// The new tensor points to the values of the this tensor (i.e., modifications
	// will affect both), as its Values slice is a view onto the original (which
	// is why only inner-most contiguous supsaces are supported).
	// Use Clone() method to separate the two.
	// Null value bits are NOT shared but are copied if present.
	SubSpace(offs []int) Tensor

	// SubSpaceTry returns a new tensor with innermost subspace at given
	// offset(s) in outermost dimension(s) (len(offs) < NumDims).
	// Try version returns an error message if the offs do not fit in tensor Shape.
	// Only valid for row or column major layouts.
	// The new tensor points to the values of the this tensor (i.e., modifications
	// will affect both), as its Values slice is a view onto the original (which
	// is why only inner-most contiguous supsaces are supported).
	// Use Clone() method to separate the two.
	// Null value bits are NOT shared but are copied if present.
	SubSpaceTry(offs []int) (Tensor, error)

	// Range returns the min, max (and associated indexes, -1 = no values) for the tensor.
	// This is needed for display and is thus in the core api in optimized form
	// Other math operations can be done using gonum/floats package.
	Range() (min, max float64, minIndex, maxIndex int)

	// Agg applies given aggregation function to each element in the tensor
	// (automatically skips IsNull and NaN elements), using float64 conversions of the values.
	// init is the initial value for the agg variable. returns final aggregate value
	Agg(ini float64, fun AggFunc) float64

	// Eval applies given function to each element in the tensor (automatically
	// skips IsNull and NaN elements), using float64 conversions of the values.
	// Puts the results into given float64 slice, which is ensured to be of the proper length.
	Eval(res *[]float64, fun EvalFunc)

	// SetFunc applies given function to each element in the tensor (automatically
	// skips IsNull and NaN elements), using float64 conversions of the values.
	// Writes the results back into the same tensor elements.
	SetFunc(fun EvalFunc)

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

	// CopyShapeFrom copies just the shape from given source tensor
	// calling SetShape with the shape params from source (see for more docs).
	CopyShapeFrom(from Tensor)

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

	// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
	// Does nothing for other stride layouts
	SetNumRows(rows int)

	// SetMetaData sets a key=value meta data (stored as a map[string]string).
	// For TensorGrid display: top-zero=+/-, odd-row=+/-, image=+/-,
	// min, max set fixed min / max values, background=color
	SetMetaData(key, val string)

	// MetaData retrieves value of given key, bool = false if not set
	MetaData(key string) (string, bool)

	// MetaDataMap returns the underlying map used for meta data
	MetaDataMap() map[string]string

	// CopyMetaData copies meta data from given source tensor
	CopyMetaData(from Tensor)
}

// Check for interface implementation
var _ Tensor = (*Float32)(nil)

// CopyDense copies a gonum mat.Dense matrix into given Tensor
// using standard Float64 interface
func CopyDense(to Tensor, dm *mat.Dense) {
	nr, nc := dm.Dims()
	to.SetShape([]int{nr, nc}, nil, nil)
	idx := 0
	for ri := 0; ri < nr; ri++ {
		for ci := 0; ci < nc; ci++ {
			v := dm.At(ri, ci)
			to.SetFloat1D(idx, v)
			idx++
		}
	}
}

// SetFloat64SliceLen is a utility function to set given slice of float64 values
// to given length, reusing existing where possible and making a new one as needed.
// For use in WriteGeom routines.
func SetFloat64SliceLen(dat *[]float64, sz int) {
	switch {
	case len(*dat) == sz:
	case len(*dat) < sz:
		if cap(*dat) >= sz {
			*dat = (*dat)[0:sz]
		} else {
			*dat = make([]float64, sz)
		}
	default:
		*dat = (*dat)[0:sz]
	}
}
