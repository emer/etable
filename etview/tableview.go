// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"fmt"
	"image"
	"reflect"
	"strings"

	"github.com/chewxy/math32"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/cursor"
	"github.com/goki/gi/oswin/mimedata"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// etview.TableView provides a GUI interface for etable.Table's
type TableView struct {
	giv.SliceViewBase
	Table      *etable.Table       `desc:"the table that we're a view of"`
	TsrDisp    TensorDisp          `desc:"overall display options for tensor display"`
	ColTsrDisp map[int]*TensorDisp `desc:"per column tensor display"`
	NCols      int                 `inactive:"+" desc:"number of columns in table (as of last update)"`
	SortIdx    int                 `desc:"current sort index"`
	SortDesc   bool                `desc:"whether current sort order is descending"`
}

var KiT_TableView = kit.Types.AddType(&TableView{}, TableViewProps)

// AddNewTableView adds a new tableview to given parent node, with given name.
func AddNewTableView(parent ki.Ki, name string) *TableView {
	return parent.AddNewChild(KiT_TableView, name).(*TableView)
}

// check for interface impl
var _ giv.SliceViewer = (*TableView)(nil)

// SetTable sets the source table that we are viewing
func (tv *TableView) SetTable(et *etable.Table, tmpSave giv.ValueView) {
	updt := false
	if et == nil {
		return
	}
	if tv.Table != et {
		if op, has := et.MetaData["read-only"]; has {
			if op == "+" || op == "true" {
				tv.SetInactive()
			} else {
				tv.ClearInactive()
			}
		}
		tv.ColTsrDisp = make(map[int]*TensorDisp)
		tv.TsrDisp.Defaults()
		if !tv.IsInactive() {
			tv.SelectedIdx = -1
		}
		tv.StartIdx = 0
		tv.SortIdx = -1
		tv.SortDesc = false
		tv.Table = et
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

var TableViewProps = ki.Props{
	"background-color": &gi.Prefs.Colors.Background,
	"color":            &gi.Prefs.Colors.Font,
	"max-width":        -1,
	"max-height":       -1,
}

// IsConfiged returns true if the widget is fully configured
func (tv *TableView) IsConfiged() bool {
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
func (tv *TableView) Config() {
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

func (tv *TableView) UpdtSliceSize() int {
	tv.SliceSize = tv.Table.Rows
	tv.NCols = tv.Table.NumCols()
	return tv.SliceSize
}

// SliceFrame returns the outer frame widget, which contains all the header,
// fields and values
func (tv *TableView) SliceFrame() *gi.Frame {
	return tv.ChildByName("frame", 0).(*gi.Frame)
}

// GridLayout returns the SliceGrid grid-layout widget, with grid and scrollbar
func (tv *TableView) GridLayout() *gi.Layout {
	return tv.SliceFrame().ChildByName("grid-lay", 0).(*gi.Layout)
}

// SliceGrid returns the SliceGrid grid frame widget, which contains all the
// fields and values, within SliceFrame
func (tv *TableView) SliceGrid() *gi.Frame {
	return tv.GridLayout().ChildByName("grid", 0).(*gi.Frame)
}

// ScrollBar returns the SliceGrid scrollbar
func (tv *TableView) ScrollBar() *gi.ScrollBar {
	return tv.GridLayout().ChildByName("scrollbar", 1).(*gi.ScrollBar)
}

// SliceHeader returns the Toolbar header for slice grid
func (tv *TableView) SliceHeader() *gi.ToolBar {
	return tv.SliceFrame().Child(0).(*gi.ToolBar)
}

// ToolBar returns the toolbar widget
func (tv *TableView) ToolBar() *gi.ToolBar {
	return tv.ChildByName("toolbar", 0).(*gi.ToolBar)
}

// RowWidgetNs returns number of widgets per row and offset for index label
func (tv *TableView) RowWidgetNs() (nWidgPerRow, idxOff int) {
	nWidgPerRow = 1 + tv.NCols
	if !tv.IsInactive() {
		if !tv.NoAdd {
			nWidgPerRow++
		}
		if !tv.NoDelete {
			nWidgPerRow++
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
func (tv *TableView) ConfigSliceGrid() {
	if tv.Table == nil {
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
	sg.SetMinPrefWidth(units.NewEm(10))
	sg.SetStretchMax() // for this to work, ALL layers above need it too

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
	sgf.SetProp("spacing", gi.StdDialogVSpaceUnits)

	// Configure Header
	hcfg := kit.TypeAndNameList{}
	if tv.ShowIndex {
		hcfg.Add(gi.KiT_Label, "head-idx")
	}
	for fli := 0; fli < tv.NCols; fli++ {
		labnm := fmt.Sprintf("head-%v", tv.Table.ColNames[fli])
		hcfg.Add(gi.KiT_Action, labnm)
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
		col := tv.Table.Cols[fli]
		colnm := tv.Table.ColNames[fli]
		hdr := sgh.Child(idxOff + fli).(*gi.Action)
		hdr.SetText(colnm)
		if fli == tv.SortIdx {
			if tv.SortDesc {
				hdr.SetIcon("wedge-down")
			} else {
				hdr.SetIcon("wedge-up")
			}
		}
		hdr.Data = fli
		hdr.Tooltip = colnm + " (click to sort / toggle sort direction by this column) Type: " + col.DataType().String()

		if dsc, has := tv.Table.MetaData[colnm+":desc"]; has {
			hdr.Tooltip += ": " + dsc
		}
		var vv giv.ValueView
		if stsr, isstr := col.(*etensor.String); isstr {
			vv = giv.ToValueView(&stsr.Values[0], "")
			vv.SetSliceValue(reflect.ValueOf(&stsr.Values[0]), stsr.Values, 0, tv.TmpSave, tv.ViewPath)
			hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TableView).(*TableView)
				act := send.(*gi.Action)
				fldIdx := act.Data.(int)
				tvv.SortSliceAction(fldIdx)
			})
		} else {
			if col.NumDims() == 1 {
				fval := 1.0
				vv = giv.ToValueView(&fval, "")
				vv.SetStandaloneValue(reflect.ValueOf(&fval))
				hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					tvv := recv.Embed(KiT_TableView).(*TableView)
					act := send.(*gi.Action)
					fldIdx := act.Data.(int)
					tvv.SortSliceAction(fldIdx)
				})
			} else {
				cell := tv.Table.CellTensorIdx(fli, 0)
				tvv := &TensorGridValueView{}
				tvv.Init(tvv)
				vv = tvv
				vv.SetStandaloneValue(reflect.ValueOf(cell))
				hdr.Tooltip = "(click to edit display parameters for this column)"
				hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					tvv := recv.Embed(KiT_TableView).(*TableView)
					act := send.(*gi.Action)
					fldIdx := act.Data.(int)
					tvv.TensorDispAction(fldIdx)
				})
			}
		}
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

	// if tv.SortIdx >= 0 {
	// 	rawIdx := tv.VisFields[tv.SortIdx].Index
	// 	kit.StructSliceSort(tv.Slice, rawIdx, !tv.SortDesc)
	// }

	tv.ConfigScroll()
}

// LayoutSliceGrid does the proper layout of slice grid depending on allocated size
// returns true if UpdateSliceGrid should be called after this
func (tv *TableView) LayoutSliceGrid() bool {
	sg := tv.SliceGrid()
	if tv.Table == nil {
		sg.DeleteChildren(true)
		return false
	}
	sz := tv.UpdtSliceSize()
	if sz == 0 {
		sg.DeleteChildren(true)
		return false
	}

	sgHt := tv.AvailHeight()
	tv.LayoutHeight = sgHt
	if sgHt == 0 {
		return false
	}

	nWidgPerRow, _ := tv.RowWidgetNs()
	tv.RowHeight = sg.GridData[gi.Row][0].AllocSize + sg.Spacing.Dots
	tv.VisRows = int(math32.Floor(sgHt / tv.RowHeight))
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
func (tv *TableView) LayoutHeader() {
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	nfld := tv.NCols + idxOff
	sgh := tv.SliceHeader()
	sgf := tv.SliceGrid()
	spc := sgf.Spacing.Dots
	if len(sgf.Kids) >= nfld {
		sumwd := float32(0)
		for fli := 0; fli < nfld; fli++ {
			lbl := sgh.Child(fli).(gi.Node2D).AsWidget()
			wd := sgf.GridData[gi.Col][fli].AllocSize
			lbl.SetMinPrefWidth(units.NewValue(wd+spc, units.Dot))
			lbl.SetProp("max-width", units.NewValue(wd+spc, units.Dot))
			sumwd += wd + spc
		}
		if !tv.IsInactive() {
			for fli := nfld; fli < nWidgPerRow; fli++ {
				lbl := sgh.Child(fli).(gi.Node2D).AsWidget()
				wd := sgf.GridData[gi.Col][fli].AllocSize
				lbl.SetMinPrefWidth(units.NewValue(wd+spc, units.Dot))
				lbl.SetProp("max-width", units.NewValue(wd+spc, units.Dot))
				sumwd += wd + spc
			}
		}
		sgh.SetMinPrefWidth(units.NewValue(sumwd, units.Dot))
	}
}

// UpdateSliceGrid updates grid display -- robust to any time calling
func (tv *TableView) UpdateSliceGrid() {
	if tv.Table == nil {
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

	if tv.Viewport != nil && tv.Viewport.Win != nil {
		wupdt := tv.Viewport.Win.UpdateStart()
		defer tv.Viewport.Win.UpdateEnd(wupdt)
	}

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

	for i := 0; i < tv.DispRows; i++ {
		ridx := i * nWidgPerRow
		si := tv.StartIdx + i // slice idx
		issel := tv.IdxIsSelected(si)

		itxt := fmt.Sprintf("%05d", i)
		sitxt := fmt.Sprintf("%05d", si)
		labnm := fmt.Sprintf("index-%v", itxt)
		if tv.ShowIndex {
			var idxlab *gi.Label
			if sg.Kids[ridx] != nil {
				idxlab = sg.Kids[ridx].(*gi.Label)
			} else {
				idxlab = &gi.Label{}
				sg.SetChild(idxlab, ridx, labnm)
				idxlab.SetProp("tv-row", i)
				idxlab.Selectable = true
				idxlab.Redrawable = true
				idxlab.Sty.Template = "View.IndexLabel"
				idxlab.WidgetSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					if sig == int64(gi.WidgetSelected) {
						wbb := send.(gi.Node2D).AsWidget()
						row := wbb.Prop("tv-row").(int)
						tvv := recv.Embed(KiT_TableView).(*TableView)
						tvv.UpdateSelectRow(row, wbb.IsSelected())
					}
				})
			}
			idxlab.CurBgColor = gi.Prefs.Colors.Background
			idxlab.SetText(sitxt)
			idxlab.SetSelectedState(issel)
		}

		for fli := 0; fli < tv.NCols; fli++ {
			col := tv.Table.Cols[fli]
			// colnm := tv.Table.ColNames[fli]
			tdsp := tv.ColTensorDisp(fli)

			vvi := i*tv.NCols + fli
			var vv giv.ValueView
			if tv.Values[vvi] == nil {
				if stsr, isstr := col.(*etensor.String); isstr {
					sval := stsr.Values[i]
					vv = giv.ToValueView(&sval, "")
					vv.SetProp("tv-row", i)
					vv.SetProp("tv-col", fli)
					vv.SetStandaloneValue(reflect.ValueOf(&sval))
					vv.AsValueViewBase().ViewSig.ConnectOnly(tv.This(),
						func(recv, send ki.Ki, sig int64, data interface{}) {
							tvv, _ := recv.Embed(KiT_TableView).(*TableView)
							tvv.SetChanged()
							vvv := send.(giv.ValueView).AsValueViewBase()
							row := vvv.Prop("tv-row").(int)
							col := vvv.Prop("tv-col").(int)
							npv := kit.NonPtrValue(vvv.Value)
							sv := kit.ToString(npv.Interface())
							tv.Table.SetCellStringIdx(col, tvv.StartIdx+row, sv)
							tvv.ViewSig.Emit(tvv.This(), 0, nil)
						})
				} else {
					if col.NumDims() == 1 {
						fval := col.FloatVal1D(si)
						vv = giv.ToValueView(&fval, "")
						vv.SetProp("tv-row", i)
						vv.SetProp("tv-col", fli)
						vv.SetStandaloneValue(reflect.ValueOf(&fval))
						vv.AsValueViewBase().ViewSig.ConnectOnly(tv.This(),
							func(recv, send ki.Ki, sig int64, data interface{}) {
								tvv, _ := recv.Embed(KiT_TableView).(*TableView)
								tvv.SetChanged()
								vvv := send.(giv.ValueView).AsValueViewBase()
								row := vvv.Prop("tv-row").(int)
								col := vvv.Prop("tv-col").(int)
								npv := kit.NonPtrValue(vvv.Value)
								fv, ok := kit.ToFloat(npv.Interface())
								if ok {
									tv.Table.SetCellFloatIdx(col, tvv.StartIdx+row, fv)
									tvv.ViewSig.Emit(tvv.This(), 0, nil)
								}
							})
					} else {
						cell := tv.Table.CellTensorIdx(fli, 0)
						tvv := &TensorGridValueView{}
						tvv.Init(tvv)
						vv = tvv
						vv.SetStandaloneValue(reflect.ValueOf(cell))
					}
				}
				tv.Values[vvi] = vv
			} else {
				vv = tv.Values[vvi]
				if stsr, isstr := col.(*etensor.String); isstr {
					vv.SetStandaloneValue(reflect.ValueOf(&stsr.Values[si]))
				} else {
					if col.NumDims() == 1 {
						fval := col.FloatVal1D(si)
						vv.SetStandaloneValue(reflect.ValueOf(&fval))
					} else {
						cell := tv.Table.CellTensorIdx(fli, si)
						vv.SetStandaloneValue(reflect.ValueOf(cell))
					}
				}
			}

			vtyp := vv.WidgetType()
			valnm := fmt.Sprintf("value-%v.%v", fli, itxt)
			cidx := ridx + idxOff + fli
			var widg gi.Node2D
			if sg.Kids[cidx] != nil {
				widg = sg.Kids[cidx].(gi.Node2D)
				wn := widg.AsNode2D()
				if tv.IsInactive() {
					wn.SetInactive()
				}
				if col.IsNull1D(i) { // todo: not working:
					wn.SetProp("background-color", gi.Prefs.Colors.Highlight)
				} else {
					wn.DeleteProp("background-color")
				}
				widg.AsNode2D().SetSelectedState(issel)
				vv.UpdateWidget()
			} else {
				widg = ki.NewOfType(vtyp).(gi.Node2D)
				sg.SetChild(widg, cidx, valnm)
				vv.ConfigWidget(widg)
				wb := widg.AsWidget()
				if wb != nil {
					wb.SetProp("tv-row", i)
					wb.SetProp("vertical-align", gi.AlignTop)
					wb.ClearSelected()
					wb.WidgetSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						if sig == int64(gi.WidgetSelected) || sig == int64(gi.WidgetFocused) {
							wbb := send.(gi.Node2D).AsWidget()
							row := wbb.Prop("tv-row").(int)
							tvv := recv.Embed(KiT_TableView).(*TableView)
							if sig != int64(gi.WidgetFocused) || !tvv.InFocusGrab {
								tvv.UpdateSelectRow(row, wbb.IsSelected())
							}
						}
					})
					if tv.IsInactive() {
						wb.SetInactive()
					}
					if col.IsNull1D(i) {
						wb.SetProp("background-color", gi.Prefs.Colors.Highlight)
					} else {
						wb.DeleteProp("background-color")
					}
				}
			}
			if tgw, istg := widg.(*TensorGrid); istg { // always update disp params
				tgw.Disp = *tdsp
			}
		}

		if !tv.IsInactive() {
			cidx := ridx + tv.NCols + idxOff
			if !tv.NoAdd {
				if sg.Kids[cidx] == nil {
					addnm := fmt.Sprintf("add-%v", itxt)
					addact := gi.Action{}
					sg.SetChild(&addact, cidx, addnm)
					addact.SetIcon("plus")
					addact.Tooltip = "insert a new element at this index"
					addact.Data = i
					addact.Sty.Template = "etview.TableView.AddAction"
					addact.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						act := send.(*gi.Action)
						tvv := recv.Embed(KiT_TableView).(*TableView)
						tvv.SliceNewAtRow(act.Data.(int) + 1)
					})
				}
				cidx++
			}
			if !tv.NoDelete {
				if sg.Kids[cidx] == nil {
					delnm := fmt.Sprintf("del-%v", itxt)
					delact := gi.Action{}
					sg.SetChild(&delact, cidx, delnm)
					delact.SetIcon("minus")
					delact.Tooltip = "delete this element"
					delact.Data = i
					delact.Sty.Template = "etview.TableView.DelAction"
					delact.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						act := send.(*gi.Action)
						tvv := recv.Embed(KiT_TableView).(*TableView)
						tvv.SliceDeleteAtRow(act.Data.(int), true)
					})
				}
				cidx++
			}
		}
	}

	if tv.IsInactive() && tv.SelectedIdx >= 0 {
		tv.SelectIdx(tv.SelectedIdx)
	}
	tv.UpdateScroll()
}

