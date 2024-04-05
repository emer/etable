// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import "fmt"

// Shape manages a tensor's shape information, including strides and dimension names
// and can compute the flat index into an underlying 1D data storage array based on an
// n-dimensional index (and vice-versa).
// This is fully compatible with (and largely taken from) apache/arrow tensors.
// except that we use plain int instead of int64, because on all relevant platforms
// int is *already* 64 and using plain int is much easier.
// Per C / Go / Python conventions (and unlike emergent) by default indexes are ordered from
// outer to inner left-to-right, so the inner-most is right-most.
// This is called Row-Major order, and is the default.
// It is also possible to use Column-Major order, which is used in R, Julia, and MATLAB,
// and emergent, where the inner-most index is first and outer-most last.  In either
// case, the internal memory is always organized with the inner-most dimension contiguous
// in memory -- thus there is no underlying difference between the different indexing
// systems in terms of underlying memory -- just in the order of indexes used to access
// this memory.
type Shape struct {

	// shape is size per dimension
	Shp []int

	// stride is offset per dimension
	Strd []int `tableview:"-" view:"-"`

	// names of each dimension
	Nms []string `tableview:"-" view:"-"`
}

// NewShape returns a new shape object initialized with params.
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewShape(shape, strides []int, names []string) *Shape {
	sh := &Shape{}
	sh.SetShape(shape, strides, names)
	return sh
}

// SetShape sets the shape parameters.
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func (sh *Shape) SetShape(shape, strides []int, names []string) {
	sh.Shp = CopyInts(shape)
	if len(strides) != len(shape) {
		sh.Strd = RowMajorStrides(shape)
	} else {
		sh.Strd = CopyInts(strides)
	}
	sh.Nms = make([]string, len(sh.Shp))
	if len(names) == len(shape) {
		copy(sh.Nms, names)
	}
}

// SetShape64 sets the shape parameters from int64 slices (e.g., arrow/tensor).
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func (sh *Shape) SetShape64(shape, strides []int64, names []string) {
	sh.Shp = CopyInts64(shape)
	if strides == nil {
		sh.Strd = RowMajorStrides(sh.Shp)
	} else {
		sh.Strd = CopyInts64(strides)
	}
	sh.Nms = make([]string, len(sh.Shp))
	if names != nil {
		copy(sh.Nms, names)
	}
}

// CopyShape copies the shape parameters from another Shape struct.
// copies the data so it is not accidentally subject to updates.
func (sh *Shape) CopyShape(cp *Shape) {
	sh.Shp = CopyInts(cp.Shp)
	sh.Strd = CopyInts(cp.Strd)
	sh.Nms = CopyStrings(cp.Nms)
}

// AddShapes returns a new shape by adding two shapes one after the other.
// uses Row / Col order of the first shape for resulting shape
func AddShapes(shape1, shape2 *Shape) *Shape {
	sh1 := shape1.Shapes()
	sh2 := shape2.Shapes()
	nsh := make([]int, len(sh1)+len(sh2))
	copy(nsh, sh1)
	copy(nsh[len(sh1):], sh2)
	rm := shape1.IsRowMajor()
	var nstr []int
	if rm {
		nstr = RowMajorStrides(nsh)
	} else {
		nstr = ColMajorStrides(nsh)
	}
	nms := make([]string, len(sh1)+len(sh2))
	copy(nms, shape1.DimNames())
	copy(nms[len(sh1):], shape2.DimNames())
	return NewShape(nsh, nstr, nms)
}

// Len returns the total length of elements in the tensor (i.e., the product of
// the shape sizes)
func (sh *Shape) Len() int {
	if len(sh.Shp) == 0 {
		return 0
	}
	o := int(1)
	for _, v := range sh.Shp {
		o *= v
	}
	return int(o)
}

// Shapes returns the slice of dimension sizes.
// This is *not* a copy -- modifications will change the shape.
func (sh *Shape) Shapes() []int { return sh.Shp }

