// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"errors"
	"log"

	"github.com/apache/arrow/go/arrow"
	"github.com/emer/etable/bitslice"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
	"gonum.org/v1/gonum/mat"
)

// BoolType not in arrow..

type BoolType struct{}

func (t *BoolType) ID() arrow.Type { return arrow.BOOL }
func (t *BoolType) Name() string   { return "bool" }
func (t *BoolType) BitWidth() int  { return 1 }

// etensor.Bits is a tensor of bits backed by a bitslice.Slice for efficient storage
// of binary data
type Bits struct {
	Shape
	Values bitslice.Slice
	Meta   map[string]string
}

// NewBits returns a new n-dimensional array of bits
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewBits(shape, strides []int, names []string) *Bits {
	bt := &Bits{}
	bt.SetShape(shape, strides, names)
	ln := bt.Len()
	bt.Values = bitslice.Make(ln, 0)
	return bt
}

// NewBitsShape returns a new n-dimensional array of bits
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewBitsShape(shape *Shape) *Bits {
	bt := &Bits{}
	bt.CopyShape(shape)
	ln := bt.Len()
	bt.Values = bitslice.Make(ln, 0)
	return bt
}

func (tsr *Bits) ShapeObj() *Shape { return &tsr.Shape }
func (tsr *Bits) DataType() Type   { return BOOl }

// Value returns value at given tensor index
func (tsr *Bits) Value(i []int) bool { j := int(tsr.Offset(i)); return tsr.Values.Index(j) }

// Value1D returns value at given tensor 1D (flat) index
func (tsr *Bits) Value1D(i int) bool { return tsr.Values.Index(i) }

func (tsr *Bits) Set(i []int, val bool) { j := int(tsr.Offset(i)); tsr.Values.Set(j, val) }
func (tsr *Bits) Set1D(i int, val bool) { tsr.Values.Set(i, val) }

// Null not supported for bits
func (tsr *Bits) IsNull(i []int) bool       { return false }
func (tsr *Bits) IsNull1D(i int) bool       { return false }
func (tsr *Bits) SetNull(i []int, nul bool) {}
func (tsr *Bits) SetNull1D(i int, nul bool) {}

func Float64ToBool(val float64) bool {
	bv := true
	if val == 0 {
		bv = false
	}
	return bv
}

func BoolToFloat64(bv bool) float64 {
	if bv {
		return 1
	} else {
		return 0
	}
}

func (tsr *Bits) FloatVal(i []int) float64 {
	j := tsr.Offset(i)
	return BoolToFloat64(tsr.Values.Index(j))
}
func (tsr *Bits) SetFloat(i []int, val float64) {
	j := tsr.Offset(i)
	tsr.Values.Set(j, Float64ToBool(val))
}

func (tsr *Bits) StringVal(i []int) string {
	j := tsr.Offset(i)
	return kit.ToString(tsr.Values.Index(j))
}

func (tsr *Bits) SetString(i []int, val string) {
	if bv, ok := kit.ToBool(val); ok {
		j := tsr.Offset(i)
		tsr.Values.Set(j, bv)
	}
}

func (tsr *Bits) FloatVal1D(off int) float64 {
	return BoolToFloat64(tsr.Values.Index(off))
}
func (tsr *Bits) SetFloat1D(off int, val float64) {
	tsr.Values.Set(off, Float64ToBool(val))
}

func (tsr *Bits) Floats(flt *[]float64) {
	sz := tsr.Len()
	if len(*flt) < sz {
		if cap(*flt) >= sz {
			*flt = (*flt)[0:sz]
		} else {
			*flt = make([]float64, sz)
		}
	}
	for j := 0; j < sz; j++ {
		(*flt)[j] = BoolToFloat64(tsr.Values.Index(j))
	}
}

// SetFloats sets tensor values from a []float64 slice (copies values).
func (tsr *Bits) SetFloats(vals []float64) {
	sz := ints.MinInt(tsr.Len(), len(vals))
	for j := 0; j < sz; j++ {
		tsr.Values.Set(j, Float64ToBool(vals[j]))
	}
}

func (tsr *Bits) StringVal1D(off int) string {
	return kit.ToString(tsr.Values.Index(off))
}

func (tsr *Bits) SetString1D(off int, val string) {
	if bv, ok := kit.ToBool(val); ok {
		tsr.Values.Set(off, bv)
	}
}

// SubSpace is not applicable to Bits tensor
func (tsr *Bits) SubSpace(subdim int, offs []int) Tensor {
	return nil
}

// SubSpaceTry is not applicable to Bits tensor
func (tsr *Bits) SubSpaceTry(subdim int, offs []int) (Tensor, error) {
	return nil, errors.New("etensor.Bits does not support SubSpace")
}

// Range is not applicable to Bits tensor
func (tsr *Bits) Range() (min, max float64, minIdx, maxIdx int) {
	minIdx = -1
	maxIdx = -1
	return
}

// Agg applies given aggregation function to each element in the tensor
// (automatically skips IsNull and NaN elements), using float64 conversions of the values.
// init is the initial value for the agg variable. returns final aggregate value
func (tsr *Bits) Agg(ini float64, fun AggFunc) float64 {
	ln := tsr.Len()
	ag := ini
	for j := 0; j < ln; j++ {
		ag = fun(j, BoolToFloat64(tsr.Values.Index(j)), ag)
	}
	return ag
}

