# etensor

Docs: [GoDoc](https://pkg.go.dev/github.com/goki/etable/v2/etensor)

`etensor` provides a basic set of tensor data structures (n-dimensional arrays of data), based on [apache arrow tensor](https://github.com/apache/arrow/tree/master/go/arrow/tensor) and intercompatible with those structures, and also with a [gonum](https://github.com/gonum/gonum) matrix interface.

The `etensor.Tensor` has all major data types available, and supports `float64` and `string` access for all types.  It provides the basis for the `etable.Table` columns.

The `Shape` of the tensor is a distinct struct that the tensor embeds, supporting *row major* ordering by default, but also *column major* or any other arbitrary ordering.  To construct a tensor, use `SetShape` method.

## Differences from arrow

* pure simple unidimensional Go slice used as the backing data array, auto allocated
* fully modifiable data -- arrow is designed to be read-only
* Shape struct is fully usable separate from the tensor data
* Everything exported, e.g., Offset method on Shape
* int used instead of int64 to make everything easier -- target platforms are all 64bit and have 64bit int in Go by default

## Updating generated code

Run the following:

```sh
go install github.com/goki/stringer@latest github.com/apache/arrow/go/arrow/_tools/tmpl@latest
PATH=$(go env GOPATH)/bin:$PATH make generate
go generate
```

The `go generate` updates type_string using the `goki` version of stringer.

Note that the `float64.go`, `int.go`, `string.go` and `bits.go` types have some amount of custom code relative to the `numeric.gen.go.tmpl` template, and thus must be updated manually with any changes.
