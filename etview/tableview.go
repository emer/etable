// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"log"
	"reflect"
	"strings"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/girl"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/cursor"
	"github.com/goki/gi/oswin/mimedata"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
	"github.com/goki/pi/filecat"
)

// etview.TableView provides a GUI interface for etable.Table's
type TableView struct {
	giv.SliceViewBase
	Table      *etable.IdxView     `desc:"the idx view of the table that we're a view of"`
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

// SetTable sets the source table that we are viewing, using a sequential IdxView
// and then configures the display
func (tv *TableView) SetTable(et *etable.Table, tmpSave giv.ValueView) {
	if et == nil {
		return
	}
	tv.Table = etable.NewIdxView(et)
	tv.TmpSave = tmpSave
	tv.TableConfig()
}

// SetTableView sets the source IdxView of a table (using a copy so original is not modified)
// and then configures the display
func (tv *TableView) SetTableView(ix *etable.IdxView, tmpSave giv.ValueView) {
	if ix == nil {
		return
	}
	tv.Table = ix.Clone() // always copy
	tv.TmpSave = tmpSave
	tv.TableConfig()
}

// TableConfig does all the configuration for a new Table view
func (tv *TableView) TableConfig() {
	if op, has := tv.Table.Table.MetaData["read-only"]; has {
		if op == "+" || op == "true" {
			tv.SetInactive()
		} else {
			tv.ClearInactive()
		}
	}
	tv.ColTsrDisp = make(map[int]*TensorDisp)
	tv.TsrDisp.Defaults()
	tv.TsrDisp.BotRtSpace.Set(4, units.Px)
	if !tv.IsInactive() {
		tv.SelectedIdx = -1
	}
	tv.StartIdx = 0
	tv.SortIdx = -1
	tv.SortDesc = false
	updt := tv.UpdateStart()
	tv.ResetSelectedIdxs()
	tv.SelectMode = false
	tv.SetFullReRender()
	tv.ShowIndex = true
	if sidxp, err := tv.PropTry("index"); err == nil {
		tv.ShowIndex, _ = kit.ToBool(sidxp)
	}
	tv.InactKeyNav = true
	if siknp, err := tv.PropTry("inact-key-nav"); err == nil {
		tv.InactKeyNav, _ = kit.ToBool(siknp)
	}
	tv.Config()
	tv.UpdateEnd(updt)
}

var TableViewProps = ki.Props{
	"EnumType:Flag":    gi.KiT_NodeFlags,
	"background-color": &gi.Prefs.Colors.Background,
	"color":            &gi.Prefs.Colors.Font,
	"max-width":        -1,
	"max-height":       -1,
}

// UpdateTable updates view of Table -- regenerates indexes and calls Update
func (tv *TableView) UpdateTable() {
	if !tv.This().(gi.Node2D).IsVisible() {
		return
	}
	if tv.Table != nil {
		tv.Table.Sequential()
		if tv.SortIdx >= 0 {
			tv.Table.SortCol(tv.SortIdx, !tv.SortDesc)
		}
	}
	tv.Update()
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
	mods, updt := tv.ConfigChildren(config)
	tv.ConfigSliceGrid()
	tv.ConfigToolbar()
	if mods {
		tv.SetFullReRender()
		tv.UpdateEnd(updt)
	}
}

func (tv *TableView) UpdtSliceSize() int {
	tv.Table.DeleteInvalid() // table could have changed
	tv.SliceSize = tv.Table.Len()
	tv.NCols = tv.Table.Table.NumCols()
	return tv.SliceSize
}

// SliceFrame returns the outer frame widget, which contains all the header,
// fields and values
func (tv *TableView) SliceFrame() *gi.Frame {
	return tv.ChildByName("frame", 0).(*gi.Frame)
}

// GridLayout returns the SliceGrid grid-layout widget, with grid and scrollbar
func (tv *TableView) GridLayout() *gi.Layout {
	gli := tv.SliceFrame().ChildByName("grid-lay", 0)
	if gli == nil {
		return nil
	}
	return gli.(*gi.Layout)
}

// SliceGrid returns the SliceGrid grid frame widget, which contains all the
// fields and values, within SliceFrame
func (tv *TableView) SliceGrid() *gi.Frame {
	gl := tv.GridLayout()
	if gl == nil {
		return nil
	}
	return gl.ChildByName("grid", 0).(*gi.Frame)
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
	sg := tv.SliceFrame()
	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)

	sgf := tv.This().(giv.SliceViewer).SliceGrid()
	if sgf != nil {
		sgf.DeleteChildren(ki.DestroyKids)
	}

	if tv.Table.Table == nil {
		return
	}

	sz := tv.This().(giv.SliceViewer).UpdtSliceSize()
	if tv.NCols == 0 {
		return
	}
	if sz == 0 {
		tv.Table.Table.SetNumRows(1) // temp
	}

	nWidgPerRow, idxOff := tv.RowWidgetNs()

	sg.Lay = gi.LayoutVert
	sg.SetMinPrefWidth(units.NewCh(20))
	sg.SetProp("overflow", gist.OverflowScroll) // this still gives it true size during PrefSize
	sg.SetStretchMax()                          // for this to work, ALL layers above need it too
	sg.SetProp("border-width", 0)
	sg.SetProp("margin", 0)
	sg.SetProp("padding", 0)

	sgcfg := kit.TypeAndNameList{}
	sgcfg.Add(gi.KiT_ToolBar, "header")
	sgcfg.Add(gi.KiT_Layout, "grid-lay")
	sg.ConfigChildren(sgcfg)

	sgh := tv.SliceHeader()
	sgh.Lay = gi.LayoutHoriz
	sgh.SetProp("overflow", gist.OverflowHidden) // no scrollbars!
	sgh.SetProp("spacing", 0)
	// sgh.SetStretchMaxWidth()

	gl := tv.GridLayout()
	gl.Lay = gi.LayoutHoriz
	gl.SetStretchMax() // for this to work, ALL layers above need it too
	gconfig := kit.TypeAndNameList{}
	gconfig.Add(gi.KiT_Frame, "grid")
	gconfig.Add(gi.KiT_ScrollBar, "scrollbar")
	gl.ConfigChildren(gconfig) // covered by above

	sgf = tv.This().(giv.SliceViewer).SliceGrid()
	sgf.Lay = gi.LayoutGrid
	sgf.Stripes = gi.RowStripes
	sgf.SetMinPrefHeight(units.NewEm(10))
	sgf.SetStretchMax() // for this to work, ALL layers above need it too
	sgf.SetProp("columns", nWidgPerRow)
	sgf.SetProp("overflow", gist.OverflowScroll) // this still gives it true size during PrefSize
	// this causes sizing / layout to fail, esp on window resize etc:
	// sgf.SetProp("spacing", gi.StdDialogVSpaceUnits)

	// Configure Header
	hcfg := kit.TypeAndNameList{}
	if tv.ShowIndex {
		hcfg.Add(gi.KiT_Action, "head-idx")
	}
	for fli := 0; fli < tv.NCols; fli++ {
		labnm := fmt.Sprintf("head-%v", tv.Table.Table.ColNames[fli])
		hcfg.Add(gi.KiT_Action, labnm)
	}
	if !tv.IsInactive() {
		hcfg.Add(gi.KiT_Label, "head-add")
		hcfg.Add(gi.KiT_Label, "head-del")
	}
	sgh.ConfigChildren(hcfg)

	// at this point, we make one dummy row to get size of widgets
	sgf.Kids = make(ki.Slice, nWidgPerRow)

	itxt := fmt.Sprintf("%05d", 0)
	labnm := fmt.Sprintf("index-%v", itxt)

	if tv.ShowIndex {
		hdr := sgh.Child(0).(*gi.Action)
		hdr.SetText("Index")
		hdr.Tooltip = "Click to sort by original native table order"
		hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			tvv := recv.Embed(KiT_TableView).(*TableView)
			tvv.SortSliceAction(-1)
		})
		idxlab := &gi.Label{}
		sgf.SetChild(idxlab, 0, labnm)
		idxlab.Text = itxt
	}

	for fli := 0; fli < tv.NCols; fli++ {
		col := tv.Table.Table.Cols[fli]
		colnm := tv.Table.Table.ColNames[fli]
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
		hdr.Tooltip = colnm + " (click to sort by) Type: " + col.DataType().String()
		if dsc, has := tv.Table.Table.MetaData[colnm+":desc"]; has {
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
				vv.SetSoloValue(reflect.ValueOf(&fval))
				hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					tvv := recv.Embed(KiT_TableView).(*TableView)
					act := send.(*gi.Action)
					fldIdx := act.Data.(int)
					tvv.SortSliceAction(fldIdx)
				})
			} else {
				cell := tv.Table.Table.CellTensorIdx(fli, 0)
				tvv := &TensorGridValueView{}
				ki.InitNode(tvv)
				vv = tvv
				vv.SetSoloValue(reflect.ValueOf(cell))
				hdr.Tooltip = "(click to edit display parameters for this column)"
				hdr.ActionSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					tvv := recv.Embed(KiT_TableView).(*TableView)
					act := send.(*gi.Action)
					fldIdx := act.Data.(int)
					tvv.TensorDispAction(fldIdx)
				})
			}
		}
		if wd, has := tv.Table.Table.MetaData[colnm+":width"]; has {
			vv.SetTag("width", wd)
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

	if sz == 0 {
		tv.Table.Table.SetNumRows(0) // revert
	}

	if tv.SortIdx >= 0 {
		tv.Table.SortCol(tv.SortIdx, !tv.SortDesc)
	}

	tv.ConfigScroll()
}

