// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"errors"
	"log"
	"math"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
)

// TableXY selects two columns from a etable.Table data table to plot in a gonum plot,
// satisfying the plotter.XYer and .Valuer interfaces (for bar charts).
// For Tensor-valued cells, Idx's specify tensor cell.
// Also satisfies the plotter.Labeler interface for labels attached to a line, and
// plotter.YErrorer for error bars.
type TableXY struct {
	Table          *etable.IdxView `desc:"the index view of data table to plot from"`
	XCol, YCol     int             `desc:"the indexes of the tensor columns to use for the X and Y data, respectively"`
	XRowSz, YRowSz int             `desc:"numer of elements in each row of data -- 1 for scalar, > 1 for multi-dimensional"`
	XIdx, YIdx     int             `desc:"the indexes of the element within each tensor cell if cells are n-dimensional, respectively"`
	LblCol         int             `desc:"the column to use for returning a label using Label interface -- for string cols"`
	ErrCol         int             `desc:"the column to use for returning errorbars (+/- given value) -- if YCol is tensor then this must also be a tensor and given YIdx used"`
	XRange         minmax.Range64
}

// NewTableXY returns a new XY plot view onto the given IdxView of etable.Table (makes a copy),
// from given column indexes, and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXY(dt *etable.IdxView, xcol, xtsrIdx, ycol, ytsrIdx int) (*TableXY, error) {
	txy := &TableXY{Table: dt.Clone(), XCol: xcol, YCol: ycol, XIdx: xtsrIdx, YIdx: ytsrIdx}
	return txy, txy.Validate()
}

// NewTableXYName returns a new XY plot view onto the given IdxView of etable.Table (makes a copy),
// from given column name and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXYName(dt *etable.IdxView, xi, xtsrIdx int, ycol string, ytsrIdx int) (*TableXY, error) {
	yi, err := dt.Table.ColIdxTry(ycol)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	txy := &TableXY{Table: dt.Clone(), XCol: xi, YCol: yi, XIdx: xtsrIdx, YIdx: ytsrIdx}
	return txy, txy.Validate()
}

// Validate returns error message if column indexes are invalid, else nil
// it also sets column indexes to 0 so nothing crashes.
func (txy *TableXY) Validate() error {
	if txy.Table == nil {
		return errors.New("eplot.TableXY table is nil")
	}
	nc := txy.Table.Table.NumCols()
	if txy.XCol >= nc || txy.XCol < 0 {
		txy.XCol = 0
		return errors.New("eplot.TableXY XCol index invalid -- reset to 0")
	}
	if txy.YCol >= nc || txy.YCol < 0 {
		txy.YCol = 0
		return errors.New("eplot.TableXY YCol index invalid -- reset to 0")
	}
	xc := txy.Table.Table.Cols[txy.XCol]
	yc := txy.Table.Table.Cols[txy.YCol]
	if xc.NumDims() > 1 {
		_, txy.XRowSz = xc.RowCellSize()
		// note: index already validated
	}
	if yc.NumDims() > 1 {
		_, txy.YRowSz = yc.RowCellSize()
		if txy.YIdx >= txy.YRowSz || txy.YIdx < 0 {
			txy.YIdx = 0
			return errors.New("eplot.TableXY Y TensorIdx invalid -- reset to 0")
		}
	}
	txy.FilterVals()
	return nil
}

// FilterVals removes items with NaN values, and out of X range
func (txy *TableXY) FilterVals() {
	txy.Table.Filter(func(et *etable.Table, row int) bool {
		xv := txy.TRowXValue(row)
		yv := txy.TRowValue(row)
		if math.IsNaN(yv) || math.IsNaN(xv) {
			return false
		}
		return true
	})
}

// Len returns the number of rows in the view of table
func (txy *TableXY) Len() int {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0
	}
	return txy.Table.Len()
}

