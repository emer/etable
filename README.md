# etable: DataTable / DataFrame structure in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/etable)](https://goreportcard.com/report/github.com/emer/etable)
[![GoDoc](https://godoc.org/github.com/emer/etable?status.svg)](https://godoc.org/github.com/emer/etable)
[![Travis](https://travis-ci.com/emer/etable.svg?branch=master)](https://travis-ci.com/emer/etable)

 **etable** (or eTable) provides a DataTable / DataFrame structure in Go (golang), similar to [pandas](https://pandas.pydata.org/) and [xarray](http://xarray.pydata.org/en/stable/) in Python, and [Apache Arrow Table](https://github.com/apache/arrow/tree/master/go/arrow/array/table.go), using `etensor` n-dimensional columns aligned by common outermost row dimension.

The e-name derives from the `emergent` neural network simulation framework, but `e` is also extra-dimensional, extended, electric, easy-to-use -- all good stuff.. :)

See `examples/dataproc` for a full demo of how to use this system for data analysis, paralleling the example in 	[Python Data Science](https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html) using pandas, to see directly how that translates into this framework.

The following packages are included:

* `bitslice` is a Go slice of bytes `[]byte` that has methods for setting individual bits, as if it was a slice of bools, while being 8x more memory efficient.  This is used for encoding null entries in  `etensor`, and as a Tensor of bool / bits there as well, and is generally very useful for binary (boolean) data.

* `etensor` is a Tensor (n-dimensional array) object.  `etensor.Tensor` is an interface that applies to many different type-specific instances, such as `etensor.Float32`.  A tensor is just a `etensor.Shape` plus a slice holding the specific data type.  Our tensor is based directly on the [Apache Arrow](https://github.com/apache/arrow/tree/master/go) project's tensor, and it fully interoperates with it.  Arrow tensors are designed to be read-only, and we needed some extra support to make our `etable.Table` work well, so we had to roll our own.  Our tensors also interoperate fully with Gonum's 2D-specific Matrix type for the 2D case.

* `etable` has the `etable.Table` DataTable / DataFrame object, which is useful for many different data analysis and database functions, and also for holding patterns to present to a neural network, and logs of output from the models, etc.  A `etable.Table` is just a slice of `etensor.Tensor` columns, that are all aligned along the outer-most *row* dimension.  Index-based indirection, which is essential for efficient Sort, Filter etc, is provided by the `etable.IdxView` type, which is an indexed view into a Table.  All data processing operations are defined on the IdxView.

* `eplot` provides an interactive 2D plotting GUI in [GoGi](https://github.com/goki/gi) for Table data, using the [gonum plot](https://github.com/gonum/plot) plotting package.  You can select which columns to plot and specify various basic plot parameters.

* `etview` provides an interactive tabular, spreadsheet-style GUI using [GoGi](https://github.com/goki/gi) for viewing and editing `etable.Table` and `etable.Tensor` objects.  The `etview.TensorGrid` also provides a colored grid display higher-dimensional tensor data.

* `agg` provides standard aggregation functions (`Sum`, `Mean`, `Var`, `Std` etc) operating over `etable.IdxView` views of Table data.  It also defines standard `AggFunc` functions such as `SumFunc` which can be used for `Agg` functions on either a Tensor or IdxView.

* `tsragg` provides the same agg functions as in `agg`, but operating on all the values in a given `Tensor`.  Because of the indexed, row-based nature of tensors in a Table, these are not the same as the `agg` functions.

