// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

// Prjn2DShape returns the size of a 2D projection of the given tensor,
// collapsing higher dimensions down to 2D (and 1D up to 2D).
// For any odd number of dimensions, the remaining outer-most dimension
// can either be multipliexed across the row or column, given the oddRow arg.
// Even multiples of inner-most dimensions are assumed to be row, then column.
// RowMajor and ColMajor layouts are handled appropriately.
// rowEx returns the number of "extra" (higher dimensional) rows
// and colEx returns the number of extra cols
func Prjn2DShape(tsr Tensor, oddRow bool) (rows, cols, rowEx, colEx int) {
	if tsr.Len() == 0 {
		return 1, 1, 0, 0
	}
	nd := tsr.NumDims()
	switch nd {
	case 1:
		if oddRow {
			return tsr.Dim(0), 1, 0, 0
		} else {
			return 1, tsr.Dim(0), 0, 0
		}
	case 2:
		if tsr.ShapeObj().IsRowMajor() {
			return tsr.Dim(0), tsr.Dim(1), 0, 0
		} else {
			return tsr.Dim(1), tsr.Dim(0), 0, 0
		}
	case 3:
		if oddRow {
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.Dim(0) * tsr.Dim(1), tsr.Dim(2), tsr.Dim(0), 0
			} else {
				return tsr.Dim(2) * tsr.Dim(1), tsr.Dim(0), tsr.Dim(2), 0
			}
		} else {
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.Dim(1), tsr.Dim(0) * tsr.Dim(2), 0, tsr.Dim(0)
			} else {
				return tsr.Dim(1), tsr.Dim(2) * tsr.Dim(0), 0, tsr.Dim(2)
			}
		}
	case 4:
		if tsr.ShapeObj().IsRowMajor() {
			return tsr.Dim(0) * tsr.Dim(2), tsr.Dim(1) * tsr.Dim(3), tsr.Dim(0), tsr.Dim(1)
		} else {
			return tsr.Dim(3) * tsr.Dim(1), tsr.Dim(2) * tsr.Dim(0), tsr.Dim(3), tsr.Dim(2)
		}
	case 5:
		if oddRow {
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.Dim(0) * tsr.Dim(1) * tsr.Dim(3), tsr.Dim(2) * tsr.Dim(4), tsr.Dim(0) * tsr.Dim(1), 0
			} else {
				return tsr.Dim(4) * tsr.Dim(3) * tsr.Dim(1), tsr.Dim(2) * tsr.Dim(0), tsr.Dim(4) * tsr.Dim(3), 0
			}
		} else {
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.Dim(1) * tsr.Dim(3), tsr.Dim(0) * tsr.Dim(2) * tsr.Dim(4), 0, tsr.Dim(0) * tsr.Dim(1)
			} else {
				return tsr.Dim(3) * tsr.Dim(1), tsr.Dim(4) * tsr.Dim(2) * tsr.Dim(0), 0, tsr.Dim(4) * tsr.Dim(2)
			}
		}
	}
	return 1, 1, 0, 0
}

