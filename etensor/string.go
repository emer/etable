// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"errors"
	"log"
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

func (tsr *String) DataType() Type { return STRING }

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

func (tsr *String) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}
func (tsr *String) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
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

func (tsr *String) StringVal1D(off int) string      { return tsr.Values[off] }
func (tsr *String) SetString1D(off int, val string) { tsr.Values[off] = val }

// AggFloat applies given aggregation function to each element in the tensor, using float64
// conversions of the values.  init is the initial value for the agg variable.  returns final
// aggregate value
func (tsr *String) AggFloat(fun func(val float64, agg float64) float64, ini float64) float64 {
	ln := tsr.Len()
	ag := ini
	for j := 0; j < ln; j++ {
		val := StringToFloat64(tsr.Values[j])
		ag = fun(val, ag)
	}
	return ag
}

// EvalFloat applies given function to each element in the tensor, using float64
// conversions of the values, and puts the results into given float64 slice, which is
// ensured to be of the proper length
func (tsr *String) EvalFloat(fun func(val float64) float64, res *[]float64) {
	ln := tsr.Len()
	if len(*res) != ln {
		*res = make([]float64, ln)
	}
	for j := 0; j < ln; j++ {
		val := StringToFloat64(tsr.Values[j])
		(*res)[j] = fun(val)
	}
}

// UpdtFloat applies given function to each element in the tensor, using float64
// conversions of the values, and writes the results back into the same tensor values
func (tsr *String) UpdtFloat(fun func(val float64) float64) {
	ln := tsr.Len()
	for j := 0; j < ln; j++ {
		val := StringToFloat64(tsr.Values[j])
		tsr.Values[j] = Float64ToString(fun(val))
	}
}

// Clone creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *String) Clone() *String {
	csr := NewStringShape(&tsr.Shape)
	copy(csr.Values, tsr.Values)
	if tsr.Nulls != nil {
		csr.Nulls = tsr.Nulls.Clone()
	}
	return csr
}

// CloneTensor creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *String) CloneTensor() Tensor {
	return tsr.Clone()
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

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *String) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	cln := tsr.Len()
	rows := tsr.Dim(0)
	inln := cln / rows // length of inner dims
	nln := (rows + n) * inln
	tsr.Shape.Shp[0] += n
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

// SubSpace returns a new tensor as a subspace of the current one, incorporating the given number
// of dimensions (0 < subdim < NumDims of this tensor).  Only valid for row or column major layouts.
// subdim are the inner, contiguous dimensions (i.e., the final dims in RowMajor and the first ones in ColMajor).
// offs are offsets for the outer dimensions (len = NDims - subdim) for the subspace to return.
// The new tensor points to the values of the this tensor (i.e., modifications will affect both),
// as its Values slice is a view onto the original (which is why only inner-most contiguous supsaces
// are supported).  Use Clone() method to separate the two.
func (tsr *String) SubSpace(subdim int, offs []int) (*String, error) {
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

// At(i, j) is the gonum/mat.Matrix interface method for returning 2D matrix element at given
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
