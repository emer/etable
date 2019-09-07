// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simat

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/metric"
)

// TableCol generates a similarity / distance matrix from given column name
// in given etable.Table, and given metric function.
// if labNm is not empty, uses given column name for labels, which are
// automatically filtered so that any sequentially repeated labels are blank
func TableCol(smat *SimMat, et *etable.Table, colNm, labNm string, mfun metric.Func64) error {
	dc, err := et.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	smat.Init()
	err = Tensor(smat.Mat, dc, mfun)
	if err != nil || labNm == "" {
		return err
	}
	lc, err := et.ColByNameTry(labNm)
	if err != nil {
		return err
	}
	rows := lc.Dim(0)
	smat.Rows = make([]string, rows)
	last := ""
	for r := 0; r < rows; r++ {
		lbl := lc.StringVal1D(r)
		if lbl == last {
			continue
		}
		smat.Rows[r] = lbl
		last = lbl
	}
	smat.Cols = smat.Rows // identical
	return nil
}
