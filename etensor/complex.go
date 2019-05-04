// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/emer/etable/bitslice"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
)

// etensor.Complex64 is a tensor of complex64 backed by a []complex64 slice
type Complex64 struct {
	Shape
	Values []complex64
	Nulls  bitslice.Slice
}

// NewComplex64 returns a new n-dimensional array of Complex64
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty Complex64 will be created.
func NewComplex64(shape, strides []int, names []string) *Complex64 {
	bt := &Complex64{}
	bt.SetShape(shape, strides, names)
	ln := bt.Len()
	bt.Values = make([]complex64, ln)
	return bt
}

// NewComplex64Shape returns a new n-dimensional array of Complex64 from given shape
func NewComplex64Shape(shape *Shape) *Complex64 {
	bt := &Complex64{}
	bt.CopyShape(shape)
	ln := bt.Len()
	bt.Values = make([]complex64, ln)
	return bt
}

func (tsr *Complex64) DataType() Type { return COMPLEX64 }

// Value returns value at given tensor index
func (tsr *Complex64) Value(i []int) complex64 {
	j := tsr.Offset(i)
	return tsr.Values[j]
}

// Value1D returns value at given 1D (flat) tensor index
func (tsr *Complex64) Value1D(i int) complex64 {
	return tsr.Values[i]
}

// Set sets value at given tensor index
func (tsr *Complex64) Set(i []int, val complex64) {
	j := int(tsr.Offset(i))
	tsr.Values[j] = val
}

func (tsr *Complex64) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}
func (tsr *Complex64) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
}

// FloatVal returns the real part of a Complex64 - see FloatValImag for getting imaginary part
func (tsr *Complex64) FloatVal(i []int) float64 {
	j := tsr.Offset(i)
	return float64(real(tsr.Values[j]))
}

// FloatValImag returns the imag portion of a Complex64 - see FloatVal for getting real part
func (tsr *Complex64) FloatValImag(i []int) float64 {
	j := tsr.Offset(i)
	return float64(imag(tsr.Values[j]))
}

// SetFloat replaces the real part of the complex number specified by the slice of matrix indices i
func (tsr *Complex64) SetFloat(i []int, val float64) {
	j := tsr.Offset(i)
	valr := float32(val)
	vali := imag(tsr.Values[j])
	tsr.Values[j] = complex(valr, vali)
}

// SetFloat replaces the imaginary part of the complex number specified by the slice of matrix indices i
func (tsr *Complex64) SetFloatImag(i []int, val float64) {
	j := tsr.Offset(i)
	valr := real(tsr.Values[j])
	vali := float32(val)
	tsr.Values[j] = complex(valr, vali)
}

// StringVal returns the numeric value as a string
func (tsr *Complex64) StringVal(i []int) string {
	j := tsr.Offset(i)
	rstr := kit.ToString(real(tsr.Values[j]))
	istr := kit.ToString(imag(tsr.Values[j]))
	return rstr + " + " + istr
}

// SetString sets the numeric value from a string represenation
func (tsr *Complex64) SetString(i []int, val string) {
	ss := strings.SplitN(val, "+", -1)
	rstr := ss[0]
	fvr, err := strconv.ParseFloat(rstr, 32)
	if err != nil {
		return
	}
	istr := ss[1]
	fvi, err := strconv.ParseFloat(istr, 32)
	if err == nil {
		j := tsr.Offset(i)
		tsr.Values[j] = complex(float32(fvr), float32(fvi))
	}
}

func (tsr *Complex64) FloatVal1D(off int) float64 {
	return float64(real(tsr.Values[off]))
}

// SetFloat replaces the imaginary part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex64) FloatVal1DImag(off int) float64 {
	return float64(imag(tsr.Values[off]))
}

// SetFloat replaces the real part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex64) SetFloat1D(off int, val float64) {
	valr := float32(val)
	vali := imag(tsr.Values[off])
	tsr.Values[off] = complex(valr, vali)
}

// SetFloat replaces the imaginary part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex64) SetFloat1DImag(off int, val float64) {
	valr := real(tsr.Values[off])
	vali := float32(val)
	tsr.Values[off] = complex(valr, vali)
}

