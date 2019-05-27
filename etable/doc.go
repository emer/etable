// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package etable provides the etable.Table structure which provides a
DataTable or DataFrame data representation, which is a collection
of columnar data all having the same number of rows.

Each column is an etensor.Tensor, so it can represent scalar or higher
dimensional data per each cell (row x column location) in the Table.
Thus, scalar data is represented using a 1D Tensor where the 1 dimension
is the rows of the table, and likewise higher dimensional data always has
the outer-most dimension as the row.

All tensors MUST have RowMajor stride layout for consistency, with the
outer-most dimension as the row dimension, which is enforced to be the
same across all columns.

The tensor columns can be individually converted to / from arrow.Tensors
and conversion between arrow.Table is planned, along with inter-conversion
with relevant gonum structures including the planned dframe.Frame.

Native support is provided for basic CSV, TSV I/O, including the
C++ emergent standard TSV format with full type information in the first
row column headers.

The etable.IdxView is an indexed view into a Table, which is used for
all data-processing operations such as Sort, Filter, Split (groupby),
and for aggregating data as in a pivot-table.

See etable/agg package for aggregation functions that operate on the
IdxView to perform standard aggregation operations such as Sum,
Mean, etc.

Other relevant examples of DataTable-like structures:
* https://github.com/apache/arrow/tree/master/go/arrow Table
* http://xarray.pydata.org/en/stable/index.html
* https://pandas.pydata.org/pandas-docs/stable/reference/frame.html
* https://www.rdocumentation.org/packages/base/versions/3.4.3/topics/data.frame
* https://github.com/tobgu/qframe
* https://github.com/kniren/gota
*/
package etable
