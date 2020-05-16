// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"fmt"
	"image"
	"reflect"

	"github.com/chewxy/math32"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin/mimedata"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// etview.TensorView provides a GUI interface for etable.Tensor's
// using a tabular rows-and-columns interface
type TensorView struct {
	giv.SliceViewBase
	Tensor etensor.Tensor `desc:"the tensor that we're a view of"`
	TsrLay TensorLayout   `desc:"layout config of the tensor"`
	NCols  int            `inactive:"+" desc:"number of columns in table (as of last update)"`
}

var KiT_TensorView = kit.Types.AddType(&TensorView{}, TensorViewProps)

// AddNewTensorView adds a new tableview to given parent node, with given name.
func AddNewTensorView(parent ki.Ki, name string) *TensorView {
	return parent.AddNewChild(KiT_TensorView, name).(*TensorView)
}

// check for interface impl
var _ giv.SliceViewer = (*TensorView)(nil)

// SetTensor sets the source tensor that we are viewing
func (tv *TensorView) SetTensor(tsr etensor.Tensor, tmpSave giv.ValueView) {
	updt := false
	if tsr == nil {
		return
	}
	if tv.Tensor != tsr {
		if !tv.IsInactive() {
			tv.SelectedIdx = -1
		}
		tv.StartIdx = 0
		tv.Tensor = tsr
		updt = tv.UpdateStart()
		tv.ResetSelectedIdxs()
		tv.SelectMode = false
		tv.SetFullReRender()
	}
	tv.ShowIndex = true
	if sidxp, err := tv.PropTry("index"); err == nil {
		tv.ShowIndex, _ = kit.ToBool(sidxp)
	}
	tv.InactKeyNav = true
	if siknp, err := tv.PropTry("inact-key-nav"); err == nil {
		tv.InactKeyNav, _ = kit.ToBool(siknp)
	}
	tv.TmpSave = tmpSave
	tv.Config()
	tv.UpdateEnd(updt)
}

var TensorViewProps = ki.Props{
	"background-color": &gi.Prefs.Colors.Background,
	"color":            &gi.Prefs.Colors.Font,
	"max-width":        -1,
	"max-height":       -1,
}

// IsConfiged returns true if the widget is fully configured
func (tv *TensorView) IsConfiged() bool {
	if len(tv.Kids) == 0 {
		return false
	}
	sf := tv.SliceFrame()
	if len(sf.Kids) == 0 {
		return false
	}
	return true
}

// Config configures the view
func (tv *TensorView) Config() {
	tv.Lay = gi.LayoutVert
	tv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_ToolBar, "toolbar")
	config.Add(gi.KiT_Frame, "frame")
	mods, updt := tv.ConfigChildren(config, true)
	tv.ConfigSliceGrid()
	tv.ConfigToolbar()
	if mods {
		tv.SetFullReRender()
		tv.UpdateEnd(updt)
	}
}

func (tv *TensorView) UpdtSliceSize() int {
	tv.SliceSize, tv.NCols, _, _ = etensor.Prjn2DShape(tv.Tensor.ShapeObj(), tv.TsrLay.OddRow)
	return tv.SliceSize
}

// SliceFrame returns the outer frame widget, which contains all the header,
// fields and values
func (tv *TensorView) SliceFrame() *gi.Frame {
	return tv.ChildByName("frame", 0).(*gi.Frame)
}

// GridLayout returns the SliceGrid grid-layout widget, with grid and scrollbar
func (tv *TensorView) GridLayout() *gi.Layout {
	return tv.SliceFrame().ChildByName("grid-lay", 0).(*gi.Layout)
}

// SliceGrid returns the SliceGrid grid frame widget, which contains all the
// fields and values, within SliceFrame
func (tv *TensorView) SliceGrid() *gi.Frame {
	return tv.GridLayout().ChildByName("grid", 0).(*gi.Frame)
}

// ScrollBar returns the SliceGrid scrollbar
func (tv *TensorView) ScrollBar() *gi.ScrollBar {
	return tv.GridLayout().ChildByName("scrollbar", 1).(*gi.ScrollBar)
}

// SliceHeader returns the Toolbar header for slice grid
func (tv *TensorView) SliceHeader() *gi.ToolBar {
	return tv.SliceFrame().Child(0).(*gi.ToolBar)
}

// ToolBar returns the toolbar widget
func (tv *TensorView) ToolBar() *gi.ToolBar {
	return tv.ChildByName("toolbar", 0).(*gi.ToolBar)
}

