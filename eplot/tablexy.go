// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"errors"
	"log"
	"math"

	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/minmax"
)

// TableXY selects two columns from a etable.Table data table to plot in a gonum plot,
// satisfying the plotter.XYer and .Valuer interfaces (for bar charts).
// For Tensor-valued cells, Index's specify tensor cell.
// Also satisfies the plotter.Labeler interface for labels attached to a line, and
// plotter.YErrorer for error bars.
type TableXY struct {

	// the index view of data table to plot from
	Table *etable.IndexView

	// the indexes of the tensor columns to use for the X and Y data, respectively
	XCol, YCol int

	// numer of elements in each row of data -- 1 for scalar, > 1 for multi-dimensional
	XRowSz, YRowSz int

	// the indexes of the element within each tensor cell if cells are n-dimensional, respectively
	XIndex, YIndex int

	// the column to use for returning a label using Label interface -- for string cols
	LblCol int

	// the column to use for returning errorbars (+/- given value) -- if YCol is tensor then this must also be a tensor and given YIndex used
	ErrCol int

	// range constraints on Y values
	YRange minmax.Range64
}

// NewTableXY returns a new XY plot view onto the given IndexView of etable.Table (makes a copy),
// from given column indexes, and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXY(dt *etable.IndexView, xcol, xtsrIndex, ycol, ytsrIndex int, yrng minmax.Range64) (*TableXY, error) {
	txy := &TableXY{Table: dt.Clone(), XCol: xcol, YCol: ycol, XIndex: xtsrIndex, YIndex: ytsrIndex, YRange: yrng}
	return txy, txy.Validate()
}

// NewTableXYName returns a new XY plot view onto the given IndexView of etable.Table (makes a copy),
// from given column name and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXYName(dt *etable.IndexView, xi, xtsrIndex int, ycol string, ytsrIndex int, yrng minmax.Range64) (*TableXY, error) {
	yi, err := dt.Table.ColIndexTry(ycol)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	txy := &TableXY{Table: dt.Clone(), XCol: xi, YCol: yi, XIndex: xtsrIndex, YIndex: ytsrIndex, YRange: yrng}
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
		if txy.YIndex >= txy.YRowSz || txy.YIndex < 0 {
			txy.YIndex = 0
			return errors.New("eplot.TableXY Y TensorIndex invalid -- reset to 0")
		}
	}
	txy.FilterValues()
	return nil
}

// FilterValues removes items with NaN values, and out of Y range
func (txy *TableXY) FilterValues() {
	txy.Table.Filter(func(et *etable.Table, row int) bool {
		xv := txy.TRowXValue(row)
		yv := txy.TRowValue(row)
		if math.IsNaN(yv) || math.IsNaN(xv) {
			return false
		}
		if txy.YRange.FixMin && yv < txy.YRange.Min {
			return false
		}
		if txy.YRange.FixMax && yv > txy.YRange.Max {
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
		if txy.YIndex < sz && txy.YIndex >= 0 {
			y = yc.FloatValueRowCell(row, txy.YIndex)
		}
	default:
		y = yc.FloatValue1D(row)
		if yc.IsNull1D(row) {
			y = math.NaN()
		}
	}
	return y
}

// Value returns the y value at given row in table view
func (txy *TableXY) Value(row int) float64 {
	if txy.Table == nil || txy.Table.Table == nil || row >= txy.Table.Len() {
		return 0
	}
	trow := txy.Table.Indexes[row] // true table row
	yc := txy.Table.Table.Cols[txy.YCol]
	y := 0.0
	switch {
	case yc.DataType() == etensor.STRING:
		y = float64(row)
	case yc.NumDims() > 1:
		_, sz := yc.RowCellSize()
		if txy.YIndex < sz && txy.YIndex >= 0 {
			y = yc.FloatValueRowCell(trow, txy.YIndex)
		}
	default:
		y = yc.FloatValue1D(trow)
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
		if txy.XIndex < sz && txy.XIndex >= 0 {
			x = xc.FloatValueRowCell(row, txy.XIndex)
		}
	default:
		x = xc.FloatValue1D(row)
		if xc.IsNull1D(row) {
			x = math.NaN()
		}
	}
	return x
}

// XValue returns an x value at given row in table view
func (txy *TableXY) XValue(row int) float64 {
	if txy.Table == nil || txy.Table.Table == nil || row >= txy.Table.Len() {
		return 0
	}
	trow := txy.Table.Indexes[row] // true table row
	xc := txy.Table.Table.Cols[txy.XCol]
	x := 0.0
	switch {
	case xc.DataType() == etensor.STRING:
		x = float64(row)
	case xc.NumDims() > 1:
		_, sz := xc.RowCellSize()
		if txy.XIndex < sz && txy.XIndex >= 0 {
			x = xc.FloatValueRowCell(trow, txy.XIndex)
		}
	default:
		x = xc.FloatValue1D(trow)
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
	if txy.Table == nil || txy.Table.Table == nil || row >= txy.Table.Len() {
		return ""
	}
	trow := txy.Table.Indexes[row] // true table row
	return txy.Table.Table.Cols[txy.LblCol].StringValue1D(trow)
}

// YError returns a error bars using ploter.YErrorer interface
func (txy *TableXY) YError(row int) (float64, float64) {
	if txy.Table == nil || txy.Table.Table == nil || row >= txy.Table.Len() {
		return 0, 0
	}
	trow := txy.Table.Indexes[row] // true table row
	ec := txy.Table.Table.Cols[txy.ErrCol]
	eval := 0.0
	switch {
	case ec.DataType() == etensor.STRING:
		eval = float64(row)
	case ec.NumDims() > 1:
		_, sz := ec.RowCellSize()
		if txy.YIndex < sz && txy.YIndex >= 0 {
			eval = ec.FloatValueRowCell(trow, txy.YIndex)
		}
	default:
		eval = ec.FloatValue1D(trow)
	}
	return -eval, eval
}
