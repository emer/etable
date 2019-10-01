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
	Table      *etable.Table `desc:"the data table to plot from"`
	XCol, YCol int           `desc:"the indexes of the tensor columns to use for the X and Y data, respectively"`
	XIdx, YIdx int           `desc:"the indexes of the element within each tensor cell if cells are n-dimensional, respectively"`
	LblCol     int           `desc:"the column to use for returning a label using Label interface -- for string cols"`
}

// NewTableXY returns a new XY plot view onto the given etable.Table, from given column indexes.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXY(dt *etable.Table, xcol, ycol int) (*TableXY, error) {
	txy := &TableXY{Table: dt, XCol: xcol, YCol: ycol}
	return txy, txy.Validate()
}

// NewTableXYNames returns a new XY plot view onto the given etable.Table, from given column names
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXYNames(dt *etable.Table, xcol, ycol string) (*TableXY, error) {
	xi, err := dt.ColIdxTry(xcol)
	if err != nil {
		log.Println(err)
	}
	yi, err := dt.ColIdxTry(ycol)
	if err != nil {
		log.Println(err)
	}
	txy := &TableXY{Table: dt, XCol: xi, YCol: yi}
	return txy, err
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
	// todo: validate Idx's
	return nil
}

// Len returns the number of rows in the table
func (txy *TableXY) Len() int {
	if txy.Table == nil || txy.Table.Rows < 0 {
		return 0
	}
	return txy.Table.NumRows()
}

// XY returns an x, y pair at given row in table
func (txy *TableXY) XY(row int) (x, y float64) {
	if txy.Table == nil {
		return 0, 0
	}
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
	return txy.Table.Cols[txy.LblCol].StringVal1D(row)
}