// LayoutSliceGrid does the proper layout of slice grid depending on allocated size
// returns true if UpdateSliceGrid should be called after this
func (tv *TableView) LayoutSliceGrid() bool {
	sg := tv.This().(giv.SliceViewer).SliceGrid()
	if sg == nil {
		return false
	}

	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)

	if tv.Table.Table == nil {
		sg.DeleteChildren(ki.DestroyKids)
		return false
	}

	tv.ViewMuLock()
	defer tv.ViewMuUnlock()

	tv.This().(giv.SliceViewer).UpdtSliceSize()
	if tv.NCols == 0 {
		sg.DeleteChildren(ki.DestroyKids)
		return false
	}

	nWidgPerRow, _ := tv.RowWidgetNs()
	if len(sg.GridData) > 0 && len(sg.GridData[gi.Row]) > 0 {
		tv.RowHeight = sg.GridData[gi.Row][0].AllocSize + sg.Spacing.Dots
	}
	if tv.Sty.Font.Face == nil {
		girl.OpenFont(&tv.Sty.Font, &tv.Sty.UnContext)
	}
	tv.RowHeight = mat32.Max(tv.RowHeight, tv.Sty.Font.Face.Metrics.Height)

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
		tv.VisRows = int(mat32.Floor(sgHt / tv.RowHeight))
	}
	tv.DispRows = ints.MinInt(tv.SliceSize, tv.VisRows)

	nWidg := nWidgPerRow * tv.DispRows

	if tv.Values == nil || sg.NumChildren() != nWidg {
		sg.DeleteChildren(ki.DestroyKids)

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
func (tv *TableView) UpdateSliceGrid() {
	sg := tv.This().(giv.SliceViewer).SliceGrid()
	if sg == nil {
		return
	}

	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := sg.UpdateStart()
	defer sg.UpdateEnd(updt)

	if tv.Table.Table == nil {
		sg.DeleteChildren(ki.DestroyKids)
		return
	}

	tv.ViewMuLock()
	defer tv.ViewMuUnlock()

	sz := tv.This().(giv.SliceViewer).UpdtSliceSize()
	if sz == 0 {
		sg.DeleteChildren(ki.DestroyKids)
		return
	}

	tv.DispRows = ints.MinInt(tv.SliceSize, tv.VisRows)

	tv.TsrDispToDots()

	nWidgPerRow, idxOff := tv.RowWidgetNs()
	nWidg := nWidgPerRow * tv.DispRows

	if tv.Values == nil || sg.NumChildren() != nWidg { // shouldn't happen..
		tv.ViewMuUnlock()
		tv.LayoutSliceGrid()
		tv.ViewMuLock()
		nWidg = nWidgPerRow * tv.DispRows
	}

	tv.UpdateStartIdx()

	for ri := 0; ri < tv.DispRows; ri++ {
		ridx := ri * nWidgPerRow
		si := tv.StartIdx + ri // slice idx
		ixi := tv.Table.Idxs[si]
		issel := tv.IdxIsSelected(si)

		itxt := fmt.Sprintf("%05d", ri)
		sitxt := fmt.Sprintf("%05d", si)
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
			col := tv.Table.Table.Cols[fli]
			colnm := tv.Table.Table.ColNames[fli]
			tdsp := tv.ColTensorDisp(fli)

			vvi := ri*tv.NCols + fli
			var vv giv.ValueView
			if tv.Values[vvi] == nil {
				if stsr, isstr := col.(*etensor.String); isstr {
					sval := stsr.Values[si]
					vv = giv.ToValueView(&sval, "")
					vv.SetProp("tv-row", ri)
					vv.SetProp("tv-col", fli)
					vv.SetSoloValue(reflect.ValueOf(&sval))
					vv.AsValueViewBase().ViewSig.ConnectOnly(tv.This(),
						func(recv, send ki.Ki, sig int64, data interface{}) {
							tvv, _ := recv.Embed(KiT_TableView).(*TableView)
							tvv.SetChanged()
							vvv := send.(giv.ValueView).AsValueViewBase()
							row := vvv.Prop("tv-row").(int)
							col := vvv.Prop("tv-col").(int)
							npv := kit.NonPtrValue(vvv.Value)
							sv := kit.ToString(npv.Interface())
							tvv.Table.Table.SetCellStringIdx(col, tvv.Table.Idxs[tvv.StartIdx+row], sv)
							tvv.ViewSig.Emit(tvv.This(), 0, nil)
						})
				} else {
					if col.NumDims() == 1 {
						fval := col.FloatVal1D(ixi)
						vv = giv.ToValueView(&fval, "")
						vv.SetProp("tv-row", ri)
						vv.SetProp("tv-col", fli)
						vv.SetSoloValue(reflect.ValueOf(&fval))
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
									tvv.Table.Table.SetCellFloatIdx(col, tvv.Table.Idxs[tvv.StartIdx+row], fv)
									tvv.ViewSig.Emit(tvv.This(), 0, nil)
								}
							})
					} else {
						cell := tv.Table.Table.CellTensorIdx(fli, si)
						tvv := &TensorGridValueView{}
						ki.InitNode(tvv)
						vv = tvv
						vv.SetSoloValue(reflect.ValueOf(cell))
					}
				}
				tv.Values[vvi] = vv
			} else {
				vv = tv.Values[vvi]
				if stsr, isstr := col.(*etensor.String); isstr {
					vv.SetSoloValue(reflect.ValueOf(&stsr.Values[ixi]))
				} else {
					if col.NumDims() == 1 {
						fval := col.FloatVal1D(ixi)
						vv.SetSoloValue(reflect.ValueOf(&fval))
					} else {
						cell := tv.Table.Table.CellTensorIdx(fli, ixi)
						vv.SetSoloValue(reflect.ValueOf(cell))
					}
				}
			}

			if wd, has := tv.Table.Table.MetaData[colnm+":width"]; has {
				vv.SetTag("width", wd)
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
				if col.IsNull1D(ri) { // todo: not working:
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
					wb.SetProp("tv-row", ri)
					wb.SetProp("vertical-align", gist.AlignTop)
					wb.ClearSelected()
					wb.WidgetSig.ConnectOnly(tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
						if sig == int64(gi.WidgetSelected) {
							wbb := send.(gi.Node2D).AsWidget()
							row := wbb.Prop("tv-row").(int)
							tvv := recv.Embed(KiT_TableView).(*TableView)
							tvv.UpdateSelectRow(row, wbb.IsSelected())
						}
					})
					if tv.IsInactive() {
						wb.SetInactive()
					}
					if col.IsNull1D(ri) {
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
					addact.Data = ri
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
					delact.Data = ri
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
		cl := tv.Table.Table.Cols[col]
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
		cl := tv.Table.Table.Cols[col]
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
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := tv.UpdateStart()
	defer tv.UpdateEnd(updt)

	tv.Table.InsertRows(idx, 1)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	tv.SetFullReRender()
	tv.This().(giv.SliceViewer).LayoutSliceGrid()
	tv.This().(giv.SliceViewer).UpdateSliceGrid()
	tv.ViewSig.Emit(tv.This(), 0, nil)
	tv.SliceViewSig.Emit(tv.This(), int64(giv.SliceViewInserted), idx)
}

// SliceDeleteAt deletes element at given index from slice -- doUpdt means
// call UpdateSliceGrid to update display
func (tv *TableView) SliceDeleteAt(idx int, doUpdt bool) {
	if idx < 0 {
		return
	}
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := tv.UpdateStart()
	defer tv.UpdateEnd(updt)

	delete(tv.SelectedIdxs, idx)

	tv.Table.DeleteRows(idx, 1)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	if doUpdt {
		tv.SetFullReRender()
		tv.This().(giv.SliceViewer).LayoutSliceGrid()
		tv.This().(giv.SliceViewer).UpdateSliceGrid()
	}
	tv.ViewSig.Emit(tv.This(), 0, nil)
	tv.SliceViewSig.Emit(tv.This(), int64(giv.SliceViewDeleted), idx)
}

// SortSliceAction sorts the slice for given field index -- toggles ascending
// vs. descending if already sorting on this dimension
func (tv *TableView) SortSliceAction(fldIdx int) {
	oswin.TheApp.Cursor(tv.ParentWindow().OSWin).Push(cursor.Wait)
	defer oswin.TheApp.Cursor(tv.ParentWindow().OSWin).Pop()

	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

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
	if fldIdx == -1 {
		tv.Table.SortIdxs()
	} else {
		tv.Table.SortCol(tv.SortIdx, !tv.SortDesc)
	}
	tv.UpdateSliceGrid()
	tv.UpdateEnd(updt)
}

// TensorDispAction allows user to select tensor display options for column
// pass -1 for global params for the entire table
func (tv *TableView) TensorDispAction(fldIdx int) {
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)

	updt := tv.UpdateStart()
	ctd := &tv.TsrDisp
	if fldIdx >= 0 {
		ctd = tv.SetColTensorDisp(fldIdx)
	}
	giv.StructViewDialog(tv.ViewportSafe(), ctd, giv.DlgOpts{Title: "TensorGrid Display Options", Ok: true, Cancel: true},
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
		tb.AddAction(gi.ActOpts{Label: "UpdtView", Icon: "update", Tooltip: "update the view to reflect current state of table: this will reset any existing filtering or sorting of the table view"},
			tv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
				tvv := recv.Embed(KiT_TableView).(*TableView)
				tvv.UpdateTable()
			})
		tb.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "configure the view -- particularly the tensor display options"},
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
	mvp := tv.ViewportSafe()
	if giv.HasToolBarView(tv.Table) && mvp != nil {
		giv.ToolBarView(tv.Table, mvp, tb)
	}
	tv.ToolbarSlice = tv.Table
}