// Prjn2DVal returns the float64 value at given row, col coords for a 2D projection
// of the given tensor, collapsing higher dimensions down to 2D (and 1D up to 2D).
// For any odd number of dimensions, the remaining outer-most dimension
// can either be multipliexed across the row or column, given the oddRow arg.
// Even multiples of inner-most dimensions are assumed to be row, then column.
// RowMajor and ColMajor layouts are handled appropriately.
func Prjn2DVal(tsr Tensor, oddRow bool, row, col int) float64 {
	nd := tsr.NumDims()
	switch nd {
	case 1:
		if oddRow {
			return tsr.FloatVal1D(row)
		} else {
			return tsr.FloatVal1D(col)
		}
	case 2:
		if tsr.ShapeObj().IsRowMajor() {
			return tsr.FloatVal([]int{row, col})
		} else {
			return tsr.FloatVal([]int{col, row})
		}
	case 3:
		if oddRow {
			ny := tsr.Dim(1)
			yy := row / ny
			y := row % ny
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.FloatVal([]int{yy, y, col})
			} else {
				return tsr.FloatVal([]int{col, y, yy})
			}
		} else {
			nx := tsr.Dim(2)
			xx := col / nx
			x := col % nx
			if tsr.ShapeObj().IsRowMajor() {
				return tsr.FloatVal([]int{xx, row, x})
			} else {
				return tsr.FloatVal([]int{x, row, xx})
			}
		}
	case 4:
		if tsr.ShapeObj().IsRowMajor() {
			ny := tsr.Dim(2)
			yy := row / ny
			y := row % ny
			nx := tsr.Dim(3)
			xx := col / nx
			x := col % nx
			return tsr.FloatVal([]int{yy, xx, y, x})
		} else {
			ny := tsr.Dim(1)
			yy := row / ny
			y := row % ny
			nx := tsr.Dim(0)
			xx := col / nx
			x := col % nx
			return tsr.FloatVal([]int{x, y, xx, yy})
		}
	case 5:
		// todo: oddRows version!
		if tsr.ShapeObj().IsRowMajor() {
			nyy := tsr.Dim(1)
			ny := tsr.Dim(3)
			yyy := row / nyy
			yy := row % nyy
			y := yy % ny
			yy = yy / ny
			nx := tsr.Dim(4)
			xx := col / nx
			x := col % nx
			return tsr.FloatVal([]int{yyy, yy, xx, y, x})
		} else {
			nyy := tsr.Dim(3)
			ny := tsr.Dim(1)
			yyy := row / nyy
			yy := row % nyy
			y := yy % ny
			yy = yy / ny
			nx := tsr.Dim(0)
			xx := col / nx
			x := col % nx
			return tsr.FloatVal([]int{x, y, xx, yy, yyy})
		}
	}
	return 0
}

// Prjn2DSet sets a float64 value at given row, col coords for a 2D projection
// of the given tensor, collapsing higher dimensions down to 2D (and 1D up to 2D).
// For any odd number of dimensions, the remaining outer-most dimension
// can either be multipliexed across the row or column, given the oddRow arg.
// Even multiples of inner-most dimensions are assumed to be row, then column.
// RowMajor and ColMajor layouts are handled appropriately.
func Prjn2DSet(tsr Tensor, oddRow bool, row, col int, val float64) {
	nd := tsr.NumDims()
	switch nd {
	case 1:
		if oddRow {
			tsr.SetFloat1D(row, val)
		} else {
			tsr.SetFloat1D(col, val)
		}
	case 2:
		if tsr.ShapeObj().IsRowMajor() {
			tsr.SetFloat([]int{row, col}, val)
		} else {
			tsr.SetFloat([]int{col, row}, val)
		}
	case 3:
		if oddRow {
			ny := tsr.Dim(1)
			yy := row / ny
			y := row % ny
			if tsr.ShapeObj().IsRowMajor() {
				tsr.SetFloat([]int{yy, y, col}, val)
			} else {
				tsr.SetFloat([]int{col, y, yy}, val)
			}
		} else {
			nx := tsr.Dim(2)
			xx := col / nx
			x := col % nx
			if tsr.ShapeObj().IsRowMajor() {
				tsr.SetFloat([]int{xx, row, x}, val)
			} else {
				tsr.SetFloat([]int{x, row, xx}, val)
			}
		}
	case 4:
		if tsr.ShapeObj().IsRowMajor() {
			ny := tsr.Dim(2)
			yy := row / ny
			y := row % ny
			nx := tsr.Dim(3)
			xx := col / nx
			x := col % nx
			tsr.SetFloat([]int{yy, xx, y, x}, val)
		} else {
			ny := tsr.Dim(1)
			yy := row / ny
			y := row % ny
			nx := tsr.Dim(0)
			xx := col / nx
			x := col % nx
			tsr.SetFloat([]int{x, y, xx, yy}, val)
		}
	case 5:
		// todo: oddRows version!
		if tsr.ShapeObj().IsRowMajor() {
			nyy := tsr.Dim(1)
			ny := tsr.Dim(3)
			yyy := row / nyy
			yy := row % nyy
			y := yy % ny
			yy = yy / ny
			nx := tsr.Dim(4)
			xx := col / nx
			x := col % nx
			tsr.SetFloat([]int{yyy, yy, xx, y, x}, val)
		} else {
			nyy := tsr.Dim(3)
			ny := tsr.Dim(1)
			yyy := row / nyy
			yy := row % nyy
			y := yy % ny
			yy = yy / ny
			nx := tsr.Dim(0)
			xx := col / nx
			x := col % nx
			tsr.SetFloat([]int{x, y, xx, yy, yyy}, val)
		}
	}
}
