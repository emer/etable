// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"sort"
	"strings"

	"github.com/emer/etable/v2/etensor"
)

// LessFunc is a function used for sort comparisons that returns
// true if Table row i is less than Table row j -- these are the
// raw row numbers, which have already been projected through
// indexes when used for sorting via Indexes.
type LessFunc func(et *Table, i, j int) bool

// FilterFunc is a function used for filtering that returns
// true if Table row should be included in the current filtered
// view of the table, and false if it should be removed.
type FilterFunc func(et *Table, row int) bool

// IndexView is an indexed wrapper around an etable.Table that provides a
// specific view onto the Table defined by the set of indexes.
// This provides an efficient way of sorting and filtering a table by only
// updating the indexes while doing nothing to the Table itself.
// To produce a table that has data actually organized according to the
// indexed order, call the NewTable method.
// IndexView views on a table can also be organized together as Splits
// of the table rows, e.g., by grouping values along a given column.
type IndexView struct { //gti:add

	// Table that we are an indexed view onto
	Table *Table

	// current indexes into Table
	Indexes []int

	// current Less function used in sorting
	lessFunc LessFunc `copier:"-" view:"-" xml:"-" json:"-"`
}

// NewIndexView returns a new IndexView based on given table, initialized with sequential idxes
func NewIndexView(et *Table) *IndexView {
	ix := &IndexView{}
	ix.SetTable(et)
	return ix
}

// SetTable sets as indexes into given table with sequential initial indexes
func (ix *IndexView) SetTable(et *Table) {
	ix.Table = et
	ix.Sequential()
}

// DeleteInvalid deletes all invalid indexes from the list.
// Call this if rows (could) have been deleted from table.
func (ix *IndexView) DeleteInvalid() {
	if ix.Table == nil || ix.Table.Rows <= 0 {
		ix.Indexes = nil
		return
	}
	ni := ix.Len()
	for i := ni - 1; i >= 0; i-- {
		if ix.Indexes[i] >= ix.Table.Rows {
			ix.Indexes = append(ix.Indexes[:i], ix.Indexes[i+1:]...)
		}
	}
}

// Sequential sets indexes to sequential row-wise indexes into table
func (ix *IndexView) Sequential() { //gti:add
	if ix.Table == nil || ix.Table.Rows <= 0 {
		ix.Indexes = nil
		return
	}
	ix.Indexes = make([]int, ix.Table.Rows)
	for i := range ix.Indexes {
		ix.Indexes[i] = i
	}
}

// Permuted sets indexes to a permuted order -- if indexes already exist
// then existing list of indexes is permuted, otherwise a new set of
// permuted indexes are generated
func (ix *IndexView) Permuted() {
	if ix.Table == nil || ix.Table.Rows <= 0 {
		ix.Indexes = nil
		return
	}
	if len(ix.Indexes) == 0 {
		ix.Indexes = rand.Perm(ix.Table.Rows)
	} else {
		rand.Shuffle(len(ix.Indexes), func(i, j int) {
			ix.Indexes[i], ix.Indexes[j] = ix.Indexes[j], ix.Indexes[i]
		})
	}
}

// AddIndex adds a new index to the list
func (ix *IndexView) AddIndex(idx int) {
	ix.Indexes = append(ix.Indexes, idx)
}

// Sort sorts the indexes into our Table using given Less function.
// The Less function operates directly on row numbers into the Table
// as these row numbers have already been projected through the indexes.
func (ix *IndexView) Sort(lessFunc func(et *Table, i, j int) bool) {
	ix.lessFunc = lessFunc
	sort.Sort(ix)
}

// SortIndexes sorts the indexes into our Table directly in
// numerical order, producing the native ordering, while preserving
// any filtering that might have occurred.
func (ix *IndexView) SortIndexes() {
	sort.Ints(ix.Indexes)
}

const (
	// Ascending specifies an ascending sort direction for etable Sort routines
	Ascending = true

	// Descending specifies a descending sort direction for etable Sort routines
	Descending = false
)

// SortColName sorts the indexes into our Table according to values in
// given column name, using either ascending or descending order.
// Only valid for 1-dimensional columns.
// Returns error if column name not found.
func (ix *IndexView) SortColName(colNm string, ascending bool) error { //gti:add
	ci, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		log.Println(err)
		return err
	}
	ix.SortCol(ci, ascending)
	return nil
}

