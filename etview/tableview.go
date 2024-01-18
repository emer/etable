// Copyright (c) 2023, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

//go:generate core generate -add-types

import (
	"fmt"
	"image"
	"reflect"
	"strconv"
	"strings"

	"cogentcore.org/core/abilities"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/gi"
	"cogentcore.org/core/giv"
	"cogentcore.org/core/grr"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/ki"
	"cogentcore.org/core/laser"
	"cogentcore.org/core/states"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/units"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
)

// etview.TableView provides a GUI interface for etable.Table's
type TableView struct {
	giv.SliceViewBase

	// the idx view of the table that we're a view of
	Table *etable.IdxView `set:"-"`

	// overall display options for tensor display
	TsrDisp TensorDisp

	// per column tensor display params
	ColTsrDisp map[int]*TensorDisp

	// per column blank tensor values
	ColTsrBlank map[int]*etensor.Float64

	// number of columns in table (as of last update)
	NCols int `edit:"-"`

	// current sort index
	SortIdx int

	// whether current sort order is descending
	SortDesc bool

	// HeaderWidths has number of characters in each header, per visfields
	HeaderWidths []int `copy:"-" view:"-" json:"-" xml:"-"`

	//	blank values for out-of-range rows
	BlankString string
	BlankFloat  float64
}

// check for interface impl
var _ giv.SliceViewer = (*TableView)(nil)

func (tv *TableView) OnInit() {
	tv.Frame.OnInit()
	tv.SliceViewBase.HandleEvents()
	tv.SetStyles()
}

func (tv *TableView) SetStyles() {
	tv.SortIdx = -1
	tv.MinRows = 4
	tv.TsrDisp.Defaults()
	tv.ColTsrDisp = make(map[int]*TensorDisp)
	tv.ColTsrBlank = make(map[int]*etensor.Float64)
	tv.SetFlag(false, giv.SliceViewSelectMode)
	tv.SetFlag(true, giv.SliceViewShowIndex)
	tv.SetFlag(true, giv.SliceViewReadOnlyKeyNav)

	tv.Style(func(s *styles.Style) {
		s.SetAbilities(true, abilities.FocusWithinable)
		s.Direction = styles.Column
		// absorb horizontal here, vertical in view
		s.Overflow.X = styles.OverflowAuto
		s.Grow.Set(1, 1)
	})
	tv.OnWidgetAdded(func(w gi.Widget) {
		switch w.PathFrom(tv) {
		case "header": // slice header
			sh := w.(*gi.Frame)
			gi.ToolbarStyles(sh)
			sh.Style(func(s *styles.Style) {
				s.Grow.Set(0, 0)
				s.Gap.Set(units.Em(0.5)) // matches grid default
			})
		case "header/head-idx": // index header
			lbl := w.(*gi.Label)
			lbl.SetText("Index").SetType(gi.LabelBodyMedium)
			w.Style(func(s *styles.Style) {
				s.Align.Self = styles.Center
			})
		case "grid": // slice grid
			sg := w.(*giv.SliceViewGrid)
			sg.Stripes = gi.RowStripes
			sg.Style(func(s *styles.Style) {
				sg.MinRows = tv.MinRows
				s.Display = styles.Grid
				nWidgPerRow, _ := tv.RowWidgetNs()
				s.Columns = nWidgPerRow
				s.Grow.Set(1, 1)
				s.Overflow.Y = styles.OverflowAuto
				s.Gap.Set(units.Em(0.5)) // note: match header
				// baseline mins:
				s.Min.X.Ch(20)
				s.Min.Y.Em(6)
			})
		}
		if w.Parent().PathFrom(tv) == "grid" {
			switch {
			case strings.HasPrefix(w.Name(), "index-"):
				w.Style(func(s *styles.Style) {
					s.Min.X.Em(1.5)
					s.Padding.Right.Dp(4)
					s.Text.Align = styles.End
					s.Min.Y.Em(1)
					s.GrowWrap = false
				})
			case strings.HasPrefix(w.Name(), "add-"):
				w.Style(func(s *styles.Style) {
					w.(*gi.Button).SetType(gi.ButtonAction)
					s.Color = colors.Scheme.Success.Base
				})
			case strings.HasPrefix(w.Name(), "del-"):
				w.Style(func(s *styles.Style) {
					w.(*gi.Button).SetType(gi.ButtonAction)
					s.Color = colors.Scheme.Error.Base
				})
			case strings.HasPrefix(w.Name(), "value-"):
				w.Style(func(s *styles.Style) {
					fstr := strings.TrimPrefix(w.Name(), "value-")
					dp := strings.Index(fstr, ".")
					fstr = fstr[:dp] // field idx is -X.
					fli := grr.Log1(strconv.Atoi(fstr))
					hw := float32(tv.HeaderWidths[fli])
					if fli == tv.SortIdx {
						hw += 6
					}
					hv := units.Ch(hw)
					s.Min.X.Val = max(s.Min.X.Val, hv.Convert(s.Min.X.Un, &s.UnContext).Val)
					s.Max.X.Val = max(s.Max.X.Val, hv.Convert(s.Max.X.Un, &s.UnContext).Val)
				})
			}
		}
		if w.Parent().PathFrom(tv) == "header" {
			w.Style(func(s *styles.Style) {
				if hdr, ok := w.(*gi.Button); ok {
					fli := hdr.Data.(int)
					if fli == tv.SortIdx {
						if tv.SortDesc {
							hdr.SetIcon(icons.KeyboardArrowDown)
						} else {
							hdr.SetIcon(icons.KeyboardArrowUp)
						}
					}
				}
			})
		}
	})
}

