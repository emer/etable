// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"goki.dev/etable/v2/etensor"
	"goki.dev/etable/v2/simat"
	"goki.dev/girl/paint"
	"goki.dev/girl/styles"
	"goki.dev/mat32/v2"
)

const LabelSpace = float32(8)

// SimMatGrid is a widget that displays a similarity / distance matrix
// with tensor values as a grid of colored squares, and labels for rows, cols
type SimMatGrid struct { //gti:add
	TensorGrid

	// the similarity / distance matrix
	SimMat *simat.SimMat `set:"-"`

	rowMaxSz    mat32.Vec2 // maximum label size
	rowMinBlank int        // minimum number of blank rows
	rowNGps     int        // number of groups in row (non-blank after blank)
	colMaxSz    mat32.Vec2 // maximum label size
	colMinBlank int        // minimum number of blank cols
	colNGps     int        // number of groups in col (non-blank after blank)
}

// Defaults sets defaults for values that are at nonsensical initial values
func (tg *SimMatGrid) OnInit() {
	tg.TensorGrid.OnInit()
	tg.Disp.GridView = &tg.TensorGrid
	tg.Disp.Defaults()
	tg.Disp.TopZero = true

}

// SetSimMat sets the similarity matrix and triggers a display update
func (tg *SimMatGrid) SetSimMat(smat *simat.SimMat) {
	tg.SimMat = smat
	tg.Tensor = smat.Mat
	if tg.Tensor != nil {
		tg.Disp.FmMeta(tg.Tensor)
	}
	tg.Update()
}

func (tg *SimMatGrid) SizeLabel(lbs []string, col bool) (minBlank, ngps int, sz mat32.Vec2) {
	mx := 0
	mxi := 0
	minBlank = len(lbs)
	if minBlank == 0 {
		return
	}
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
				minBlank = min(minBlank, curblk)
			}
			curblk = 0
			if l > mx {
				mx = l
				mxi = i
			}
		}
	}
	minBlank = min(minBlank, curblk)
	tr := paint.Text{}
	fr := tg.Styles.FontRender()
	if col {
		tr.SetStringRot90(lbs[mxi], fr, &tg.Styles.UnContext, &tg.Styles.Text, true, 0)
	} else {
		tr.SetString(lbs[mxi], fr, &tg.Styles.UnContext, &tg.Styles.Text, true, 0, 0)
	}
	tsz := tg.Geom.Size.Actual.Content
	if !col {
		tr.LayoutStdLR(&tg.Styles.Text, fr, &tg.Styles.UnContext, tsz)
	}
	return minBlank, ngps, tr.Size
}

func (tg *SimMatGrid) MinSize() mat32.Vec2 {
	tg.rowMinBlank, tg.rowNGps, tg.rowMaxSz = tg.SizeLabel(tg.SimMat.Rows, false)
	tg.colMinBlank, tg.colNGps, tg.colMaxSz = tg.SizeLabel(tg.SimMat.Cols, true)

	tg.colMaxSz.Y += tg.rowMaxSz.Y // needs one more for some reason

	rtxtsz := tg.rowMaxSz.Y / float32(tg.rowMinBlank+1)
	ctxtsz := tg.colMaxSz.X / float32(tg.colMinBlank+1)
	txtsz := mat32.Max(rtxtsz, ctxtsz)

	rows, cols, rowEx, colEx := etensor.Prjn2DShape(tg.Tensor.ShapeObj(), tg.Disp.OddRow)
	rowEx = tg.rowNGps
	colEx = tg.colNGps
	frw := float32(rows) + float32(rowEx)*tg.Disp.DimExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*tg.Disp.DimExtra // extra spacing
	max := float32(mat32.Max(frw, fcl))
	gsz := tg.Disp.TotPrefSize / max
	gsz = mat32.Max(gsz, tg.Disp.GridMinSize)
	gsz = mat32.Max(gsz, txtsz)
	gsz = mat32.Min(gsz, tg.Disp.GridMaxSize)
	return mat32.Vec2{tg.rowMaxSz.X + LabelSpace + gsz*float32(cols), tg.colMaxSz.Y + LabelSpace + gsz*float32(rows)}
}

