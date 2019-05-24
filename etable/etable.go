// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import (
	"fmt"
	"math"

	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// etable.Table is the emer DataTable structure, containing columns of etensor tensors.
// All tensors MUST have RowMajor stride layout!
type Table struct {
	Cols       []etensor.Tensor  `view:"no-inline" desc:"columns of data, as etensor.Tensor tensors"`
	ColNames   []string          `desc:"the names of the columns"`
	Rows       int               `inactive:"+" desc:"number of rows, which is enforced to be the size of the outer-most dimension of the column tensors"`
	ColNameMap map[string]int    `view:"-" desc:"the map of column names to column numbers"`
	MetaData   map[string]string `desc:"misc meta data for the table.  Name is key for name of table. For Column-specific data, we look for ColName: prefix, specifically ColName:desc is a key used to provide a description of the column contents, which is shown as tooltip in the etview.View"`
}

var KiT_Table = kit.Types.AddType(&Table{}, TableProps)

// NumRows returns the number of rows (arrow / dframe api)
func (dt *Table) NumRows() int {
	return dt.Rows
}

// NumCols returns the number of columns (arrow / dframe api)
func (dt *Table) NumCols() int {
	return len(dt.Cols)
}

// Col returns the tensor at given column index
func (dt *Table) Col(i int) etensor.Tensor {
	return dt.Cols[i]
}

// ColNameIndex returns the index of the given column name, along with an error if not found
func (dt *Table) ColNameIndex(name string) (int, error) {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return 0, fmt.Errorf("etable.Table ColNameIndex: column named: %v not found", name)
	}
	return i, nil
}

// ColByName returns the tensor at given column name without any error messages -- just
// returns nil if not found
func (dt *Table) ColByName(name string) etensor.Tensor {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return nil
	}
	return dt.Cols[i]
}

// ColByNameTry returns the tensor at given column name, if not found, returns error
func (dt *Table) ColByNameTry(name string) (etensor.Tensor, error) {
	i, err := dt.ColNameIndex(name)
	if err != nil {
		return nil, err
	}
	return dt.Cols[i], nil
}

// ColName returns the name of given column
func (dt *Table) ColName(i int) string {
	return dt.ColNames[i]
}

// UpdateColNameMap updates the column name map
func (dt *Table) UpdateColNameMap() {
	nc := dt.NumCols()
	dt.ColNameMap = make(map[string]int, nc)
	for i := range dt.ColNames {
		dt.ColNameMap[dt.ColNames[i]] = i
	}
}

// AddCol adds the given tensor as a column to the table.
// returns error if it is not a RowMajor organized tensor, and automatically
// adjusts the shape to fit the current number of rows.
func (dt *Table) AddCol(tsr etensor.Tensor, name string) error {
	if !tsr.IsRowMajor() {
		return fmt.Errorf("tensor must be RowMajor organized")
	}
	dt.Cols = append(dt.Cols, tsr)
	dt.ColNames = append(dt.ColNames, name)
	dt.UpdateColNameMap()
	tsr.SetNumRows(dt.Rows)
	return nil
}

// AddRows adds n rows to each of the columns
func (dt *Table) AddRows(n int) {
	for _, tsr := range dt.Cols {
		tsr.AddRows(n)
	}
	dt.Rows += n
}

// SetNumRows sets the number of rows in the table, across all columns
// if rows = 0 then effective number of rows in tensors is 1, as this dim cannot be 0
func (dt *Table) SetNumRows(rows int) {
	dt.Rows = rows // can be 0
	rows = ints.MaxInt(1, rows)
	for _, tsr := range dt.Cols {
		tsr.SetNumRows(rows)
	}
}

// SetFromSchema configures table from given Schema.
// The actual tensor number of rows is enforced to be > 0, because we
// cannot have a null dimension in tensor shape.
// does not preserve any existing columns / data.
func (dt *Table) SetFromSchema(sc Schema, rows int) {
	nc := len(sc)
	dt.Cols = make([]etensor.Tensor, nc)
	dt.ColNames = make([]string, nc)
	dt.Rows = rows // can be 0
	rows = ints.MaxInt(1, rows)
	for i := range dt.Cols {
		cl := &sc[i]
		dt.ColNames[i] = cl.Name
		sh := append([]int{rows}, cl.CellShape...)
		dn := append([]string{"row"}, cl.DimNames...)
		tsr := etensor.New(cl.Type, sh, nil, dn)
		dt.Cols[i] = tsr
	}
	dt.UpdateColNameMap()
}

// New returns a new Table constructed from given Schema.
// The actual tensor number of rows is enforced to be > 0, because we
// cannot have a null dimension in tensor shape
func New(sc Schema, rows int) *Table {
	dt := &Table{}
	dt.SetFromSchema(sc, rows)
	return dt
}

//////////////////////////////////////////////////////////////////////////////////////
//  Cell convenience access methods

// CellFloat returns the float64 value of cell at given column, row index
// for columns that have 1-dimensional tensors.
// Returns NaN if column is not a 1-dimensional tensor.
func (dt *Table) CellFloat(col, row int) float64 {
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return math.NaN()
	}
	return ct.FloatVal1D(row)
}