// SetTable sets the source table that we are viewing, using a sequential IdxView
// and then configures the display
func (tv *TableView) SetTable(et *etable.Table) *TableView {
	if et == nil {
		return nil
	}
	updt := tv.UpdateStart()
	defer tv.UpdateEndLayout(updt)

	tv.Table = etable.NewIdxView(et)

	tv.This().(giv.SliceViewer).UpdtSliceSize()
	tv.SetFlag(false, giv.SliceViewConfiged)
	tv.StartIdx = 0
	tv.VisRows = tv.MinRows
	if !tv.IsReadOnly() {
		tv.SelIdx = -1
	}
	tv.ResetSelectedIdxs()
	tv.SetFlag(false, giv.SliceViewSelectMode)
	tv.ConfigIter = 0
	tv.Update()
	return tv
}

// GoUpdateView updates the display for asynchronous updating from
// other goroutines.  Also updates indexview (calling Sequential).
func (tv *TableView) GoUpdateView() {
	updt := tv.UpdateStartAsync()
	tv.Table.Sequential()
	tv.ScrollToIdxNoUpdt(tv.SliceSize - 1)
	tv.UpdateWidgets()
	tv.UpdateEndAsyncLayout(updt)
}

// SetTableView sets the source IdxView of a table (using a copy so original is not modified)
// and then configures the display
func (tv *TableView) SetTableView(ix *etable.IdxView) *TableView {
	if ix == nil {
		return tv
	}
	updt := tv.UpdateStart()
	defer tv.UpdateEndLayout(updt)

	tv.Table = ix.Clone() // always copy

	tv.This().(giv.SliceViewer).UpdtSliceSize()
	tv.SetFlag(false, giv.SliceViewConfiged)
	tv.StartIdx = 0
	tv.VisRows = tv.MinRows
	if !tv.IsReadOnly() {
		tv.SelIdx = -1
	}
	tv.ResetSelectedIdxs()
	tv.SetFlag(false, giv.SliceViewSelectMode)
	tv.ConfigIter = 0
	tv.Update()
	return tv
}

func (tv *TableView) UpdtSliceSize() int {
	tv.Table.DeleteInvalid() // table could have changed
	tv.SliceSize = tv.Table.Len()
	tv.NCols = tv.Table.Table.NumCols()
	return tv.SliceSize
}

// Config configures the view
func (tv *TableView) ConfigWidget() {
	tv.ConfigTableView()
}