// SortFieldName returns the name of the field being sorted, along with :up or
// :down depending on descending
func (tv *TableView) SortFieldName() string {
	if tv.SortIdx >= 0 && tv.SortIdx < tv.NCols {
		nm := tv.Table.Table.ColNames[tv.SortIdx]
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
		colnm := tv.Table.Table.ColNames[fli]
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

func (tv *TableView) TsrDispToDots() {
	tv.TsrDisp.ToDots(&tv.Sty.UnContext)
	for _, ctd := range tv.ColTsrDisp {
		ctd.ToDots(&tv.Sty.UnContext)
	}
}

func (tv *TableView) Style2D() {
	tv.SliceViewBase.Style2D()
	tv.TsrDispToDots()
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

//////////////////////////////////////////////////////////////////////////////
//    Copy / Cut / Paste

func (tv *TableView) MimeDataType() string {
	return filecat.DataCsv
}

// CopySelToMime copies selected rows to mime data
func (tv *TableView) CopySelToMime() mimedata.Mimes {
	nitms := len(tv.SelectedIdxs)
	if nitms == 0 {
		return nil
	}
	ix := &etable.IdxView{}
	ix.Table = tv.Table.Table
	idx := tv.SelectedIdxsList(false) // ascending
	iidx := make([]int, len(idx))
	for i, di := range idx {
		iidx[i] = tv.Table.Idxs[di]
	}
	ix.Idxs = iidx
	var b bytes.Buffer
	ix.WriteCSV(&b, etable.Tab, etable.Headers)
	md := mimedata.NewTextBytes(b.Bytes())
	md[0].Type = filecat.DataCsv
	return md
}

// FromMimeData returns records from csv of mime data
func (tv *TableView) FromMimeData(md mimedata.Mimes) [][]string {
	var recs [][]string
	for _, d := range md {
		if d.Type == filecat.DataCsv {
			b := bytes.NewBuffer(d.Data)
			cr := csv.NewReader(b)
			cr.Comma = etable.Tab.Rune()
			rec, err := cr.ReadAll()
			if err != nil || len(rec) == 0 {
				log.Printf("Error reading CSV from clipboard: %s\n", err)
				return nil
			}
			recs = append(recs, rec...)
		}
	}
	return recs
}

// PasteAssign assigns mime data (only the first one!) to this idx
func (tv *TableView) PasteAssign(md mimedata.Mimes, idx int) {
	recs := tv.FromMimeData(md)
	if len(recs) == 0 {
		return
	}
	updt := tv.UpdateStart()
	tv.Table.Table.ReadCSVRow(recs[1], tv.Table.Idxs[idx])
	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	tv.This().(giv.SliceViewer).UpdateSliceGrid()
	tv.UpdateEnd(updt)
}

// PasteAtIdx inserts object(s) from mime data at (before) given slice index
// adds to end of table
func (tv *TableView) PasteAtIdx(md mimedata.Mimes, idx int) {
	recs := tv.FromMimeData(md)
	nr := len(recs) - 1
	if nr <= 0 {
		return
	}
	wupdt := tv.TopUpdateStart()
	defer tv.TopUpdateEnd(wupdt)
	updt := tv.UpdateStart()
	tv.Table.InsertRows(idx, nr)
	for ri := 0; ri < nr; ri++ {
		rec := recs[1+ri]
		rw := tv.Table.Idxs[idx+ri]
		tv.Table.Table.ReadCSVRow(rec, rw)
	}
	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.SetChanged()
	tv.This().(giv.SliceViewer).UpdateSliceGrid()
	tv.UpdateEnd(updt)
	tv.SelectIdxAction(idx, mouse.SelectOne)
}

func (tv *TableView) ItemCtxtMenu(idx int) {
	var men gi.Menu
	tv.StdCtxtMenu(&men, idx)
	if len(men) > 0 {
		pos := tv.IdxPos(idx)
		gi.PopupMenu(men, pos.X, pos.Y, tv.ViewportSafe(), tv.Nm+"-menu")
	}
}
