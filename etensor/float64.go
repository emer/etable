// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/emer/etable/bitslice"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
	"gonum.org/v1/gonum/mat"
)

// Float64 is an n-dim array of float64s.
type Float64 struct {
	Shape
	Values []float64
	Nulls  bitslice.Slice
	Meta   map[string]string
}

// NewFloat64 returns a new n-dimensional array of float64s.
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
// Nulls are initialized to nil.
func NewFloat64(shape, strides []int, names []string) *Float64 {
	tsr := &Float64{}
	tsr.SetShape(shape, strides, names)
	tsr.Values = make([]float64, tsr.Len())
	return tsr
}

// NewFloat64Shape returns a new n-dimensional array of float64s.
// Using shape structure instead of separate slices, and optionally
// existing values if vals != nil (must be of proper length) -- we
// directly set our internal Values = vals, thereby sharing the same
// underlying data. Nulls are initialized to nil.
func NewFloat64Shape(shape *Shape, vals []float64) *Float64 {
	tsr := &Float64{}
	tsr.CopyShape(shape)
	if vals != nil {
		if len(vals) != tsr.Len() {
			log.Printf("etensor.New*Shape: length of provided vals: %d not proper length: %d", len(vals), tsr.Len())
			tsr.Values = make([]float64, tsr.Len())
		} else {
			tsr.Values = vals
		}
	} else {
		tsr.Values = make([]float64, tsr.Len())
	}
	return tsr
}

func (tsr *Float64) ShapeObj() *Shape         { return &tsr.Shape }
func (tsr *Float64) DataType() Type           { return FLOAT64 }
func (tsr *Float64) Value(i []int) float64    { j := tsr.Offset(i); return tsr.Values[j] }
func (tsr *Float64) Value1D(i int) float64    { return tsr.Values[i] }
func (tsr *Float64) Set(i []int, val float64) { j := tsr.Offset(i); tsr.Values[j] = val }

// IsNull returns true if the given index has been flagged as a Null
// (undefined, not present) value
func (tsr *Float64) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}

// IsNull1D returns true if the given 1-dimensional index has been flagged as a Null
// (undefined, not present) value
func (tsr *Float64) IsNull1D(i int) bool {
	if tsr.Nulls == nil {
		return false
	}
	return tsr.Nulls.Index(i)
}

// SetNull sets whether given index has a null value or not.
// All values are assumed valid (non-Null) until marked otherwise, and calling
// this method creates a Null bitslice map if one has not already been set yet.
func (tsr *Float64) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
}

// SetNull1D sets whether given 1-dimensional index has a null value or not.
// All values are assumed valid (non-Null) until marked otherwise, and calling
// this method creates a Null bitslice map if one has not already been set yet.
func (tsr *Float64) SetNull1D(i int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	tsr.Nulls.Set(i, nul)
}

func (tsr *Float64) FloatVal(i []int) float64      { j := tsr.Offset(i); return float64(tsr.Values[j]) }
func (tsr *Float64) SetFloat(i []int, val float64) { j := tsr.Offset(i); tsr.Values[j] = float64(val) }

func (tsr *Float64) StringVal(i []int) string { j := tsr.Offset(i); return kit.ToString(tsr.Values[j]) }
func (tsr *Float64) SetString(i []int, val string) {
	if fv, err := strconv.ParseFloat(val, 64); err == nil {
		j := tsr.Offset(i)
		tsr.Values[j] = float64(fv)
	}
}

func (tsr *Float64) FloatVal1D(off int) float64      { return float64(tsr.Values[off]) }
func (tsr *Float64) SetFloat1D(off int, val float64) { tsr.Values[off] = float64(val) }

// Floats sets []float64 slice of all elements in the tensor
// (length is ensured to be sufficient).
// This can be used for all of the gonum/floats methods
// for basic math, gonum/stats, etc.
func (tsr *Float64) Floats(flt *[]float64) {
	sz := len(tsr.Values)
	if len(*flt) < sz {
		if cap(*flt) >= sz {
			*flt = (*flt)[0:sz]
		} else {
			*flt = make([]float64, sz)
		}
	}
	copy(*flt, tsr.Values) // diff: blit from values directly
}

// SetFloats sets tensor values from a []float64 slice (copies values).
func (tsr *Float64) SetFloats(vals []float64) {
	copy(tsr.Values, vals) // diff: blit from values directly
}

func (tsr *Float64) StringVal1D(off int) string { return kit.ToString(tsr.Values[off]) }
func (tsr *Float64) SetString1D(off int, val string) {
	if fv, err := strconv.ParseFloat(val, 64); err == nil {
		tsr.Values[off] = float64(fv)
	}
}

// Range returns the min, max (and associated indexes, -1 = no values) for the tensor.
// This is needed for display and is thus in the core api in optimized form
// Other math operations can be done using gonum/floats package.
func (tsr *Float64) Range() (min, max float64, minIdx, maxIdx int) {
	minIdx = -1
	maxIdx = -1
	for j, vl := range tsr.Values {
		fv := float64(vl)
		if math.IsNaN(fv) {
			continue
		}
		if fv < min || minIdx < 0 {
			min = fv
			minIdx = j
		}
		if fv > max || maxIdx < 0 {
			max = fv
			maxIdx = j
		}
	}
	return
}

