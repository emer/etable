// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import (
	"fmt"
	"math"
	"strings"

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
	MetaData   map[string]string `desc:"misc meta data for the table.  We use lower-case key names following the struct tag convention:  name = name of table; desc = description; read-only = gui is read-only; precision = n for precision to write out floats in csv.  For Column-specific data, we look for ColName: prefix, specifically ColName:desc = description of the column contents, which is shown as tooltip in the etview.TableView"`
}

var KiT_Table = kit.Types.AddType(&Table{}, TableProps)

// NumRows returns the number of rows (arrow / dframe api)
func (dt *Table) NumRows() int {
	return dt.Rows
}

// IsValidRow returns true if the row is valid
func (dt *Table) IsValidRow(row int) bool {
	if row < 0 || row >= dt.Rows {
		return false
	}
	return true
}

// IsValidRowTry returns an error message if the row is not valid.
func (dt *Table) IsValidRowTry(row int) error {
	if row < 0 || row >= dt.Rows {
		return fmt.Errorf("etable.Table Row: %v is not valid for table with Rows: %v\n", row, dt.Rows)
	}
	return nil
}

// NumCols returns the number of columns (arrow / dframe api)
func (dt *Table) NumCols() int {
	return len(dt.Cols)
}

// Col returns the tensor at given column index
func (dt *Table) Col(i int) etensor.Tensor {
	return dt.Cols[i]
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
	i, err := dt.ColIdxTry(name)
	if err != nil {
		return nil, err
	}
	return dt.Cols[i], nil
}

// ColIdx returns the index of the given column name.
// returns -1 if name not found -- see Try version for error message.
func (dt *Table) ColIdx(name string) int {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return -1
	}
	return i
}

// ColIdxTry returns the index of the given column name,
// along with an error if not found.
func (dt *Table) ColIdxTry(name string) (int, error) {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return 0, fmt.Errorf("etable.Table ColIdx: column named: %v not found", name)
	}
	return i, nil
}

// ColIdxsByNames returns the indexes of the given column names.
// idxs have -1 if name not found -- see Try version for error message.
func (dt *Table) ColIdxsByNames(names []string) []int {
	nc := len(names)
	if nc == 0 {
		return nil
	}
	cidx := make([]int, nc)
	for i, cn := range names {
		cidx[i] = dt.ColIdx(cn)
	}
	return cidx
}

