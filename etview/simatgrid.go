// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"github.com/chewxy/math32"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/simat"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/mat32"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

const LabelSpace = float32(8)

// SimMatGrid is a widget that displays a similarity / distance matrix
// with tensor values as a grid of colored squares, and labels for rows, cols
type SimMatGrid struct {
	TensorGrid
	SimMat      *simat.SimMat `desc:"the similarity / distance matrix"`
	rowMaxSz    gi.Vec2D      // maximum label size
	rowMinBlank int           // minimum number of blank rows
	rowNGps     int           // number of groups in row (non-blank after blank)
	colMaxSz    gi.Vec2D      // maximum label size
	colMinBlank int           // minimum number of blank cols
	colNGps     int           // number of groups in col (non-blank after blank)
}

var KiT_SimMatGrid = kit.Types.AddType(&SimMatGrid{}, nil)

// AddNewSimMatGrid adds a new tensor grid to given parent node, with given name.
func AddNewSimMatGrid(parent ki.Ki, name string, smat *simat.SimMat) *SimMatGrid {
	tg := parent.AddNewChild(KiT_SimMatGrid, name).(*SimMatGrid)
	tg.SimMat = smat
	tg.Tensor = smat.Mat
	return tg
}

// Defaults sets defaults for values that are at nonsensical initial values
func (tg *SimMatGrid) Defaults() {
	tg.Disp.GridView = &tg.TensorGrid
	tg.Disp.Defaults()
	tg.Disp.TopZero = true
	if tg.Tensor != nil {
		tg.Disp.FmMeta(tg.Tensor)
	}
}

// func (tg *SimMatGrid) Disconnect() {
// 	tg.WidgetBase.Disconnect()
// 	tg.ColorMapSig.DisconnectAll()
// }

// SetSimMat sets the similarity matrix and triggers a display update
func (tg *SimMatGrid) SetSimMat(smat *simat.SimMat) {
	tg.SimMat = smat
	tg.Tensor = smat.Mat
	tg.Defaults()
	tg.UpdateSig()
}

// MouseEvent handles button MouseEvent
func (tg *SimMatGrid) MouseEvent() {
	tg.ConnectEvent(oswin.MouseEvent, gi.RegPri, func(retg, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		tgv := retg.(*SimMatGrid)
		switch {
		case me.Button == mouse.Right && me.Action == mouse.Press:
			giv.StructViewDialog(tgv.Viewport, &tgv.Disp, giv.DlgOpts{Title: "SimMatGrid Display Options", Ok: true, Cancel: true}, nil, nil)
		case me.Button == mouse.Left && me.Action == mouse.Press:
			me.SetProcessed()
			tgv.OpenTensorView()
		}
	})
}

func (tg *SimMatGrid) ConnectEvents2D() {
	tg.MouseEvent()
	tg.HoverTooltipEvent()
}

func (tg *SimMatGrid) Style2D() {
	tg.WidgetBase.Style2D()
	tg.Disp.Defaults()
	tg.Disp.ToDots(&tg.Sty.UnContext)
}

func (tg *SimMatGrid) Size2DLabel(lbs []string, col bool) (minBlank, ngps int, sz gi.Vec2D) {
	mx := 0
	mxi := 0
	minBlank = len(lbs)
	curblk := 0
	ngps = 0
	for i, lb := range lbs {
		l := len(lb)
		if l == 0 {
			curblk++
		} else {
			if curblk > 0 {
				ngps++
			}
			if i > 0 {
				minBlank = ints.MinInt(minBlank, curblk)
			}
			curblk = 0
			if l > mx {
				mx = l
				mxi = i
			}
		}
	}
	minBlank = ints.MinInt(minBlank, curblk)
	tr := gi.TextRender{}
	if col {
		tr.SetStringRot90(lbs[mxi], &tg.Sty.Font, &tg.Sty.UnContext, &tg.Sty.Text, true, 0)
	} else {
		tr.SetString(lbs[mxi], &tg.Sty.Font, &tg.Sty.UnContext, &tg.Sty.Text, true, 0, 0)
	}
	tsz := tg.LayData.SizePrefOrMax()
	if !col {
		tr.LayoutStdLR(&tg.Sty.Text, &tg.Sty.Font, &tg.Sty.UnContext, tsz)
	}
	return minBlank, ngps, tr.Size
}

func (tg *SimMatGrid) Size2D(iter int) {
	if iter > 0 {
		return // already updated in previous iter, don't redo!
	} else {
		tg.rowMinBlank, tg.rowNGps, tg.rowMaxSz = tg.Size2DLabel(tg.SimMat.Rows, false)
		tg.colMinBlank, tg.colNGps, tg.colMaxSz = tg.Size2DLabel(tg.SimMat.Cols, true)

		tg.colMaxSz.Y += tg.rowMaxSz.Y // needs one more for some reason

		rtxtsz := tg.rowMaxSz.Y / float32(tg.rowMinBlank+1)
		ctxtsz := tg.colMaxSz.X / float32(tg.colMinBlank+1)
		txtsz := mat32.Max(rtxtsz, ctxtsz)

		tg.InitLayout2D()
		rows, cols, rowEx, colEx := etensor.Prjn2DShape(tg.Tensor, tg.Disp.OddRow)
		rowEx = tg.rowNGps
		colEx = tg.colNGps
		frw := float32(rows) + float32(rowEx)*GridExtra // extra spacing
		fcl := float32(cols) + float32(colEx)*GridExtra // extra spacing
		tg.Disp.ToDots(&tg.Sty.UnContext)
		max := float32(math32.Max(frw, fcl))
		gsz := tg.Disp.TotPrefSize.Dots / max
		gsz = math32.Max(gsz, tg.Disp.GridMinSize.Dots)
		gsz = math32.Max(gsz, txtsz)
		gsz = math32.Min(gsz, tg.Disp.GridMaxSize.Dots)
		tg.Size2DFromWH(tg.rowMaxSz.X+LabelSpace+gsz*float32(cols), tg.colMaxSz.Y+LabelSpace+gsz*float32(rows))
	}
}