func (tv *TableView) ConfigTableView() {
	if tv.Is(giv.SliceViewConfiged) {
		tv.This().(giv.SliceViewer).UpdateWidgets()
		return
	}
	updt := tv.UpdateStart()
	tv.ConfigFrame()
	tv.This().(giv.SliceViewer).ConfigRows()
	tv.This().(giv.SliceViewer).UpdateWidgets()
	tv.ApplyStyleTree()
	tv.UpdateEndLayout(updt)
}

func (tv *TableView) ConfigFrame() {
	if tv.HasChildren() {
		return
	}
	tv.SetFlag(true, giv.SliceViewConfiged)
	gi.NewFrame(tv, "header")
	giv.NewSliceViewGrid(tv, "grid")
	tv.ConfigHeader()
}

func (tv *TableView) ConfigHeader() {
	sgh := tv.SliceHeader()
	hcfg := ki.Config{}
	if tv.Is(giv.SliceViewShowIndex) {
		hcfg.Add(gi.LabelType, "head-idx")
	}
	tv.HeaderWidths = make([]int, tv.NCols)
	for fli := 0; fli < tv.NCols; fli++ {
		fld := tv.Table.Table.ColNames[fli]
		labnm := "head-" + fld
		hcfg.Add(gi.ButtonType, labnm)
	}
	if !tv.IsReadOnly() {
		hcfg.Add(gi.LabelType, "head-add")
		hcfg.Add(gi.LabelType, "head-del")
	}
	sgh.ConfigChildren(hcfg) // headers SHOULD be unique, but with labels..
	_, idxOff := tv.RowWidgetNs()
	nfld := tv.NCols
	for fli := 0; fli < nfld; fli++ {
		fli := fli
		field := tv.Table.Table.ColNames[fli]
		hdr := sgh.Child(idxOff + fli).(*gi.Button)
		hdr.SetType(gi.ButtonMenu)
		hdr.SetText(field)
		hdr.Data = fli
		tv.HeaderWidths[fli] = len(field)
		if fli == tv.SortIdx {
			if tv.SortDesc {
				hdr.SetIcon(icons.KeyboardArrowDown)
			} else {
				hdr.SetIcon(icons.KeyboardArrowUp)
			}
		}
		hdr.Tooltip = field + " (click to sort by)"
		hdr.OnClick(func(e events.Event) {
			tv.SortSliceAction(fli)
		})
	}
	if !tv.IsReadOnly() {
		cidx := tv.NCols + idxOff
		if !tv.Is(giv.SliceViewNoAdd) {
			lbl := sgh.Child(cidx).(*gi.Label)
			lbl.Text = "+"
			lbl.Tooltip = "insert row"
			cidx++
		}
		if !tv.Is(giv.SliceViewNoDelete) {
			lbl := sgh.Child(cidx).(*gi.Label)
			lbl.Text = "-"
			lbl.Tooltip = "delete row"
		}
	}
}

// SliceGrid returns the SliceGrid grid frame widget, which contains all the
// fields and values, within SliceFrame
func (tv *TableView) SliceGrid() *giv.SliceViewGrid {
	return tv.Child(1).(*giv.SliceViewGrid)
}

// SliceHeader returns the Frame header for slice grid
func (tv *TableView) SliceHeader() *gi.Frame {
	return tv.Child(0).(*gi.Frame)
}

// RowWidgetNs returns number of widgets per row and offset for index label
func (tv *TableView) RowWidgetNs() (nWidgPerRow, idxOff int) {
	nWidgPerRow = 1 + tv.NCols
	if !tv.IsReadOnly() {
		if !tv.Is(giv.SliceViewNoAdd) {
			nWidgPerRow += 1
		}
		if !tv.Is(giv.SliceViewNoDelete) {
			nWidgPerRow += 1
		}
	}
	idxOff = 1
	if !tv.Is(giv.SliceViewShowIndex) {
		nWidgPerRow -= 1
		idxOff = 0
	}
	return
}

