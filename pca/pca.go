// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pca

import (
	"fmt"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
	"gonum.org/v1/gonum/mat"
)

// PCA computes the eigenvalue decomposition of a square similarity matrix,
// typically generated using the correlation metric.
type PCA struct {
	Covar   *etensor.Float64 `view:"no-inline" desc:"the covariance matrix computed on original data, which is then eigen-factored"`
	Vectors *etensor.Float64 `view:"no-inline" desc:"the eigenvectors, in same size as Covar - each eigenvector is a column in this 2D square matrix, ordered *lowest* to *highest* across the columns -- i.e., maximum eigenvector is the last column"`
	Values  []float64        `view:"no-inline" desc:"the eigenvalues, ordered *lowest* to *highest*"`
}

func (pca *PCA) Init() {
	pca.Covar = &etensor.Float64{}
	pca.Vectors = &etensor.Float64{}
	pca.Values = nil
}

// TableCol is a convenience method that computes a covariance matrix
// on given column of table and then performs the PCA on the resulting matrix.
// If no error occurs, the results can be read out from Vectors and Values
// or used in Projection methods.
// mfun is metric function, typically Covariance or Correlation -- use Covar
// if vars have similar overall scaling, which is typical in neural network models,
// and use Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix, which extracts the eigenvectors as directions with maximal
// variance in this matrix.
func (pca *PCA) TableCol(ix *etable.IdxView, colNm string, mfun metric.Func64) error {
	if pca.Covar == nil {
		pca.Init()
	}
	err := CovarTableCol(pca.Covar, ix, colNm, mfun)
	if err != nil {
		return err
	}
	return pca.PCA()
}

// Tensor is a convenience method that computes a covariance matrix
// on given tensor and then performs the PCA on the resulting matrix.
// If no error occurs, the results can be read out from Vectors and Values
// or used in Projection methods.
// mfun is metric function, typically Covariance or Correlation -- use Covar
// if vars have similar overall scaling, which is typical in neural network models,
// and use Correl if they are on very different scales -- Correl effectively rescales).
// A Covariance matrix computes the *row-wise* vector similarities for each
// pairwise combination of column cells -- i.e., the extent to which each
// cell co-varies in its value with each other cell across the rows of the table.
// This is the input to the PCA eigenvalue decomposition of the resulting
// covariance matrix, which extracts the eigenvectors as directions with maximal
// variance in this matrix.
func (pca *PCA) Tensor(tsr etensor.Tensor, mfun metric.Func64) error {
	if pca.Covar == nil {
		pca.Init()
	}
	err := CovarTensor(pca.Covar, tsr, mfun)
	if err != nil {
		return err
	}
	return pca.PCA()
}

// PCA performs the eigen decomposition of the existing Covar matrix.
// Vectors and Values fields contain the results.
func (pca *PCA) PCA() error {
	if pca.Covar == nil || pca.Covar.NumDims() != 2 {
		return fmt.Errorf("pca.PCA: Covar matrix is nil or not 2D")
	}
	var eig mat.EigenSym
	// note: MUST be a Float64 otherwise doesn't have Symmetric function
	ok := eig.Factorize(pca.Covar, true)
	if !ok {
		return fmt.Errorf("gonum EigenSym Factorize failed")
	}
	if pca.Vectors == nil {
		pca.Vectors = &etensor.Float64{}
	}
	ev := eig.VectorsTo(nil)
	etensor.CopyDense(pca.Vectors, ev)
	nr := pca.Vectors.Dim(0)
	if len(pca.Values) != nr {
		pca.Values = make([]float64, nr)
	}
	eig.Values(pca.Values)
	return nil
}

// ProjectCol projects values from the given colNm of given table (via IdxView)
// onto the idx'th eigenvector (0 = largest eigenvalue, 1 = next, etc).
// Must have already called PCA() method.
func (pca *PCA) ProjectCol(vals *[]float64, ix *etable.IdxView, colNm string, idx int) error {
	col, err := ix.Table.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	if pca.Vectors == nil {
		return fmt.Errorf("PCA.ProjectCol Vectors are nil -- must call PCA first")
	}
	nr := pca.Vectors.Dim(0)
	if idx >= nr {
		return fmt.Errorf("PCA.ProjectCol eigenvector index > rank of matrix")
	}
	cvec := make([]float64, nr)
	eidx := nr - 1 - idx // eigens in reverse order
	for ri := 0; ri < nr; ri++ {
		cvec[ri] = pca.Vectors.Value([]int{ri, eidx}) // vecs are in columns, reverse magnitude order
	}
	rows := ix.Len()
	if len(*vals) != rows {
		*vals = make([]float64, rows)
	}
	nd := col.NumDims()
	ln := col.Len()
	sz := ln / col.Dim(0) // size of cell
	if sz != nr {
		return fmt.Errorf("PCA.ProjectCol column cell size != pca eigenvectors")
	}
	rdim := []int{0}
	for row := 0; row < rows; row++ {
		sum := 0.0
		rdim[0] = ix.Idxs[row]
		rt := col.SubSpace(nd-1, rdim)
		for ci := 0; ci < sz; ci++ {
			sum += cvec[ci] * rt.FloatVal1D(ci)
		}
		(*vals)[row] = sum
	}
	return nil
}

// ProjectColToTable projects values from the given colNm of given table (via IdxView)
// onto the given set of eigenvectors (idxs, 0 = largest eigenvalue, 1 = next, etc),
// and stores results along with labels from column labNm into results table.
// Must have already called PCA() method.
func (pca *PCA) ProjectColToTable(prjns *etable.Table, ix *etable.IdxView, colNm, labNm string, idxs []int) error {
	_, err := ix.Table.ColByNameTry(colNm)
	if err != nil {
		return err
	}
	if pca.Vectors == nil {
		return fmt.Errorf("PCA.ProjectCol Vectors are nil -- must call PCA first")
	}
	rows := ix.Len()
	sch := etable.Schema{}
	pcolSt := 0
	if labNm != "" {
		sch = append(sch, etable.Column{labNm, etensor.STRING, nil, nil})
		pcolSt = 1
	}
	for _, idx := range idxs {
		sch = append(sch, etable.Column{fmt.Sprintf("Prjn%v", idx), etensor.FLOAT64, nil, nil})
	}
	prjns.SetFromSchema(sch, rows)

	for ii, idx := range idxs {
		pcol := prjns.Cols[pcolSt+ii].(*etensor.Float64)
		pca.ProjectCol(&pcol.Values, ix, colNm, idx)
	}

	if labNm != "" {
		lcol, err := ix.Table.ColByNameTry(labNm)
		if err == nil {
			plcol := prjns.Cols[0]
			for row := 0; row < rows; row++ {
				plcol.SetString1D(row, lcol.StringVal1D(row))
			}
		}
	}
	return nil
}
