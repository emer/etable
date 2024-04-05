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
	"unsafe"

	"cogentcore.org/core/laser"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/emer/etable/v2/bitslice"
	"gonum.org/v1/gonum/mat"
)

// Int is an n-dim array of ints, which are assumed to be 64bit in various places
// so usage on non-64bit platforms is not recommended.  However, convenience of
// avoiding the constant casting is worth it in some cases.
type Int struct {
	Shape
	Values []int
	Nulls  bitslice.Slice
	Meta   map[string]string
}

// NewInt returns a new n-dimensional array of ints.
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
// Nulls are initialized to nil.
func NewInt(shape, strides []int, names []string) *Int {
	tsr := &Int{}
	tsr.SetShape(shape, strides, names)
	tsr.Values = make([]int, tsr.Len())
	return tsr
}

// NewIntShape returns a new n-dimensional array of ints.
// Using shape structure instead of separate slices, and optionally
// existing values if vals != nil (must be of proper length) -- we
// directly set our internal Values = vals, thereby sharing the same
// underlying data. Nulls are initialized to nil.
func NewIntShape(shape *Shape, vals []int) *Int {
	tsr := &Int{}
	tsr.CopyShape(shape)
	if vals != nil {
		if len(vals) != tsr.Len() {
			log.Printf("etensor.NewIntShape: length of provided vals: %d not proper length: %d", len(vals), tsr.Len())
			tsr.Values = make([]int, tsr.Len())
		} else {
			tsr.Values = vals
		}
	} else {
		tsr.Values = make([]int, tsr.Len())
	}
	return tsr
}

func (tsr *Int) ShapeObj() *Shape     { return &tsr.Shape }
func (tsr *Int) DataType() Type       { return INT }
func (tsr *Int) Value(i []int) int    { j := tsr.Offset(i); return tsr.Values[j] }
func (tsr *Int) Value1D(i int) int    { return tsr.Values[i] }
func (tsr *Int) Set(i []int, val int) { j := tsr.Offset(i); tsr.Values[j] = val }
func (tsr *Int) Set1D(i int, val int) { tsr.Values[i] = val }
func (tsr *Int) AddScalar(i []int, val int) int {
	j := tsr.Offset(i)
	tsr.Values[j] += val
	return tsr.Values[j]
}
func (tsr *Int) MulScalar(i []int, val int) int {
	j := tsr.Offset(i)
	tsr.Values[j] *= val
	return tsr.Values[j]
}

// IsNull returns true if the given index has been flagged as a Null
// (undefined, not present) value
func (tsr *Int) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}

// IsNull1D returns true if the given 1-dimensional index has been flagged as a Null
// (undefined, not present) value
func (tsr *Int) IsNull1D(i int) bool {
	if tsr.Nulls == nil {
		return false
	}
	return tsr.Nulls.Index(i)
}

// SetNull sets whether given index has a null value or not.
// All values are assumed valid (non-Null) until marked otherwise, and calling
// this method creates a Null bitslice map if one has not already been set yet.
func (tsr *Int) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
}

// SetNull1D sets whether given 1-dimensional index has a null value or not.
// All values are assumed valid (non-Null) until marked otherwise, and calling
// this method creates a Null bitslice map if one has not already been set yet.
func (tsr *Int) SetNull1D(i int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	tsr.Nulls.Set(i, nul)
}

func (tsr *Int) FloatValue(i []int) float64    { j := tsr.Offset(i); return float64(tsr.Values[j]) }
func (tsr *Int) SetFloat(i []int, val float64) { j := tsr.Offset(i); tsr.Values[j] = int(val) }

func (tsr *Int) StringValue(i []int) string { j := tsr.Offset(i); return laser.ToString(tsr.Values[j]) }
func (tsr *Int) SetString(i []int, val string) {
	if fv, err := strconv.ParseFloat(val, 64); err == nil {
		j := tsr.Offset(i)
		tsr.Values[j] = int(fv)
	}
}

func (tsr *Int) FloatValue1D(off int) float64    { return float64(tsr.Values[off]) }
func (tsr *Int) SetFloat1D(off int, val float64) { tsr.Values[off] = int(val) }

