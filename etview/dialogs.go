// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
)

//gopy:interface=handle TensorViewDialog is for editing an etensor.Tensor using a TensorView --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
func TensorViewDialog(avp *gi.Viewport2D, tsr etensor.Tensor, opts giv.DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TensorView, prIdx+1, "tensor-view").(*TensorView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	if opts.Inactive {
		sv.SetInactive()
	}
	sv.NoAdd = opts.NoAdd
	sv.NoDelete = opts.NoDelete
	sv.SetTensor(tsr, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewEm(60))
	dlg.SetProp("min-height", units.NewEm(30))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		giv.MainMenuView(tsr, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

//gopy:interface=handle TensorGridDialog is for viewing a etensor.Tensor using a TensorGrid --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
func TensorGridDialog(avp *gi.Viewport2D, tsr etensor.Tensor, opts giv.DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TensorGrid, prIdx+1, "tensor-grid").(*TensorGrid)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	if opts.Inactive {
		sv.SetInactive()
	}
	sv.SetStretchMaxHeight()
	sv.SetStretchMaxWidth()
	sv.SetTensor(tsr)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewEm(60))
	dlg.SetProp("min-height", units.NewEm(30))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		giv.MainMenuView(tsr, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

//gopy:interface=handle TableViewDialog is for editing an etable.Table using a TableView --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
func TableViewDialog(avp *gi.Viewport2D, et *etable.Table, opts giv.DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TableView, prIdx+1, "table-view").(*TableView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	if opts.Inactive {
		sv.SetInactive()
	}
	sv.NoAdd = opts.NoAdd
	sv.NoDelete = opts.NoDelete
	sv.SetTable(et, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewEm(60))
	dlg.SetProp("min-height", units.NewEm(30))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		giv.MainMenuView(et, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}