// TRowValue returns the y value at given true table row in table view
func (txy *TableXY) TRowValue(row int) float64 {
	yc := txy.Table.Table.Cols[txy.YCol]
	y := 0.0
	switch {
	case yc.DataType() == etensor.STRING:
		y = float64(row)
	case yc.NumDims() > 1:
		_, sz := yc.RowCellSize()
		if txy.YIdx < sz && txy.YIdx >= 0 {
			y = yc.FloatValRowCell(row, txy.YIdx)
		}
	default:
		y = yc.FloatVal1D(row)
		if yc.IsNull1D(row) {
			y = math.NaN()
		}
	}
	return y
}

// Value returns the y value at given row in table view
func (txy *TableXY) Value(row int) float64 {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0
	}
	trow := txy.Table.Idxs[row] // true table row
	yc := txy.Table.Table.Cols[txy.YCol]
	y := 0.0
	switch {
	case yc.DataType() == etensor.STRING:
		y = float64(row)
	case yc.NumDims() > 1:
		_, sz := yc.RowCellSize()
		if txy.YIdx < sz && txy.YIdx >= 0 {
			y = yc.FloatValRowCell(trow, txy.YIdx)
		}
	default:
		y = yc.FloatVal1D(trow)
		if yc.IsNull1D(trow) {
			y = math.NaN()
		}
	}
	return y
}

// TRowXValue returns an x value at given actual row in table
func (txy *TableXY) TRowXValue(row int) float64 {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0
	}
	xc := txy.Table.Table.Cols[txy.XCol]
	x := 0.0
	switch {
	case xc.DataType() == etensor.STRING:
		x = float64(row)
	case xc.NumDims() > 1:
		_, sz := xc.RowCellSize()
		if txy.XIdx < sz && txy.XIdx >= 0 {
			x = xc.FloatValRowCell(row, txy.XIdx)
		}
	default:
		x = xc.FloatVal1D(row)
		if xc.IsNull1D(row) {
			x = math.NaN()
		}
	}
	return x
}

// XValue returns an x value at given row in table view
func (txy *TableXY) XValue(row int) float64 {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0
	}
	trow := txy.Table.Idxs[row] // true table row
	xc := txy.Table.Table.Cols[txy.XCol]
	x := 0.0
	switch {
	case xc.DataType() == etensor.STRING:
		x = float64(row)
	case xc.NumDims() > 1:
		_, sz := xc.RowCellSize()
		if txy.XIdx < sz && txy.XIdx >= 0 {
			x = xc.FloatValRowCell(trow, txy.XIdx)
		}
	default:
		x = xc.FloatVal1D(trow)
		if xc.IsNull1D(trow) {
			x = math.NaN()
		}
	}
	return x
}

// XY returns an x, y pair at given row in table
func (txy *TableXY) XY(row int) (x, y float64) {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0, 0
	}
	x = txy.XValue(row)
	y = txy.Value(row)
	return
}

// Label returns a label for given row in table, using plotter.Labeler interface
func (txy *TableXY) Label(row int) string {
	if txy.Table == nil || txy.Table.Table == nil {
		return ""
	}
	trow := txy.Table.Idxs[row] // true table row
	return txy.Table.Table.Cols[txy.LblCol].StringVal1D(trow)
}

// YError returns a error bars using ploter.YErrorer interface
func (txy *TableXY) YError(row int) (float64, float64) {
	if txy.Table == nil || txy.Table.Table == nil {
		return 0, 0
	}
	trow := txy.Table.Idxs[row] // true table row
	ec := txy.Table.Table.Cols[txy.ErrCol]
	eval := 0.0
	switch {
	case ec.DataType() == etensor.STRING:
		eval = float64(row)
	case ec.NumDims() > 1:
		_, sz := ec.RowCellSize()
		if txy.YIdx < sz && txy.YIdx >= 0 {
			eval = ec.FloatValRowCell(trow, txy.YIdx)
		}
	default:
		eval = ec.FloatVal1D(trow)
	}
	return -eval, eval
}