// RowWidgetNs returns number of widgets per row and offset for index label
func (tv *TensorView) RowWidgetNs() (nWidgPerRow, idxOff int) {
	nWidgPerRow = 1 + tv.NCols
	if !tv.IsInactive() {
		if !tv.NoAdd {
			nWidgPerRow += 1
		}
		if !tv.NoDelete {
			nWidgPerRow += 1
		}
	}
	idxOff = 1
	if !tv.ShowIndex {
		nWidgPerRow -= 1
		idxOff = 0
	}
	return
}

// ConfigSliceGrid configures the SliceGrid for the current slice
// this is only called by global Config and updates are guarded by that
func (tv *TensorView) ConfigSliceGrid() {
	if tv.Tensor == nil {
		return
	}

	sz := tv.UpdtSliceSize()
	if sz == 0 {
		return
	}

	nWidgPerRow, idxOff := tv.RowWidgetNs()

	sg := tv.SliceFrame()
	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)

	sg.Lay = gi.LayoutVert
	sg.SetMinPrefWidth(units.NewCh(20))
	sg.SetProp("overflow", gi.OverflowScroll) // this still gives it true size during PrefSize
	sg.SetStretchMax()                        // for this to work, ALL layers above need it too

	sgcfg := kit.TypeAndNameList{}
	sgcfg.Add(gi.KiT_ToolBar, "header")
	sgcfg.Add(gi.KiT_Layout, "grid-lay")
	sg.ConfigChildren(sgcfg, true)

	sgh := tv.SliceHeader()
	sgh.Lay = gi.LayoutHoriz
	sgh.SetProp("overflow", gi.OverflowHidden) // no scrollbars!
	sgh.SetProp("spacing", 0)
	// sgh.SetStretchMaxWidth()

	gl := tv.GridLayout()
	gl.Lay = gi.LayoutHoriz
	gl.SetStretchMax() // for this to work, ALL layers above need it too
	gconfig := kit.TypeAndNameList{}
	gconfig.Add(gi.KiT_Frame, "grid")
	gconfig.Add(gi.KiT_ScrollBar, "scrollbar")
	gl.ConfigChildren(gconfig, true) // covered by above

	sgf := tv.SliceGrid()
	sgf.Lay = gi.LayoutGrid
	sgf.Stripes = gi.RowStripes
	sgf.SetMinPrefHeight(units.NewEm(10))
	sgf.SetStretchMax() // for this to work, ALL layers above need it too
	sgf.SetProp("columns", nWidgPerRow)
	sgf.SetProp("overflow", gi.OverflowScroll) // this still gives it true size during PrefSize

	// Configure Header
	hcfg := kit.TypeAndNameList{}
	if tv.ShowIndex {
		hcfg.Add(gi.KiT_Label, "head-idx")
	}
	for fli := 0; fli < tv.NCols; fli++ {
		labnm := fmt.Sprintf("head-%03d", fli)
		hcfg.Add(gi.KiT_Label, labnm)
	}
	if !tv.IsInactive() {
		hcfg.Add(gi.KiT_Label, "head-add")
		hcfg.Add(gi.KiT_Label, "head-del")
	}
	sgh.ConfigChildren(hcfg, false) // headers SHOULD be unique, but with labels..

	// at this point, we make one dummy row to get size of widgets

	sgf.DeleteChildren(true)
	sgf.Kids = make(ki.Slice, nWidgPerRow)

	itxt := fmt.Sprintf("%05d", 0)
	labnm := fmt.Sprintf("index-%v", itxt)

	if tv.ShowIndex {
		lbl := sgh.Child(0).(*gi.Label)
		lbl.Text = "Index"

		idxlab := &gi.Label{}
		sgf.SetChild(idxlab, 0, labnm)
		idxlab.Text = itxt
	}

	for fli := 0; fli < tv.NCols; fli++ {
		_, cc := etensor.Prjn2DCoords(tv.Tensor.ShapeObj(), tv.TsrLay.OddRow, 0, fli)
		sitxt := ""
		for i, ccc := range cc {
			sitxt += fmt.Sprintf("%03d", ccc)
			if i < len(cc)-1 {
				sitxt += ","
			}
		}
		hdr := sgh.Child(idxOff + fli).(*gi.Label)
		hdr.SetText(sitxt)

		fval := 1.0
		vv := giv.ToValueView(&fval, "")
		vv.SetSoloValue(reflect.ValueOf(&fval))
		vtyp := vv.WidgetType()
		valnm := fmt.Sprintf("value-%v.%v", fli, itxt)
		cidx := idxOff + fli
		widg := ki.NewOfType(vtyp).(gi.Node2D)
		sgf.SetChild(widg, cidx, valnm)
		vv.ConfigWidget(widg)
	}

	if !tv.IsInactive() {
		cidx := tv.NCols + idxOff
		if !tv.NoAdd {
			lbl := sgh.Child(cidx).(*gi.Label)
			lbl.Text = "+"
			lbl.Tooltip = "insert row"
			addnm := fmt.Sprintf("add-%v", itxt)
			addact := gi.Action{}
			sgf.SetChild(&addact, cidx, addnm)
			addact.SetIcon("plus")
			cidx++
		}
		if !tv.NoDelete {
			lbl := sgh.Child(cidx).(*gi.Label)
			lbl.Text = "-"
			lbl.Tooltip = "delete row"
			delnm := fmt.Sprintf("del-%v", itxt)
			delact := gi.Action{}
			sgf.SetChild(&delact, cidx, delnm)
			delact.SetIcon("minus")
			cidx++
		}
	}

	tv.ConfigScroll()
}

