// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etview

import (
	"github.com/chewxy/math32"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// TensorGrid is a widget that displays tensor values as a grid of
// colored squares.
type TensorGrid struct {
	gi.WidgetBase
	Tensor   etensor.Tensor   `desc:"the tensor that we view"`
	ColorMap giv.ColorMapName `desc:"the name of the color map to use in translating values to colors"`
	Map      *giv.ColorMap    `desc:"the actual colormap"`
}

var KiT_TensorGrid = kit.Types.AddType(&TensorGrid{}, nil)

// AddNewTensorGrid adds a new colorview to given parent node, with given name.
func AddNewTensorGrid(parent ki.Ki, name string, cmap *ColorMap) *TensorGrid {
	cv := parent.AddNewChild(KiT_TensorGrid, name).(*TensorGrid)
	cv.Map = cmap
	return cv
}

func (cv *TensorGrid) Disconnect() {
	cv.WidgetBase.Disconnect()
	cv.ColorMapSig.DisconnectAll()
}

// SetColorMap sets the color map and triggers a display update
func (cv *TensorGrid) SetColorMap(cmap *ColorMap) {
	cv.Map = cmap
	cv.UpdateSig()
}

// SetColorMapAction sets the color map and triggers a display update
// and signals the ColorMapSig signal
func (cv *TensorGrid) SetColorMapAction(cmap *ColorMap) {
	cv.Map = cmap
	cv.ColorMapSig.Emit(cv.This(), 0, nil)
	cv.UpdateSig()
}

// ChooseColorMap pulls up a chooser to select a color map
func (cv *TensorGrid) ChooseColorMap() {
	sl := AvailColorMapsList()
	cur := ""
	if cv.Map != nil {
		cur = cv.Map.Name
	}
	SliceViewSelectDialog(cv.Viewport, &sl, cur, DlgOpts{Title: "Select a ColorMap", Prompt: "choose color map to use from among available list"}, nil,
		cv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.DialogAccepted) {
				ddlg := send.Embed(gi.KiT_Dialog).(*gi.Dialog)
				si := SliceViewSelectDialogValue(ddlg)
				if si >= 0 {
					nmap, ok := AvailColorMaps[sl[si]]
					if ok {
						cv.SetColorMapAction(nmap)
					}
				}
			}
		})
}

// MouseEvent handles button MouseEvent
func (cv *TensorGrid) MouseEvent() {
	cv.ConnectEvent(oswin.MouseEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		cvv := recv.(*TensorGrid)
		if me.Button == mouse.Left {
			switch me.Action {
			case mouse.DoubleClick: // we just count as a regular click
				fallthrough
			case mouse.Press:
				me.SetProcessed()
				cvv.ChooseColorMap()
			}
		}
	})
}

func (cv *TensorGrid) ConnectEvents2D() {
	cv.MouseEvent()
	cv.HoverTooltipEvent()
}

func (cv *TensorGrid) RenderColorMap() {
	if cv.Map == nil {
		cv.Map = StdColorMaps["ColdHot"]
	}
	rs := &cv.Viewport.Render
	rs.Lock()
	pc := &rs.Paint

	pos := cv.LayData.AllocPos
	sz := cv.LayData.AllocSize

	lsz := sz.Dim(cv.Orient)
	inc := math32.Ceil(lsz / 100)
	if inc < 2 {
		inc = 2
	}
	for p := float32(0); p < lsz; p += inc {
		val := p / (lsz - 1)
		clr := cv.Map.Map(float64(val))
		if cv.Orient == gi.X {
			pr := pos
			pr.X += p
			sr := sz
			sr.X = inc
			pc.FillBoxColor(rs, pr, sr, clr)
		} else {
			pr := pos
			pr.Y += p
			sr := sz
			sr.Y = inc
			pc.FillBoxColor(rs, pr, sr, clr)
		}
	}
	rs.Unlock()
}

func (cv *TensorGrid) Render2D() {
	if cv.FullReRenderIfNeeded() {
		return
	}
	if cv.PushBounds() {
		cv.This().(gi.Node2D).ConnectEvents2D()
		cv.RenderColorMap()
		cv.Render2DChildren()
		cv.PopBounds()
	} else {
		cv.DisconnectAllEvents(gi.RegPri)
	}
}