// CellFloatByName returns the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns NaN if column is not a 1-dimensional tensor or col name not found.
func (dt *Table) CellFloatByName(colNm string, row int) float64 {
	ct := dt.ColByName(colNm)
	if ct == nil {
		return math.NaN()
	}
	if ct.NumDims() != 1 {
		return math.NaN()
	}
	return ct.FloatVal1D(row)
}

// CellFloatByNameTry returns the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) CellFloatByNameTry(colNm string, row int) (float64, error) {
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return math.NaN(), err
	}
	if ct.NumDims() != 1 {
		return math.NaN(), fmt.Errorf("etable.Table: CellFloatByNameTry called on column named: %v which is not 1-dimensional", colNm)
	}
	return ct.FloatVal1D(row), nil
}

// CellString returns the string value of cell at given column, row index
// for columns that have 1-dimensional tensors.
// Returns NaN if column is not a 1-dimensional tensor.
func (dt *Table) CellString(col, row int) string {
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return ""
	}
	return ct.StringVal1D(row)
}

// CellStringByName returns the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
func (dt *Table) CellStringByName(colNm string, row int) string {
	ct := dt.ColByName(colNm)
	if ct == nil {
		return ""
	}
	if ct.NumDims() != 1 {
		return ""
	}
	return ct.StringVal1D(row)
}

// CellStringByNameTry returns the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) CellStringByNameTry(colNm string, row int) (string, error) {
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return "", err
	}
	if ct.NumDims() != 1 {
		return "", fmt.Errorf("etable.Table: CellStringByNameTry called on column named: %v which is not 1-dimensional", colNm)
	}
	return ct.StringVal1D(row), nil
}

// CellTensor returns the tensor SubSpace for given column, row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns nil if column is a 1-dimensional
// tensor or there is any error from the etensor.Tensor.SubSpace call.
func (dt *Table) CellTensor(col, row int) etensor.Tensor {
	ct := dt.Cols[col]
	if ct.NumDims() == 1 {
		return nil
	}
	return ct.SubSpace(ct.NumDims()-1, []int{row})
}

// CellTensorByName returns the tensor SubSpace for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns nil on any error -- see Try version for
// error returns.
func (dt *Table) CellTensorByName(colNm string, row int) etensor.Tensor {
	ct := dt.ColByName(colNm)
	if ct == nil {
		return nil
	}
	if ct.NumDims() == 1 {
		return nil
	}
	return ct.SubSpace(ct.NumDims()-1, []int{row})
}

// CellTensorByNameTry returns the tensor SubSpace for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns an error if column is a 1-dimensional
// tensor or any error from the etensor.Tensor.SubSpace call.
func (dt *Table) CellTensorByNameTry(colNm string, row int) (etensor.Tensor, error) {
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	if ct.NumDims() == 1 {
		return nil, fmt.Errorf("etable.Table: CellTensorByNameTry called on column named: %v which is 1-dimensional", colNm)
	}
	return ct.SubSpaceTry(ct.NumDims()-1, []int{row})
}

// Schema returns the Schema (column properties) for this table
func (dt *Table) Schema() Schema {
	nc := dt.NumCols()
	sc := make(Schema, nc)
	for i := range dt.Cols {
		cl := &sc[i]
		tsr := dt.Cols[i]
		cl.Name = dt.ColNames[i]
		//cl.Type = etensor.Type(tsr.DataType().ID())
		cl.Type = etensor.Type(tsr.DataType())
		cl.CellShape = tsr.Shapes()[1:]
		cl.DimNames = tsr.DimNames()[1:]
	}
	return sc
}

//////////////////////////////////////////////////////////////////////////////////////
//  Table props for gui

var TableProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"OpenCSV", ki.Props{
			"label": "Open CSV File...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted data (or any delimeter -- default is tab (9), comma = 44) -- also recognizes emergent-style headers",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{}},
				{"Delimiter", ki.Props{
					"default": ',',
					"desc":    "can use any single-character rune here -- default is tab (9) b/c otherwise hard to type, comma = 44",
				}},
			},
		}},
		{"SaveCSV", ki.Props{
			"label": "Save CSV File...",
			"icon":  "file-save",
			"desc":  "Save CSV-formatted data (or any delimiter -- default is tab (9), comma = 44) -- header outputs emergent-style header data",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{}},
				{"Delimiter", ki.Props{
					"default": '\t',
					"desc":    "can use any single-character rune here -- default is tab (9) b/c otherwise hard to type, comma = 44",
				}},
				{"Headers", ki.Props{
					"desc": "output C++ emergent-style headers that have type and tensor geometry information",
				}},
			},
		}},
		{"sep-file", ki.BlankProp{}},
		{"AddRows", ki.Props{
			"icon": "new",
			"Args": ki.PropSlice{
				{"N Rows", ki.Props{
					"default": 1,
				}},
			},
		}},
		{"SetNumRows", ki.Props{
			"label": "Set N Rows",
			"icon":  "new",
			"Args": ki.PropSlice{
				{"N Rows", ki.Props{
					"default-field": "Rows",
				}},
			},
		}},
	},
}