// ColsIdxsByNamesTry returns the indexes of the given column names,
// along with an error if any not found.
func (dt *Table) ColIdxsByNamesTry(names []string) ([]int, error) {
	nc := len(names)
	if nc == 0 {
		return nil, fmt.Errorf("etable.Table ColsByNamesIdxs: no column names provided")
	}
	cidx := make([]int, nc)
	var err error
	for i, cn := range names {
		cidx[i], err = dt.ColIdxTry(cn)
		if err != nil {
			return nil, err
		}
	}
	return cidx, nil
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
	dt.SetNumRows(dt.Rows + n)
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

func NewTable(name string) *Table {
	et := &Table{}
	et.SetMetaData("name", name)
	return et
}

// New returns a new Table constructed from given Schema.
// The actual tensor number of rows is enforced to be > 0, because we
// cannot have a null dimension in tensor shape
func New(sc Schema, rows int) *Table {
	dt := &Table{}
	dt.SetFromSchema(sc, rows)
	return dt
}

// Schema returns the Schema (column properties) for this table
func (dt *Table) Schema() Schema {
	nc := dt.NumCols()
	sc := make(Schema, nc)
	for i := range dt.Cols {
		cl := &sc[i]
		tsr := dt.Cols[i]
		cl.Name = dt.ColNames[i]
		cl.Type = tsr.DataType()
		cl.CellShape = tsr.Shapes()[1:]
		cl.DimNames = tsr.DimNames()[1:]
	}
	return sc
}

// note: no really clean definition of CopyFrom -- no point of re-using existing
// table -- just clone it.

// Clone returns a complete copy of this table
func (dt *Table) Clone() *Table {
	sc := dt.Schema()
	cp := New(sc, dt.Rows)
	for i, cl := range dt.Cols {
		ccl := cp.Cols[i]
		ccl.CopyFrom(cl)
	}
	cp.CopyMetaDataFrom(dt)
	return cp
}

// SetMetaData sets given meta-data key to given value, safely creating the
// map if not yet initialized.  Standard Keys are:
// * name -- name of table
// * desc -- description of table
// * read-only  -- makes gui read-only (inactive edits) for etview.TableView
// * ColName:* -- prefix for all column-specific meta-data
//     + desc -- description of column
func (dt *Table) SetMetaData(key, val string) {
	if dt.MetaData == nil {
		dt.MetaData = make(map[string]string)
	}
	dt.MetaData[key] = val
}

// CopyMetaDataFrom copies meta data from other table
func (dt *Table) CopyMetaDataFrom(cp *Table) {
	nm := len(cp.MetaData)
	if nm == 0 {
		return
	}
	if dt.MetaData == nil {
		dt.MetaData = make(map[string]string, nm)
	}
	for k, v := range cp.MetaData {
		dt.MetaData[k] = v
	}
}

// Named arg values for Contains, IgnoreCase
const (
	// Contains means the string only needs to contain the target string (see Equals)
	Contains bool = true
	// Equals means the string must equal the target string (see Contains)
	Equals = false
	// IgnoreCase means that differences in case are ignored in comparing strings
	IgnoreCase = true
	// UseCase means that case matters when comparing strings
	UseCase = false
)

// RowsByStringIdx returns the list of rows that have given
// string value in given column index.
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
func (dt *Table) RowsByStringIdx(colIdx int, str string, contains, ignoreCase bool) []int {
	col := dt.Cols[colIdx]
	lowstr := strings.ToLower(str)
	var idxs []int
	for i := 0; i < dt.Rows; i++ {
		val := col.StringVal1D(i)
		has := false
		switch {
		case contains && ignoreCase:
			has = strings.Contains(strings.ToLower(val), lowstr)
		case contains:
			has = strings.Contains(val, str)
		case ignoreCase:
			has = strings.EqualFold(val, str)
		default:
			has = (val == str)
		}
		if has {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// RowsByString returns the list of rows that have given
// string value in given column name.  returns nil if name invalid -- see also Try.
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
func (dt *Table) RowsByString(colNm string, str string, contains, ignoreCase bool) []int {
	ci := dt.ColIdx(colNm)
	if ci < 0 {
		return nil
	}
	return dt.RowsByStringIdx(ci, str, contains, ignoreCase)
}

// RowsByStringTry returns the list of rows that have given
// string value in given column name.  returns error message for invalid column name.
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
func (dt *Table) RowsByStringTry(colNm string, str string, contains, ignoreCase bool) ([]int, error) {
	ci, err := dt.ColIdxTry(colNm)
	if err != nil {
		return nil, err
	}
	return dt.RowsByStringIdx(ci, str, contains, ignoreCase), nil
}

//////////////////////////////////////////////////////////////////////////////////////
//  Cell convenience access methods

// CellFloatIdx returns the float64 value of cell at given column, row index
// for columns that have 1-dimensional tensors.
// Returns NaN if column is not a 1-dimensional tensor or row not valid.
func (dt *Table) CellFloatIdx(col, row int) float64 {
	if !dt.IsValidRow(row) {
		return math.NaN()
	}
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return math.NaN()
	}
	return ct.FloatVal1D(row)
}

// CellFloat returns the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns NaN if column is not a 1-dimensional tensor or col name not found, or row not valid.
func (dt *Table) CellFloat(colNm string, row int) float64 {
	if !dt.IsValidRow(row) {
		return math.NaN()
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return math.NaN()
	}
	if ct.NumDims() != 1 {
		return math.NaN()
	}
	return ct.FloatVal1D(row)
}

// CellFloatTry returns the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor, or row not valid.
func (dt *Table) CellFloatTry(colNm string, row int) (float64, error) {
	if err := dt.IsValidRowTry(row); err != nil {
		return math.NaN(), err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return math.NaN(), err
	}
	if ct.NumDims() != 1 {
		return math.NaN(), fmt.Errorf("etable.Table: CellFloatTry called on column named: %v which is not 1-dimensional", colNm)
	}
	return ct.FloatVal1D(row), nil
}

// CellStringIdx returns the string value of cell at given column, row index
// for columns that have 1-dimensional tensors.
// Returns "" if column is not a 1-dimensional tensor or row not valid.
func (dt *Table) CellStringIdx(col, row int) string {
	if !dt.IsValidRow(row) {
		return ""
	}
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return ""
	}
	return ct.StringVal1D(row)
}

// CellString returns the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns "" if column is not a 1-dimensional tensor or row not valid.
func (dt *Table) CellString(colNm string, row int) string {
	if !dt.IsValidRow(row) {
		return ""
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return ""
	}
	if ct.NumDims() != 1 {
		return ""
	}
	return ct.StringVal1D(row)
}

// CellStringTry returns the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor, or row not valid.
func (dt *Table) CellStringTry(colNm string, row int) (string, error) {
	if err := dt.IsValidRowTry(row); err != nil {
		return "", err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return "", err
	}
	if ct.NumDims() != 1 {
		return "", fmt.Errorf("etable.Table: CellStringTry called on column named: %v which is not 1-dimensional", colNm)
	}
	return ct.StringVal1D(row), nil
}

// CellTensorIdx returns the tensor SubSpace for given column, row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns nil if column is a 1-dimensional
// tensor or there is any error from the etensor.Tensor.SubSpace call.
func (dt *Table) CellTensorIdx(col, row int) etensor.Tensor {
	if !dt.IsValidRow(row) {
		return nil
	}
	ct := dt.Cols[col]
	if ct.NumDims() == 1 {
		return nil
	}
	return ct.SubSpace([]int{row})
}

// CellTensor returns the tensor SubSpace for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns nil on any error -- see Try version for
// error returns.
func (dt *Table) CellTensor(colNm string, row int) etensor.Tensor {
	if !dt.IsValidRow(row) {
		return nil
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return nil
	}
	if ct.NumDims() == 1 {
		return nil
	}
	return ct.SubSpace([]int{row})
}

// CellTensorTry returns the tensor SubSpace for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns an error if column is a 1-dimensional
// tensor or any error from the etensor.Tensor.SubSpace call.
func (dt *Table) CellTensorTry(colNm string, row int) (etensor.Tensor, error) {
	if err := dt.IsValidRowTry(row); err != nil {
		return nil, err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return nil, err
	}
	if ct.NumDims() == 1 {
		return nil, fmt.Errorf("etable.Table: CellTensorTry called on column named: %v which is 1-dimensional", colNm)
	}
	return ct.SubSpaceTry([]int{row})
}

// CellTensorFloat1D returns the float value of a Tensor cell's cell at given
// 1D offset within cell, for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns 0 on any error -- see Try version for
// error returns.
func (dt *Table) CellTensorFloat1D(colNm string, row int, idx int) float64 {
	if !dt.IsValidRow(row) {
		return 0
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return 0
	}
	if ct.NumDims() == 1 {
		return 0
	}
	_, sz := ct.RowCellSize()
	if idx >= sz || idx < 0 {
		return 0
	}
	off := row*sz + idx
	return ct.FloatVal1D(off)
}

// CellTensorFloat1DTry returns the float value of a Tensor cell's cell at given
// 1D offset within cell, for given column (by name), row index
// for columns that have higher-dimensional tensors so each row is
// represented by an n-1 dimensional tensor, with the outer dimension
// being the row number.  Returns any error.
func (dt *Table) CellTensorFloat1DTry(colNm string, row int, idx int) (float64, error) {
	if err := dt.IsValidRowTry(row); err != nil {
		return 0, err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return 0, err
	}
	if ct.NumDims() == 1 {
		return 0, fmt.Errorf("etable.Table: CellTensorFloat1DTry called on column named: %v which is 1-dimensional", colNm)
	}
	_, sz := ct.RowCellSize()
	if idx >= sz || idx < 0 {
		return 0, fmt.Errorf("etable.Table: CellTensorFloat1DTry index out of range for cell size")
	}
	off := row*sz + idx
	return ct.FloatVal1D(off), nil
}

/////////////////////////////////////////////////////////////////////////////////////
//  Set

// SetCellFloatIdx sets the float64 value of cell at given column, row index
// for columns that have 1-dimensional tensors.  Returns true if set.
func (dt *Table) SetCellFloatIdx(col, row int, val float64) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return false
	}
	ct.SetFloat1D(row, val)
	return true
}