// ColTensorDisp returns tensor display parameters for this column
// either the overall defaults or the per-column if set
func (tv *TableView) ColTensorDisp(col int) *TensorDisp {
	if ctd, has := tv.ColTsrDisp[col]; has {
		return ctd
	}
	if tv.Table != nil {
		cl := tv.Table.Cols[col]
		if len(cl.MetaDataMap()) > 0 {
			return tv.SetColTensorDisp(col)
		}
	}
	return &tv.TsrDisp
}

// SetColTensorDisp sets per-column tensor display params and returns them
// if already set, just returns them
func (tv *TableView) SetColTensorDisp(col int) *TensorDisp {
	if ctd, has := tv.ColTsrDisp[col]; has {
		return ctd
	}
	ctd := &TensorDisp{}
	*ctd = tv.TsrDisp
	if tv.Table != nil {
		cl := tv.Table.Cols[col]
		ctd.FmMeta(cl)
	}
	tv.ColTsrDisp[col] = ctd
	return ctd
}

func (tv *TableView) StyleRow(svnp reflect.Value, widg gi.Node2D, idx, fidx int, vv giv.ValueView) {
}

// SliceNewAt inserts a new blank element at given index in the slice -- -1
// means the end
func (tv *TableView) SliceNewAt(idx int) {
	wupdt := tv.Viewport.Win.UpdateStart()
	defer tv.Viewport.Win.UpdateEnd(wupdt)

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
func (tv *TableView) SliceDeleteAt(idx int, doupdt bool) {
	if idx < 0 {
		return
	}
	wupdt := tv.Viewport.Win.UpdateStart()
	defer tv.Viewport.Win.UpdateEnd(wupdt)

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

// SortSliceAction sorts the slice for given field index -- toggles ascending
// vs. descending if already sorting on this dimension
func (tv *TableView) SortSliceAction(fldIdx int) {
	oswin.TheApp.Cursor(tv.Viewport.Win.OSWin).Push(cursor.Wait)
	defer oswin.TheApp.Cursor(tv.Viewport.Win.OSWin).Pop()

	wupdt := tv.Viewport.Win.UpdateStart()
	defer tv.Viewport.Win.UpdateEnd(wupdt)

	updt := tv.UpdateStart()
	sgh := tv.SliceHeader()
	sgh.SetFullReRender()
	_, idxOff := tv.RowWidgetNs()

	ascending := true

	for fli := 0; fli < tv.NCols; fli++ {
		hdr := sgh.Child(idxOff + fli).(*gi.Action)
		if fli == fldIdx {
			if tv.SortIdx == fli {
				tv.SortDesc = !tv.SortDesc
				ascending = !tv.SortDesc
			} else {
				tv.SortDesc = false
			}
			if ascending {
				hdr.SetIcon("wedge-up")
			} else {
				hdr.SetIcon("wedge-down")
			}
		} else {
			hdr.SetIcon("none")
		}
	}

	tv.SortIdx = fldIdx

	// kit.StructSliceSort(tv.Slice, rawIdx, !tv.SortDesc)
	tv.UpdateSliceGrid()
	tv.UpdateEnd(updt)
}

// TensorDispAction allows user to select tensor display options for column
// pass -1 for global params for the entire table
func (tv *TableView) TensorDispAction(fldIdx int) {
	wupdt := tv.Viewport.Win.UpdateStart()
	defer tv.Viewport.Win.UpdateEnd(wupdt)

	updt := tv.UpdateStart()
	ctd := &tv.TsrDisp
	if fldIdx >= 0 {
		ctd = tv.SetColTensorDisp(fldIdx)
	}
	giv.StructViewDialog(tv.Viewport, ctd, giv.DlgOpts{Title: "TensorGrid Display Options", Ok: true, Cancel: true},
		tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			tvv := recv.Embed(KiT_TableView).(*TableView)
			tvv.UpdateSliceGrid()
		})

	tv.UpdateSliceGrid()
	tv.UpdateEnd(updt)
}

// ConfigToolbar configures the toolbar actions
func (tv *TableView) ConfigToolbar() {
	if tv.Table == nil {
		return
	}
	if tv.ToolbarSlice == tv.Table {
		return
	}
	tb := tv.ToolBar()
	if len(*tb.Children()) == 0 {
		tb.SetStretchMaxWidth()
		tb.AddAction(gi.ActOpts{Label: "UpdtView", Icon: "update", Tooltip: "update the view to reflect current state of table"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TableView).(*TableView)
				tvv.Update()
			})
		tb.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "configure the view"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TableView).(*TableView)
				tvv.TensorDispAction(-1)
			})
	}
	ndef := 2
	sz := len(*tb.Children())
	if sz > ndef {
		for i := sz - 1; i >= ndef; i-- {
			tb.DeleteChildAtIndex(i, true)
		}
	}
	if giv.HasToolBarView(tv.Table) && tv.Viewport != nil {
		giv.ToolBarView(tv.Table, tv.Viewport, tb)
	}
	tv.ToolbarSlice = tv.Table
}

