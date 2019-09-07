// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

// Func32 is a distance / similarity metric operating on slices of float32 numbers
type Func32 func(a, b []float32) float32

// Func64 is a distance / similarity metric operating on slices of float64 numbers
type Func64 func(a, b []float64) float64