// SortCol sorts the indexes into our Table according to values in
// given column index, using either ascending or descending order.
// Only valid for 1-dimensional columns.
func (ix *IndexView) SortCol(colIndex int, ascending bool) {
	cl := ix.Table.Cols[colIndex]
	if cl.DataType() == etensor.STRING {
		ix.Sort(func(et *Table, i, j int) bool {
			if ascending {
				return cl.StringValue1D(i) < cl.StringValue1D(j)
			} else {
				return cl.StringValue1D(i) > cl.StringValue1D(j)
			}
		})
	} else {
		ix.Sort(func(et *Table, i, j int) bool {
			if ascending {
				return cl.FloatValue1D(i) < cl.FloatValue1D(j)
			} else {
				return cl.FloatValue1D(i) > cl.FloatValue1D(j)
			}
		})
	}
}

// SortColNames sorts the indexes into our Table according to values in
// given column names, using either ascending or descending order.
// Only valid for 1-dimensional columns.
// Returns error if column name not found.
func (ix *IndexView) SortColNames(colNms []string, ascending bool) error {
	nc := len(colNms)
	if nc == 0 {
		return fmt.Errorf("etable.IndexView.SortColNames: no column names provided")
	}
	cis := make([]int, nc)
	for i, cn := range colNms {
		ci, err := ix.Table.ColIndexTry(cn)
		if err != nil {
			log.Println(err)
			return err
		}
		cis[i] = ci
	}
	ix.SortCols(cis, ascending)
	return nil
}

// SortCols sorts the indexes into our Table according to values in
// given list of column indexes, using either ascending or descending order for
// all of the columns.  Only valid for 1-dimensional columns.
func (ix *IndexView) SortCols(colIndexes []int, ascending bool) {
	ix.Sort(func(et *Table, i, j int) bool {
		for _, ci := range colIndexes {
			cl := ix.Table.Cols[ci]
			if cl.DataType() == etensor.STRING {
				if ascending {
					if cl.StringValue1D(i) < cl.StringValue1D(j) {
						return true
					} else if cl.StringValue1D(i) > cl.StringValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				} else {
					if cl.StringValue1D(i) > cl.StringValue1D(j) {
						return true
					} else if cl.StringValue1D(i) < cl.StringValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				}
			} else {
				if ascending {
					if cl.FloatValue1D(i) < cl.FloatValue1D(j) {
						return true
					} else if cl.FloatValue1D(i) > cl.FloatValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				} else {
					if cl.FloatValue1D(i) > cl.FloatValue1D(j) {
						return true
					} else if cl.FloatValue1D(i) < cl.FloatValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				}
			}
		}
		return false
	})
}

/////////////////////////////////////////////////////////////////////////
//  Stable sorts -- sometimes essential..

// SortStable stably sorts the indexes into our Table using given Less function.
// The Less function operates directly on row numbers into the Table
// as these row numbers have already been projected through the indexes.
// It is *essential* that it always returns false when the two are equal
// for the stable function to actually work.
func (ix *IndexView) SortStable(lessFunc func(et *Table, i, j int) bool) {
	ix.lessFunc = lessFunc
	sort.Stable(ix)
}

// SortStableColName sorts the indexes into our Table according to values in
// given column name, using either ascending or descending order.
// Only valid for 1-dimensional columns.
// Returns error if column name not found.
func (ix *IndexView) SortStableColName(colNm string, ascending bool) error {
	ci, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		log.Println(err)
		return err
	}
	ix.SortStableCol(ci, ascending)
	return nil
}

// SortStableCol sorts the indexes into our Table according to values in
// given column index, using either ascending or descending order.
// Only valid for 1-dimensional columns.
func (ix *IndexView) SortStableCol(colIndex int, ascending bool) {
	cl := ix.Table.Cols[colIndex]
	if cl.DataType() == etensor.STRING {
		ix.SortStable(func(et *Table, i, j int) bool {
			if ascending {
				return cl.StringValue1D(i) < cl.StringValue1D(j)
			} else {
				return cl.StringValue1D(i) > cl.StringValue1D(j)
			}
		})
	} else {
		ix.SortStable(func(et *Table, i, j int) bool {
			if ascending {
				return cl.FloatValue1D(i) < cl.FloatValue1D(j)
			} else {
				return cl.FloatValue1D(i) > cl.FloatValue1D(j)
			}
		})
	}
}

