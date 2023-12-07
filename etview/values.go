// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"reflect"

	"goki.dev/etable/v2/etable"
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/simat"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
	"goki.dev/gti"
	"goki.dev/laser"
)

func init() {
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etensor.Float32{})), func() giv.Value {
		return &TensorValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etensor.Float64{})), func() giv.Value {
		return &TensorValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etensor.Int64{})), func() giv.Value {
		return &TensorValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etensor.Int32{})), func() giv.Value {
		return &TensorValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etensor.String{})), func() giv.Value {
		return &TensorValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(etable.Table{})), func() giv.Value {
		return &TableValue{}
	})
	giv.ValueMapAdd(laser.LongTypeName(reflect.TypeOf(simat.SimMat{})), func() giv.Value {
		return &SimMatValue{}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorGridValue

// TensorGridValue manages a TensorGrid view of an etensor.Tensor
type TensorGridValue struct {
	giv.ValueBase
}

func (vv *TensorGridValue) WidgetType() *gti.Type {
	vv.WidgetTyp = TensorGridType
	return vv.WidgetTyp
}

func (vv *TensorGridValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
}

func (vv *TensorGridValue) ConfigWidget(w gi.Widget) {
	if vv.Widget == w {
		vv.UpdateWidget()
		return
	}
	vv.Widget = w
	vv.StdConfigWidget(w)
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
	vv.UpdateWidget()
}

func (vv *TensorGridValue) HasDialog() bool { return false }

////////////////////////////////////////////////////////////////////////////////////////
//  TensorValue

// TensorValue presents a button that pulls up the TensorView viewer for an etensor.Tensor
type TensorValue struct {
	giv.ValueBase
}

func (vv *TensorValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *TensorValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		bt.SetText("nil")
	} else {
		// opv := laser.OnePtrUnderlyingValue(vv.Value)
		bt.SetText("etensor.Tensor")
	}
}

func (vv *TensorValue) ConfigWidget(w gi.Widget) {
	if vv.Widget == w {
		vv.UpdateWidget()
		return
	}
	vv.Widget = w
	vv.StdConfigWidget(w)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config()
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *TensorValue) HasDialog() bool                      { return true }
func (vv *TensorValue) OpenDialog(ctx gi.Widget, fun func()) { giv.OpenValueDialog(vv, ctx, fun) }

func (vv *TensorValue) ConfigDialog(d *gi.Body) (bool, func()) {
	opv := laser.OnePtrUnderlyingValue(vv.Value)
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
	giv.ValueBase
}

func (vv *TableValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *TableValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		bt.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		et := opv.Interface().(*etable.Table)
		if et != nil {
			if nm, has := et.MetaData["name"]; has {
				bt.SetText(nm)
			} else {
				bt.SetText("etable.Table")
			}
		}
	}
}

func (vv *TableValue) ConfigWidget(w gi.Widget) {
	if vv.Widget == w {
		vv.UpdateWidget()
		return
	}
	vv.Widget = w
	vv.StdConfigWidget(w)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config()
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *TableValue) HasDialog() bool                      { return true }
func (vv *TableValue) OpenDialog(ctx gi.Widget, fun func()) { giv.OpenValueDialog(vv, ctx, fun) }

func (vv *TableValue) ConfigDialog(d *gi.Body) (bool, func()) {
	opv := laser.OnePtrUnderlyingValue(vv.Value)
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
	giv.ValueBase
}

func (vv *SimMatValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *SimMatValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		bt.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		smat := opv.Interface().(*simat.SimMat)
		if smat != nil && smat.Mat != nil {
			if nm, has := smat.Mat.MetaData("name"); has {
				bt.SetText(nm)
			} else {
				bt.SetText("simat.SimMat")
			}
		} else {
			bt.SetText("simat.SimMat")
		}
	}
}

func (vv *SimMatValue) ConfigWidget(w gi.Widget) {
	if vv.Widget == w {
		vv.UpdateWidget()
		return
	}
	vv.Widget = w
	bt := vv.Widget.(*gi.Button)
	vv.StdConfigWidget(w)
	bt.SetType(gi.ButtonTonal)
	bt.Config()
	bt.OnClick(func(e events.Event) {
		if !vv.IsReadOnly() {
			vv.OpenDialog(bt, nil)
		}
	})
	vv.UpdateWidget()
}

func (vv *SimMatValue) HasDialog() bool                      { return true }
func (vv *SimMatValue) OpenDialog(ctx gi.Widget, fun func()) { giv.OpenValueDialog(vv, ctx, fun) }

func (vv *SimMatValue) ConfigDialog(d *gi.Body) (bool, func()) {
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	smat := opv.Interface().(*simat.SimMat)
	if smat == nil || smat.Mat == nil {
		return false, nil
	}
	NewSimMatGrid(d).SetSimMat(smat)
	return true, nil
}