// todo:
// AggFunc applies given aggregation function to each element in the tensor, using float64
// conversions of the values.  init is the initial value for the agg variable.  returns final
// aggregate value
func (tsr *Complex64) AggFunc(fun func(val float64, agg float64) float64, ini float64) float64 {
	//ln := tsr.Len()
	//ag := ini
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	ag = fun(val, ag)
	//}
	//return ag
	return float64(0.0)
}

// todo:
// EvalFunc applies given function to each element in the tensor, using float64
// conversions of the values, and puts the results into given float64 slice, which is
// ensured to be of the proper length
func (tsr *Complex64) EvalFunc(fun func(val float64) float64, res *[]float64) {
	//ln := tsr.Len()
	//if len(*res) != ln {
	//	*res = make([]float64, ln)
	//}
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	(*res)[j] = fun(val)
	//}
}

// todo:
// SetFunc applies given function to each element in the tensor, using float64
// conversions of the values, and writes the results back into the same tensor values
func (tsr *Complex64) SetFunc(fun func(val float64) float64) {
	//ln := tsr.Len()
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	tsr.Values[j] = complex64(fun(val))
	//}
}

// Clone creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Complex64) Clone() *Complex64 {
	csr := NewComplex64Shape(&tsr.Shape)
	copy(csr.Values, tsr.Values)
	if tsr.Nulls != nil {
		csr.Nulls = tsr.Nulls.Clone()
	}
	return csr
}

// CloneTensor creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Complex64) CloneTensor() Tensor {
	return nil
	//return tsr.Clone()
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Complex64) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]complex64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *Complex64) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows, cells := tsr.RowCellSize()
	nln := (rows + n) * cells
	tsr.Shape.Shp[0] += n
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]complex64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Complex64) SetNumRows(rows int) {
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
		nv := make([]complex64, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SubSlice returns a new tensor as a sub-slice of the current one, incorporating the given number
// of dimensions (0 < subdim < NumDims of this tensor).  Only valid for row or column major layouts.
// subdim are the inner, contiguous dimensions (i.e., the final dims in RowMajor and the first ones in ColMajor).
// offs are offsets for the outer dimensions (len = NDims - subdim) for the subslice to return.
// The new tensor points to the values of the this tensor (i.e., modifications will affect both).
// Use Clone() method to separate the two.
// todo: not getting nulls yet.
func (tsr *Complex64) SubSlice(subdim int, offs []int) (*Complex64, error) {
	nd := tsr.NumDims()
	od := nd - subdim
	if od <= 0 {
		return nil, errors.New("SubSlice number of sub dimensions was >= NumDims -- must be less")
	}
	if tsr.IsRowMajor() {
		stsr := &Complex64{}
		stsr.SetShape(tsr.Shp[od:], nil, tsr.Nms[od:]) // row major def
		sti := make([]int, nd)
		copy(sti, offs)
		stoff := tsr.Offset(sti)
		stsr.Values = tsr.Values[stoff:]
		return stsr, nil
	} else if tsr.IsColMajor() {
		stsr := &Complex64{}
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
	return nil, errors.New("SubSlice only valid for RowMajor or ColMajor tensors")
}

// At(i, j) is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Complex64) At(i, j int) complex64 {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0
	} else if nd == 2 {
		return tsr.Value([]int{i, j})
	} else {
		ix := make([]int, nd)
		ix[nd-2] = i
		ix[nd-1] = j
		return tsr.Value(ix)
	}
}

// Dims is the gonum/mat.Matrix interface method for returning the dimensionality of the
// 2D Matrix.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Complex64) Dims() (r, c int) {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0, 0
	}
	return tsr.Dim(nd - 2), tsr.Dim(nd - 1)
}

// Todo: this needs to be implemented - but not sure how now
// H is the gonum/mat.Matrix transpose method.
// It performs an implicit transpose by returning the receiver inside a Transpose.
//func (tsr *Complex64) H() mat.Matrix {
//	return mat.Transpose{tsr}
//}

// etensor.Complex128 is a tensor of complex128 backed by a []complex128 slice
type Complex128 struct {
	Shape
	Values []complex128
	Nulls  bitslice.Slice
}

// NewComplex128 returns a new n-dimensional array of Complex128
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty Complex128 will be created.
func NewComplex128(shape, strides []int, names []string) *Complex128 {
	bt := &Complex128{}
	bt.SetShape(shape, strides, names)
	ln := bt.Len()
	bt.Values = make([]complex128, ln)
	return bt
}