// SortStableColNames sorts the indexes into our Table according to values in
// given column names, using either ascending or descending order.
// Only valid for 1-dimensional columns.
// Returns error if column name not found.
func (ix *IndexView) SortStableColNames(colNms []string, ascending bool) error {
	nc := len(colNms)
	if nc == 0 {
		return fmt.Errorf("etable.IndexView.SortStableColNames: no column names provided")
	}
	cis := make([]int, nc)
	for i, cn := range colNms {
		ci, err := ix.Table.ColIndexTry(cn)
		if err != nil {
			log.Println(err)
			return err
		}
		cis[i] = ci
	}
	ix.SortStableCols(cis, ascending)
	return nil
}

// SortStableCols sorts the indexes into our Table according to values in
// given list of column indexes, using either ascending or descending order for
// all of the columns.  Only valid for 1-dimensional columns.
func (ix *IndexView) SortStableCols(colIndexes []int, ascending bool) {
	ix.SortStable(func(et *Table, i, j int) bool {
		for _, ci := range colIndexes {
			cl := ix.Table.Cols[ci]
			if cl.DataType() == etensor.STRING {
				if ascending {
					if cl.StringValue1D(i) < cl.StringValue1D(j) {
						return true
					} else if cl.StringValue1D(i) > cl.StringValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				} else {
					if cl.StringValue1D(i) > cl.StringValue1D(j) {
						return true
					} else if cl.StringValue1D(i) < cl.StringValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				}
			} else {
				if ascending {
					if cl.FloatValue1D(i) < cl.FloatValue1D(j) {
						return true
					} else if cl.FloatValue1D(i) > cl.FloatValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				} else {
					if cl.FloatValue1D(i) > cl.FloatValue1D(j) {
						return true
					} else if cl.FloatValue1D(i) < cl.FloatValue1D(j) {
						return false
					} // if equal, fallthrough to next col
				}
			}
		}
		return false
	})
}

// Filter filters the indexes into our Table using given Filter function.
// The Filter function operates directly on row numbers into the Table
// as these row numbers have already been projected through the indexes.
func (ix *IndexView) Filter(filterFunc func(et *Table, row int) bool) {
	sz := len(ix.Indexes)
	for i := sz - 1; i >= 0; i-- { // always go in reverse for filtering
		if !filterFunc(ix.Table, ix.Indexes[i]) { // delete
			ix.Indexes = append(ix.Indexes[:i], ix.Indexes[i+1:]...)
		}
	}
}

// FilterColName filters the indexes into our Table according to values in
// given column name, using string representation of column values.
// Includes rows with matching values unless exclude is set.
// If contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
// Only valid for 1-dimensional columns.
// Returns error if column name not found.
func (ix *IndexView) FilterColName(colNm string, str string, exclude, contains, ignoreCase bool) error { //gti:add
	ci, err := ix.Table.ColIndexTry(colNm)
	if err != nil {
		log.Println(err)
		return err
	}
	ix.FilterCol(ci, str, exclude, contains, ignoreCase)
	return nil
}

// FilterCol sorts the indexes into our Table according to values in
// given column index, using string representation of column values.
// Includes rows with matching values unless exclude is set.
// If contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
// Only valid for 1-dimensional columns.
func (ix *IndexView) FilterCol(colIndex int, str string, exclude, contains, ignoreCase bool) {
	col := ix.Table.Cols[colIndex]
	lowstr := strings.ToLower(str)
	ix.Filter(func(et *Table, row int) bool {
		val := col.StringValue1D(row)
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
		if exclude {
			return !has
		}
		return has
	})
}

// NewTable returns a new table with column data organized according to
// the indexes
func (ix *IndexView) NewTable() *Table {
	rows := len(ix.Indexes)
	sc := ix.Table.Schema()
	nt := New(sc, rows)
	if rows == 0 {
		return nt
	}
	for ci := range nt.Cols {
		scl := ix.Table.Cols[ci]
		tcl := nt.Cols[ci]
		_, csz := tcl.RowCellSize()
		for i, srw := range ix.Indexes {
			tcl.CopyCellsFrom(scl, i*csz, srw*csz, csz)
		}
	}
	return nt
}

