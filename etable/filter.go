// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

// FilterNull is a FilterFunc that filters out all rows that have a Null value
// in a 1D (scalar) column, according to the IsNull flag
func FilterNull(et *Table, row int) bool {
	for _, cl := range et.Cols {
		if cl.NumDims() > 1 {
			continue
		}
		if cl.IsNull1D(row) {
			return false
		}
	}
	return true
}