// Agg applies given aggregation function to each element in the tensor
// (automatically skips IsNull and NaN elements), using float64 conversions of the values.
// init is the initial value for the agg variable. returns final aggregate value
func (tsr *Float64) Agg(ini float64, fun AggFunc) float64 {
	ag := ini
	for j, vl := range tsr.Values {
		val := float64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			ag = fun(j, val, ag)
		}
	}
	return ag
}

// Eval applies given function to each element in the tensor (automatically
// skips IsNull and NaN elements), using float64 conversions of the values.
// Puts the results into given float64 slice, which is ensured to be of the proper length.
func (tsr *Float64) Eval(res *[]float64, fun EvalFunc) {
	ln := tsr.Len()
	if len(*res) != ln {
		*res = make([]float64, ln)
	}
	for j, vl := range tsr.Values {
		val := float64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			(*res)[j] = fun(j, val)
		}
	}
}

// SetFunc applies given function to each element in the tensor (automatically
// skips IsNull and NaN elements), using float64 conversions of the values.
// Writes the results back into the same tensor elements.
func (tsr *Float64) SetFunc(fun EvalFunc) {
	for j, vl := range tsr.Values {
		val := float64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			tsr.Values[j] = float64(fun(j, val))
		}
	}
}

// SetZeros is simple convenience function initialize all values to 0
func (tsr *Float64) SetZeros() {
	for j := range tsr.Values {
		tsr.Values[j] = 0
	}
}

// Clone clones this tensor, creating a duplicate copy of itself with its
// own separate memory representation of all the values, and returns
// that as a Tensor (which can be converted into the known type as needed).
func (tsr *Float64) Clone() Tensor {
	csr := NewFloat64Shape(&tsr.Shape, nil)
	copy(csr.Values, tsr.Values)
	if tsr.Nulls != nil {
		csr.Nulls = tsr.Nulls.Clone()
	}
	return csr
}

// CopyFrom copies all avail values from other tensor into this tensor, with an
// optimized implementation if the other tensor is of the same type, and
// otherwise it goes through appropriate standard type.
// Copies Null state as well if present.
func (tsr *Float64) CopyFrom(frm Tensor) {
	if fsm, ok := frm.(*Float64); ok {
		copy(tsr.Values, fsm.Values)
		if fsm.Nulls != nil {
			if tsr.Nulls == nil {
				tsr.Nulls = bitslice.Make(tsr.Len(), 0)
			}
			copy(tsr.Nulls, fsm.Nulls)
		}
		return
	}
	sz := ints.MinInt(len(tsr.Values), frm.Len())
	for i := 0; i < sz; i++ {
		tsr.Values[i] = float64(frm.FloatVal1D(i))
		if frm.IsNull1D(i) {
			tsr.SetNull1D(i, true)
		}
	}
}

// CopyCellsFrom copies given range of values from other tensor into this tensor,
// using flat 1D indexes: to = starting index in this Tensor to start copying into,
// start = starting index on from Tensor to start copying from, and n = number of
// values to copy.  Uses an optimized implementation if the other tensor is
// of the same type, and otherwise it goes through appropriate standard type.
func (tsr *Float64) CopyCellsFrom(frm Tensor, to, start, n int) {
	if fsm, ok := frm.(*Float64); ok {
		for i := 0; i < n; i++ {
			tsr.Values[to+i] = fsm.Values[start+i]
			if fsm.IsNull1D(start + i) {
				tsr.SetNull1D(to+i, true)
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		tsr.Values[to+i] = float64(frm.FloatVal1D(start + i))
		if frm.IsNull1D(start + i) {
			tsr.SetNull1D(to+i, true)
		}
	}
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Float64) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]float64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *Float64) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows, cells := tsr.RowCellSize()
	nln := (rows + n) * cells
	tsr.Shape.Shp[0] += n
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]float64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Float64) SetNumRows(rows int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows = ints.MaxInt(1, rows) // must be > 0
	_, cells := tsr.RowCellSize()
	nln := rows * cells
	tsr.Shape.Shp[0] = rows
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]float64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

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
func (tsr *Float64) SubSpace(subdim int, offs []int) Tensor {
	ss, _ := tsr.SubSpaceTry(subdim, offs)
	return ss
}

// SubSpaceTry is SubSpace but returns an error message if the subdim and offs
// do not match the tensor Shape.
func (tsr *Float64) SubSpaceTry(subdim int, offs []int) (Tensor, error) {
	nd := tsr.NumDims()
	od := nd - subdim
	if od <= 0 {
		return nil, errors.New("SubSpace number of sub dimensions was >= NumDims -- must be less")
	}
	if tsr.IsRowMajor() {
		stsr := &Float64{}
		stsr.SetShape(tsr.Shp[od:], nil, tsr.Nms[od:]) // row major def
		sti := make([]int, nd)
		copy(sti, offs)
		stoff := tsr.Offset(sti)
		sln := stsr.Len()
		stsr.Values = tsr.Values[stoff : stoff+sln]
		return stsr, nil
	} else if tsr.IsColMajor() {
		stsr := &Float64{}
		stsr.SetShape(tsr.Shp[:subdim], nil, tsr.Nms[:subdim])
		stsr.Strd = ColMajorStrides(stsr.Shp)
		sti := make([]int, nd)
		for i := subdim; i < nd; i++ {
			sti[i] = offs[i-subdim]
		}
		stoff := tsr.Offset(sti)
		sln := stsr.Len()
		stsr.Values = tsr.Values[stoff : stoff+sln]
		return stsr, nil
	}
	return nil, errors.New("SubSpace only valid for RowMajor or ColMajor tensors")
}

