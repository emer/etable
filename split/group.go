// Copyright (c) 2019, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package split

import (
	"log"
	"slices"

	"goki.dev/etable/v2/etable"
)

// All returns a single "split" with all of the rows in given view
// useful for leveraging the aggregation management functions in splits
func All(ix *etable.IdxView) *etable.Splits {
	spl := &etable.Splits{}
	spl.Levels = []string{"All"}
	spl.New(ix.Table, []string{"All"}, ix.Idxs...)
	return spl
}

// GroupByIdx returns a new Splits set based on the groups of values
// across the given set of column indexes.
// Uses a stable sort on columns, so ordering of other dimensions is preserved.
func GroupByIdx(ix *etable.IdxView, colIdxs []int) *etable.Splits {
	nc := len(colIdxs)
	if nc == 0 || ix.Table == nil {
		return nil
	}
	if ix.Table.ColNames == nil {
		log.Println("split.GroupBy: Table does not have any column names -- will not work")
		return nil
	}
	spl := &etable.Splits{}
	spl.Levels = make([]string, nc)
	for i, ci := range colIdxs {
		spl.Levels[i] = ix.Table.ColNames[ci]
	}
	srt := ix.Clone()
	srt.SortStableCols(colIdxs, true) // important for consistency
	lstVals := make([]string, nc)
	curVals := make([]string, nc)
	var curIx *etable.IdxView
	for _, rw := range srt.Idxs {
		diff := false
		for i, ci := range colIdxs {
			cl := ix.Table.Cols[ci]
			cv := cl.StringVal1D(rw)
			curVals[i] = cv
			if cv != lstVals[i] {
				diff = true
			}
		}
		if diff || curIx == nil {
			curIx = spl.New(ix.Table, curVals, rw)
			copy(lstVals, curVals)
		} else {
			curIx.AddIndex(rw)
		}
	}
	return spl
}

// GroupBy returns a new Splits set based on the groups of values
// across the given set of column names (see Try for version with error)
// Uses a stable sort on columns, so ordering of other dimensions is preserved.
func GroupBy(ix *etable.IdxView, colNms []string) *etable.Splits {
	return GroupByIdx(ix, ix.Table.ColIdxsByNames(colNms))
}

// GroupByTry returns a new Splits set based on the groups of values
// across the given set of column names.  returns error for bad column names.
// Uses a stable sort on columns, so ordering of other dimensions is preserved.
func GroupByTry(ix *etable.IdxView, colNms []string) (*etable.Splits, error) {
	cidx, err := ix.Table.ColIdxsByNamesTry(colNms)
	if err != nil {
		return nil, err
	}
	return GroupByIdx(ix, cidx), nil
}

// GroupByFunc returns a new Splits set based on the given function
// which returns value(s) to group on for each row of the table.
// The function should always return the same number of values -- if
// it doesn't behavior is undefined.
// Uses a stable sort on columns, so ordering of other dimensions is preserved.
func GroupByFunc(ix *etable.IdxView, fun func(row int) []string) *etable.Splits {
	if ix.Table == nil {
		return nil
	}

	// save function values
	funvals := make(map[int][]string, ix.Len())
	nv := 0 // number of valeus
	for _, rw := range ix.Idxs {
		sv := fun(rw)
		if nv == 0 {
			nv = len(sv)
		}
		funvals[rw] = slices.Clone(sv)
	}

	srt := ix.Clone()
	srt.SortStable(func(et *etable.Table, i, j int) bool { // sort based on given function values
		fvi := funvals[i]
		fvj := funvals[j]
		for fi := 0; fi < nv; fi++ {
			if fvi[fi] < fvj[fi] {
				return true
			} else if fvi[fi] > fvj[fi] {
				return false
			}
		}
		return false
	})

	// now do our usual grouping operation
	spl := &etable.Splits{}
	lstVals := make([]string, nv)
	var curIx *etable.IdxView
	for _, rw := range srt.Idxs {
		curVals := funvals[rw]
		diff := (curIx == nil)
		if !diff {
			for fi := 0; fi < nv; fi++ {
				if lstVals[fi] != curVals[fi] {
					diff = true
					break
				}
			}
		}
		if diff {
			curIx = spl.New(ix.Table, curVals, rw)
			copy(lstVals, curVals)
		} else {
			curIx.AddIndex(rw)
		}
	}
	return spl
}