// ConfigRows configures VisRows worth of widgets
// to display slice data.  It should only be called
// when NeedsConfigRows is true: when VisRows changes.
func (tv *TableView) ConfigRows() {
	sg := tv.This().(giv.SliceViewer).SliceGrid()
	if sg == nil {
		return
	}
	tv.SetFlag(true, giv.SliceViewConfiged)
	sg.SetFlag(true, gi.LayoutNoKeys)

	tv.ViewMuLock()
	defer tv.ViewMuUnlock()

	sg.DeleteChildren(ki.DestroyKids)
	tv.Values = nil

	if tv.Table == nil {
		return
	}

	tv.This().(giv.SliceViewer).UpdtSliceSize()

	nWidgPerRow, idxOff := tv.RowWidgetNs()
	nWidg := nWidgPerRow * tv.VisRows
	sg.Styles.Columns = nWidgPerRow

	tv.Values = make([]giv.Value, tv.NCols*tv.VisRows)
	sg.Kids = make(ki.Slice, nWidg)

	for i := 0; i < tv.VisRows; i++ {
		i := i
		si := i
		ridx := i * nWidgPerRow

		idxlab := &gi.Label{}
		itxt := strconv.Itoa(i)
		sitxt := strconv.Itoa(si)
		labnm := "index-" + itxt
		if tv.Is(giv.SliceViewShowIndex) {
			idxlab = &gi.Label{}
			sg.SetChild(idxlab, ridx, labnm)
			idxlab.OnSelect(func(e events.Event) {
				e.SetHandled()
				tv.UpdateSelectRow(i)
			})
			idxlab.SetText(sitxt)
		}

		vpath := tv.ViewPath + "[" + sitxt + "]"
		if lblr, ok := tv.Slice.(gi.SliceLabeler); ok {
			slbl := lblr.ElemLabel(si)
			if slbl != "" {
				vpath = tv.ViewPath + "[" + slbl + "]"
			}
		}
		for fli := 0; fli < tv.NCols; fli++ {
			fli := fli
			col := tv.Table.Table.Cols[fli]
			vvi := i*tv.NCols + fli
			tags := ""
			var vv giv.Value
			if _, isstr := col.(*etensor.String); isstr {
				vv = giv.ToValue(&tv.BlankString, tags)
				vv.SetSoloValue(reflect.ValueOf(&tv.BlankString))
				vvb := vv.AsValueBase()
				vvb.OnChange(func(e events.Event) {
					tv.SetChanged()
					npv := laser.NonPtrValue(vvb.Value)
					sv := laser.ToString(npv.Interface())
					si := tv.StartIdx + i
					if si < len(tv.Table.Idxs) {
						tv.Table.Table.SetCellStringIdx(fli, tv.Table.Idxs[si], sv)
					}
				})
			} else {
				if col.NumDims() == 1 {
					vv = giv.ToValue(&tv.BlankFloat, "")
					vv.SetSoloValue(reflect.ValueOf(&tv.BlankFloat))
					vvb := vv.AsValueBase()
					vvb.OnChange(func(e events.Event) {
						tv.SetChanged()
						npv := laser.NonPtrValue(vvb.Value)
						fv := grr.Log1(laser.ToFloat(npv.Interface()))
						si := tv.StartIdx + i
						if si < len(tv.Table.Idxs) {
							tv.Table.Table.SetCellFloatIdx(fli, tv.Table.Idxs[si], fv)
						}
					})
				} else {
					// tdsp := tv.ColTensorDisp(fli)
					cell := tv.ColTensorBlank(fli, col)
					tvv := &TensorGridValue{}
					vv = tvv
					tvv.ViewPath = vpath
					vv.SetSoloValue(reflect.ValueOf(cell))
				}
			}
			tv.Values[vvi] = vv
			vv.SetReadOnly(tv.IsReadOnly())
			vtyp := vv.WidgetType()
			valnm := fmt.Sprintf("value-%v.%v", fli, itxt)
			cidx := ridx + idxOff + fli
			w := ki.NewOfType(vtyp).(gi.Widget)
			sg.SetChild(w, cidx, valnm)
			vv.ConfigWidget(w)
			if tv.IsReadOnly() {
				w.AsWidget().SetReadOnly(true)
			}
			wb := w.AsWidget()
			wb.OnSelect(func(e events.Event) {
				e.SetHandled()
				tv.UpdateSelectRow(i)
			})
			if col.NumDims() > 1 {
				tgw := w.This().(*TensorGrid)
				tgw.Style(func(s *styles.Style) {
					s.Grow.Set(0, 0)
				})
			}
		}

		if !tv.IsReadOnly() {
			cidx := ridx + tv.NCols + idxOff
			if !tv.Is(giv.SliceViewNoAdd) {
				addnm := fmt.Sprintf("add-%v", itxt)
				addact := gi.Button{}
				sg.SetChild(&addact, cidx, addnm)
				addact.SetType(gi.ButtonAction).SetIcon(icons.Add).
					SetTooltip("insert a new element at this index").OnClick(func(e events.Event) {
					tv.SliceNewAtRow(i + 1)
				})
				cidx++
			}
			if !tv.Is(giv.SliceViewNoDelete) {
				delnm := fmt.Sprintf("del-%v", itxt)
				delact := gi.Button{}
				sg.SetChild(&delact, cidx, delnm)
				delact.SetType(gi.ButtonAction).SetIcon(icons.Delete).
					SetTooltip("delete this element").OnClick(func(e events.Event) {
					tv.SliceDeleteAtRow(i)
				})
				cidx++
			}
		}
	}
	tv.ConfigTree()
	tv.ApplyStyleTree()
}