func (tg *SimMatGrid) RenderSimMat() {
	if tg.SimMat == nil || tg.SimMat.Mat.Len() == 0 {
		return
	}
	tg.EnsureColorMap()
	tg.UpdateRange()
	rs, pc, _ := tg.RenderLock()
	defer tg.RenderUnlock(rs)

	pos := tg.Geom.Pos.Content
	sz := tg.Geom.Size.Actual.Content

	effsz := sz
	effsz.X -= tg.rowMaxSz.X + LabelSpace
	effsz.Y -= tg.colMaxSz.Y + LabelSpace

	pc.FillBoxColor(rs, pos, sz, tg.Styles.BackgroundColor.Solid)

	tsr := tg.SimMat.Mat

	rows, cols, rowEx, colEx := etensor.Prjn2DShape(tsr.ShapeObj(), tg.Disp.OddRow)
	rowEx = tg.rowNGps
	colEx = tg.colNGps
	frw := float32(rows) + float32(rowEx)*tg.Disp.DimExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*tg.Disp.DimExtra // extra spacing
	tsz := mat32.Vec2{fcl, frw}
	gsz := effsz.Div(tsz)

	// Render Rows
	epos := pos
	epos.Y += tg.colMaxSz.Y + LabelSpace
	nr := len(tg.SimMat.Rows)
	mx := min(nr, rows)
	tr := paint.Text{}
	txsty := tg.Styles.Text
	txsty.AlignV = styles.Start
	ygp := 0
	prvyblk := false
	fr := tg.Styles.FontRender()
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
		yex := float32(ygp) * tg.Disp.DimExtra
		tr.SetString(lb, fr, &tg.Styles.UnContext, &txsty, true, 0, 0)
		tr.LayoutStdLR(&txsty, fr, &tg.Styles.UnContext, tg.rowMaxSz)
		cr := mat32.Vec2{0, float32(y) + yex}
		pr := epos.Add(cr.Mul(gsz))
		tr.Render(rs, pr)
	}

	// Render Cols
	epos = pos
	epos.X += tg.rowMaxSz.X + LabelSpace
	nc := len(tg.SimMat.Cols)
	mx = min(nc, cols)
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
		xex := float32(xgp) * tg.Disp.DimExtra
		tr.SetStringRot90(lb, fr, &tg.Styles.UnContext, &tg.Styles.Text, true, 0)
		cr := mat32.Vec2{float32(x) + xex, 0}
		pr := epos.Add(cr.Mul(gsz))
		tr.Render(rs, pr)
	}

	pos.X += tg.rowMaxSz.X + LabelSpace
	pos.Y += tg.colMaxSz.Y + LabelSpace
	ssz := gsz.MulScalar(tg.Disp.GridFill) // smaller size with margin
	prvyblk = false
	ygp = 0
	for y := 0; y < rows; y++ {
		ylb := tg.SimMat.Rows[y]
		if len(ylb) > 0 && prvyblk {
			ygp++
			prvyblk = false
		}
		yex := float32(ygp) * tg.Disp.DimExtra
		prvxblk = false
		xgp = 0
		for x := 0; x < cols; x++ {
			xlb := tg.SimMat.Cols[x]
			if len(xlb) > 0 && prvxblk {
				xgp++
				prvxblk = false
			}
			xex := float32(xgp) * tg.Disp.DimExtra
			ey := y
			if !tg.Disp.TopZero {
				ey = (rows - 1) - y
			}
			val := etensor.Prjn2DVal(tsr, tg.Disp.OddRow, ey, x)
			cr := mat32.Vec2{float32(x) + xex, float32(y) + yex}
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
}

func (tg *SimMatGrid) Render() {
	if tg.PushBounds() {
		tg.RenderSimMat()
		tg.RenderChildren()
		tg.PopBounds()
	}
}