// LayoutSliceGrid does the proper layout of slice grid depending on allocated size
// returns true if UpdateSliceGrid should be called after this
func (tv *TensorView) LayoutSliceGrid() bool {
	sg := tv.SliceGrid()
	if tv.Tensor == nil {
		sg.DeleteChildren(true)
		return false
	}
	sz := tv.UpdtSliceSize()
	if sz == 0 {
		sg.DeleteChildren(true)
		return false
	}

	nWidgPerRow, _ := tv.RowWidgetNs()
	if len(sg.GridData) > 0 && len(sg.GridData[gi.Row]) > 0 {
		tv.RowHeight = sg.GridData[gi.Row][0].AllocSize + sg.Spacing.Dots
	}
	if tv.Sty.Font.Face == nil {
		tv.Sty.Font.OpenFont(&tv.Sty.UnContext)
	}
	tv.RowHeight = math32.Max(tv.RowHeight, tv.Sty.Font.Face.Metrics.Height)

	mvp := tv.ViewportSafe()
	if mvp != nil && mvp.HasFlag(int(gi.VpFlagPrefSizing)) {
		tv.VisRows = ints.MinInt(gi.LayoutPrefMaxRows, tv.SliceSize)
		tv.LayoutHeight = float32(tv.VisRows) * tv.RowHeight
	} else {
		sgHt := tv.AvailHeight()
		tv.LayoutHeight = sgHt
		if sgHt == 0 {
			return false
		}
		tv.VisRows = int(math32.Floor(sgHt / tv.RowHeight))
	}
	tv.DispRows = ints.MinInt(tv.SliceSize, tv.VisRows)

	nWidg := nWidgPerRow * tv.DispRows

	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)
	if tv.Values == nil || sg.NumChildren() != nWidg {
		sg.DeleteChildren(true)

		tv.Values = make([]giv.ValueView, tv.NCols*tv.DispRows)
		sg.Kids = make(ki.Slice, nWidg)
	}
	tv.ConfigScroll()
	tv.LayoutHeader()
	return true
}

// LayoutHeader updates the header layout based on field widths
func (tv *TensorView) LayoutHeader() {
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	nfld := tv.NCols + idxOff
	sgh := tv.SliceHeader()
	sgf := tv.SliceGrid()
	spc := sgh.Spacing.Dots
	gd := sgf.GridData[gi.Col]
	if gd == nil {
		return
	}
	sumwd := float32(0)
	for fli := 0; fli < nfld; fli++ {
		lbl := sgh.Child(fli).(gi.Node2D).AsWidget()
		wd := gd[fli].AllocSize - spc
		if fli == 0 {
			wd += spc
		}
		lbl.SetMinPrefWidth(units.NewValue(wd, units.Dot))
		lbl.SetProp("max-width", units.NewValue(wd, units.Dot))
		sumwd += wd
	}
	if !tv.IsInactive() {
		for fli := nfld; fli < nWidgPerRow; fli++ {
			lbl := sgh.Child(fli).(gi.Node2D).AsWidget()
			wd := gd[fli].AllocSize - spc
			lbl.SetMinPrefWidth(units.NewValue(wd, units.Dot))
			lbl.SetProp("max-width", units.NewValue(wd, units.Dot))
			sumwd += wd
		}
	}
	sgh.SetMinPrefWidth(units.NewValue(sumwd+spc, units.Dot))
}

