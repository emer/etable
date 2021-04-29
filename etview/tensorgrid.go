// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"log"
	"strconv"

	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// TensorLayout are layout options for displaying tensors
type TensorLayout struct {
	OddRow  bool `desc:"even-numbered dimensions are displayed as Y*X rectangles -- this determines along which dimension to display any remaining odd dimension: OddRow = true = organize vertically along row dimension, false = organize horizontally across column dimension"`
	TopZero bool `desc:"if true, then the Y=0 coordinate is displayed from the top-down; otherwise the Y=0 coordinate is displayed from the bottom up, which is typical for emergent network patterns."`
	Image   bool `desc:"display the data as a bitmap image.  if a 2D tensor, then it will be a greyscale image.  if a 3D tensor with size of either the first or last dim = either 3 or 4, then it is a RGB(A) color image"`
}

// TensorDisp are options for displaying tensors
type TensorDisp struct {
	TensorLayout
	Range       minmax.Range64   `view:"inline" desc:"range to plot"`
	MinMax      minmax.F64       `view:"inline" desc:"if not using fixed range, this is the actual range of data"`
	ColorMap    giv.ColorMapName `desc:"the name of the color map to use in translating values to colors"`
	Background  gist.Color       `desc:"background color"`
	GridFill    float32          `min:"0.1" max:"1" step:"0.1" def:"0.9,1" desc:"what proportion of grid square should be filled by color block -- 1 = all, .5 = half, etc"`
	DimExtra    float32          `min:"0" max:"1" step:"0.02" def:"0.1,0.3" desc:"amount of extra space to add at dimension boundaries, as a proportion of total grid size"`
	BotRtSpace  units.Value      `desc:"extra space to add at bottom of grid -- needed when included in TableView for example"`
	GridMinSize units.Value      `desc:"minimum size for grid squares -- they will never be smaller than this"`
	GridMaxSize units.Value      `desc:"maximum size for grid squares -- they will never be larger than this"`
	TotPrefSize units.Value      `desc:"total preferred display size along largest dimension -- grid squares will be sized to fit within this size, subject to harder GridMin / Max size constraints"`
	FontSize    float32          `desc:"font size in standard point units for labels (e.g., SimMat)"`
	GridView    *TensorGrid      `copy:"-" json:"-" xml:"-" view:"-" desc:"our gridview, for update method"`
}

var KiT_TensorDisp = kit.Types.AddType(&TensorDisp{}, TensorDispProps)