// UpdateWidgets updates the row widget display to
// represent the current state of the slice data,
// including which range of data is being displayed.
// This is called for scrolling, navigation etc.
func (tv *TableView) UpdateWidgets() {
	sg := tv.This().(giv.SliceViewer).SliceGrid()
	if sg == nil || tv.VisRows == 0 || sg.VisRows == 0 || !sg.HasChildren() {
		return
	}
	// sc := tv.Sc

	updt := sg.UpdateStart()
	defer sg.UpdateEndRender(updt)

	tv.ViewMuLock()
	defer tv.ViewMuUnlock()

	tv.This().(giv.SliceViewer).UpdtSliceSize()

	nWidgPerRow, idxOff := tv.RowWidgetNs()

	tv.UpdateStartIdx()
	for i := 0; i < tv.VisRows; i++ {
		i := i
		ridx := i * nWidgPerRow
		si := tv.StartIdx + i // slice idx
		ixi := -1
		if si < len(tv.Table.Idxs) {
			ixi = tv.Table.Idxs[si]
		}
		invis := si >= tv.SliceSize

		var idxlab *gi.Label
		if tv.Is(giv.SliceViewShowIndex) {
			idxlab = sg.Kids[ridx].(*gi.Label)
			idxlab.SetTextUpdate(strconv.Itoa(si))
			idxlab.SetState(invis, states.Invisible)
		}

		sitxt := strconv.Itoa(si)
		vpath := tv.ViewPath + "[" + sitxt + "]"
		if lblr, ok := tv.Slice.(gi.SliceLabeler); ok {
			slbl := lblr.ElemLabel(si)
			if slbl != "" {
				vpath = tv.ViewPath + "[" + slbl + "]"
			}
		}
		for fli := 0; fli < tv.NCols; fli++ {
			fli := fli
			col := tv.Table.Table.Cols[fli]
			cidx := ridx + idxOff + fli
			w := sg.Kids[cidx].(gi.Widget)
			wb := w.AsWidget()
			vvi := i*tv.NCols + fli
			vv := tv.Values[vvi]
			vvb := vv.AsValueBase()
			vvb.ViewPath = vpath

			if stsr, isstr := col.(*etensor.String); isstr {
				sval := &tv.BlankString
				if ixi >= 0 {
					sval = &stsr.Values[ixi]
				}
				vv.SetSoloValue(reflect.ValueOf(sval))
			} else {
				if col.NumDims() == 1 {
					fval := 0.0
					if ixi >= 0 {
						fval = col.FloatVal1D(ixi)
					}
					vv.SetSoloValue(reflect.ValueOf(&fval))
				} else {
					tdsp := tv.ColTensorDisp(fli)
					var cell etensor.Tensor
					cell = tv.ColTensorBlank(fli, col)
					if ixi >= 0 {
						cell = tv.Table.Table.CellTensorIdx(fli, ixi)
					}
					vv.SetSoloValue(reflect.ValueOf(cell))
					tgw := w.This().(*TensorGrid)
					tgw.Disp = *tdsp
				}
			}
			vv.SetReadOnly(tv.IsReadOnly())
			vv.UpdateWidget()

			w.SetState(invis, states.Invisible)
			if !invis {
				issel := tv.IdxIsSelected(si)
				if tv.IsReadOnly() {
					wb.SetReadOnly(true)
				}
				wb.SetSelected(issel)
			} else {
				wb.SetSelected(false)
				if tv.Is(giv.SliceViewShowIndex) {
					idxlab.SetSelected(false)
				}
			}
		}
		if !tv.IsReadOnly() {
			cidx := ridx + tv.NCols + idxOff
			if !tv.Is(giv.SliceViewNoAdd) {
				addact := sg.Kids[cidx].(*gi.Button)
				addact.SetState(invis, states.Invisible)
				cidx++
			}
			if !tv.Is(giv.SliceViewNoDelete) {
				delact := sg.Kids[cidx].(*gi.Button)
				delact.SetState(invis, states.Invisible)
				cidx++
			}
		}
	}

	if tv.InitSelIdx >= 0 {
		tv.SelIdx = tv.InitSelIdx
		tv.InitSelIdx = -1
		tv.ScrollToIdx(tv.SelIdx)
		// tv.SetFocusEvent()
	}

	if tv.IsReadOnly() && tv.SelIdx >= 0 {
		tv.SelectIdx(tv.SelIdx)
	}
}

