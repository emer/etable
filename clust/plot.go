// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clust

import (
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/simat"
)

// Plot sets the rows of given data table to trace out lines with labels that
// will render cluster plot starting at root node when plotted with a standard plotting package.
// The lines double-back on themselves to form a continuous line to be plotted.
func Plot(pt *etable.Table, root *Node, smat *simat.SimMat) {
	sch := etable.Schema{
		{"X", etensor.FLOAT64, nil, nil},
		{"Y", etensor.FLOAT64, nil, nil},
		{"Label", etensor.STRING, nil, nil},
	}
	pt.SetFromSchema(sch, 0)
	nextY := 0.5
	root.SetYs(&nextY)
	root.SetParDist(0.0)
	root.Plot(pt, smat)
}

// Plot sets the rows of given data table to trace out lines with labels that
// will render this node in a cluster plot when plotted with a standard plotting package.
// The lines double-back on themselves to form a continuous line to be plotted.
func (nn *Node) Plot(pt *etable.Table, smat *simat.SimMat) {
	row := pt.Rows
	if nn.IsLeaf() {
		pt.SetNumRows(row + 1)
		pt.SetCellFloatIndex(0, row, nn.ParDist)
		pt.SetCellFloatIndex(1, row, nn.Y)
		if len(smat.Rows) > nn.Index {
			pt.SetCellStringIndex(2, row, smat.Rows[nn.Index])
		}
	} else {
		for _, kn := range nn.Kids {
			pt.SetNumRows(row + 2)
			pt.SetCellFloatIndex(0, row, nn.ParDist)
			pt.SetCellFloatIndex(1, row, kn.Y)
			row++
			pt.SetCellFloatIndex(0, row, nn.ParDist+nn.Dist)
			pt.SetCellFloatIndex(1, row, kn.Y)
			kn.Plot(pt, smat)
			row = pt.Rows
			pt.SetNumRows(row + 1)
			pt.SetCellFloatIndex(0, row, nn.ParDist)
			pt.SetCellFloatIndex(1, row, kn.Y)
			row++
		}
		pt.SetNumRows(row + 1)
		pt.SetCellFloatIndex(0, row, nn.ParDist)
		pt.SetCellFloatIndex(1, row, nn.Y)
	}
}

// SetYs sets the Y-axis values for the nodes in preparation for plotting.
func (nn *Node) SetYs(nextY *float64) {
	if nn.IsLeaf() {
		nn.Y = *nextY
		(*nextY) += 1.0
	} else {
		avgy := 0.0
		for _, kn := range nn.Kids {
			kn.SetYs(nextY)
			avgy += kn.Y
		}
		avgy /= float64(len(nn.Kids))
		nn.Y = avgy
	}
}

// SetParDist sets the parent distance for the nodes in preparation for plotting.
func (nn *Node) SetParDist(pard float64) {
	nn.ParDist = pard
	if !nn.IsLeaf() {
		pard += nn.Dist
		for _, kn := range nn.Kids {
			kn.SetParDist(pard)
		}
	}
}