// SetCellFloat sets the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
func (dt *Table) SetCellFloat(colNm string, row int, val float64) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return false
	}
	if ct.NumDims() != 1 {
		return false
	}
	ct.SetFloat1D(row, val)
	return true
}

// SetCellFloatTry sets the float64 value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) SetCellFloatTry(colNm string, row int, val float64) error {
	if err := dt.IsValidRowTry(row); err != nil {
		return err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	if ct.NumDims() != 1 {
		return fmt.Errorf("etable.Table: SetCellFloatTry called on column named: %v which is not 1-dimensional", colNm)
	}
	ct.SetFloat1D(row, val)
	return nil
}

// SetCellStringIdx sets the string value of cell at given column, row index
// for columns that have 1-dimensional tensors.  Returns true if set.
func (dt *Table) SetCellStringIdx(col, row int, val string) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.Cols[col]
	if ct.NumDims() != 1 {
		return false
	}
	ct.SetString1D(row, val)
	return true
}

// SetCellString sets the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.  Returns true if set.
func (dt *Table) SetCellString(colNm string, row int, val string) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return false
	}
	if ct.NumDims() != 1 {
		return false
	}
	ct.SetString1D(row, val)
	return true
}

// SetCellStringTry sets the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) SetCellStringTry(colNm string, row int, val string) error {
	if err := dt.IsValidRowTry(row); err != nil {
		return err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	if ct.NumDims() != 1 {
		return fmt.Errorf("etable.Table: SetCellStringTry called on column named: %v which is not 1-dimensional", colNm)
	}
	ct.SetString1D(row, val)
	return nil
}

