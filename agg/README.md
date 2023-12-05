# agg

Docs: [GoDoc](https://pkg.go.dev/goki.dev/etable/v2/agg)

This package provides aggregation functions operating on `IdxView` indexed views of `etable.Table` data, along with standard AggFunc functions that can be used at any level of aggregation from etensor on up.

The main functions use names to specify columns, and `*Idx` and `*Try` versions are available that operate on column indexes and return errors, respectively.

See the `tsragg` package for functions that operate directly on a `etensor.Tensor` without the index view indirection.