// Defaults sets defaults for values that are at nonsensical initial values
func (td *TensorDisp) Defaults() {
	if td.ColorMap == "" {
		td.ColorMap = "ColdHot"
		td.Background.SetName("white")
	}
	if td.Range.Max == 0 && td.Range.Min == 0 {
		td.Range.SetMin(-1)
		td.Range.SetMax(1)
	}
	if td.GridMinSize.Val == 0 {
		td.GridMinSize.Set(4, units.Px)
	}
	if td.GridMaxSize.Val == 0 {
		td.GridMaxSize.Set(2, units.Em)
	}
	if td.TotPrefSize.Val == 0 {
		td.TotPrefSize.Set(20, units.Em)
	}
	if td.GridFill == 0 {
		td.GridFill = 0.9
		td.DimExtra = 0.3
	}
	if td.FontSize == 0 {
		td.FontSize = 24
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (td *TensorDisp) Update() {
	if td.GridView != nil {
		td.GridView.UpdateSig()
	}
}

func (td *TensorDisp) ToDots(uc *units.Context) {
	td.BotRtSpace.ToDots(uc)
	td.GridMinSize.ToDots(uc)
	td.GridMaxSize.ToDots(uc)
	td.TotPrefSize.ToDots(uc)
}

// FmMeta sets display options from Tensor meta-data
func (td *TensorDisp) FmMeta(tsr etensor.Tensor) {
	if op, has := tsr.MetaData("top-zero"); has {
		if op == "+" || op == "true" {
			td.TopZero = true
		}
	}
	if op, has := tsr.MetaData("odd-row"); has {
		if op == "+" || op == "true" {
			td.OddRow = true
		}
	}
	if op, has := tsr.MetaData("image"); has {
		if op == "+" || op == "true" {
			td.Image = true
		}
	}
	if op, has := tsr.MetaData("min"); has {
		mv, _ := strconv.ParseFloat(op, 64)
		td.Range.Min = mv
	}
	if op, has := tsr.MetaData("max"); has {
		mv, _ := strconv.ParseFloat(op, 64)
		td.Range.Max = mv
	}
	if op, has := tsr.MetaData("fix-min"); has {
		if op == "+" || op == "true" {
			td.Range.FixMin = true
		} else {
			td.Range.FixMin = false
		}
	}
	if op, has := tsr.MetaData("fix-max"); has {
		if op == "+" || op == "true" {
			td.Range.FixMax = true
		} else {
			td.Range.FixMax = false
		}
	}
	if op, has := tsr.MetaData("background"); has {
		td.Background.SetString(op, nil)
	}
	if op, has := tsr.MetaData("colormap"); has {
		td.ColorMap = giv.ColorMapName(op)
	}
	if op, has := tsr.MetaData("grid-fill"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridFill = float32(mv)
	}
	if op, has := tsr.MetaData("grid-min"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridMinSize.Val = float32(mv)
	}
	if op, has := tsr.MetaData("grid-max"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridMaxSize.Val = float32(mv)
	}
	if op, has := tsr.MetaData("dim-extra"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.DimExtra = float32(mv)
	}
	if op, has := tsr.MetaData("font-size"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.FontSize = float32(mv)
	}
}

// TensorGrid is a widget that displays tensor values as a grid of colored squares.
type TensorGrid struct {
	gi.WidgetBase
	Tensor etensor.Tensor `desc:"the tensor that we view"`
	Disp   TensorDisp     `desc:"display options"`
	Map    *giv.ColorMap  `desc:"the actual colormap"`
}

var KiT_TensorGrid = kit.Types.AddType(&TensorGrid{}, nil)

// AddNewTensorGrid adds a new tensor grid to given parent node, with given name.
func AddNewTensorGrid(parent ki.Ki, name string, tsr etensor.Tensor) *TensorGrid {
	tg := parent.AddNewChild(KiT_TensorGrid, name).(*TensorGrid)
	tg.Tensor = tsr
	return tg
}

// Defaults sets defaults for values that are at nonsensical initial values
func (tg *TensorGrid) Defaults() {
	tg.Disp.GridView = tg
	tg.Disp.Defaults()
}

// func (tg *TensorGrid) Disconnect() {
// 	tg.WidgetBase.Disconnect()
// 	tg.ColorMapSig.DisconnectAll()
// }

// SetTensor sets the tensor and triggers a display update
func (tg *TensorGrid) SetTensor(tsr etensor.Tensor) {
	if _, ok := tsr.(*etensor.String); ok {
		log.Printf("TensorGrid: String tensors cannot be displayed using TensorGrid\n")
		return
	}
	tg.Tensor = tsr
	tg.Defaults()
	if tg.Tensor != nil {
		tg.Disp.FmMeta(tg.Tensor)
	}
	tg.UpdateSig()
}

// OpenTensorView pulls up a TensorView of our tensor
func (tg *TensorGrid) OpenTensorView() {
	dlg := TensorViewDialog(tg.ViewportSafe(), tg.Tensor, giv.DlgOpts{Title: "Edit Tensor", Prompt: "", NoAdd: true, NoDelete: true}, nil, nil)
	tvk := dlg.Frame().ChildByType(KiT_TensorView, true, 2)
	if tvk != nil {
		tv := tvk.(*TensorView)
		tv.TsrLay = tg.Disp.TensorLayout
		tv.SetInactiveState(tg.IsInactive())
		tv.ViewSig.Connect(tg.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			tgg, _ := recv.Embed(KiT_TensorGrid).(*TensorGrid)
			tgg.UpdateSig()
		})
	}
}

// MouseEvent handles button MouseEvent
func (tg *TensorGrid) MouseEvent() {
	tg.ConnectEvent(oswin.MouseEvent, gi.RegPri, func(retg, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		tgv := retg.(*TensorGrid)
		switch {
		case me.Button == mouse.Right && me.Action == mouse.Press:
			giv.StructViewDialog(tgv.ViewportSafe(), &tgv.Disp, giv.DlgOpts{Title: "TensorGrid Display Options", Ok: true, Cancel: true}, nil, nil)
		case me.Button == mouse.Left && me.Action == mouse.DoubleClick:
			me.SetProcessed()
			tgv.OpenTensorView()
		}
	})
}

func (tg *TensorGrid) ConnectEvents2D() {
	tg.MouseEvent()
	tg.HoverTooltipEvent()
}

func (tg *TensorGrid) Style2D() {
	tg.WidgetBase.Style2D()
	tg.Disp.Defaults()
	tg.Disp.ToDots(&tg.Sty.UnContext)
}

func (tg *TensorGrid) Size2D(iter int) {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return
	}
	if iter > 0 {
		return // already updated in previous iter, don't redo!
	} else {
		if tg.Disp.Image {
			tg.Size2DFromWH(float32(tg.Tensor.Dim(1)), float32(tg.Tensor.Dim(0)))
		} else {
			tg.InitLayout2D()
			rows, cols, rowEx, colEx := etensor.Prjn2DShape(tg.Tensor.ShapeObj(), tg.Disp.OddRow)
			frw := float32(rows) + float32(rowEx)*tg.Disp.DimExtra // extra spacing
			fcl := float32(cols) + float32(colEx)*tg.Disp.DimExtra // extra spacing
			tg.Disp.ToDots(&tg.Sty.UnContext)
			max := float32(mat32.Max(frw, fcl))
			gsz := tg.Disp.TotPrefSize.Dots / max
			gsz = mat32.Max(gsz, tg.Disp.GridMinSize.Dots)
			gsz = mat32.Min(gsz, tg.Disp.GridMaxSize.Dots)
			tg.Size2DFromWH(gsz*float32(cols)+tg.Disp.BotRtSpace.Dots, gsz*float32(rows)+tg.Disp.BotRtSpace.Dots)
		}
	}
}

