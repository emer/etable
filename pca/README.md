# pca

Docs: [GoDoc](https://pkg.go.dev/github.com/goki/etable/v2/pca)

This performs principal component's analysis and associated covariance matrix computations, operating on `etable.Table` or `etensor.Tensor` data, using the [gonum](https://github.com/gonum/gonum) matrix interface.

There is support for the SVD version, which is much faster and produces the same results, with options for how much information to compute trading off with compute time.