func (tsr *Int) FloatValueRowCell(row, cell int) float64 {
	_, sz := tsr.RowCellSize()
	return float64(tsr.Values[row*sz+cell])
}
func (tsr *Int) SetFloatRowCell(row, cell int, val float64) {
	_, sz := tsr.RowCellSize()
	tsr.Values[row*sz+cell] = int(val)
}

// Floats sets []float64 slice of all elements in the tensor
// (length is ensured to be sufficient).
// This can be used for all of the gonum/floats methods
// for basic math, gonum/stats, etc.
func (tsr *Int) Floats(flt *[]float64) {
	SetFloat64SliceLen(flt, len(tsr.Values))
	for j, vl := range tsr.Values {
		(*flt)[j] = float64(vl)
	}
}

// SetFloats sets tensor values from a []float64 slice (copies values).
func (tsr *Int) SetFloats(vals []float64) {
	sz := min(len(tsr.Values), len(vals))
	for j := 0; j < sz; j++ {
		tsr.Values[j] = int(vals[j])
	}
}

func (tsr *Int) StringValue1D(off int) string { return laser.ToString(tsr.Values[off]) }
func (tsr *Int) SetString1D(off int, val string) {
	if fv, err := strconv.ParseFloat(val, 64); err == nil {
		tsr.Values[off] = int(fv)
	}
}

func (tsr *Int) StringValueRowCell(row, cell int) string {
	_, sz := tsr.RowCellSize()
	return laser.ToString(tsr.Values[row*sz+cell])
}
func (tsr *Int) SetStringRowCell(row, cell int, val string) {
	if fv, err := strconv.ParseFloat(val, 64); err == nil {
		_, sz := tsr.RowCellSize()
		tsr.Values[row*sz+cell] = int(fv)
	}
}

// Range returns the min, max (and associated indexes, -1 = no values) for the tensor.
// This is needed for display and is thus in the core api in optimized form
// Other math operations can be done using gonum/floats package.
func (tsr *Int) Range() (min, max float64, minIndex, maxIndex int) {
	minIndex = -1
	maxIndex = -1
	for j, vl := range tsr.Values {
		fv := float64(vl)
		if math.IsNaN(fv) {
			continue
		}
		if fv < min || minIndex < 0 {
			min = fv
			minIndex = j
		}
		if fv > max || maxIndex < 0 {
			max = fv
			maxIndex = j
		}
	}
	return
}

