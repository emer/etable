// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simat

import (
	"fmt"

	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
)

// Tensor computes a similarity / distance matrix on two tensors
// using given metric function.   Outer-most dimension ("rows") is
// used as "indexical" dimension and all other dimensions within that
// are compared.  Resulting reduced 2D shape of two tensors must be
// the same (returns error if not).
// Results go in smat which is ensured to have proper square 2D shape
// (rows * rows).
func Tensor(smat etensor.Tensor, a, b etensor.Tensor, mfun metric.Func64) error {
	if a.Dim(0) != b.Dim(0) {
		return fmt.Errorf("simat.Tensor: number of rows must be same")
	}
	nrows := a.Dim(0)
	and := a.NumDims()
	orgShpA := make([]int, and)
	copy(orgShpA, a.Shapes())
	aln := 1
	for i, d := range orgShpA {
		if i > 0 {
			aln *= d
		}
	}
	bnd := b.NumDims()
	orgShpB := make([]int, bnd)
	copy(orgShpB, b.Shapes())
	bln := 1
	for i, d := range orgShpB {
		if i > 0 {
			bln *= d
		}
	}
	if aln != bln {
		return fmt.Errorf("simat.Tensor: size of inner dimensions must be same")
	}

	sshp := []int{nrows, nrows}
	smat.SetShape(sshp, nil, nil)

	av := make([]float64, aln)
	bv := make([]float64, bln)
	ardim := []int{0}
	brdim := []int{0}
	sdim := []int{0, 0}
	for ai := 0; ai < nrows; ai++ {
		ardim[0] = ai
		sdim[0] = ai
		ar := a.SubSpace(and-1, ardim)
		ar.Floats(&av)
		for bi := 0; bi <= ai; bi++ { // lower diag
			brdim[0] = bi
			sdim[1] = bi
			br := b.SubSpace(bnd-1, brdim)
			br.Floats(&bv)
			sv := mfun(av, bv)
			smat.SetFloat(sdim, sv)
		}
	}
	// now fill in upper diagonal with values from lower diagonal
	// note: assumes symmetric distance function
	fdim := []int{0, 0}
	for ai := 0; ai < nrows; ai++ {
		sdim[0] = ai
		fdim[1] = ai
		for bi := ai + 1; bi < nrows; bi++ { // upper diag
			fdim[0] = bi
			sdim[1] = bi
			sv := smat.FloatVal(fdim)
			smat.SetFloat(sdim, sv)
		}
	}
	return nil
}
