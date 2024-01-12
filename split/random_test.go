// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package split

import (
	"fmt"
	"testing"

	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
)

func TestPermuted(t *testing.T) {
	dt := etable.New(etable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
		{"Output", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
	}, 25)
	ix := etable.NewIdxView(dt)
	spl, err := Permuted(ix, []float64{.5, .5}, nil)
	if err != nil {
		t.Error(err)
	}
	for i, sp := range spl.Splits {
		fmt.Printf("split: %v name: %v len: %v idxs: %v\n", i, spl.Values[i], len(sp.Idxs), sp.Idxs)
	}

	spl, err = Permuted(ix, []float64{.25, .5, .25}, []string{"test", "train", "validate"})
	if err != nil {
		t.Error(err)
	}
	for i, sp := range spl.Splits {
		fmt.Printf("split: %v name: %v len: %v idxs: %v\n", i, spl.Values[i], len(sp.Idxs), sp.Idxs)
	}
}
