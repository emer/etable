// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"cogentcore.org/core/gi"
	"cogentcore.org/core/giv"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/laser"
	"cogentcore.org/core/styles"
)

func init() {
	giv.AddValue(Plot2D{}, func() giv.Value {
		return &Plot2DValue{}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  Plot2DValue

// Plot2DValue presents a button that pulls up the Plot2D in a dialog
type Plot2DValue struct {
	giv.ValueBase[*gi.Button]
}

func (v *Plot2DValue) Config() {
	v.Widget.SetType(gi.ButtonTonal).SetIcon(icons.Edit)
	giv.ConfigDialogWidget(v, true)
}

func (v *Plot2DValue) Update() {
	npv := laser.NonPtrValue(v.Value)
	if !v.Value.IsValid() || v.Value.IsZero() || !npv.IsValid() || npv.IsZero() {
		v.Widget.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(v.Value)
		plot := opv.Interface().(*Plot2D)
		if plot != nil && plot.Table != nil && plot.Table.Table != nil {
			if nm, has := plot.Table.Table.MetaData["name"]; has {
				v.Widget.SetText(nm)
			} else {
				v.Widget.SetText("eplot.Plot2D")
			}
		} else {
			v.Widget.SetText("eplot.Plot2D")
		}
	}
}

func (v *Plot2DValue) ConfigDialog(d *gi.Body) (bool, func()) {
	opv := laser.OnePtrUnderlyingValue(v.Value)
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