// NewComplex128Shape returns a new n-dimensional array of Complex128 from given shape
func NewComplex128Shape(shape *Shape) *Complex128 {
	bt := &Complex128{}
	bt.CopyShape(shape)
	ln := bt.Len()
	bt.Values = make([]complex128, ln)
	return bt
}

func (tsr *Complex128) DataType() Type { return COMPLEX64 }

// Value returns value at given tensor index
func (tsr *Complex128) Value(i []int) complex128 {
	j := tsr.Offset(i)
	return tsr.Values[j]
}

// Value1D returns value at given 1D (flat) tensor index
func (tsr *Complex128) Value1D(i int) complex128 {
	return tsr.Values[i]
}

// Set sets value at given tensor index
func (tsr *Complex128) Set(i []int, val complex128) {
	j := int(tsr.Offset(i))
	tsr.Values[j] = val
}

func (tsr *Complex128) IsNull(i []int) bool {
	if tsr.Nulls == nil {
		return false
	}
	j := tsr.Offset(i)
	return tsr.Nulls.Index(j)
}
func (tsr *Complex128) SetNull(i []int, nul bool) {
	if tsr.Nulls == nil {
		tsr.Nulls = bitslice.Make(tsr.Len(), 0)
	}
	j := tsr.Offset(i)
	tsr.Nulls.Set(j, nul)
}

// FloatVal returns the real part of a Complex128 - see FloatValImag for getting imaginary part
func (tsr *Complex128) FloatVal(i []int) float64 {
	j := tsr.Offset(i)
	return float64(real(tsr.Values[j]))
}

// FloatValImag returns the imag portion of a Complex128 - see FloatVal for getting real part
func (tsr *Complex128) FloatValImag(i []int) float64 {
	j := tsr.Offset(i)
	return float64(imag(tsr.Values[j]))
}

// SetFloat replaces the real part of the complex number specified by the slice of matrix indices i
func (tsr *Complex128) SetFloat(i []int, val float64) {
	j := tsr.Offset(i)
	valr := float64(val)
	vali := imag(tsr.Values[j])
	tsr.Values[j] = complex(valr, vali)
}

// SetFloat replaces the imaginary part of the complex number specified by the slice of matrix indices i
func (tsr *Complex128) SetFloatImag(i []int, val float64) {
	j := tsr.Offset(i)
	valr := real(tsr.Values[j])
	vali := float64(val)
	tsr.Values[j] = complex(valr, vali)
}

// StringVal returns the numeric value as a string
func (tsr *Complex128) StringVal(i []int) string {
	j := tsr.Offset(i)
	rstr := kit.ToString(real(tsr.Values[j]))
	istr := kit.ToString(imag(tsr.Values[j]))
	return rstr + " + " + istr
}

// SetString sets the numeric value from a string represenation
func (tsr *Complex128) SetString(i []int, val string) {
	ss := strings.SplitN(val, "+", -1)
	rstr := ss[0]
	fvr, err := strconv.ParseFloat(rstr, 32)
	if err != nil {
		return
	}
	istr := ss[1]
	fvi, err := strconv.ParseFloat(istr, 32)
	if err == nil {
		j := tsr.Offset(i)
		tsr.Values[j] = complex(float64(fvr), float64(fvi))
	}
}

func (tsr *Complex128) FloatVal1D(off int) float64 {
	return float64(real(tsr.Values[off]))
}

// SetFloat replaces the imaginary part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex128) FloatVal1DImag(off int) float64 {
	return float64(imag(tsr.Values[off]))
}

// SetFloat replaces the real part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex128) SetFloat1D(off int, val float64) {
	valr := float64(val)
	vali := imag(tsr.Values[off])
	tsr.Values[off] = complex(valr, vali)
}

// SetFloat replaces the imaginary part of the complex number at the offset "off" for a 1D matrix
func (tsr *Complex128) SetFloat1DImag(off int, val float64) {
	valr := real(tsr.Values[off])
	vali := float64(val)
	tsr.Values[off] = complex(valr, vali)
}

// todo:
// AggFunc applies given aggregation function to each element in the tensor, using float64
// conversions of the values.  init is the initial value for the agg variable.  returns final
// aggregate value
func (tsr *Complex128) AggFunc(fun func(val float64, agg float64) float64, ini float64) float64 {
	//ln := tsr.Len()
	//ag := ini
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	ag = fun(val, ag)
	//}
	//return ag
	return float64(0.0)
}

