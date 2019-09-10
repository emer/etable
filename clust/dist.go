// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clust

import "math"

// DistFunc is a clustering distance function that evaluates aggregate distance
// between nodes
type DistFunc func(na, nb *Node, tot int, smat []float64) float64

// MinDist is the minimum-distance or single-linkage weighting function for comparing
// two nodes.  ntot is total number of nodes, and smat is the square similarity matrix [ntot x ntot]
func MinDist(na, nb *Node, ntot int, smat []float64) float64 {
	md := math.MaxFloat64
	if na.IsLeaf() {
		if nb.IsLeaf() {
			return smat[na.Idx*ntot+nb.Idx]
		}
		for _, kb := range nb.Kids {
			kd := MinDist(na, kb, ntot, smat)
			if kd < md {
				md = kd
			}
		}
	} else if nb.IsLeaf() {
		for _, ka := range na.Kids {
			kd := MinDist(nb, ka, ntot, smat)
			if kd < md {
				md = kd
			}
		}
	} else { // all pairwise :(
		for _, ka := range na.Kids {
			for _, kb := range nb.Kids {
				kd := MinDist(ka, kb, ntot, smat)
				if kd < md {
					md = kd
				}
			}
		}
	}
	return md
}

// MaxDist is the maximum-distance or complete-linkage weighting function for comparing
// two nodes.  ntot is total number of nodes, and smat is the square similarity matrix [ntot x ntot]
func MaxDist(na, nb *Node, ntot int, smat []float64) float64 {
	md := -math.MaxFloat64
	if na.IsLeaf() {
		if nb.IsLeaf() {
			return smat[na.Idx*ntot+nb.Idx]
		}
		for _, kb := range nb.Kids {
			kd := MinDist(na, kb, ntot, smat)
			if kd > md {
				md = kd
			}
		}
	} else if nb.IsLeaf() {
		for _, ka := range na.Kids {
			kd := MinDist(nb, ka, ntot, smat)
			if kd > md {
				md = kd
			}
		}
	} else { // all pairwise :(
		for _, ka := range na.Kids {
			for _, kb := range nb.Kids {
				kd := MinDist(ka, kb, ntot, smat)
				if kd > md {
					md = kd
				}
			}
		}
	}
	return md
}

// AvgDist is the average-distance or average-linkage weighting function for comparing
// two nodes.  ntot is total number of nodes, and smat is the square similarity matrix [ntot x ntot]
func AvgDist(na, nb *Node, ntot int, smat []float64) float64 {
	md := 0.0
	n := 0
	if na.IsLeaf() {
		if nb.IsLeaf() {
			return smat[na.Idx*ntot+nb.Idx]
		}
		for _, kb := range nb.Kids {
			kd := MinDist(na, kb, ntot, smat)
			md += kd
			n++
		}
	} else if nb.IsLeaf() {
		for _, ka := range na.Kids {
			kd := MinDist(nb, ka, ntot, smat)
			md += kd
			n++
		}
	} else { // all pairwise :(
		for _, ka := range na.Kids {
			for _, kb := range nb.Kids {
				kd := MinDist(ka, kb, ntot, smat)
				md += kd
				n++
			}
		}
	}
	if n > 0 {
		md /= float64(n)
	}
	return md
}