// ColTensorBlank returns tensor blanks for given tensor col
func (tv *TableView) ColTensorBlank(cidx int, col etensor.Tensor) *etensor.Float64 {
	if ctb, has := tv.ColTsrBlank[cidx]; has {
		return ctb
	}
	ctb := etensor.NewFloat64Shape(col.ShapeObj(), nil)
	tv.ColTsrBlank[cidx] = ctb
	return ctb
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

// SliceNewAt inserts a new blank element at given index in the slice -- -1
// means the end
func (tv *TableView) SliceNewAt(idx int) {
	tv.ViewMuLock()
	updt := tv.UpdateStart()
	defer tv.UpdateEndLayout(updt)

	tv.SliceNewAtSel(idx)

	tv.Table.InsertRows(idx, 1)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.ViewMuUnlock()
	tv.SetChanged()
	tv.This().(giv.SliceViewer).UpdateWidgets()
}

// SliceDeleteAt deletes element at given index from slice -- doUpdt means
// call UpdateSliceGrid to update display
func (tv *TableView) SliceDeleteAt(idx int) {
	if idx < 0 || idx >= tv.SliceSize {
		return
	}
	tv.ViewMuLock()
	updt := tv.UpdateStart()
	defer tv.UpdateEndLayout(updt)

	tv.SliceDeleteAtSel(idx)

	tv.Table.DeleteRows(idx, 1)

	if tv.TmpSave != nil {
		tv.TmpSave.SaveTmp()
	}
	tv.ViewMuUnlock()
	tv.SetChanged()
	tv.This().(giv.SliceViewer).UpdateWidgets()
}

// SortSliceAction sorts the slice for given field index -- toggles ascending
// vs. descending if already sorting on this dimension
func (tv *TableView) SortSliceAction(fldIdx int) {
	updt := tv.UpdateStart()
	defer tv.UpdateEndLayout(updt)

	sgh := tv.SliceHeader()
	_, idxOff := tv.RowWidgetNs()

	ascending := true

	for fli := 0; fli < tv.NCols; fli++ {
		hdr := sgh.Child(idxOff + fli).(*gi.Button)
		hdr.SetType(gi.ButtonAction)
		if fli == fldIdx {
			if tv.SortIdx == fli {
				tv.SortDesc = !tv.SortDesc
				ascending = !tv.SortDesc
			} else {
				tv.SortDesc = false
			}
			if ascending {
				hdr.SetIcon(icons.KeyboardArrowUp)
			} else {
				hdr.SetIcon(icons.KeyboardArrowDown)
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
	tv.Update() // requires full update due to sort button icon
}

// TensorDispAction allows user to select tensor display options for column
// pass -1 for global params for the entire table
func (tv *TableView) TensorDispAction(fldIdx int) {
	updt := tv.UpdateStart()
	defer tv.UpdateEndRender(updt)

	ctd := &tv.TsrDisp
	if fldIdx >= 0 {
		ctd = tv.SetColTensorDisp(fldIdx)
	}
	d := gi.NewBody().AddTitle("Tensor Grid Display Options")
	giv.NewStructView(d).SetStruct(ctd)
	d.NewFullDialog(tv).Run()
	// tv.UpdateSliceGrid()
}

func (tv *TableView) StyleRow(w gi.Widget, idx, fidx int) {
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
	got := false
	for fli := 0; fli < tv.NCols; fli++ {
		fld := tv.Table.Table.ColNames[fli]
		if fld == spnm[0] {
			got = true
			// fmt.Println("sorting on:", fld.Name, fli, "from:", nm)
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
	_ = got
	// if got {
	// 	tv.SortSlice()
	// }
}

// RowFirstVisWidget returns the first visible widget for given row (could be
// index or not) -- false if out of range
func (tv *TableView) RowFirstVisWidget(row int) (*gi.WidgetBase, bool) {
	if !tv.IsRowInBounds(row) {
		return nil, false
	}
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	sg := tv.SliceGrid()
	w := sg.Kids[row*nWidgPerRow].(gi.Widget).AsWidget()
	if w.Geom.TotalBBox != (image.Rectangle{}) {
		return w, true
	}
	ridx := nWidgPerRow * row
	for fli := 0; fli < tv.NCols; fli++ {
		w := sg.Child(ridx + idxOff + fli).(gi.Widget).AsWidget()
		if w.Geom.TotalBBox != (image.Rectangle{}) {
			return w, true
		}
	}
	return nil, false
}

// RowGrabFocus grabs the focus for the first focusable widget in given row --
// returns that element or nil if not successful -- note: grid must have
// already rendered for focus to be grabbed!
func (tv *TableView) RowGrabFocus(row int) *gi.WidgetBase {
	if !tv.IsRowInBounds(row) || tv.Is(giv.SliceViewInFocusGrab) { // range check
		return nil
	}
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	ridx := nWidgPerRow * row
	sg := tv.SliceGrid()
	// first check if we already have focus
	for fli := 0; fli < tv.NCols; fli++ {
		w := sg.Child(ridx + idxOff + fli).(gi.Widget).AsWidget()
		if w.StateIs(states.Focused) || w.ContainsFocus() {
			return w
		}
	}
	tv.SetFlag(true, giv.SliceViewInFocusGrab)
	defer func() { tv.SetFlag(false, giv.SliceViewInFocusGrab) }()
	for fli := 0; fli < tv.NCols; fli++ {
		w := sg.Child(ridx + idxOff + fli).(gi.Widget).AsWidget()
		if w.CanFocus() {
			w.SetFocusEvent()
			return w
		}
	}
	return nil
}

// SelectRowWidgets sets the selection state of given row of widgets
func (tv *TableView) SelectRowWidgets(row int, sel bool) {
	if row < 0 {
		return
	}
	updt := tv.UpdateStart()
	defer tv.UpdateEndRender(updt)

	sg := tv.SliceGrid()
	nWidgPerRow, idxOff := tv.RowWidgetNs()
	ridx := row * nWidgPerRow
	for fli := 0; fli < tv.NCols; fli++ {
		seldx := ridx + idxOff + fli
		if sg.Kids.IsValidIndex(seldx) == nil {
			w := sg.Child(seldx).(gi.Widget).AsWidget()
			w.SetSelected(sel)
		}
	}
	if tv.Is(giv.SliceViewShowIndex) {
		if sg.Kids.IsValidIndex(ridx) == nil {
			w := sg.Child(ridx).(gi.Widget).AsWidget()
			w.SetSelected(sel)
		}
	}
}

func (tv *TableView) EditIdx(idx int) {
	val := laser.OnePtrUnderlyingValue(tv.SliceNPVal.Index(idx))
	stru := val.Interface()
	tynm := laser.NonPtrType(val.Type()).Name()
	lbl := gi.ToLabel(stru)
	if lbl != "" {
		tynm += ": " + lbl
	}
	d := gi.NewBody().AddTitle(tynm)
	giv.NewStructView(d).SetStruct(stru)
	// d.AddBottomBar(func(pw gi.Widget) {
	// 	d.AddCancel(pw)
	// 	d.AddOk(pw)
	// })
	d.NewFullDialog(tv).Run()
}

//////////////////////////////////////////////////////
// 	Header layout

func (tv *TableView) SizeFinal() {
	tv.SliceViewBase.SizeFinal()
	sg := tv.This().(giv.SliceViewer).SliceGrid()
	sh := tv.SliceHeader()
	sh.WidgetKidsIter(func(i int, kwi gi.Widget, kwb *gi.WidgetBase) bool {
		_, sgb := gi.AsWidget(sg.Child(i))
		gsz := &sgb.Geom.Size
		ksz := &kwb.Geom.Size
		ksz.Actual.Total.X = gsz.Actual.Total.X
		ksz.Actual.Content.X = gsz.Actual.Content.X
		ksz.Alloc.Total.X = gsz.Alloc.Total.X
		ksz.Alloc.Content.X = gsz.Alloc.Content.X
		return ki.Continue
	})
	gsz := &sg.Geom.Size
	ksz := &sh.Geom.Size
	ksz.Actual.Total.X = gsz.Actual.Total.X
	ksz.Actual.Content.X = gsz.Actual.Content.X
	ksz.Alloc.Total.X = gsz.Alloc.Total.X
	ksz.Alloc.Content.X = gsz.Alloc.Content.X
}

//////////////////////////////////////////////////////////////////////////////
//    Copy / Cut / Paste

func (tv *TableView) ConfigToolbar(tb *gi.Toolbar) {
	if tv.Table == nil || tv.Table.Table == nil {
		return
	}
	giv.NewFuncButton(tb, tv.Table.AddRows).SetIcon(icons.Add)
	giv.NewFuncButton(tb, tv.Table.SortColName).SetText("Sort").SetIcon(icons.Sort)
	giv.NewFuncButton(tb, tv.Table.FilterColName).SetText("Filter").SetIcon(icons.FilterAlt)
	giv.NewFuncButton(tb, tv.Table.Sequential).SetText("Unfilter").SetIcon(icons.FilterAltOff)
	giv.NewFuncButton(tb, tv.Table.OpenCSV).SetIcon(icons.Open)
	giv.NewFuncButton(tb, tv.Table.SaveCSV).SetIcon(icons.Save)
}

/*
func (tv *TableView) MimeDataType() string {
	return fi.DataCsv
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
	md[0].Type = fi.DataCsv
	return md
}

// FromMimeData returns records from csv of mime data
func (tv *TableView) FromMimeData(md mimedata.Mimes) [][]string {
	var recs [][]string
	for _, d := range md {
		if d.Type == fi.DataCsv {
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
	tv.SelectIdxAction(idx, events.SelectOne)
}

func (tv *TableView) ItemCtxtMenu(idx int) {
	var men gi.Menu
	tv.StdCtxtMenu(&men, idx)
	if len(men) > 0 {
		pos := tv.IdxPos(idx)
		gi.PopupMenu(men, pos.X, pos.Y, tv.ViewportSafe(), tv.Nm+"-menu")
	}
}
*/
