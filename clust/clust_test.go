// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clust

import (
	"fmt"
	"testing"

	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/metric"
	"goki.dev/etable/v2/simat"
)

func TestClust(t *testing.T) {
	dt := &etable.Table{}
	err := dt.OpenCSV("test_data/faces.dat", etable.Tab)
	if err != nil {
		t.Error(err)
	}
	ix := etable.NewIdxView(dt)
	smat := &simat.SimMat{}
	smat.TableCol(ix, "Input", "Name", false, metric.Euclidean64)

	// fmt.Printf("%v\n", smat.Mat)
	// cl := Glom(smat, MinDist)
	cl := Glom(smat, AvgDist)
	s := cl.Sprint(smat, 0)
	fmt.Println(s)
}
