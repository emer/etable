// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/emer/etable/bitslice"
	"github.com/goki/ki/ints"
	"gonum.org/v1/gonum/mat"
)

// etensor.String is a tensor of strings backed by a []string slice
type String struct {
	Shape
	Values []string
	Nulls  bitslice.Slice
	Meta   map[string]string
}

// NewString returns a new n-dimensional array of strings
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewString(shape, strides []int, names []string) *String {
	bt := &String{}
	bt.SetShape(shape, strides, names)
	ln := bt.Len()
	bt.Values = make([]string, ln)
	return bt
}

// NewStringShape returns a new n-dimensional array of strings from given shape
func NewStringShape(shape *Shape) *String {
	bt := &String{}
	bt.CopyShape(shape)
	ln := bt.Len()
	bt.Values = make([]string, ln)
	return bt
}

func (tsr *String) ShapeObj() *Shape { return &tsr.Shape }
func (tsr *String) DataType() Type   { return STRING }

// Value returns value at given tensor index
func (tsr *String) Value(i []int) string {
	j := int(tsr.Offset(i))
	return tsr.Values[j]
}

// Value1D returns value at given 1D (flat) tensor index
func (tsr *String) Value1D(i int) string {
	return tsr.Values[i]
}

// Set sets value at given tensor index
func (tsr *String) Set(i []int, val string) {
	j := int(tsr.Offset(i))
	tsr.Values[j] = val
}

// Set1D sets value at given 1D (flat) tensor index
func (tsr *String) Set1D(i int, val string) {
	tsr.Values[i] = val
}

func (tsr *String) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}

func (tsr *String) IsNull1D(i int) bool {
	if tsr.Nulls == nil {
		return false
	}
	return tsr.Nulls.Index(i)
}

func (tsr *String) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
}

func (tsr *String) SetNull1D(i int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	tsr.Nulls.Set(i, nul)
}

func StringToFloat64(str string) float64 {
	if fv, err := strconv.ParseFloat(str, 64); err == nil {
		return fv
	}
	return 0
}

func Float64ToString(val float64) string {
	return strconv.FormatFloat(val, 'g', -1, 64)
}

func (tsr *String) FloatVal(i []int) float64 {
	j := tsr.Offset(i)
	return StringToFloat64(tsr.Values[j])
}

func (tsr *String) SetFloat(i []int, val float64) {
	j := tsr.Offset(i)
	tsr.Values[j] = Float64ToString(val)
}

func (tsr *String) StringVal(i []int) string      { j := tsr.Offset(i); return tsr.Values[j] }
func (tsr *String) SetString(i []int, val string) { j := tsr.Offset(i); tsr.Values[j] = val }

func (tsr *String) FloatVal1D(off int) float64 {
	return StringToFloat64(tsr.Values[off])
}

func (tsr *String) SetFloat1D(off int, val float64) {
	tsr.Values[off] = Float64ToString(val)
}

func (tsr *String) Floats(flt *[]float64) {
	sz := len(tsr.Values)
	if len(*flt) < sz {
		if cap(*flt) >= sz {
			*flt = (*flt)[0:sz]
		} else {
			*flt = make([]float64, sz)
		}
	}
	for j, vl := range tsr.Values {
		(*flt)[j] = StringToFloat64(vl)
	}
}

// SetFloats sets tensor values from a []float64 slice (copies values).
func (tsr *String) SetFloats(vals []float64) {
	sz := ints.MinInt(len(tsr.Values), len(vals))
	for j := 0; j < sz; j++ {
		tsr.Values[j] = Float64ToString(vals[j])
	}
}

func (tsr *String) StringVal1D(off int) string      { return tsr.Values[off] }
func (tsr *String) SetString1D(off int, val string) { tsr.Values[off] = val }

// Range is not applicable to String tensor
func (tsr *String) Range() (min, max float64, minIdx, maxIdx int) {
	minIdx = -1
	maxIdx = -1
	return
}

// Agg applies given aggregation function to each element in the tensor
// (automatically skips IsNull and NaN elements), using float64 conversions of the values.
// init is the initial value for the agg variable. returns final aggregate value
func (tsr *String) Agg(ini float64, fun AggFunc) float64 {
	ag := ini
	for j, vl := range tsr.Values {
		val := StringToFloat64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			ag = fun(j, val, ag)
		}
	}
	return ag
}

// Eval applies given function to each element in the tensor (automatically
// skips IsNull and NaN elements), using float64 conversions of the values.
// Puts the results into given float64 slice, which is ensured to be of the proper length.
func (tsr *String) Eval(res *[]float64, fun EvalFunc) {
	ln := tsr.Len()
	if len(*res) != ln {
		*res = make([]float64, ln)
	}
	for j, vl := range tsr.Values {
		val := StringToFloat64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			(*res)[j] = fun(j, val)
		}
	}
}

// SetFunc applies given function to each element in the tensor (automatically
// skips IsNull and NaN elements), using float64 conversions of the values.
// Writes the results back into the same tensor elements.
func (tsr *String) SetFunc(fun EvalFunc) {
	for j, vl := range tsr.Values {
		val := StringToFloat64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			tsr.Values[j] = Float64ToString(fun(j, val))
		}
	}
}

// SetZeros is simple convenience function initialize all values to ""
func (tsr *String) SetZeros() {
	ln := tsr.Len()
	for j := 0; j < ln; j++ {
		tsr.Values[j] = ""
	}
}