// Agg applies given aggregation function to each element in the tensor
// (automatically skips IsNull and NaN elements), using float64 conversions of the values.
// init is the initial value for the agg variable. returns final aggregate value
func (tsr *Int) Agg(ini float64, fun AggFunc) float64 {
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
func (tsr *Int) Eval(res *[]float64, fun EvalFunc) {
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
func (tsr *Int) SetFunc(fun EvalFunc) {
	for j, vl := range tsr.Values {
		val := float64(vl)
		if !tsr.IsNull1D(j) && !math.IsNaN(val) {
			tsr.Values[j] = int(fun(j, val))
		}
	}
}

// SetZeros is simple convenience function initialize all values to 0
func (tsr *Int) SetZeros() {
	for j := range tsr.Values {
		tsr.Values[j] = 0
	}
}

// Clone clones this tensor, creating a duplicate copy of itself with its
// own separate memory representation of all the values, and returns
// that as a Tensor (which can be converted into the known type as needed).
func (tsr *Int) Clone() Tensor {
	csr := NewIntShape(&tsr.Shape, nil)
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
func (tsr *Int) CopyFrom(frm Tensor) {
	if fsm, ok := frm.(*Int); ok {
		copy(tsr.Values, fsm.Values)
		if fsm.Nulls != nil {
			if tsr.Nulls == nil {
				tsr.Nulls = bitslice.Make(tsr.Len(), 0)
			}
			copy(tsr.Nulls, fsm.Nulls)
		}
		return
	}
	sz := min(len(tsr.Values), frm.Len())
	for i := 0; i < sz; i++ {
		tsr.Values[i] = int(frm.FloatValue1D(i))
		if frm.IsNull1D(i) {
			tsr.SetNull1D(i, true)
		}
	}
}

// CopyShapeFrom copies just the shape from given source tensor
// calling SetShape with the shape params from source (see for more docs).
func (tsr *Int) CopyShapeFrom(frm Tensor) {
	tsr.SetShape(frm.Shapes(), frm.Strides(), frm.DimNames())
}

// CopyCellsFrom copies given range of values from other tensor into this tensor,
// using flat 1D indexes: to = starting index in this Tensor to start copying into,
// start = starting index on from Tensor to start copying from, and n = number of
// values to copy.  Uses an optimized implementation if the other tensor is
// of the same type, and otherwise it goes through appropriate standard type.
func (tsr *Int) CopyCellsFrom(frm Tensor, to, start, n int) {
	if fsm, ok := frm.(*Int); ok {
		for i := 0; i < n; i++ {
			tsr.Values[to+i] = fsm.Values[start+i]
			if fsm.IsNull1D(start + i) {
				tsr.SetNull1D(to+i, true)
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		tsr.Values[to+i] = int(frm.FloatValue1D(start + i))
		if frm.IsNull1D(start + i) {
			tsr.SetNull1D(to+i, true)
		}
	}
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Int) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]int, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
	if tsr.Nulls != nil {
		tsr.Nulls.SetLen(nln)
	}
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Int) SetNumRows(rows int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows = max(1, rows) // must be > 0
	_, cells := tsr.RowCellSize()
	nln := rows * cells
	tsr.Shape.Shp[0] = rows
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]int, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
	if tsr.Nulls != nil {
		tsr.Nulls.SetLen(nln)
	}
}

// SubSpace returns a new tensor with innermost subspace at given
// offset(s) in outermost dimension(s) (len(offs) < NumDims).
// Only valid for row or column major layouts.
// The new tensor points to the values of the this tensor (i.e., modifications
// will affect both), as its Values slice is a view onto the original (which
// is why only inner-most contiguous supsaces are supported).
// Use Clone() method to separate the two.
// Null value bits are NOT shared but are copied if present.
func (tsr *Int) SubSpace(offs []int) Tensor {
	ss, _ := tsr.SubSpaceTry(offs)
	return ss
}

// SubSpaceTry returns a new tensor with innermost subspace at given
// offset(s) in outermost dimension(s) (len(offs) < NumDims).
// Try version returns an error message if the offs do not fit in tensor Shape.
// Only valid for row or column major layouts.
// The new tensor points to the values of the this tensor (i.e., modifications
// will affect both), as its Values slice is a view onto the original (which
// is why only inner-most contiguous supsaces are supported).
// Use Clone() method to separate the two.
// Null value bits are NOT shared but are copied if present.
func (tsr *Int) SubSpaceTry(offs []int) (Tensor, error) {
	nd := tsr.NumDims()
	od := len(offs)
	if od >= nd {
		return nil, errors.New("SubSpace len(offsets) for outer dimensions was >= NumDims -- must be less")
	}
	id := nd - od
	if tsr.IsRowMajor() {
		stsr := &Int{}
		stsr.SetShape(tsr.Shp[od:], nil, tsr.Nms[od:]) // row major def
		sti := make([]int, nd)
		copy(sti, offs)
		stoff := tsr.Offset(sti)
		sln := stsr.Len()
		stsr.Values = tsr.Values[stoff : stoff+sln]
		if tsr.Nulls != nil {
			stsr.Nulls = tsr.Nulls.SubSlice(stoff, stoff+sln)
		}
		return stsr, nil
	} else if tsr.IsColMajor() {
		stsr := &Int{}
		stsr.SetShape(tsr.Shp[:id], nil, tsr.Nms[:id])
		stsr.Strd = ColMajorStrides(stsr.Shp)
		sti := make([]int, nd)
		for i := id; i < nd; i++ {
			sti[i] = offs[i-id]
		}
		stoff := tsr.Offset(sti)
		sln := stsr.Len()
		stsr.Values = tsr.Values[stoff : stoff+sln]
		if tsr.Nulls != nil {
			stsr.Nulls = tsr.Nulls.SubSlice(stoff, stoff+sln)
		}
		return stsr, nil
	}
	return nil, errors.New("SubSpace only valid for RowMajor or ColMajor tensors")
}

// Label satisfies the gi.Labeler interface for a summary description of the tensor
func (tsr *Int) Label() string {
	return fmt.Sprintf("Int: %s", tsr.Shape.String())
}

// String satisfies the fmt.Stringer interface for string of tensor data
func (tsr *Int) String() string {
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
		str += tsr.StringValue1D(i) + " "
	}
	str += "\n"
	return str
}

// ToArrow returns the apache arrow equivalent of the tensor
func (tsr *Int) ToArrow() *tensor.Int64 {
	bld := array.NewInt64Builder(memory.DefaultAllocator)
	if tsr.Nulls != nil {
		bld.AppendValues(*(*[]int64)(unsafe.Pointer(&tsr.Values)), tsr.Nulls.ToBools())
	} else {
		bld.AppendValues(*(*[]int64)(unsafe.Pointer(&tsr.Values)), nil)
	}
	vec := bld.NewInt64Array()
	return tensor.NewInt64(vec.Data(), tsr.Shape64(), tsr.Strides64(), tsr.DimNames())
}

// FromArrow intializes this tensor from an arrow tensor of same type
// cpy = true means make a copy of the arrow data, otherwise it directly
// refers to its values slice -- we do not Retain() on that data so it is up
// to the go GC and / or your own memory management policies to ensure the data
// remains intact!
func (tsr *Int) FromArrow(arw *tensor.Int64, cpy bool) {
	nms := make([]string, arw.NumDims()) // note: would be nice if it exposed DimNames()
	for i := range nms {
		nms[i] = arw.DimName(i)
	}
	tsr.SetShape64(arw.Shape(), arw.Strides(), nms)
	vls := arw.Int64Values()
	if cpy {
		tsr.Values = make([]int, tsr.Len())
		copy(*(*[]int64)(unsafe.Pointer(&tsr.Values)), vls)
	} else {
		tsr.Values = *(*[]int)(unsafe.Pointer(&vls))
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
func (tsr *Int) Dims() (r, c int) {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0, 0
	}
	return tsr.Dim(nd - 2), tsr.Dim(nd - 1)
}

// At is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Int) At(i, j int) float64 {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0
	} else if nd == 2 {
		return tsr.FloatValue([]int{i, j})
	} else {
		ix := make([]int, nd)
		ix[nd-2] = i
		ix[nd-1] = j
		return tsr.FloatValue(ix)
	}
}

