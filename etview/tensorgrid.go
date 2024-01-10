// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"image/color"
	"log"
	"strconv"

	"goki.dev/colors"
	"goki.dev/colors/colormap"
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/minmax"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/mat32/v2"
)

// TensorLayout are layout options for displaying tensors
type TensorLayout struct { //gti:add

	// even-numbered dimensions are displayed as Y*X rectangles -- this determines along which dimension to display any remaining odd dimension: OddRow = true = organize vertically along row dimension, false = organize horizontally across column dimension
	OddRow bool

	// if true, then the Y=0 coordinate is displayed from the top-down; otherwise the Y=0 coordinate is displayed from the bottom up, which is typical for emergent network patterns.
	TopZero bool

	// display the data as a bitmap image.  if a 2D tensor, then it will be a greyscale image.  if a 3D tensor with size of either the first or last dim = either 3 or 4, then it is a RGB(A) color image
	Image bool
}

// TensorDisp are options for displaying tensors
type TensorDisp struct { //gti:add
	TensorLayout

	// range to plot
	Range minmax.Range64 `view:"inline"`

	// if not using fixed range, this is the actual range of data
	MinMax minmax.F64 `view:"inline"`

	// the name of the color map to use in translating values to colors
	ColorMap giv.ColorMapName

	// what proportion of grid square should be filled by color block -- 1 = all, .5 = half, etc
	GridFill float32 `min:"0.1" max:"1" step:"0.1" def:"0.9,1"`

	// amount of extra space to add at dimension boundaries, as a proportion of total grid size
	DimExtra float32 `min:"0" max:"1" step:"0.02" def:"0.1,0.3"`

	// minimum size for grid squares -- they will never be smaller than this
	GridMinSize float32

	// maximum size for grid squares -- they will never be larger than this
	GridMaxSize float32

	// total preferred display size along largest dimension.
	// grid squares will be sized to fit within this size,
	// subject to harder GridMin / Max size constraints
	TotPrefSize float32

	// font size in standard point units for labels (e.g., SimMat)
	FontSize float32

	// our gridview, for update method
	GridView *TensorGrid `copy:"-" json:"-" xml:"-" view:"-"`
}

