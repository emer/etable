// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/reflectx"
	"cogentcore.org/core/views"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/simat"
)

func init() {
	views.AddValue(etensor.Float32{}, func() views.Value {
		return &TensorValue{}
	})
	views.AddValue(etensor.Float64{}, func() views.Value {
		return &TensorValue{}
	})
	views.AddValue(etensor.Int64{}, func() views.Value {
		return &TensorValue{}
	})
	views.AddValue(etensor.Int32{}, func() views.Value {
		return &TensorValue{}
	})
	views.AddValue(etensor.String{}, func() views.Value {
		return &TensorValue{}
	})
	views.AddValue(etable.Table{}, func() views.Value {
		return &TableValue{}
	})
	views.AddValue(simat.SimMat{}, func() views.Value {
		return &SimMatValue{}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorGridValue

// TensorGridValue manages a TensorGrid view of an etensor.Tensor
type TensorGridValue struct {
	views.ValueBase[*TensorGrid]
}

func (v *TensorGridValue) Config() {
	tsr := v.Value.Interface().(etensor.Tensor)
	v.Widget.SetTensor(tsr)
}

func (v *TensorGridValue) Update() {
	tsr := v.Value.Interface().(etensor.Tensor)
	v.Widget.SetTensor(tsr)
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorValue

// TensorValue presents a button that pulls up the TensorView viewer for an etensor.Tensor
type TensorValue struct {
	views.ValueBase[*core.Button]
}

func (v *TensorValue) Config() {
	v.Widget.SetType(core.ButtonTonal).SetIcon(icons.Edit)
	views.ConfigDialogWidget(v, true)
}

func (v *TensorValue) Update() {
	npv := reflectx.NonPointerValue(v.Value)
	if !v.Value.IsValid() || v.Value.IsZero() || !npv.IsValid() || npv.IsZero() {
		v.Widget.SetText("nil")
	} else {
		// opv := reflectx.OnePointerUnderlyingValue(vv.Value)
		v.Widget.SetText("etensor.Tensor")
	}
}

func (v *TensorValue) ConfigDialog(d *core.Body) (bool, func()) {
	opv := reflectx.OnePointerUnderlyingValue(v.Value)
	et := opv.Interface().(etensor.Tensor)
	if et == nil {
		return false, nil
	}
	NewTensorGrid(d).SetTensor(et)
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////////////
//  TableValue

// TableValue presents a button that pulls up the TableView viewer for an etable.Table
type TableValue struct {
	views.ValueBase[*core.Button]
}

func (v *TableValue) Config() {
	v.Widget.SetType(core.ButtonTonal).SetIcon(icons.Edit)
	views.ConfigDialogWidget(v, true)
}

func (v *TableValue) Update() {
	npv := reflectx.NonPointerValue(v.Value)
	if !v.Value.IsValid() || v.Value.IsZero() || !npv.IsValid() || npv.IsZero() {
		v.Widget.SetText("nil")
	} else {
		opv := reflectx.OnePointerUnderlyingValue(v.Value)
		et := opv.Interface().(*etable.Table)
		if et != nil {
			if nm, has := et.MetaData["name"]; has {
				v.Widget.SetText(nm)
			} else {
				v.Widget.SetText("etable.Table")
			}
		}
	}
}

func (v *TableValue) ConfigDialog(d *core.Body) (bool, func()) {
	opv := reflectx.OnePointerUnderlyingValue(v.Value)
	et := opv.Interface().(*etable.Table)
	if et == nil {
		return false, nil
	}
	NewTableView(d).SetTable(et)
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////////////
//  SimMatValue

// SimMatValue presents a button that pulls up the SimMatGridView viewer for an etable.Table
type SimMatValue struct {
	views.ValueBase[*core.Button]
}

func (v *SimMatValue) Config() {
	v.Widget.SetType(core.ButtonTonal).SetIcon(icons.Edit)
	views.ConfigDialogWidget(v, false)
}

func (v *SimMatValue) Update() {
	npv := reflectx.NonPointerValue(v.Value)
	if !v.Value.IsValid() || v.Value.IsZero() || !npv.IsValid() || npv.IsZero() {
		v.Widget.SetText("nil")
	} else {
		opv := reflectx.OnePointerUnderlyingValue(v.Value)
		smat := opv.Interface().(*simat.SimMat)
		if smat != nil && smat.Mat != nil {
			if nm, has := smat.Mat.MetaData("name"); has {
				v.Widget.SetText(nm)
			} else {
				v.Widget.SetText("simat.SimMat")
			}
		} else {
			v.Widget.SetText("simat.SimMat")
		}
	}
}

func (v *SimMatValue) ConfigDialog(d *core.Body) (bool, func()) {
	opv := reflectx.OnePointerUnderlyingValue(v.Value)
	smat := opv.Interface().(*simat.SimMat)
	if smat == nil || smat.Mat == nil {
		return false, nil
	}
	NewSimMatGrid(d).SetSimMat(smat)
	return true, nil
}