func (tg *SimMatGrid) RenderSimMat() {
	if tg.SimMat == nil || tg.SimMat.Mat.Len() == 0 {
		return
	}
	tg.Defaults()
	tg.EnsureColorMap()
	tg.UpdateRange()
	rs := &tg.Viewport.Render
	rs.Lock()
	pc := &rs.Paint

	pos := tg.LayData.AllocPos
	sz := tg.LayData.AllocSize
	effsz := sz
	effsz.X -= tg.rowMaxSz.X + LabelSpace
	effsz.Y -= tg.colMaxSz.Y + LabelSpace

	pc.FillBoxColor(rs, pos, sz, tg.Disp.Background)

	tsr := tg.SimMat.Mat

	rows, cols, rowEx, colEx := etensor.Prjn2DShape(tsr, tg.Disp.OddRow)
	rowEx = tg.rowNGps
	colEx = tg.colNGps
	frw := float32(rows) + float32(rowEx)*GridExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*GridExtra // extra spacing
	tsz := gi.Vec2D{fcl, frw}
	gsz := effsz.Div(tsz)

	// Render Rows
	epos := pos
	epos.Y += tg.colMaxSz.Y + LabelSpace
	nr := len(tg.SimMat.Rows)
	mx := ints.MinInt(nr, rows)
	tr := gi.TextRender{}
	txsty := tg.Sty.Text
	txsty.AlignV = gi.AlignTop
	ygp := 0
	prvyblk := false
	for y := 0; y < mx; y++ {
		lb := tg.SimMat.Rows[y]
		if len(lb) == 0 {
			prvyblk = true
			continue
		}
		if prvyblk {
			ygp++
			prvyblk = false
		}
		yex := float32(ygp) * GridExtra
		tr.SetString(lb, &tg.Sty.Font, &tg.Sty.UnContext, &txsty, true, 0, 0)
		tr.LayoutStdLR(&txsty, &tg.Sty.Font, &tg.Sty.UnContext, tg.rowMaxSz)
		cr := gi.Vec2D{0, float32(y) + yex}
		pr := epos.Add(cr.Mul(gsz))
		tr.Render(rs, pr)
	}

	// Render Cols
	epos = pos
	epos.X += tg.rowMaxSz.X + LabelSpace
	nc := len(tg.SimMat.Cols)
	mx = ints.MinInt(nc, cols)
	xgp := 0
	prvxblk := false
	for x := 0; x < mx; x++ {
		lb := tg.SimMat.Cols[x]
		if len(lb) == 0 {
			prvxblk = true
			continue
		}
		if prvxblk {
			xgp++
			prvxblk = false
		}
		xex := float32(xgp) * GridExtra
		tr.SetStringRot90(lb, &tg.Sty.Font, &tg.Sty.UnContext, &tg.Sty.Text, true, 0)
		cr := gi.Vec2D{float32(x) + xex, 0}
		pr := epos.Add(cr.Mul(gsz))
		tr.Render(rs, pr)
	}

	pos.X += tg.rowMaxSz.X + LabelSpace
	pos.Y += tg.colMaxSz.Y + LabelSpace
	ssz := gsz.MulVal(.9) // smaller size with margin
	prvyblk = false
	ygp = 0
	for y := 0; y < rows; y++ {
		ylb := tg.SimMat.Rows[y]
		if len(ylb) > 0 && prvyblk {
			ygp++
			prvyblk = false
		}
		yex := float32(ygp) * GridExtra
		prvxblk = false
		xgp = 0
		for x := 0; x < cols; x++ {
			xlb := tg.SimMat.Cols[x]
			if len(xlb) > 0 && prvxblk {
				xgp++
				prvxblk = false
			}
			xex := float32(xgp) * GridExtra
			ey := y
			if !tg.Disp.TopZero {
				ey = (rows - 1) - y
			}
			val := etensor.Prjn2DVal(tsr, tg.Disp.OddRow, ey, x)
			cr := gi.Vec2D{float32(x) + xex, float32(y) + yex}
			pr := pos.Add(cr.Mul(gsz))
			_, clr := tg.Color(val)
			pc.FillBoxColor(rs, pr, ssz, clr)
			if len(xlb) == 0 {
				prvxblk = true
			}
		}
		if len(ylb) == 0 {
			prvyblk = true
		}
	}

	rs.Unlock()
}

func (tg *SimMatGrid) Render2D() {
	if tg.FullReRenderIfNeeded() {
		return
	}
	if tg.PushBounds() {
		tg.This().(gi.Node2D).ConnectEvents2D()
		tg.RenderSimMat()
		tg.Render2DChildren()
		tg.PopBounds()
	} else {
		tg.DisconnectAllEvents(gi.RegPri)
	}
}
