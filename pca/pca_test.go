// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pca

import (
	"fmt"
	"math"
	"testing"

	"github.com/goki/etable/v2/etable"
	"github.com/goki/etable/v2/etensor"
	"github.com/goki/etable/v2/metric"
)

func TestPCAIris(t *testing.T) {
	// sch := etable.Schema{
	// 	{"sepal_len", etensor.FLOAT64, nil, nil},
	// 	{"sepal_wid", etensor.FLOAT64, nil, nil},
	// 	{"petal_len", etensor.FLOAT64, nil, nil},
	// 	{"petal_wid", etensor.FLOAT64, nil, nil},
	// 	{"class", etensor.STRING, nil, nil},
	// }

	// note: these results are verified against this example:
	// https://plot.ly/ipython-notebooks/principal-component-analysis/

	sch := etable.Schema{
		{"data", etensor.FLOAT64, []int{4}, nil},
		{"class", etensor.STRING, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, 0)
	err := dt.OpenCSV("test_data/iris.data", etable.Comma)
	if err != nil {
		t.Error(err)
	}
	ix := etable.NewIdxView(dt)
	pc := &PCA{}
	// pc.TableCol(ix, "data", metric.Covariance64)
	// fmt.Printf("covar: %v\n", pc.Covar)
	err = pc.TableCol(ix, "data", metric.Correlation64)
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("correl: %v\n", pc.Covar)
	// fmt.Printf("correl vec: %v\n", pc.Vectors)
	// fmt.Printf("correl val: %v\n", pc.Values)

	errtol := 1.0e-9
	corvals := []float64{0.020607707235624825, 0.14735327830509573, 0.9212209307072254, 2.910818083752054}
	for i, v := range pc.Values {
		dif := math.Abs(corvals[i] - v)
		if dif > errtol {
			err = fmt.Errorf("eigenvalue: %v  differs from correct: %v  was:  %v", i, corvals[i], v)
			t.Error(err)
		}
	}

	prjt := &etable.Table{}
	err = pc.ProjectColToTable(prjt, ix, "data", "class", []int{0, 1})
	if err != nil {
		t.Error(err)
	}
	// prjt.SaveCSV("test_data/projection01.csv", etable.Comma, true)
}
