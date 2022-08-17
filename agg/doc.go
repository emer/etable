// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package agg provides aggregation functions operating on IdxView indexed views
of etable.Table data, along with standard AggFunc functions that can be used
at any level of aggregation from etensor on up.

The main functions use names to specify columns, and *Idx and *Try versions
are available that operate on column indexes and return errors, respectively.

See tsragg package for functions that operate directly on a etensor.Tensor
without the indexview indirection.

*/
package agg