// Strides returns the slice of strides
// This is *not* a copy -- modifications will change the shape.
func (sh *Shape) Strides() []int { return sh.Strd }

// Shape64 returns a slice of int64 containing the dimensions sizes.
// This is a copy -- modifications will not change shape.
func (sh *Shape) Shape64() []int64 { return IntTo64(sh.Shp) }

// Strides64 returns a slice of int64 containing strides
// This is a copy -- modifications will not change shape.
func (sh *Shape) Strides64() []int64 { return IntTo64(sh.Strd) }

// DimNames returns slice of dimension names
// This is *not* a copy -- modifications will change the shape.
func (sh *Shape) DimNames() []string { return sh.Nms }

// NumDims returns the total number of dimensions.
func (sh *Shape) NumDims() int { return len(sh.Shp) }

// Dim returns the size of given dimension.
func (sh *Shape) Dim(i int) int { return sh.Shp[i] }

// DimName returns the name of given dimension.
func (sh *Shape) DimName(i int) string { return sh.Nms[i] }

// DimIndexByNameTry returns the index of the given dimension name.
// and an error if name not found.
func (sh *Shape) DimIndexByNameTry(name string) (int, error) {
	for i, nm := range sh.Nms {
		if nm == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("etensor.Shape:DimIndexByNameTry -- dimension name not found: %s", name)
}

// DimIndexByName returns the index of the given dimension name.
// returns -1 if not found.
func (sh *Shape) DimIndexByName(name string) int {
	idx, _ := sh.DimIndexByNameTry(name)
	return idx
}

// DimByName returns the size of given dimension, specified by name.
// will crash if name not found -- use DimIndexByNameTry if not sure.
func (sh *Shape) DimByName(name string) int {
	return sh.Dim(sh.DimIndexByName(name))
}

// IsContiguous returns true if shape is either row or column major
func (sh *Shape) IsContiguous() bool {
	return sh.IsRowMajor() || sh.IsColMajor()
}

// IndexIsValid() returns true if given index is valid (within ranges for all dimensions)
func (sh *Shape) IndexIsValid(idx []int) bool {
	if len(idx) != sh.NumDims() {
		return false
	}
	for i, v := range sh.Shp {
		if idx[i] < 0 || idx[i] >= v {
			return false
		}
	}
	return true
}

// todo: cache rowmajor vs. colmajor as flags?  much faster, and this is frequently
// accessed

// IsRowMajor returns true if shape is row-major organized:
// first dimension is the row or outer-most storage dimension.
// Values *along a row* are contiguous, as you increment along
// the minor, inner-most (column) dimension.
// Importantly: ColMajor and RowMajor both have the *same*
// actual memory storage arrangement, with values along a row
// (across columns) contiguous in memory -- the only difference
// is in the order of the indexes used to access this memory.
func (sh *Shape) IsRowMajor() bool {
	strides := RowMajorStrides(sh.Shp)
	return EqualInts(strides, sh.Strd)
}

// IsColMajor returns true if shape is column-major organized:
// first dimension is column, i.e., inner-most storage dimension.
// Values *along a row* are contiguous, as you increment along
// the major inner-most (column) dimension.
// Importantly: ColMajor and RowMajor both have the *same*
// actual memory storage arrangement, with values along a row
// (across columns) contiguous in memory -- the only difference
// is in the order of the indexes used to access this memory.
func (sh *Shape) IsColMajor() bool {
	strides := ColMajorStrides(sh.Shp)
	return EqualInts(strides, sh.Strd)
}

// IsEqual returns true if this shape is same as other (does not compare names)
func (sh *Shape) IsEqual(oth *Shape) bool {
	if !EqualInts(sh.Shp, oth.Shp) {
		return false
	}
	if !EqualInts(sh.Strd, oth.Strd) {
		return false
	}
	return true
}

// RowCellSize returns the size of the outer-most Row shape dimension, and the size of all the
// remaining inner dimensions (the "cell" size) -- e.g., for Tensors that are columns in a
// data table.  Only valid for RowMajor organization.
func (sh *Shape) RowCellSize() (rows, cells int) {
	rows = sh.Shp[0]
	if len(sh.Shp) == 1 {
		cells = 1
	} else {
		cells = sh.Len() / rows
	}
	return
}

// Offset returns the "flat" 1D array index into an element at the given n-dimensional index
// No checking is done on the length or size of the index values relative to the shape of the tensor.
func (sh *Shape) Offset(index []int) int {
	var offset int
	for i, v := range index {
		offset += v * sh.Strd[i]
	}
	return offset
}

// OffsetByName returns the "flat" 1D array index into an element at the given n-dimensional index
// using given list of names to order the indexes.  Both names and index must be same length which
// is the dimensionality of the shape -- none of that is checked.
func (sh *Shape) OffsetByName(names []string, index []int) int {
	oidx := CopyInts(index)
	for i, nm := range names {
		idx := sh.DimIndexByName(nm)
		oidx[idx] = index[i]
	}
	return sh.Offset(oidx)
}

// Index returns the n-dimensional index from a "flat" 1D array index.  Only works for RowMajor
// or ColMajor organization.
func (sh *Shape) Index(offset int) []int {
	nd := len(sh.Shp)
	index := make([]int, nd)
	if sh.IsRowMajor() {
		rem := offset
		for i := nd - 1; i >= 0; i-- {
			s := sh.Shp[i]
			iv := rem % s
			rem /= s
			index[i] = iv
		}
	} else if sh.IsColMajor() {
		rem := offset
		for i := 0; i < nd; i++ {
			s := sh.Shp[i]
			iv := rem % s
			rem /= s
			index[i] = iv
		}
	}
	return index
}

// String satisfies the fmt.Stringer interface
func (sh *Shape) String() string {
	str := "["
	for i := range sh.Shp {
		nm := sh.Nms[i]
		if nm != "" {
			str += nm + ": "
		}
		str += fmt.Sprintf("%d", sh.Shp[i])
		if i < len(sh.Shp)-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}

// RowMajorStrides returns strides for shape where the first dimension is outer-most
// and subsequent dimensions are progressively inner
func RowMajorStrides(shape []int) []int {
	rem := int(1)
	for _, v := range shape {
		rem *= v
	}

	if rem == 0 {
		strides := make([]int, len(shape))
		rem := int(1)
		for i := range strides {
			strides[i] = rem
		}
		return strides
	}

	strides := make([]int, len(shape))
	for i, v := range shape {
		rem /= v
		strides[i] = rem
	}
	return strides
}

// ColMajorStrides returns strides for shape where the first dimension is inner-most
// and subsequent dimensions are progressively outer
func ColMajorStrides(shape []int) []int {
	total := int(1)
	for _, v := range shape {
		if v == 0 {
			strides := make([]int, len(shape))
			for i := range strides {
				strides[i] = total
			}
			return strides
		}
	}

	strides := make([]int, len(shape))
	for i, v := range shape {
		strides[i] = total
		total *= v
	}
	return strides
}

// EqualInts compares two int slices and returns true if they are equal
func EqualInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CopyInts makes a copy of an int slice
func CopyInts(a []int) []int {
	ns := make([]int, len(a))
	copy(ns, a)
	return ns
}

// CopyInts64 makes a copy of an int64 slice to an int slice
func CopyInts64(a []int64) []int {
	ln := len(a)
	ns := make([]int, ln)
	for i := 0; i < ln; i++ {
		ns[i] = int(a[i])
	}
	return ns
}

// IntTo64 converts an []int slice to an []int64 slice
func IntTo64(isl []int) []int64 {
	is := make([]int64, len(isl))
	for i := range isl {
		is[i] = int64(isl[i])
	}
	return is
}

// CopyStrings makes a copy of a string slice
func CopyStrings(a []string) []string {
	ns := make([]string, len(a))
	copy(ns, a)
	return ns
}