// SortFieldName returns the name of the field being sorted, along with :up or
// :down depending on descending
func (tv *TableView) SortFieldName() string {
	if tv.SortIdx >= 0 && tv.SortIdx < tv.NCols {
		nm := tv.Table.ColNames[tv.SortIdx]
		if tv.SortDesc {
			nm += ":down"
		} else {
			nm += ":up"
		}
		return nm
	}
	return ""
}

// SetSortField sets sorting to happen on given field and direction -- see
// SortFieldName for details
func (tv *TableView) SetSortFieldName(nm string) {
	if nm == "" {
		return
	}
	spnm := strings.Split(nm, ":")
	for fli := 0; fli < tv.NCols; fli++ {
		colnm := tv.Table.ColNames[fli]
		if colnm == spnm[0] {
			tv.SortIdx = fli
		}
	}
	if len(spnm) == 2 {
		if spnm[1] == "down" {
			tv.SortDesc = true
		} else {
			tv.SortDesc = false
		}
	}
}

func (tv *TableView) Layout2D(parBBox image.Rectangle, iter int) bool {
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
func (tv *TableView) RowFirstVisWidget(row int) (*gi.WidgetBase, bool) {
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
func (tv *TableView) RowGrabFocus(row int) *gi.WidgetBase {
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
func (tv *TableView) SelectRowWidgets(row int, sel bool) {
	if row < 0 {
		return
	}
	wupdt := tv.Viewport.Win.UpdateStart()
	defer tv.Viewport.Win.UpdateEnd(wupdt)

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
func (tv *TableView) CopySelToMime() mimedata.Mimes {
	return nil
}

// PasteAssign assigns mime data (only the first one!) to this idx
func (tv *TableView) PasteAssign(md mimedata.Mimes, idx int) {
	// todo
}

// PasteAtIdx inserts object(s) from mime data at (before) given slice index
func (tv *TableView) PasteAtIdx(md mimedata.Mimes, idx int) {
	// todo
}

func (tv *TableView) ItemCtxtMenu(idx int) {
}

// // SelectFieldVal sets SelField and SelVal and attempts to find corresponding
// // row, setting SelectedIdx and selecting row if found -- returns true if
// // found, false otherwise
// func (tv *TableView) SelectFieldVal(fld, val string) bool {
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
