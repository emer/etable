// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"reflect"

	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/simat"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

func init() {
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(etensor.Float32{})), func() giv.ValueView {
		vv := &TensorValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(etensor.Float64{})), func() giv.ValueView {
		vv := &TensorValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(etensor.Int64{})), func() giv.ValueView {
		vv := &TensorValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(etensor.Int32{})), func() giv.ValueView {
		vv := &TensorValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(etensor.String{})), func() giv.ValueView {
		vv := &TensorValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(etable.KiT_Table), func() giv.ValueView {
		vv := &TableValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(reflect.TypeOf(simat.SimMat{})), func() giv.ValueView {
		vv := &SimMatValueView{}
		vv.Init(vv)
		return vv
	})
	giv.ValueViewMapAdd(kit.LongTypeName(eplot.KiT_Plot2D), func() giv.ValueView {
		vv := &Plot2DValueView{}
		vv.Init(vv)
		return vv
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorGridValueView

// TensorGridValueView manages a TensorGrid view of an etensor.Tensor
type TensorGridValueView struct {
	giv.ValueViewBase
}

var KiT_TensorGridValueView = kit.Types.AddType(&TensorGridValueView{}, nil)

func (vv *TensorGridValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = KiT_TensorGrid
	return vv.WidgetTyp
}

func (vv *TensorGridValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
	tg.UpdateSig()
}

func (vv *TensorGridValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	tg := vv.Widget.(*TensorGrid)
	tsr := vv.Value.Interface().(etensor.Tensor)
	tg.SetTensor(tsr)
	vv.UpdateWidget()
}

func (vv *TensorGridValueView) HasAction() bool {
	return false
}

////////////////////////////////////////////////////////////////////////////////////////
//  TensorValueView

// TensorValueView presents a button that pulls up the TensorView viewer for an etensor.Tensor
type TensorValueView struct {
	giv.ValueViewBase
}

var KiT_TensorValueView = kit.Types.AddType(&TensorValueView{}, nil)

func (vv *TensorValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *TensorValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := kit.NonPtrValue(vv.Value)
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		// opv := kit.OnePtrUnderlyingValue(vv.Value)
		ac.SetText("etensor.Tensor")
	}
}

func (vv *TensorValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_TensorValueView).(*TensorValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *TensorValueView) HasAction() bool {
	return true
}

func (vv *TensorValueView) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(kit.NonPtrValue(vv.Value)) {
		return
	}
	opv := kit.OnePtrUnderlyingValue(vv.Value)
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
	if vv.This().(giv.ValueView).IsInactive() {
		inact = true
	}
	TensorViewDialog(vp, et, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}

////////////////////////////////////////////////////////////////////////////////////////
//  TableValueView

// TableValueView presents a button that pulls up the TableView viewer for an etable.Table
type TableValueView struct {
	giv.ValueViewBase
}

var KiT_TableValueView = kit.Types.AddType(&TableValueView{}, nil)

func (vv *TableValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *TableValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := kit.NonPtrValue(vv.Value)
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		opv := kit.OnePtrUnderlyingValue(vv.Value)
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

func (vv *TableValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_TableValueView).(*TableValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *TableValueView) HasAction() bool {
	return true
}

func (vv *TableValueView) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(kit.NonPtrValue(vv.Value)) {
		return
	}
	opv := kit.OnePtrUnderlyingValue(vv.Value)
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
	if vv.This().(giv.ValueView).IsInactive() {
		inact = true
	}
	TableViewDialog(vp, et, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}

////////////////////////////////////////////////////////////////////////////////////////
//  SimMatValueView

// SimMatValueView presents a button that pulls up the SimMatGridView viewer for an etable.Table
type SimMatValueView struct {
	giv.ValueViewBase
}

var KiT_SimMatValueView = kit.Types.AddType(&SimMatValueView{}, nil)

func (vv *SimMatValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *SimMatValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := kit.NonPtrValue(vv.Value)
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		opv := kit.OnePtrUnderlyingValue(vv.Value)
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

func (vv *SimMatValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_SimMatValueView).(*SimMatValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *SimMatValueView) HasAction() bool {
	return true
}

func (vv *SimMatValueView) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(kit.NonPtrValue(vv.Value)) {
		return
	}
	opv := kit.OnePtrUnderlyingValue(vv.Value)
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
	if vv.This().(giv.ValueView).IsInactive() {
		inact = true
	}
	SimMatGridDialog(vp, smat, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact}, recv, dlgFunc)
}

////////////////////////////////////////////////////////////////////////////////////////
//  Plot2DValueView

// Plot2DValueView presents a button that pulls up the Plot2D in a dialog
type Plot2DValueView struct {
	giv.ValueViewBase
}

var KiT_Plot2DValueView = kit.Types.AddType(&Plot2DValueView{}, nil)

func (vv *Plot2DValueView) WidgetType() reflect.Type {
	vv.WidgetTyp = gi.KiT_Action
	return vv.WidgetTyp
}

func (vv *Plot2DValueView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Action)
	npv := kit.NonPtrValue(vv.Value)
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(npv) {
		ac.SetText("nil")
	} else {
		opv := kit.OnePtrUnderlyingValue(vv.Value)
		plot := opv.Interface().(*eplot.Plot2D)
		if plot != nil && plot.Table != nil && plot.Table.Table != nil {
			if nm, has := plot.Table.Table.MetaData["name"]; has {
				ac.SetText(nm)
			} else {
				ac.SetText("eplot.Plot2D")
			}
		} else {
			ac.SetText("eplot.Plot2D")
		}
	}
}

func (vv *Plot2DValueView) ConfigWidget(widg gi.Node2D) {
	vv.Widget = widg
	ac := vv.Widget.(*gi.Action)
	ac.Tooltip, _ = vv.Tag("desc")
	ac.SetProp("padding", units.NewPx(2))
	ac.SetProp("margin", units.NewPx(2))
	ac.SetProp("border-radius", units.NewPx(4))
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		vvv, _ := recv.Embed(KiT_Plot2DValueView).(*Plot2DValueView)
		ac := vvv.Widget.(*gi.Action)
		vvv.Activate(ac.ViewportSafe(), nil, nil)
	})
	vv.UpdateWidget()
}

func (vv *Plot2DValueView) HasAction() bool {
	return true
}

func (vv *Plot2DValueView) Activate(vp *gi.Viewport2D, recv ki.Ki, dlgFunc ki.RecvFunc) {
	if kit.ValueIsZero(vv.Value) || kit.ValueIsZero(kit.NonPtrValue(vv.Value)) {
		return
	}
	opv := kit.OnePtrUnderlyingValue(vv.Value)
	plot := opv.Interface().(*eplot.Plot2D)
	if plot == nil || plot.Table == nil {
		return
	}
	tynm := "eplot.Plot2D"
	olbl := vv.OwnerLabel()
	if olbl != "" {
		tynm += " " + olbl
	}
	desc, _ := plot.Table.Table.MetaData["desc"]
	if td, has := vv.Tag("desc"); has {
		desc += " " + td
	}
	// _, inact := smat.Mat.MetaData("read-only")
	// if vv.This().(giv.ValueView).IsInactive() {
	// 	inact = true
	// }
	Plot2DDialog(vp, plot, giv.DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave}, recv, dlgFunc)
}