// SetCellTensorIdx sets the tensor value of cell at given column, row index
// for columns that have n-dimensional tensors.  Returns true if set.
func (dt *Table) SetCellTensorIdx(col, row int, val etensor.Tensor) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.Cols[col]
	_, csz := ct.RowCellSize()
	st := row * csz
	sz := ints.MinInt(csz, val.Len())
	if ct.DataType() == etensor.STRING {
		for j := 0; j < sz; j++ {
			ct.SetString1D(st+j, val.StringVal1D(j))
		}
	} else {
		for j := 0; j < sz; j++ {
			ct.SetFloat1D(st+j, val.FloatVal1D(j))
		}
	}
	return true
}

// SetCellTensor sets the tensor value of cell at given column (by name), row index
// for columns that have n-dimensional tensors.  Returns true if set.
func (dt *Table) SetCellTensor(colNm string, row int, val etensor.Tensor) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ci := dt.ColIdx(colNm)
	if ci < 0 {
		return false
	}
	return dt.SetCellTensorIdx(ci, row, val)
}

// SetCellTensorTry sets the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) SetCellTensorTry(colNm string, row int, val etensor.Tensor) error {
	if err := dt.IsValidRowTry(row); err != nil {
		return err
	}
	ci, err := dt.ColIdxTry(colNm)
	if err != nil {
		return err
	}
	dt.SetCellTensorIdx(ci, row, val)
	return nil
}

// SetCellTensorFloat1D sets the tensor cell's float cell value at given 1D index within cell,
// at given column (by name), row index for columns that have n-dimensional tensors.
// Returns true if set.
func (dt *Table) SetCellTensorFloat1D(colNm string, row int, idx int, val float64) bool {
	if !dt.IsValidRow(row) {
		return false
	}
	ct := dt.ColByName(colNm)
	if ct == nil {
		return false
	}
	_, sz := ct.RowCellSize()
	if idx >= sz || idx < 0 {
		return false
	}
	off := row*sz + idx
	ct.SetFloat1D(off, val)
	return true
}

// SetCellTensorFloat1DTry sets the string value of cell at given column (by name), row index
// for columns that have 1-dimensional tensors.
// Returns an error if column not found, or column is not a 1-dimensional tensor.
func (dt *Table) SetCellTensorFloat1DTry(colNm string, row int, idx int, val float64) error {
	if err := dt.IsValidRowTry(row); err != nil {
		return err
	}
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	_, sz := ct.RowCellSize()
	if idx >= sz || idx < 0 {
		return fmt.Errorf("etable.Table: SetCellTensorFloat1DTry index out of range for cell size")
	}
	off := row*sz + idx
	ct.SetFloat1D(off, val)
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////
//  Copy Cell

// CopyCell copies into cell at given col, row from cell in other table.
// It is robust to differences in type -- uses destination cell type.
// Returns error if column names are invalid.
func (dt *Table) CopyCell(colNm string, row int, cpt *Table, cpColNm string, cpRow int) error {
	ct, err := dt.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	cpct, err := cpt.ColByNameTry(cpColNm)
	if err != nil {
		return err
	}
	_, sz := ct.RowCellSize()
	if sz == 1 {
		if ct.DataType() == etensor.STRING {
			ct.SetString1D(row, cpct.StringVal1D(cpRow))
		} else {
			ct.SetFloat1D(row, cpct.FloatVal1D(cpRow))
		}
	} else {
		_, cpsz := cpct.RowCellSize()
		st := row * sz
		cst := cpRow * cpsz
		msz := ints.MinInt(sz, cpsz)
		if ct.DataType() == etensor.STRING {
			for j := 0; j < msz; j++ {
				ct.SetString1D(st+j, cpct.StringVal1D(cst+j))
			}
		} else {
			for j := 0; j < msz; j++ {
				ct.SetFloat1D(st+j, cpct.FloatVal1D(cst+j))
			}
		}
	}
	return nil
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
					"default": '\t',
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
					"default": true,
					"desc":    "output C++ emergent-style headers that have type and tensor geometry information",
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