// UpdateSliceGrid updates grid display -- robust to any time calling
func (tv *TensorView) UpdateSliceGrid() {
	if tv.Tensor == nil {
		return
	}
	sz := tv.UpdtSliceSize()
	if sz == 0 {
		return
	}
	sg := tv.SliceGrid()
	tv.DispRows = ints.MinInt(tv.SliceSize, tv.VisRows)

	nWidgPerRow, idxOff := tv.RowWidgetNs()
	nWidg := nWidgPerRow * tv.DispRows

	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)

	if tv.Values == nil || sg.NumChildren() != nWidg { // shouldn't happen..
		tv.LayoutSliceGrid()
		nWidg = nWidgPerRow * tv.DispRows
	}

	if sz > tv.DispRows {
		sb := tv.ScrollBar()
		tv.StartIdx = int(sb.Value)
		lastSt := sz - tv.DispRows
		tv.StartIdx = ints.MinInt(lastSt, tv.StartIdx)
		tv.StartIdx = ints.MaxInt(0, tv.StartIdx)
	} else {
		tv.StartIdx = 0
	}

	for ri := 0; ri < tv.DispRows; ri++ {
		ridx := ri * nWidgPerRow
		si := tv.StartIdx + ri // slice idx
		if !tv.TsrLay.TopZero {
			si = (tv.SliceSize - 1) - si
		}
		issel := tv.IdxIsSelected(si)
		itxt := fmt.Sprintf("%05d", ri)
		cr, _ := etensor.Prjn2DCoords(tv.Tensor.ShapeObj(), tv.TsrLay.OddRow, si, 0)
		sitxt := ""
		for i, crc := range cr {
			sitxt += fmt.Sprintf("%03d", crc)
			if i < len(cr)-1 {
				sitxt += ","
			}
		}
		labnm := fmt.Sprintf("index-%v", itxt)
		if tv.ShowIndex {
			var idxlab *gi.Label
			if sg.Kids[ridx] != nil {
				idxlab = sg.Kids[ridx].(*gi.Label)
			} else {
				idxlab = &gi.Label{}
				sg.SetChild(idxlab, ridx, labnm)
				idxlab.SetProp("tv-row", ri)
				idxlab.Selectable = true
				idxlab.Redrawable = true
				idxlab.Sty.Template = "View.IndexLabel"
				idxlab.WidgetSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					if sig == int64(gi.WidgetSelected) {
						wbb := send.(gi.Node2D).AsWidget()
						row := wbb.Prop("tv-row").(int)
						tvv := recv.Embed(KiT_TensorView).(*TensorView)
						tvv.UpdateSelectRow(row, wbb.IsSelected())
					}
				})
			}
			idxlab.CurBgColor = gi.Prefs.Colors.Background
			idxlab.SetText(sitxt)
			idxlab.SetSelectedState(issel)
		}

		for fli := 0; fli < tv.NCols; fli++ {
			fval := etensor.Prjn2DVal(tv.Tensor, tv.TsrLay.OddRow, si, fli)
			vvi := ri*tv.NCols + fli
			var vv giv.ValueView
			if tv.Values[vvi] == nil {
				vv = giv.ToValueView(&fval, "")
				vv.SetSoloValue(reflect.ValueOf(&fval))
				tv.Values[vvi] = vv
				vv.SetProp("tv-row", ri)
				vv.SetProp("tv-col", fli)
			} else {
				vv = tv.Values[vvi]
				vv.SetSoloValue(reflect.ValueOf(&fval))
			}

			vtyp := vv.WidgetType()
			valnm := fmt.Sprintf("value-%v.%v", fli, itxt)
			cidx := ridx + idxOff + fli
			var widg gi.Node2D
			if sg.Kids[cidx] != nil {
				widg = sg.Kids[cidx].(gi.Node2D)
				vv.UpdateWidget()
				if tv.IsInactive() {
					widg.AsNode2D().SetInactive()
				}
				widg.AsNode2D().SetSelectedState(issel)
			} else {
				widg = ki.NewOfType(vtyp).(gi.Node2D)
				sg.SetChild(widg, cidx, valnm)
				vv.ConfigWidget(widg)
				wb := widg.AsWidget()
				if wb != nil {
					wb.SetProp("tv-row", ri)
					wb.ClearSelected()
					wb.WidgetSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						if sig == int64(gi.WidgetSelected) || sig == int64(gi.WidgetFocused) {
							wbb := send.(gi.Node2D).AsWidget()
							row := wbb.Prop("tv-row").(int)
							tvv := recv.Embed(KiT_TensorView).(*TensorView)
							if sig != int64(gi.WidgetFocused) || !tvv.InFocusGrab {
								tvv.UpdateSelectRow(row, wbb.IsSelected())
							}
						}
					})
				}
				if tv.IsInactive() {
					widg.AsNode2D().SetInactive()
				} else {
					vvb := vv.AsValueViewBase()
					vvb.ViewSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						tvv, _ := recv.Embed(KiT_TensorView).(*TensorView)
						tvv.SetChanged()
						vvv := send.(giv.ValueView).AsValueViewBase()
						row := vvv.Prop("tv-row").(int)
						rsi := (tvv.StartIdx + row)
						if !tvv.TsrLay.TopZero {
							rsi = (tvv.SliceSize - 1) - rsi
						}
						col := vvv.Prop("tv-col").(int)
						npv := kit.NonPtrValue(vvv.Value)
						fv, ok := kit.ToFloat(npv.Interface())
						if ok {
							etensor.Prjn2DSet(tvv.Tensor, tvv.TsrLay.OddRow, rsi, col, fv)
							tvv.ViewSig.Emit(tvv.This(), 0, nil)
						}
					})
				}
			}
		}
	}

	if tv.IsInactive() && tv.SelectedIdx >= 0 {
		tv.SelectIdx(tv.SelectedIdx)
	}
	tv.UpdateScroll()
}

