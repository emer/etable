// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package split

import (
	"github.com/emer/etable/etable"
	"github.com/goki/ki/sliceclone"
)

// GroupByIdx returns a new Splits set based on the groups of values
// across the given set of column indexes
func GroupByIdx(ix *etable.IdxView, colIdxs []int) *etable.Splits {
	nc := len(colIdxs)
	if nc == 0 || ix.Table == nil {
		return nil
	}
	spl := &etable.Splits{}
	spl.Levels = make([]string, nc)
	for ci := range colIdxs {
		spl.Levels[ci] = ix.Table.ColNames[ci]
	}
	srt := ix.Clone()
	srt.SortCols(colIdxs, true)
	lstVals := make([]string, nc)
	curVals := make([]string, nc)
	var curIx *etable.IdxView
	for _, rw := range srt.Idxs {
		diff := false
		for ci := range colIdxs {
			cl := ix.Table.Cols[ci]
			cv := cl.StringVal1D(rw)
			curVals[ci] = cv
			if cv != lstVals[ci] {
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
func GroupBy(ix *etable.IdxView, colNms []string) *etable.Splits {
	return GroupByIdx(ix, ix.Table.ColIdxsByNames(colNms))
}

// GroupByTry returns a new Splits set based on the groups of values
// across the given set of column names.  returns error for bad column names.
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
		funvals[rw] = sliceclone.String(sv)
	}

	srt := ix.Clone()
	srt.Sort(func(et *etable.Table, i, j int) bool { // sort based on given function values
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