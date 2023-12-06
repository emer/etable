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
	"goki.dev/girl/units"
	"goki.dev/ki/v2"
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
	giv.ValueMapAdd(laser.LongTypeName(etable.KiT_Table), func() giv.Value {
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

func (vv *TensorGridValue) WidgetType() reflect.Type {
	vv.WidgetTyp = KiT_TensorGrid
	return vv.WidgetTyp
}

func (vv *TensorGridValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
	tg.UpdateSig()
}

func (vv *TensorGridValue) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
	vv.UpdateWidget()
}

func (vv *TensorGridValue) HasAction() bool {
	return false
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorValue

// TensorValue presents a button that pulls up the TensorView viewer for an etensor.Tensor
type TensorValue struct {
	giv.ValueBase
}

func (vv *TensorValue) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *TensorValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		// opv := laser.OnePtrUnderlyingValue(vv.Value)
		ac.SetText("etensor.Tensor")
	}
}

func (vv *TensorValue) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_TensorValue).(*TensorValue)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *TensorValue) HasAction() bool {
	return true
}

func (vv *TensorValue) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(laser.NonPtrValue(vv.Value)) {
		return
	}
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	et := opv.Interface().(etensor.Tensor)
	if et == nil {
		return
	}
	tynm := "etensor.Tensor"
	olbl := vv.OwnerLabel()
	if olbl != "" {
		tynm += " " + olbl
	}
	desc, _ := vv.Tag("desc")
	_, inact := vv.Tag("inactive")
	if vv.This().(giv.Value).IsInactive() {
		inact = true
	}
	TensorGridDialog(vp, et, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}

////////////////////////////////////////////////////////////////////////////////////////
//  TableValue

// TableValue presents a button that pulls up the TableView viewer for an etable.Table
type TableValue struct {
	giv.ValueBase
}

func (vv *TableValue) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *TableValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		et := opv.Interface().(*etable.Table)
		if et != nil {
			if nm, has := et.MetaData["name"]; has {
				ac.SetText(nm)
			} else {
				ac.SetText("etable.Table")
			}
		}
	}
}

func (vv *TableValue) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_TableValue).(*TableValue)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *TableValue) HasAction() bool {
	return true
}

func (vv *TableValue) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(laser.NonPtrValue(vv.Value)) {
		return
	}
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	et := opv.Interface().(*etable.Table)
	if et == nil {
		return
	}
	tynm := "etable.Table"
	olbl := vv.OwnerLabel()
	if olbl != "" {
		tynm += " " + olbl
	}
	desc := et.MetaData["desc"]
	if td, has := vv.Tag("desc"); has {
		desc += " " + td
	}
	_, inact := et.MetaData["read-only"]
	if vv.This().(giv.Value).IsInactive() {
		inact = true
	}
	TableViewDialog(vp, et, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}

////////////////////////////////////////////////////////////////////////////////////////
//  SimMatValue

// SimMatValue presents a button that pulls up the SimMatGridView viewer for an etable.Table
type SimMatValue struct {
	giv.ValueBase
}

func (vv *SimMatValue) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *SimMatValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		smat := opv.Interface().(*simat.SimMat)
		if smat != nil && smat.Mat != nil {
			if nm, has := smat.Mat.MetaData("name"); has {
				ac.SetText(nm)
			} else {
				ac.SetText("simat.SimMat")
			}
		} else {
			ac.SetText("simat.SimMat")
		}
	}
}

func (vv *SimMatValue) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_SimMatValue).(*SimMatValue)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *SimMatValue) HasAction() bool {
	return true
}

func (vv *SimMatValue) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(laser.NonPtrValue(vv.Value)) {
		return
	}
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	smat := opv.Interface().(*simat.SimMat)
	if smat == nil || smat.Mat == nil {
		return
	}
	tynm := "simat.SimMat"
	olbl := vv.OwnerLabel()
	if olbl != "" {
		tynm += " " + olbl
	}
	desc, _ := smat.Mat.MetaData("desc")
	if td, has := vv.Tag("desc"); has {
		desc += " " + td
	}
	_, inact := smat.Mat.MetaData("read-only")
	if vv.This().(giv.Value).IsInactive() {
		inact = true
	}
	SimMatGridDialog(vp, smat, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}