// EnsureColorMap makes sure there is a valid color map that matches specified name
func (tg *TensorGrid) EnsureColorMap() {
	if tg.Map != nil && tg.Map.Name != string(tg.Disp.ColorMap) {
		tg.Map = nil
	}
	if tg.Map == nil {
		ok := false
		tg.Map, ok = giv.AvailColorMaps[string(tg.Disp.ColorMap)]
		if !ok {
			tg.Disp.ColorMap = ""
			tg.Disp.Defaults()
		}
		tg.Map = giv.AvailColorMaps[string(tg.Disp.ColorMap)]
	}
}

func (tg *TensorGrid) Color(val float64) (norm float64, clr gist.Color) {
	if tg.Map.Indexed {
		clr = tg.Map.MapIndex(int(val))
	} else {
		norm = tg.Disp.Range.ClipNormVal(val)
		clr = tg.Map.Map(norm)
	}
	return
}

func (tg *TensorGrid) UpdateRange() {
	if !tg.Disp.Range.FixMin || !tg.Disp.Range.FixMax {
		min, max, _, _ := tg.Tensor.Range()
		if !tg.Disp.Range.FixMin {
			nmin := minmax.NiceRoundNumber(min, true) // true = below #
			tg.Disp.Range.Min = nmin
		}
		if !tg.Disp.Range.FixMax {
			nmax := minmax.NiceRoundNumber(max, false) // false = above #
			tg.Disp.Range.Max = nmax
		}
	}
}

func (tg *TensorGrid) RenderTensor() {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return
	}
	tg.Defaults()
	tg.EnsureColorMap()
	tg.UpdateRange()

	rs, pc, _ := tg.RenderLock()
	defer tg.RenderUnlock(rs)

	pos := tg.LayState.Alloc.Pos
	sz := tg.LayState.Alloc.Size
	sz.SetSubScalar(tg.Disp.BotRtSpace.Dots)

	pc.FillBoxColor(rs, pos, sz, tg.Disp.Background)

	tsr := tg.Tensor

	if tg.Disp.Image {
		posx := int(pos.X)
		posy := int(pos.Y)
		ysz := tsr.Dim(0)
		xsz := tsr.Dim(1)
		nclr := 1
		if tsr.NumDims() == 3 {
			nclr = tsr.Dim(2)
		}
		for y := 0; y < ysz; y++ {
			for x := 0; x < xsz; x++ {
				if nclr > 1 {
					var r, g, b, a float64
					a = 1
					r = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 0}))
					g = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 1}))
					b = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 2}))
					if nclr > 3 {
						a = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 3}))
					}
					pc.StrokeStyle.Color.Color.SetFloat64(r, g, b, a)
					pc.SetPixel(rs, posx+x, posy+y)
				} else {
					val := tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x}))
					pc.StrokeStyle.Color.Color.SetFloat64(val, val, val, 1)
					pc.SetPixel(rs, posx+x, posy+y)
				}
			}
		}
		return
	}
	rows, cols, rowEx, colEx := etensor.Prjn2DShape(tsr.ShapeObj(), tg.Disp.OddRow)
	frw := float32(rows) + float32(rowEx)*tg.Disp.DimExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*tg.Disp.DimExtra // extra spacing
	rowsInner := rows
	colsInner := cols
	if rowEx > 0 {
		rowsInner = rows / rowEx
	}
	if colEx > 0 {
		colsInner = cols / colEx
	}
	tsz := mat32.Vec2{fcl, frw}
	gsz := sz.Div(tsz)

	ssz := gsz.MulScalar(tg.Disp.GridFill) // smaller size with margin
	for y := 0; y < rows; y++ {
		yex := float32(int(y/rowsInner)) * tg.Disp.DimExtra
		for x := 0; x < cols; x++ {
			xex := float32(int(x/colsInner)) * tg.Disp.DimExtra
			ey := y
			if !tg.Disp.TopZero {
				ey = (rows - 1) - y
			}
			val := etensor.Prjn2DVal(tsr, tg.Disp.OddRow, ey, x)
			cr := mat32.Vec2{float32(x) + xex, float32(y) + yex}
			pr := pos.Add(cr.Mul(gsz))
			_, clr := tg.Color(val)
			pc.FillBoxColor(rs, pr, ssz, clr)
		}
	}
}

func (tg *TensorGrid) Render2D() {
	if tg.FullReRenderIfNeeded() {
		return
	}
	if tg.PushBounds() {
		tg.This().(gi.Node2D).ConnectEvents2D()
		tg.RenderTensor()
		tg.Render2DChildren()
		tg.PopBounds()
	} else {
		tg.DisconnectAllEvents(gi.RegPri)
	}
}

var TensorDispProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"Update", ki.Props{
			"label": "Update Grid",
			"icon":  "update",
		}},
	},
}