// Label satisfies the gi.Labeler interface for a summary description of the tensor
func (tsr *Float64) Label() string {
	return fmt.Sprintf("Float64: %s", tsr.Shape.String())
}

// String satisfies the fmt.Stringer interface for string of tensor data
func (tsr *Float64) String() string {
	str := tsr.Label()
	sz := len(tsr.Values)
	if sz > 1000 {
		return str
	}
	for i := 0; i < sz; i++ {
		idx := tsr.Index(i)
		for j := 1; j < len(idx); j++ {
			if idx[j] == 0 {
				str += "\n["
				for k := 0; k < len(idx); k++ {
					str += fmt.Sprintf("%d", idx[k])
					if k < len(idx)-1 {
						str += ","
					}
				}
				str += "]: "
				break
			}
		}
		str += tsr.StringVal1D(i) + " "
	}
	str += "\n"
	return str
}

// ToArrow returns the apache arrow equivalent of the tensor
func (tsr *Float64) ToArrow() *tensor.Float64 {
	bld := array.NewFloat64Builder(memory.DefaultAllocator)
	if tsr.Nulls != nil {
		bld.AppendValues(tsr.Values, tsr.Nulls.ToBools())
	} else {
		bld.AppendValues(tsr.Values, nil)
	}
	vec := bld.NewFloat64Array()
	return tensor.NewFloat64(vec.Data(), tsr.Shape64(), tsr.Strides64(), tsr.DimNames())
}

// FromArrow intializes this tensor from an arrow tensor of same type
// cpy = true means make a copy of the arrow data, otherwise it directly
// refers to its values slice -- we do not Retain() on that data so it is up
// to the go GC and / or your own memory management policies to ensure the data
// remains intact!
func (tsr *Float64) FromArrow(arw *tensor.Float64, cpy bool) {
	nms := make([]string, arw.NumDims()) // note: would be nice if it exposed DimNames()
	for i := range nms {
		nms[i] = arw.DimName(i)
	}
	tsr.SetShape64(arw.Shape(), arw.Strides(), nms)
	if cpy {
		vls := arw.Float64Values()
		tsr.Values = make([]float64, tsr.Len())
		copy(tsr.Values, vls)
	} else {
		tsr.Values = arw.Float64Values()
	}
	// note: doesn't look like the Data() exposes the nulls themselves so it is not
	// clear we can copy the null values -- nor does it seem that the tensor class
	// exposes it either!  https://github.com/apache/arrow/issues/3496
	// nln := arw.Data().NullN()
	// if nln > 0 {
	// }
}

// Dims is the gonum/mat.Matrix interface method for returning the dimensionality of the
// 2D Matrix.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Float64) Dims() (r, c int) {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0, 0
	}
	return tsr.Dim(nd - 2), tsr.Dim(nd - 1)
}

// At is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Float64) At(i, j int) float64 {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0
	} else if nd == 2 {
		return tsr.FloatVal([]int{i, j})
	} else {
		ix := make([]int, nd)
		ix[nd-2] = i
		ix[nd-1] = j
		return tsr.FloatVal(ix)
	}
}

// T is the gonum/mat.Matrix transpose method.
// It performs an implicit transpose by returning the receiver inside a Transpose.
func (tsr *Float64) T() mat.Matrix {
	return mat.Transpose{tsr}
}

// Symmetric is the gonum/mat.Matrix interface method for returning the dimensionality of a symmetric
// 2D Matrix.
func (tsr *Float64) Symmetric() (r int) {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Symmetric gonum Matrix call made on Tensor with dims < 2")
		return 0
	}
	if tsr.Dim(nd-2) != tsr.Dim(nd-1) {
		log.Println("etensor Symmatrics gonum Matrix call made on Tensor that is not symmetric")
		return 0
	}
	return tsr.Dim(nd - 1)
}

// SetMetaData sets a key=value meta data (stored as a map[string]string).
// For TensorGrid display: top-zero=+/-, odd-row=+/-, image=+/-,
// min, max set fixed min / max values, background=color
func (tsr *Float64) SetMetaData(key, val string) {
	if tsr.Meta == nil {
		tsr.Meta = make(map[string]string)
	}
	tsr.Meta[key] = val
}

// MetaData retrieves value of given key, bool = false if not set
func (tsr *Float64) MetaData(key string) (string, bool) {
	if tsr.Meta == nil {
		return "", false
	}
	val, ok := tsr.Meta[key]
	return val, ok
}
