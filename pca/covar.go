// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pca

import (
	"fmt"

	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/metric"
)

// CovarTableCol generates a covariance matrix from given column name
// in given IdxView of an etable.Table, and given metric function
// (typically Covariance or Correlation -- use Covar if vars have similar
// overall scaling, which is typical in neural network models, and use
// Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix.
func CovarTableCol(cmat etensor.Tensor, ix *etable.IdxView, colNm string, mfun metric.Func64) error {
	col, err := ix.Table.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	rows := ix.Len()
	nd := col.NumDims()
	if nd < 2 || rows == 0 {
		return fmt.Errorf("pca.CovarTableCol: must have 2 or more dims and rows != 0")
	}
	ln := col.Len()
	sz := ln / col.Dim(0) // size of cell

	cshp := []int{sz, sz}
	cmat.SetShape(cshp, nil, nil)

	av := make([]float64, rows)
	bv := make([]float64, rows)
	sdim := []int{0, 0}
	for ai := 0; ai < sz; ai++ {
		sdim[0] = ai
		TableColRowsVec(av, ix, col, ai)
		for bi := 0; bi <= ai; bi++ { // lower diag
			sdim[1] = bi
			TableColRowsVec(bv, ix, col, bi)
			cv := mfun(av, bv)
			cmat.SetFloat(sdim, cv)
		}
	}
	// now fill in upper diagonal with values from lower diagonal
	// note: assumes symmetric distance function
	fdim := []int{0, 0}
	for ai := 0; ai < sz; ai++ {
		sdim[0] = ai
		fdim[1] = ai
		for bi := ai + 1; bi < sz; bi++ { // upper diag
			fdim[0] = bi
			sdim[1] = bi
			cv := cmat.FloatVal(fdim)
			cmat.SetFloat(sdim, cv)
		}
	}

	if nm, has := ix.Table.MetaData["name"]; has {
		cmat.SetMetaData("name", nm+"_"+colNm)
	} else {
		cmat.SetMetaData("name", colNm)
	}
	if ds, has := ix.Table.MetaData["desc"]; has {
		cmat.SetMetaData("desc", ds)
	}
	return nil
}

// CovarTensor generates a covariance matrix from given etensor.Tensor,
// where the outer-most dimension is rows, and all other dimensions within that
// are covaried against each other, using given metric function
// (typically Covariance or Correlation -- use Covar if vars have similar
// overall scaling, which is typical in neural network models, and use
// Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix.
func CovarTensor(cmat etensor.Tensor, tsr etensor.Tensor, mfun metric.Func64) error {
	rows := tsr.Dim(0)
	nd := tsr.NumDims()
	if nd < 2 || rows == 0 {
		return fmt.Errorf("pca.CovarTensor: must have 2 or more dims and rows != 0")
	}
	ln := tsr.Len()
	sz := ln / rows

	cshp := []int{sz, sz}
	cmat.SetShape(cshp, nil, nil)

	av := make([]float64, rows)
	bv := make([]float64, rows)
	sdim := []int{0, 0}
	for ai := 0; ai < sz; ai++ {
		sdim[0] = ai
		TensorRowsVec(av, tsr, ai)
		for bi := 0; bi <= ai; bi++ { // lower diag
			sdim[1] = bi
			TensorRowsVec(bv, tsr, bi)
			cv := mfun(av, bv)
			cmat.SetFloat(sdim, cv)
		}
	}
	// now fill in upper diagonal with values from lower diagonal
	// note: assumes symmetric distance function
	fdim := []int{0, 0}
	for ai := 0; ai < sz; ai++ {
		sdim[0] = ai
		fdim[1] = ai
		for bi := ai + 1; bi < sz; bi++ { // upper diag
			fdim[0] = bi
			sdim[1] = bi
			cv := cmat.FloatVal(fdim)
			cmat.SetFloat(sdim, cv)
		}
	}

	if nm, has := tsr.MetaData("name"); has {
		cmat.SetMetaData("name", nm+"Covar")
	} else {
		cmat.SetMetaData("name", "Covar")
	}
	if ds, has := tsr.MetaData("desc"); has {
		cmat.SetMetaData("desc", ds)
	}
	return nil
}

// TableColRowsVec extracts row-wise vector from given cell index into vec.
// vec must be of size ix.Len() -- number of rows
func TableColRowsVec(vec []float64, ix *etable.IdxView, col etensor.Tensor, cidx int) {
	rows := ix.Len()
	ln := col.Len()
	sz := ln / col.Dim(0) // size of cell
	for ri := 0; ri < rows; ri++ {
		coff := ix.Idxs[ri]*sz + cidx
		vec[ri] = col.FloatVal1D(coff)
	}
}

// TensorRowsVec extracts row-wise vector from given cell index into vec.
// vec must be of size tsr.Dim(0) -- number of rows
func TensorRowsVec(vec []float64, tsr etensor.Tensor, cidx int) {
	rows := tsr.Dim(0)
	ln := tsr.Len()
	sz := ln / rows
	for ri := 0; ri < rows; ri++ {
		coff := ri*sz + cidx
		vec[ri] = tsr.FloatVal1D(coff)
	}
}

// CovarTableColStd generates a covariance matrix from given column name
// in given IdxView of an etable.Table, and given metric function
// (typically Covariance or Correlation -- use Covar if vars have similar
// overall scaling, which is typical in neural network models, and use
// Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix.
// This Std version is usable e.g., in Python where the func cannot be passed.
func CovarTableColStd(cmat etensor.Tensor, ix *etable.IdxView, colNm string, met metric.StdMetrics) error {
	return CovarTableCol(cmat, ix, colNm, metric.StdFunc64(met))
}

// CovarTensorStd generates a covariance matrix from given etensor.Tensor,
// where the outer-most dimension is rows, and all other dimensions within that
// are covaried against each other, using given metric function
// (typically Covariance or Correlation -- use Covar if vars have similar
// overall scaling, which is typical in neural network models, and use
// Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix.
// This Std version is usable e.g., in Python where the func cannot be passed.
func CovarTensorStd(cmat etensor.Tensor, tsr etensor.Tensor, met metric.StdMetrics) error {
	return CovarTensor(cmat, tsr, metric.StdFunc64(met))
}
