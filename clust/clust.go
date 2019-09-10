// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clust

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/emer/etable/etensor"
	"github.com/emer/etable/simat"
	"github.com/goki/ki/indent"
)

// Node is one node in the cluster
type Node struct {
	Idx     int     `desc:"index into original distance matrix -- only valid for for terminal leaves"`
	Dist    float64 `desc:"distance for this node -- how far apart were all the kids from each other when this node was created -- is 0 for leaf nodes"`
	ParDist float64 `desc:"total aggregate distance from parents -- the X axis offset at which our cluster starts"`
	Y       float64 `desc:"y-axis value for this node -- if a parent, it is the average of its kids Y's, otherwise it counts down"`
	Kids    []*Node `desc:"child nodes under this one"`
}

// IsLeaf returns true if node is a leaf of the tree with no kids
func (nn *Node) IsLeaf() bool {
	return len(nn.Kids) == 0
}

// Sprint prints to string
func (nn *Node) Sprint(smat *simat.SimMat, depth int) string {
	if nn.IsLeaf() {
		return smat.Rows[nn.Idx] + " "
	}
	sv := fmt.Sprintf("\n%v%v: ", indent.Tabs(depth), nn.Dist)
	for _, kn := range nn.Kids {
		sv += kn.Sprint(smat, depth+1)
	}
	return sv
}

// NewNode merges two nodes into a new node
func NewNode(na, nb *Node, dst float64) *Node {
	nn := &Node{Dist: dst}
	nn.Kids = []*Node{na, nb}
	return nn
}

// Glom implements basic agglomerative clustering, based on a raw similarity matrix as given.
// This calls GlomInit to initialize the root node with all of the leaves, and the calls
// GlomClust to do the iterative clustering process.  If you want to start with pre-defined
// initial clusters, then call GlomClust with a root node so-initialized.
// The smat.Mat matrix must be an etensor.Float64.
func Glom(smat *simat.SimMat, dfunc DistFunc) *Node {
	ntot := smat.Mat.Dim(0) // number of leaves
	root := GlomInit(ntot)
	return GlomClust(root, smat, dfunc)
}

// GlomInit returns a standard root node initialized with all of the leaves
func GlomInit(ntot int) *Node {
	root := &Node{}
	root.Kids = make([]*Node, ntot)
	for i := 0; i < ntot; i++ {
		root.Kids[i] = &Node{Idx: i}
	}
	return root
}

// GlomClust does the iterative agglomerative clustering, based on a raw similarity matrix as given,
// using a root node that has already been initialized with the starting clusters (all of the
// leaves by default, but could be anything if you want to start with predefined clusters).
// The smat.Mat matrix must be an etensor.Float64.
func GlomClust(root *Node, smat *simat.SimMat, dfunc DistFunc) *Node {
	ntot := smat.Mat.Dim(0) // number of leaves
	smatf := smat.Mat.(*etensor.Float64).Values
	for {
		var ma, mb []int
		mval := math.MaxFloat64
		for ai, ka := range root.Kids {
			for bi := 0; bi < ai; bi++ {
				kb := root.Kids[bi]
				dv := dfunc(ka, kb, ntot, smatf)
				if dv < mval {
					mval = dv
					ma = []int{ai}
					mb = []int{bi}
				} else if dv == mval { // do all ties at same time
					ma = append(ma, ai)
					mb = append(mb, bi)
				}
			}
		}
		ni := 0
		if len(ma) > 1 {
			ni = rand.Intn(len(ma))
		}
		na := ma[ni]
		nb := mb[ni]
		// fmt.Printf("merging nodes at dist: %v: %v and %v\nA: %v\nB: %v\n", mval, na, nb, root.Kids[na].Sprint(smat, 0), root.Kids[nb].Sprint(smat, 0))
		nn := NewNode(root.Kids[na], root.Kids[nb], mval)
		for i := len(root.Kids) - 1; i >= 0; i-- {
			if i == na || i == nb {
				root.Kids = append(root.Kids[:i], root.Kids[i+1:]...)
			}
		}
		root.Kids = append(root.Kids, nn)
		if len(root.Kids) == 1 {
			break
		}
	}
	return root
}
