// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"reflect"

	"github.com/goki/ki/kit"
	"goki.dev/goki/gi/v2/gi"
	"goki.dev/goki/gi/v2/giv"
	"goki.dev/goki/gi/v2/units"
	"goki.dev/goki/ki/v2"
)

func init() {
	giv.ValueViewMapAdd(kit.LongTypeName(KiT_Plot2D), func() giv.ValueView {
		vv := &Plot2DValueView{}
		ki.InitNode(vv)
		return vv
	})
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
		plot := opv.Interface().(*Plot2D)
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
	ac.ActionSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data any) {
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
	plot := opv.Interface().(*Plot2D)
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

// Plot2DDialog is for viewing an eplot.Plot2D --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
// gopy:interface=handle
func Plot2DDialog(avp *gi.Viewport2D, plot *Plot2D, opts giv.DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	if plot == nil || plot.Table == nil {
		return nil
	}

	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	clplot := plot.Clone().(*Plot2D)
	frame.InsertChild(clplot, prIdx+1)
	clplot.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	if opts.Inactive {
		clplot.SetInactive()
	}
	clplot.SetStretchMaxHeight()
	clplot.SetStretchMaxWidth()

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewEm(60))
	dlg.SetProp("min-height", units.NewEm(30))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		giv.MainMenuView(clplot, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}