// todo:
// EvalFunc applies given function to each element in the tensor, using float64
// conversions of the values, and puts the results into given float64 slice, which is
// ensured to be of the proper length
func (tsr *Complex128) EvalFunc(fun func(val float64) float64, res *[]float64) {
	//ln := tsr.Len()
	//if len(*res) != ln {
	//	*res = make([]float64, ln)
	//}
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	(*res)[j] = fun(val)
	//}
}

// SetFunc applies given function to each element in the tensor, using float64
// conversions of the values, and writes the results back into the same tensor values
func (tsr *Complex128) SetFunc(fun func(val float64) float64) {
	//ln := tsr.Len()
	//for j := 0; j < ln; j++ {
	//	val := float64(tsr.Values[j])
	//	tsr.Values[j] = complex128(fun(val))
	//}
}

// Clone creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Complex128) Clone() *Complex128 {
	csr := NewComplex128Shape(&tsr.Shape)
	copy(csr.Values, tsr.Values)
	if tsr.Nulls != nil {
		csr.Nulls = tsr.Nulls.Clone()
	}
	return csr
}

// CloneTensor creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Complex128) CloneTensor() Tensor {
	return nil
	//return tsr.Clone()
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Complex128) SetShape(shape, strides []int, names []string) {
	tsr.Shape.SetShape(shape, strides, names)
	nln := tsr.Len()
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]complex128, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *Complex128) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows, cells := tsr.RowCellSize()
	nln := (rows + n) * cells
	tsr.Shape.Shp[0] += n
	if cap(tsr.Values) >= nln {
		tsr.Values = tsr.Values[0:nln]
	} else {
		nv := make([]complex128, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Complex128) SetNumRows(rows int) {
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
		nv := make([]complex128, nln)
		copy(nv, tsr.Values)
		tsr.Values = nv
	}
}

// SubSlice returns a new tensor as a sub-slice of the current one, incorporating the given number
// of dimensions (0 < subdim < NumDims of this tensor).  Only valid for row or column major layouts.
// subdim are the inner, contiguous dimensions (i.e., the final dims in RowMajor and the first ones in ColMajor).
// offs are offsets for the outer dimensions (len = NDims - subdim) for the subslice to return.
// The new tensor points to the values of the this tensor (i.e., modifications will affect both).
// Use Clone() method to separate the two.
// todo: not getting nulls yet.
func (tsr *Complex128) SubSlice(subdim int, offs []int) (*Complex128, error) {
	nd := tsr.NumDims()
	od := nd - subdim
	if od <= 0 {
		return nil, errors.New("SubSlice number of sub dimensions was >= NumDims -- must be less")
	}
	if tsr.IsRowMajor() {
		stsr := &Complex128{}
		stsr.SetShape(tsr.Shp[od:], nil, tsr.Nms[od:]) // row major def
		sti := make([]int, nd)
		copy(sti, offs)
		stoff := tsr.Offset(sti)
		stsr.Values = tsr.Values[stoff:]
		return stsr, nil
	} else if tsr.IsColMajor() {
		stsr := &Complex128{}
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
	return nil, errors.New("SubSlice only valid for RowMajor or ColMajor tensors")
}

// At(i, j) is the gonum/mat.Matrix interface method for returning 2D matrix element at given
// row, column index.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Complex128) At(i, j int) complex128 {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0
	} else if nd == 2 {
		return tsr.Value([]int{i, j})
	} else {
		ix := make([]int, nd)
		ix[nd-2] = i
		ix[nd-1] = j
		return tsr.Value(ix)
	}
}

// Dims is the gonum/mat.Matrix interface method for returning the dimensionality of the
// 2D Matrix.  Assumes Row-major ordering and logs an error if NumDims < 2.
func (tsr *Complex128) Dims() (r, c int) {
	nd := tsr.NumDims()
	if nd < 2 {
		log.Println("etensor Dims gonum Matrix call made on Tensor with dims < 2")
		return 0, 0
	}
	return tsr.Dim(nd - 2), tsr.Dim(nd - 1)
}

// Todo: this needs to be implemented - but not sure how now
// H is the gonum/mat.Matrix transpose method.
// It performs an implicit transpose by returning the receiver inside a Transpose.
//func (tsr *Complex128) H() mat.Matrix {
//	return mat.Transpose{tsr}
//}
