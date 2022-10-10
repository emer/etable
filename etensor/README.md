# etensor

Docs: [GoDoc](https://pkg.go.dev/github.com/emer/etable/etensor)

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

TODO: This is not complete. How to generate e.g. float64.go from numeric.gen.go?

```sh
go install golang.org/x/tools/cmd/stringer github.com/apache/arrow/go/arrow/_tools/tmpl
PATH=$GOROOT/bin:$PATH make generate
go generate
```