func (tv *TensorView) StyleRow(svnp reflect.Value, widg gi.Node2D, idx, fidx int, vv giv.ValueView) {
}

// SliceNewAt inserts a new blank element at given index in the slice -- -1
// means the end
func (tv *TensorView) SliceNewAt(idx int) {
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := tv.UpdateStart()
	defer tv.UpdateEnd(updt)

	// todo: insert row -- do we even have this??  no!
	// kit.SliceNewAt(tv.Slice, idx)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	tv.This().(giv.SliceViewer).LayoutSliceGrid()
	tv.This().(giv.SliceViewer).UpdateSliceGrid()
	tv.ViewSig.Emit(tv.This(), 0, nil)
}

// SliceDeleteAt deletes element at given index from slice -- doupdt means
// call UpdateSliceGrid to update display
func (tv *TensorView) SliceDeleteAt(idx int, doupdt bool) {
	if idx < 0 {
		return
	}
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := tv.UpdateStart()
	defer tv.UpdateEnd(updt)

	// kit.SliceDeleteAt(tv.Slice, idx)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	if doupdt {
		tv.This().(giv.SliceViewer).LayoutSliceGrid()
		tv.This().(giv.SliceViewer).UpdateSliceGrid()
	}
	tv.ViewSig.Emit(tv.This(), 0, nil)
}

// ConfigToolbar configures the toolbar actions
func (tv *TensorView) ConfigToolbar() {
	if tv.Tensor == nil {
		return
	}
	if tv.ToolbarSlice == tv.Tensor {
		return
	}
	tb := tv.ToolBar()
	if len(*tb.Children()) == 0 {
		tb.SetStretchMaxWidth()
		tb.AddAction(gi.ActOpts{Label: "UpdtView", Icon: "update", Tooltip: "update the view to reflect current state of tensor"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TensorView).(*TensorView)
				tvv.Update()
			})
		tb.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "configure the view"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TensorView).(*TensorView)
				giv.StructViewDialog(tv.ViewportSafe(), &tvv.TsrLay, giv.DlgOpts{Title: "TensorView Display Options", Ok: true, Cancel: true},
					tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						tvvv := recv.Embed(KiT_TensorView).(*TensorView)
						tvvv.UpdateSliceGrid()
					})
			})
		tb.AddAction(gi.ActOpts{Label: "Grid", Icon: "file-sheet", Tooltip: "open a grid view of the tensor -- with a grid of colored squares representing values"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TensorView).(*TensorView)
				TensorGridDialog(tv.ViewportSafe(), tvv.Tensor, giv.DlgOpts{Title: "TensorGrid", Ok: false, Cancel: false},
					tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						tvvv := recv.Embed(KiT_TensorView).(*TensorView)
						tvvv.UpdateSliceGrid()
					})
			})
	}
	ndef := 3
	sz := len(*tb.Children())
	if sz > ndef {
		for i := sz - 1; i >= ndef; i-- {
			tb.DeleteChildAtIndex(i, true)
		}
	}
	mvp := tv.ViewportSafe()
	if giv.HasToolBarView(tv.Slice) && mvp != nil {
		giv.ToolBarView(tv.Slice, mvp, tb)
	}
	tv.ToolbarSlice = tv.Tensor
}

