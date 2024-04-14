// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"cogentcore.org/core/core"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/etable/v2/simat"
)

/*
// TensorViewDialog is for editing an etensor.Tensor using a TensorView --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
// gopy:interface=handle
func TensorViewDialog(avp *core.Viewport2D, tsr etensor.Tensor, opts views.DlgOpts, recv tree.Ki, dlgFunc tree.RecvFunc) *core.Body {
	dlg, recyc := core.RecycleStdDialog(tsr, opts.ToGiOpts(), opts.Ok, opts.Cancel)
	if recyc {
		return dlg
	}
	dlg.Data = tsr

	frame := dlg.Frame()
	_, prIndex := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TensorView, prIndex+1, "tensor-view").(*TensorView)
	sv.Viewport = dlg.Embed(core.KiT_Viewport2D).(*core.Viewport2D)
	if opts.Inactive {
		sv.SetInactive()
	}
	sv.NoAdd = opts.NoAdd
	sv.NoDelete = opts.NoDelete
	sv.SetTensor(tsr, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		views.MainMenuView(tsr, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}
*/

// TensorGridDialog is for viewing a etensor.Tensor using a TensorGrid.
// gopy:interface=handle
func TensorGridDialog(ctx core.Widget, tsr etensor.Tensor, title string) {
	d := core.NewBody()
	if title != "" {
		d.SetTitle(title)
	}
	NewTensorGrid(d).SetTensor(tsr)
	d.NewDialog(ctx).SetNewWindow(true).Run()
}

// TableViewDialog is for editing an etable.Table using a TableView.
// gopy:interface=handle
func TableViewDialog(ctx core.Widget, ix *etable.IndexView, title string) {
	d := core.NewBody()
	if title != "" {
		d.SetTitle(title)
	}
	NewTableView(d).SetTableView(ix)
	d.NewDialog(ctx).SetNewWindow(true).Run()
}

// SimMatGridDialog is for viewing a etensor.Tensor using a SimMatGrid.
// dialog signals (nil to ignore)
// gopy:interface=handle
func SimMatGridDialog(ctx core.Widget, smat *simat.SimMat, title string) {
	d := core.NewBody()
	if title != "" {
		d.SetTitle(title)
	}
	NewSimMatGrid(d).SetSimMat(smat)
	d.NewDialog(ctx).SetNewWindow(true).Run()
}
