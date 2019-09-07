// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

type TstSort struct {
	raw []int
	idx []int
}

func (ts *TstSort) Len() int {
	return len(ts.idx)
}

func (ts *TstSort) Less(i, j int) bool {
	return ts.raw[ts.idx[i]] < ts.raw[ts.idx[j]]
}

func (ts *TstSort) Swap(i, j int) {
	ts.idx[i], ts.idx[j] = ts.idx[j], ts.idx[i]
}

func TestSort(t *testing.T) {
	n := 20
	ts := &TstSort{}
	ts.raw = rand.Perm(n)
	ts.idx = make([]int, n)
	for i := range ts.idx {
		ts.idx[i] = i
	}
	sort.Sort(ts)

	for i := range ts.idx {
		fmt.Printf("i: %d\t idx[i]: %d\t raw[idx[i]]: %d\n", i, ts.idx[i], ts.raw[ts.idx[i]])
	}
}