// AggCol applies given aggregation function to each element in the given column, using float64
// conversions of the values.  init is the initial value for the agg variable.
// Operates independently over each cell on n-dimensional columns and returns the result as a slice
// of values per cell.
func (ix *IndexView) AggCol(colIndex int, ini float64, fun etensor.AggFunc) []float64 {
	cl := ix.Table.Cols[colIndex]
	_, csz := cl.RowCellSize()

	ag := make([]float64, csz)
	for i := range ag {
		ag[i] = ini
	}
	if csz == 1 {
		for _, srw := range ix.Indexes {
			val := cl.FloatValue1D(srw)
			if !cl.IsNull1D(srw) && !math.IsNaN(val) {
				ag[0] = fun(srw, val, ag[0])
			}
		}
	} else {
		for _, srw := range ix.Indexes {
			si := srw * csz
			for j := range ag {
				val := cl.FloatValue1D(si + j)
				if !cl.IsNull1D(si+j) && !math.IsNaN(val) {
					ag[j] = fun(si+j, val, ag[j])
				}
			}
		}
	}
	return ag
}

// Clone returns a copy of the current index view with its own index memory
func (ix *IndexView) Clone() *IndexView {
	nix := &IndexView{}
	nix.CopyFrom(ix)
	return nix
}

// CopyFrom copies from given other IndexView (we have our own unique copy of indexes)
func (ix *IndexView) CopyFrom(oix *IndexView) {
	ix.Table = oix.Table
	ix.Indexes = slices.Clone(oix.Indexes)
}

// AddRows adds n rows to end of underlying Table, and to the indexes in this view
func (ix *IndexView) AddRows(n int) { //gti:add
	stidx := ix.Table.Rows
	ix.Table.SetNumRows(stidx + n)
	for i := stidx; i < stidx+n; i++ {
		ix.Indexes = append(ix.Indexes, i)
	}
}

// InsertRows adds n rows to end of underlying Table, and to the indexes starting at
// given index in this view
func (ix *IndexView) InsertRows(at, n int) {
	stidx := ix.Table.Rows
	ix.Table.SetNumRows(stidx + n)
	nw := make([]int, n, n+len(ix.Indexes)-at)
	for i := 0; i < n; i++ {
		nw[i] = stidx + i
	}
	ix.Indexes = append(ix.Indexes[:at], append(nw, ix.Indexes[at:]...)...)
}

// DeleteRows deletes n rows of indexes starting at given index in the list of indexes
func (ix *IndexView) DeleteRows(at, n int) {
	ix.Indexes = append(ix.Indexes[:at], ix.Indexes[at+n:]...)
}

// RowsByStringIndex returns the list of *our indexes* whose row in the table has
// given string value in given column index (de-reference our indexes to get actual row).
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// Use named args for greater clarity.
func (ix *IndexView) RowsByStringIndex(colIndex int, str string, contains, ignoreCase bool) []int {
	dt := ix.Table
	col := dt.Cols[colIndex]
	lowstr := strings.ToLower(str)
	var idxs []int
	for idx, srw := range ix.Indexes {
		val := col.StringValue1D(srw)
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
			idxs = append(idxs, idx)
		}
	}
	return idxs
}

// RowsByString returns the list of *our indexes* whose row in the table has
// given string value in given column name (de-reference our indexes to get actual row).
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// returns nil if name invalid -- see also Try.
// Use named args for greater clarity.
func (ix *IndexView) RowsByString(colNm string, str string, contains, ignoreCase bool) []int {
	dt := ix.Table
	ci := dt.ColIndex(colNm)
	if ci < 0 {
		return nil
	}
	return ix.RowsByStringIndex(ci, str, contains, ignoreCase)
}

// RowsByStringTry returns the list of *our indexes* whose row in the table has
// given string value in given column name (de-reference our indexes to get actual row).
// if contains, only checks if row contains string; if ignoreCase, ignores case.
// returns error message for invalid column name.
// Use named args for greater clarity.
func (ix *IndexView) RowsByStringTry(colNm string, str string, contains, ignoreCase bool) ([]int, error) {
	dt := ix.Table
	ci, err := dt.ColIndexTry(colNm)
	if err != nil {
		return nil, err
	}
	return ix.RowsByStringIndex(ci, str, contains, ignoreCase), nil
}

// Len returns the length of the index list
func (ix *IndexView) Len() int {
	return len(ix.Indexes)
}

// Less calls the LessFunc for sorting
func (ix *IndexView) Less(i, j int) bool {
	return ix.lessFunc(ix.Table, ix.Indexes[i], ix.Indexes[j])
}

// Swap switches the indexes for i and j
func (ix *IndexView) Swap(i, j int) {
	ix.Indexes[i], ix.Indexes[j] = ix.Indexes[j], ix.Indexes[i]
}
