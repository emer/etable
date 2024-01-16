// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"goki.dev/gi"
	"goki.dev/giv"
	"goki.dev/gti"
	"goki.dev/laser"
	"goki.dev/styles"
)

func init() {
	giv.ValueMapAdd(Plot2D{}, func() giv.Value {
		return &Plot2DValue{}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  Plot2DValue

// Plot2DValue presents a button that pulls up the Plot2D in a dialog
type Plot2DValue struct {
	giv.ValueBase
}

func (vv *Plot2DValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *Plot2DValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	if vv.Value.IsZero() || npv.IsZero() {
		ac.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		plot := opv.Interface().(*Plot2D)
		if plot != nil && plot.Table != nil && plot.Table.Table != nil {
			if nm, has := plot.Table.Table.MetaData["name"]; has {
				ac.SetText(nm)
			} else {
				ac.SetText("eplot.Plot2D")
			}
		} else {
			ac.SetText("eplot.Plot2D")
		}
	}
}

func (vv *Plot2DValue) ConfigWidget(w gi.Widget) {
	if vv.Widget == w {
		vv.UpdateWidget()
		return
	}
	vv.Widget = w
	vv.StdConfigWidget(w)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config()
	giv.ConfigDialogWidget(vv, bt, true)
	vv.UpdateWidget()
}

func (vv *Plot2DValue) HasDialog() bool                      { return true }
func (vv *Plot2DValue) OpenDialog(ctx gi.Widget, fun func()) { giv.OpenValueDialog(vv, ctx, fun) }

func (vv *Plot2DValue) ConfigDialog(d *gi.Body) (bool, func()) {
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	plot := opv.Interface().(*Plot2D)
	if plot == nil || plot.Table == nil {
		return false, nil
	}
	clplot := plot.Clone().(*Plot2D)
	d.AddChild(clplot)
	d.Style(func(s *styles.Style) {
		s.Min.X.Em(60)
		s.Min.Y.Em(30)
	})
	return true, nil
}