func (tv *TensorView) Layout2D(parBBox image.Rectangle, iter int) bool {
	redo := tv.Frame.Layout2D(parBBox, iter)
	if !tv.IsConfiged() {
		return redo
	}
	tv.LayoutHeader()
	tv.SliceHeader().Layout2D(parBBox, iter)
	return redo
}

// RowFirstVisWidget returns the first visible widget for given row (could be
// index or not) -- false if out of range
func (tv *TensorView) RowFirstVisWidget(row int) (*gi.WidgetBase, bool) {
	if !tv.IsRowInBounds(row) {
		return nil, false
	}
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	sg := tv.SliceGrid()
	widg := sg.Kids[row*nWidgPerRow].(gi.Node2D).AsWidget()
	if widg.VpBBox != image.ZR {
		return widg, true
	}
	ridx := nWidgPerRow * row
	for fli := 0; fli < tv.NCols; fli++ {
		widg := sg.Child(ridx + idxOff + fli).(gi.Node2D).AsWidget()
		if widg.VpBBox != image.ZR {
			return widg, true
		}
	}
	return nil, false
}

// RowGrabFocus grabs the focus for the first focusable widget in given row --
// returns that element or nil if not successful -- note: grid must have
// already rendered for focus to be grabbed!
func (tv *TensorView) RowGrabFocus(row int) *gi.WidgetBase {
	if !tv.IsRowInBounds(row) || tv.InFocusGrab { // range check
		return nil
	}
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	ridx := nWidgPerRow * row
	sg := tv.SliceGrid()
	// first check if we already have focus
	for fli := 0; fli < tv.NCols; fli++ {
		widg := sg.Child(ridx + idxOff + fli).(gi.Node2D).AsWidget()
		if widg.HasFocus() || widg.ContainsFocus() {
			return widg
		}
	}
	tv.InFocusGrab = true
	defer func() { tv.InFocusGrab = false }()
	for fli := 0; fli < tv.NCols; fli++ {
		widg := sg.Child(ridx + idxOff + fli).(gi.Node2D).AsWidget()
		if widg.CanFocus() {
			widg.GrabFocus()
			return widg
		}
	}
	return nil
}

// SelectRowWidgets sets the selection state of given row of widgets
func (tv *TensorView) SelectRowWidgets(row int, sel bool) {
	if row < 0 {
		return
	}
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	sg := tv.SliceGrid()
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	ridx := row * nWidgPerRow
	for fli := 0; fli < tv.NCols; fli++ {
		seldx := ridx + idxOff + fli
		if sg.Kids.IsValidIndex(seldx) == nil {
			widg := sg.Child(seldx).(gi.Node2D).AsNode2D()
			widg.SetSelectedState(sel)
			widg.UpdateSig()
		}
	}
	if tv.ShowIndex {
		if sg.Kids.IsValidIndex(ridx) == nil {
			widg := sg.Child(ridx).(gi.Node2D).AsNode2D()
			widg.SetSelectedState(sel)
			widg.UpdateSig()
		}
	}
}

// CopySelToMime copies selected rows to mime data
func (tv *TensorView) CopySelToMime() mimedata.Mimes {
	return nil
}

// PasteAssign assigns mime data (only the first one!) to this idx
func (tv *TensorView) PasteAssign(md mimedata.Mimes, idx int) {
	// todo
}

// PasteAtIdx inserts object(s) from mime data at (before) given slice index
func (tv *TensorView) PasteAtIdx(md mimedata.Mimes, idx int) {
	// todo
}

func (tv *TensorView) ItemCtxtMenu(idx int) {
}

// // SelectFieldVal sets SelField and SelVal and attempts to find corresponding
// // row, setting SelectedIdx and selecting row if found -- returns true if
// // found, false otherwise
// func (tv *TensorView) SelectFieldVal(fld, val string) bool {
// 	tv.SelField = fld
// 	tv.SelVal = val
// 	if tv.SelField != "" && tv.SelVal != nil {
// 		idx, _ := StructSliceIdxByValue(tv.Slice, tv.SelField, tv.SelVal)
// 		if idx >= 0 {
// 			tv.ScrollToIdx(idx)
// 			tv.UpdateSelectIdx(idx, true)
// 			return true
// 		}
// 	}
// 	return false
// }
