// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"errors"
	"log"

	"github.com/emer/etable/etable"
)

// TableXY selects two columns from a etable.Table data table to plot in a gonum plot,
// satisfying the plotter.XYer interface.  For Tensor-valued cells, Idx's specify tensor cell.
type TableXY struct {
	Table          *etable.Table `desc:"the data table to plot from"`
	stRow, edRow   int           `desc:"starting, ending row numbers"`
	XCol, YCol     int           `desc:"the indexes of the tensor columns to use for the X and Y data, respectively"`
	XRowSz, YRowSz int           `desc:"numer of elements in each row of data -- 1 for scalar, > 1 for multi-dimensional"`
	XIdx, YIdx     int           `desc:"the indexes of the element within each tensor cell if cells are n-dimensional, respectively"`
	LblCol         int           `desc:"the column to use for returning a label using Label interface -- for string cols"`
	RowSt          int
}

// NewTableXY returns a new XY plot view onto the given etable.Table, from given column indexes,
// and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXY(dt *etable.Table, strow, edrow int, xcol, xtsrIdx, ycol, ytsrIdx int) (*TableXY, error) {
	txy := &TableXY{Table: dt, stRow: strow, edRow: edrow, XCol: xcol, YCol: ycol, XIdx: xtsrIdx, YIdx: ytsrIdx}
	return txy, txy.Validate()
}

// NewTableXYName returns a new XY plot view onto the given etable.Table, from given column name
// and tensor indexes within each cell.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXYName(dt *etable.Table, strow, edrow int, xi, xtsrIdx int, ycol string, ytsrIdx int) (*TableXY, error) {
	yi, err := dt.ColIdxTry(ycol)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	txy := &TableXY{Table: dt, stRow: strow, edRow: edrow, XCol: xi, YCol: yi, XIdx: xtsrIdx, YIdx: ytsrIdx}
	return txy, txy.Validate()
}

// Validate returns error message if column indexes are invalid, else nil
// it also sets column indexes to 0 so nothing crashes.
func (txy *TableXY) Validate() error {
	if txy.Table == nil {
		return errors.New("eplot.TableXY table is nil")
	}
	nc := txy.Table.NumCols()
	if txy.XCol >= nc || txy.XCol < 0 {
		txy.XCol = 0
		return errors.New("eplot.TableXY XCol index invalid -- reset to 0")
	}
	if txy.YCol >= nc || txy.YCol < 0 {
		txy.YCol = 0
		return errors.New("eplot.TableXY YCol index invalid -- reset to 0")
	}
	xc := txy.Table.Cols[txy.XCol]
	yc := txy.Table.Cols[txy.YCol]
	if xc.NumDims() > 1 {
		txy.XRowSz = xc.Len() / xc.Dim(0)
		// note: index already validated
	}
	if yc.NumDims() > 1 {
		txy.YRowSz = yc.Len() / yc.Dim(0)
		if txy.YIdx >= txy.YRowSz || txy.YIdx < 0 {
			txy.YIdx = 0
			return errors.New("eplot.TableXY Y TensorIdx invalid -- reset to 0")
		}
	}
	return nil
}

// Len returns the number of rows in the table
func (txy *TableXY) Len() int {
	if txy.Table == nil || txy.Table.Rows < 0 {
		return 0
	}
	return txy.edRow - txy.stRow
}

// XY returns an x, y pair at given row in table
func (txy *TableXY) XY(row int) (x, y float64) {
	if txy.Table == nil {
		return 0, 0
	}
	row += txy.stRow
	xc := txy.Table.Cols[txy.XCol]
	yc := txy.Table.Cols[txy.YCol]
	if xc.NumDims() > 1 {
		sz := xc.Len() / xc.Dim(0)
		if txy.XIdx < sz && txy.XIdx >= 0 {
			off := row*sz + txy.XIdx
			x = xc.FloatVal1D(off)
		}
	} else {
		x = xc.FloatVal1D(row)
	}
	if yc.NumDims() > 1 {
		sz := yc.Len() / yc.Dim(0)
		if txy.YIdx < sz && txy.YIdx >= 0 {
			off := row*sz + txy.YIdx
			y = yc.FloatVal1D(off)
		}
	} else {
		y = yc.FloatVal1D(row)
	}
	return
}

// Label returns a label for given row in table
func (txy *TableXY) Label(row int) string {
	if txy.Table == nil {
		return ""
	}
	row += txy.stRow
	return txy.Table.Cols[txy.LblCol].StringVal1D(row)
}
