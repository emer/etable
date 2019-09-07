// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simat

import (
	"github.com/emer/etable/etensor"
)

// SimMat is a similarity / distance matrix with additional row and column
// labels for display purposes.
type SimMat struct {
	Mat  etensor.Tensor `desc:"the similarity / distance matrix (typically an etensor.Float64)"`
	Rows []string       `desc:"labels for the rows -- blank rows trigger generation of grouping lines"`
	Cols []string       `desc:"labels for the columns -- blank columns trigger generation of grouping lines"`
}

// Init initializes SimMat with default Matrix and nil rows, cols
func (sm *SimMat) Init() {
	sm.Mat = &etensor.Float64{}
	sm.Rows = nil
	sm.Cols = nil
}
