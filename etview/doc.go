// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package etview provides GUI Views of etable Table and Tensor structures
using the Cogent Core View framework: https://github.com/goki/gi

* TableView provides a row-and-column tabular GUI interface, similar to
a spreadsheet, for viewing and editing Table data.  Any higher-dimensional
tensor columns are shown as TensorGrid elements that can be clicked to
open a TensorView editor with actual numeric values in a similar spreadsheet-
like GUI.

* TensorView provides a spreadsheet-like GUI for viewing and editing
tensor data.

* TensorGrid provides a 2D colored grid display of tensor data, collapsing
any higher dimensions down to 2D.  Different giv.ColorMaps can be used
to translate values into colors.
*/
package etview
