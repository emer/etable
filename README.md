# dtable: DataTable / DataFrame structure in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/dtable)](https://goreportcard.com/report/github.com/emer/dtable)
[![GoDoc](https://godoc.org/github.com/emer/dtable?status.svg)](https://godoc.org/github.com/emer/dtable)

 **dtable** provides a DataTable / DataFrame structure in Go (golang), similar to pandas and xarray in Python, using etensor columns aligned by common row dimension.

The following packages are included:

* `bitslice` is a Go slice of bytes `[]byte` that has methods for setting individual bits, as if it was a slice of bools, while being 8x more memory efficient.  This is used in `prjn` for representing the pattern of connectivity, for encoding null entries in  `etensor`, and as a Tensor of bool / bits there as well.

* `etensor` is our own implementation of a Tensor object, which corresponds to the `Matrix` type in C++ emergent.  `etensor.Tensor` is an interface that applies to many different type-specific instances, such as `etensor.Float32`.  A tensor is just a `etensor.Shape` plus a slice holding the specific data type.  Our tensor is based directly on the [Apache Arrow](https://github.com/apache/arrow/tree/master/go) project's tensor, and it fully interoperates with it.  Arrow tensors are designed to be read-only, and we needed some extra support to make our `dtable.Table` work well, so we had to roll our own.  Our tensors will also interoperate fully with Gonum's 2D-specific Matrix type.

* `dtable` is our Go version of `DataTable` from C++ emergent, which is widely useful for holding input patterns to present to the network, and logs of output from the network, among many other uses.  A `dtable.Table` is a collection of `etensor.Tensor` columns, that are all aligned along the outer-most *row* dimension.  We are keeping the index-based indirection outside of the core type, which greatly simplifies many things.  The `dtable.Table` should interoperate with the under-development gonum `DataFrame` structure among others.  The use of this data structure is always optional and orthogonal to the core network algorithm code -- in Python the `pandas` library has a suitable `DataFrame` structure that can be used instead.