// Eval applies given function to each element in the tensor, using float64
// conversions of the values, and puts the results into given float64 slice, which is
// ensured to be of the proper length
func (tsr *Bits) Eval(res *[]float64, fun EvalFunc) {
	ln := tsr.Len()
	if len(*res) != ln {
		*res = make([]float64, ln)
	}
	for j := 0; j < ln; j++ {
		(*res)[j] = fun(j, BoolToFloat64(tsr.Values.Index(j)))
	}
}

// SetFunc applies given function to each element in the tensor, using float64
// conversions of the values, and writes the results back into the same tensor values
func (tsr *Bits) SetFunc(fun EvalFunc) {
	ln := tsr.Len()
	for j := 0; j < ln; j++ {
		tsr.Values.Set(j, Float64ToBool(fun(j, BoolToFloat64(tsr.Values.Index(j)))))
	}
}

// SetZeros is simple convenience function initialize all values to 0
func (tsr *Bits) SetZeros() {
	ln := tsr.Len()
	for j := 0; j < ln; j++ {
		tsr.Values.Set(j, false)
	}
}

// Clone clones this tensor, creating a duplicate copy of itself with its
// own separate memory representation of all the values, and returns
// that as a Tensor (which can be converted into the known type as needed).
func (tsr *Bits) Clone() Tensor {
	csr := NewBitsShape(&tsr.Shape)
	csr.Values = tsr.Values.Clone()
	return csr
}

// CopyFrom copies all avail values from other tensor into this tensor, with an
// optimized implementation if the other tensor is of the same type, and
// otherwise it goes through appropriate standard type.
// Copies Null state as well if present.
func (tsr *Bits) CopyFrom(frm Tensor) {
	if fsm, ok := frm.(*Bits); ok {
		copy(tsr.Values, fsm.Values)
		return
	}
	sz := ints.MinInt(len(tsr.Values), frm.Len())
	for i := 0; i < sz; i++ {
		tsr.Values.Set(i, Float64ToBool(frm.FloatVal1D(i)))
	}
}

// CopyShapeFrom copies just the shape from given source tensor
// calling SetShape with the shape params from source (see for more docs).
func (tsr *Bits) CopyShapeFrom(frm Tensor) {
	tsr.SetShape(frm.Shapes(), frm.Strides(), frm.DimNames())
}

// CopyCellsFrom copies given range of values from other tensor into this tensor,
// using flat 1D indexes: to = starting index in this Tensor to start copying into,
// start = starting index on from Tensor to start copying from, and n = number of
// values to copy.  Uses an optimized implementation if the other tensor is
// of the same type, and otherwise it goes through appropriate standard type.
func (tsr *Bits) CopyCellsFrom(frm Tensor, to, start, n int) {
	if fsm, ok := frm.(*Bits); ok {
		for i := 0; i < n; i++ {
			tsr.Values.Set(to+i, fsm.Values.Index(start+i))
		}
		return
	}
	for i := 0; i < n; i++ {
		tsr.Values.Set(to+i, Float64ToBool(frm.FloatVal1D(start+i)))
	}
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Bits) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	tsr.Values.SetLen(nln)
}

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *Bits) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	cln := tsr.Len()
	rows := tsr.Dim(0)
	inln := cln / rows // length of inner dims
	nln := (rows + n) * inln
	tsr.Shape.Shp[0] += n
	tsr.Values.SetLen(nln)
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Bits) SetNumRows(rows int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows = ints.MaxInt(1, rows) // must be > 0
	cln := tsr.Len()
	crows := tsr.Dim(0)
	inln := cln / crows // length of inner dims
	nln := rows * inln
	tsr.Shape.Shp[0] = rows
	tsr.Values.SetLen(nln)
}

// Dims is the gonum/mat.Matrix interface method for returning the dimensionality of the
// 2D Matrix.  Not supported for Bits -- do not call!
func (tsr *Bits) Dims() (r, c int) {
	log.Println("etensor Dims gonum Matrix call made on Bits Tensor -- not supported")
	return 0, 0
}

// At is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Not supported for Bits -- do not call!
func (tsr *Bits) At(i, j int) float64 {
	log.Println("etensor At gonum Matrix call made on Bits Tensor -- not supported")
	return 0
}

// T is the gonum/mat.Matrix transpose method.
// Not supported for Bits -- do not call!
func (tsr *Bits) T() mat.Matrix {
	log.Println("etensor T gonum Matrix call made on Bits Tensor -- not supported")
	return mat.Transpose{tsr}
}

// SetMetaData sets a key=value meta data (stored as a map[string]string).
// For TensorGrid display: top-zero=+/-, odd-row=+/-, image=+/-,
// min, max set fixed min / max values, background=color
func (tsr *Bits) SetMetaData(key, val string) {
	if tsr.Meta == nil {
		tsr.Meta = make(map[string]string)
	}
	tsr.Meta[key] = val
}

// MetaData retrieves value of given key, bool = false if not set
func (tsr *Bits) MetaData(key string) (string, bool) {
	if tsr.Meta == nil {
		return "", false
	}
	val, ok := tsr.Meta[key]
	return val, ok
}

// MetaDataMap returns the underlying map used for meta data
func (tsr *Bits) MetaDataMap() map[string]string {
	return tsr.Meta
}

// CopyMetaData copies meta data from given source tensor
func (tsr *Bits) CopyMetaData(frm Tensor) {
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
