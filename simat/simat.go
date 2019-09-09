// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simat

import (
	"fmt"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
)

// SimMat is a similarity / distance matrix with additional row and column
// labels for display purposes.
type SimMat struct {
	Mat  etensor.Tensor `desc:"the similarity / distance matrix (typically an etensor.Float64)"`
	Rows []string       `desc:"labels for the rows -- blank rows trigger generation of grouping lines"`
	Cols []string       `desc:"labels for the columns -- blank columns trigger generation of grouping lines"`
}

// Init initializes SimMat with default Matrix and nil rows, cols
func (smat *SimMat) Init() {
	smat.Mat = &etensor.Float64{}
	smat.Rows = nil
	smat.Cols = nil
}

// TableCol generates a similarity / distance matrix from given column name
// in given IdxView of an etable.Table, and given metric function.
// if labNm is not empty, uses given column name for labels, which are
// automatically filtered so that any sequentially repeated labels are blank
func (smat *SimMat) TableCol(ix *etable.IdxView, colNm, labNm string, mfun metric.Func64) error {
	col, err := ix.Table.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	smat.Init()
	sm := smat.Mat

	rows := ix.Len()
	nd := col.NumDims()
	if nd < 2 || rows == 0 {
		return fmt.Errorf("simat.Tensor: must have 2 or more dims and rows != 0")
	}
	ln := col.Len()
	sz := ln / col.Dim(0) // size of cell

	sshp := []int{rows, rows}
	sm.SetShape(sshp, nil, nil)

	av := make([]float64, sz)
	bv := make([]float64, sz)
	ardim := []int{0}
	brdim := []int{0}
	sdim := []int{0, 0}
	for ai := 0; ai < rows; ai++ {
		ardim[0] = ix.Idxs[ai]
		sdim[0] = ai
		ar := col.SubSpace(nd-1, ardim)
		ar.Floats(&av)
		for bi := 0; bi <= ai; bi++ { // lower diag
			brdim[0] = ix.Idxs[bi]
			sdim[1] = bi
			br := col.SubSpace(nd-1, brdim)
			br.Floats(&bv)
			sv := mfun(av, bv)
			sm.SetFloat(sdim, sv)
		}
	}
	// now fill in upper diagonal with values from lower diagonal
	// note: assumes symmetric distance function
	fdim := []int{0, 0}
	for ai := 0; ai < rows; ai++ {
		sdim[0] = ai
		fdim[1] = ai
		for bi := ai + 1; bi < rows; bi++ { // upper diag
			fdim[0] = bi
			sdim[1] = bi
			sv := sm.FloatVal(fdim)
			sm.SetFloat(sdim, sv)
		}
	}

	if nm, has := ix.Table.MetaData["name"]; has {
		sm.SetMetaData("name", nm+"_"+colNm)
	} else {
		sm.SetMetaData("name", colNm)
	}
	if ds, has := ix.Table.MetaData["desc"]; has {
		sm.SetMetaData("desc", ds)
	}

	if labNm == "" {
		return nil
	}
	lc, err := ix.Table.ColByNameTry(labNm)
	if err != nil {
		return err
	}
	smat.Rows = make([]string, rows)
	last := ""
	for r := 0; r < rows; r++ {
		lbl := lc.StringVal1D(ix.Idxs[r])
		if lbl == last {
			continue
		}
		smat.Rows[r] = lbl
		last = lbl
	}
	smat.Cols = smat.Rows // identical
	return nil
}