// Defaults sets defaults for values that are at nonsensical initial values
func (td *TensorDisp) Defaults() {
	if td.ColorMap == "" {
		td.ColorMap = "ColdHot"
	}
	if td.Range.Max == 0 && td.Range.Min == 0 {
		td.Range.SetMin(-1)
		td.Range.SetMax(1)
	}
	if td.GridMinSize == 0 {
		td.GridMinSize = 2
	}
	if td.GridMaxSize == 0 {
		td.GridMaxSize = 16
	}
	if td.TotPrefSize == 0 {
		td.TotPrefSize = 100
	}
	if td.GridFill == 0 {
		td.GridFill = 0.9
		td.DimExtra = 0.3
	}
	if td.FontSize == 0 {
		td.FontSize = 24
	}
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
	if op, has := tsr.MetaData("colormap"); has {
		td.ColorMap = giv.ColorMapName(op)
	}
	if op, has := tsr.MetaData("grid-fill"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridFill = float32(mv)
	}
	if op, has := tsr.MetaData("grid-min"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridMinSize = float32(mv)
	}
	if op, has := tsr.MetaData("grid-max"); has {
		mv, _ := strconv.ParseFloat(op, 32)
		td.GridMaxSize = float32(mv)
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

////////////////////////////////////////////////////////////////////////////
//  	TensorGrid

// TensorGrid is a widget that displays tensor values as a grid of colored squares.
type TensorGrid struct {
	gi.WidgetBase

	// the tensor that we view
	Tensor etensor.Tensor `set:"-"`

	// display options
	Disp TensorDisp

	// the actual colormap
	ColorMap *colormap.Map
}

func (tg *TensorGrid) OnInit() {
	tg.WidgetBase.OnInit()
	tg.Disp.GridView = tg
	tg.Disp.Defaults()
	tg.HandleEvents()
	tg.SetStyles()
}

func (tg *TensorGrid) SetStyles() {
	tg.Style(func(s *styles.Style) {
		ms := tg.MinSize()
		s.Min.X.Dot(ms.X)
		s.Min.Y.Dot(ms.Y)
		s.Grow.Set(1, 1)
	})
}

// SetTensor sets the tensor and triggers a display update
func (tg *TensorGrid) SetTensor(tsr etensor.Tensor) *TensorGrid {
	if _, ok := tsr.(*etensor.String); ok {
		log.Printf("TensorGrid: String tensors cannot be displayed using TensorGrid\n")
		return tg
	}
	tg.Tensor = tsr
	if tg.Tensor != nil {
		tg.Disp.FmMeta(tg.Tensor)
	}
	tg.Update()
	return tg
}

// OpenTensorView pulls up a TensorView of our tensor
func (tg *TensorGrid) OpenTensorView() {
	/*
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
	*/
}

func (tg *TensorGrid) HandleEvents() {
	tg.OnDoubleClick(func(e events.Event) {
		tg.OpenTensorView()
	})
	tg.On(events.ContextMenu, func(e events.Event) {
		d := gi.NewBody().AddTitle("Tensor Grid Display Options")
		giv.NewStructView(d).SetStruct(&tg.Disp)
		d.NewFullDialog(tg).Run()
		// todo: update view
	})
}

// MinSize returns minimum size based on tensor and display settings
func (tg *TensorGrid) MinSize() mat32.Vec2 {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return mat32.Vec2{}
	}
	if tg.Disp.Image {
		return mat32.V2(float32(tg.Tensor.Dim(1)), float32(tg.Tensor.Dim(0)))
	}
	rows, cols, rowEx, colEx := etensor.Prjn2DShape(tg.Tensor.ShapeObj(), tg.Disp.OddRow)
	frw := float32(rows) + float32(rowEx)*tg.Disp.DimExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*tg.Disp.DimExtra // extra spacing
	mx := float32(max(frw, fcl))
	gsz := tg.Disp.TotPrefSize / mx
	gsz = max(gsz, tg.Disp.GridMinSize)
	gsz = min(gsz, tg.Disp.GridMaxSize)
	gsz = max(gsz, 2)
	return mat32.V2(gsz*float32(fcl), gsz*float32(frw))
}

// EnsureColorMap makes sure there is a valid color map that matches specified name
func (tg *TensorGrid) EnsureColorMap() {
	if tg.ColorMap != nil && tg.ColorMap.Name != string(tg.Disp.ColorMap) {
		tg.ColorMap = nil
	}
	if tg.ColorMap == nil {
		ok := false
		tg.ColorMap, ok = colormap.AvailMaps[string(tg.Disp.ColorMap)]
		if !ok {
			tg.Disp.ColorMap = ""
			tg.Disp.Defaults()
		}
		tg.ColorMap = colormap.AvailMaps[string(tg.Disp.ColorMap)]
	}
}

func (tg *TensorGrid) Color(val float64) (norm float64, clr color.Color) {
	if tg.ColorMap.Indexed {
		clr = tg.ColorMap.MapIndex(int(val))
	} else {
		norm = tg.Disp.Range.ClipNormVal(val)
		clr = tg.ColorMap.Map(float32(norm))
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
	tg.EnsureColorMap()
	tg.UpdateRange()

	pc, _ := tg.RenderLock()
	defer tg.RenderUnlock()

	pos := tg.Geom.Pos.Content
	sz := tg.Geom.Size.Actual.Content
	// sz.SetSubScalar(tg.Disp.BotRtSpace.Dots)

	pc.FillBox(pos, sz, tg.Styles.Background)

	tsr := tg.Tensor

	if tg.Disp.Image {
		ysz := tsr.Dim(0)
		xsz := tsr.Dim(1)
		nclr := 1
		outclr := false // outer dimension is color
		if tsr.NumDims() == 3 {
			if tsr.Dim(0) == 3 || tsr.Dim(0) == 4 {
				outclr = true
				ysz = tsr.Dim(1)
				xsz = tsr.Dim(2)
				nclr = tsr.Dim(0)
			} else {
				nclr = tsr.Dim(2)
			}
		}
		tsz := mat32.V2(float32(xsz), float32(ysz))
		gsz := sz.Div(tsz)
		for y := 0; y < ysz; y++ {
			for x := 0; x < xsz; x++ {
				ey := y
				if !tg.Disp.TopZero {
					ey = (ysz - 1) - y
				}
				switch {
				case outclr:
					var r, g, b, a float64
					a = 1
					r = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{0, y, x}))
					g = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{1, y, x}))
					b = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{2, y, x}))
					if nclr > 3 {
						a = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{3, y, x}))
					}
					cr := mat32.V2(float32(x), float32(ey))
					pr := pos.Add(cr.Mul(gsz))
					pc.StrokeStyle.Color = colors.C(colors.FromFloat64(r, g, b, a))
					pc.FillBox(pr, gsz, pc.StrokeStyle.Color)
				case nclr > 1:
					var r, g, b, a float64
					a = 1
					r = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 0}))
					g = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 1}))
					b = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 2}))
					if nclr > 3 {
						a = tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x, 3}))
					}
					cr := mat32.V2(float32(x), float32(ey))
					pr := pos.Add(cr.Mul(gsz))
					pc.StrokeStyle.Color = colors.C(colors.FromFloat64(r, g, b, a))
					pc.FillBox(pr, gsz, pc.StrokeStyle.Color)
				default:
					val := tg.Disp.Range.ClipNormVal(tsr.FloatVal([]int{y, x}))
					cr := mat32.V2(float32(x), float32(ey))
					pr := pos.Add(cr.Mul(gsz))
					pc.StrokeStyle.Color = colors.C(colors.FromFloat64(val, val, val, 1))
					pc.FillBox(pr, gsz, pc.StrokeStyle.Color)
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
	tsz := mat32.V2(fcl, frw)
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
			cr := mat32.V2(float32(x)+xex, float32(y)+yex)
			pr := pos.Add(cr.Mul(gsz))
			_, clr := tg.Color(val)
			pc.FillBoxColor(pr, ssz, clr)
		}
	}
}

func (tg *TensorGrid) Render() {
	if tg.PushBounds() {
		tg.RenderTensor()
		tg.RenderChildren()
		tg.PopBounds()
	}
}