// T is the gonum/mat.Matrix transpose method.
// It performs an implicit transpose by returning the receiver inside a Transpose.
func (tsr *Int) T() mat.Matrix {
	return mat.Transpose{tsr}
}

// Symmetric is the gonum/mat.Matrix interface method for returning the dimensionality of a symmetric
// 2D Matrix.  Logs error if called on non-symmetric matrix.
func (tsr *Int) Symmetric() (r int) {
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

// SymmetricDim is the gonum/mat.Matrix interface method for returning the dimensionality of a symmetric
// 2D Matrix.  Logs error if called on non-symmetric matrix.
func (tsr *Int) SymmetricDim() (r int) {
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
func (tsr *Int) SetMetaData(key, val string) {
	if tsr.Meta == nil {
		tsr.Meta = make(map[string]string)
	}
	tsr.Meta[key] = val
}

// MetaData retrieves value of given key, bool = false if not set
func (tsr *Int) MetaData(key string) (string, bool) {
	if tsr.Meta == nil {
		return "", false
	}
	val, ok := tsr.Meta[key]
	return val, ok
}

// MetaDataMap returns the underlying map used for meta data
func (tsr *Int) MetaDataMap() map[string]string {
	return tsr.Meta
}

// CopyMetaData copies meta data from given source tensor
func (tsr *Int) CopyMetaData(frm Tensor) {
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