// Clone clones this tensor, creating a duplicate copy of itself with its
// own separate memory representation of all the values, and returns
// that as a Tensor (which can be converted into the known type as needed).
func (tsr *String) Clone() Tensor {
	csr := NewStringShape(&tsr.Shape)
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
func (tsr *String) CopyFrom(frm Tensor) {
	if fsm, ok := frm.(*String); ok {
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
		tsr.Values[i] = frm.StringVal1D(i)
		if frm.IsNull1D(i) {
			tsr.SetNull1D(i, true)
		}
	}
}

// CopyShapeFrom copies just the shape from given source tensor
// calling SetShape with the shape params from source (see for more docs).
func (tsr *String) CopyShapeFrom(frm Tensor) {
	tsr.SetShape(frm.Shapes(), frm.Strides(), frm.DimNames())
}

// CopyCellsFrom copies given range of values from other tensor into this tensor,
// using flat 1D indexes: to = starting index in this Tensor to start copying into,
// start = starting index on from Tensor to start copying from, and n = number of
// values to copy.  Uses an optimized implementation if the other tensor is
// of the same type, and otherwise it goes through appropriate standard type.
func (tsr *String) CopyCellsFrom(frm Tensor, to, start, n int) {
	if fsm, ok := frm.(*String); ok {
		for i := 0; i < n; i++ {
			tsr.Values[to+i] = fsm.Values[start+i]
			if fsm.IsNull1D(start + i) {
				tsr.SetNull1D(to+i, true)
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		tsr.Values[to+i] = frm.StringVal1D(start + i)
		if frm.IsNull1D(start + i) {
			tsr.SetNull1D(to+i, true)
		}
	}
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *String) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]string, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *String) SetNumRows(rows int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows = ints.MaxInt(1, rows) // must be > 0
	cln := tsr.Len()
	crows := tsr.Dim(0)
	inln := cln / crows // length of inner dims
	nln := rows * inln
	tsr.Shape.Shp[0] = rows
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]string, nln)
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
func (tsr *String) SubSpace(subdim int, offs []int) Tensor {
	ss, _ := tsr.SubSpaceTry(subdim, offs)
	return ss
}

// SubSpaceTry is SubSpace but returns an error message if the subdim and offs
// do not match the tensor Shape.
func (tsr *String) SubSpaceTry(subdim int, offs []int) (Tensor, error) {
	nd := tsr.NumDims()
	od := nd - subdim
	if od <= 0 {
		return nil, errors.New("SubSpace number of sub dimensions was >= NumDims -- must be less")
	}
	if tsr.IsRowMajor() {
		stsr := &String{}
		stsr.SetShape(tsr.Shp[od:], nil, tsr.Nms[od:]) // row major def
		sti := make([]int, nd)
		copy(sti, offs)
		stoff := tsr.Offset(sti)
		stsr.Values = tsr.Values[stoff:]
		return stsr, nil
	} else if tsr.IsColMajor() {
		stsr := &String{}
		stsr.SetShape(tsr.Shp[:subdim], nil, tsr.Nms[:subdim])
		stsr.Strd = ColMajorStrides(stsr.Shp)
		sti := make([]int, nd)
		for i := subdim; i < nd; i++ {
			sti[i] = offs[i-subdim]
		}
		stoff := tsr.Offset(sti)
		stsr.Values = tsr.Values[stoff:]
		return stsr, nil
	}
	return nil, errors.New("SubSpace only valid for RowMajor or ColMajor tensors")
}

// Dims is the gonum/mat.Matrix interface method for returning the dimensionality of the
// 2D Matrix.  Not supported for String -- do not call!
func (tsr *String) Dims() (r, c int) {
	log.Println("etensor Dims gonum Matrix call made on String Tensor -- not supported")
	return 0, 0
}

// At is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Not supported for String -- do not call!
func (tsr *String) At(i, j int) float64 {
	log.Println("etensor At gonum Matrix call made on String Tensor -- not supported")
	return 0
}

// T is the gonum/mat.Matrix transpose method.
// Not supported for String -- do not call!
func (tsr *String) T() mat.Matrix {
	log.Println("etensor T gonum Matrix call made on String Tensor -- not supported")
	return mat.Transpose{tsr}
}

// SetMetaData sets a key=value meta data (stored as a map[string]string).
// For TensorGrid display: top-zero=+/-, odd-row=+/-, image=+/-,
// min, max set fixed min / max values, background=color
func (tsr *String) SetMetaData(key, val string) {
	if tsr.Meta == nil {
		tsr.Meta = make(map[string]string)
	}
	tsr.Meta[key] = val
}

// MetaData retrieves value of given key, bool = false if not set
func (tsr *String) MetaData(key string) (string, bool) {
	if tsr.Meta == nil {
		return "", false
	}
	val, ok := tsr.Meta[key]
	return val, ok
}

// MetaDataMap returns the underlying map used for meta data
func (tsr *String) MetaDataMap() map[string]string {
	return tsr.Meta
}

// CopyMetaData copies meta data from given source tensor
func (tsr *String) CopyMetaData(frm Tensor) {
	fmap := frm.MetaDataMap()
	if len(fmap) == 0 {
		return
	}
	if tsr.Meta == nil {
		tsr.Meta = make(map[string]string)
	}
	for k, v := range fmap {
		tsr.Meta[k] = v
	}
}
