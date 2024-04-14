// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/reflectx"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/views"
)

func init() {
	views.AddValue(Plot2D{}, func() views.Value {
		return &Plot2DValue{}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  Plot2DValue

// Plot2DValue presents a button that pulls up the Plot2D in a dialog
type Plot2DValue struct {
	views.ValueBase[*core.Button]
}

func (v *Plot2DValue) Config() {
	v.Widget.SetType(core.ButtonTonal).SetIcon(icons.Edit)
	views.ConfigDialogWidget(v, true)
}

func (v *Plot2DValue) Update() {
	npv := reflectx.NonPointerValue(v.Value)
	if !v.Value.IsValid() || v.Value.IsZero() || !npv.IsValid() || npv.IsZero() {
		v.Widget.SetText("nil")
	} else {
		opv := reflectx.OnePointerUnderlyingValue(v.Value)
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

func (v *Plot2DValue) ConfigDialog(d *core.Body) (bool, func()) {
	opv := reflectx.OnePointerUnderlyingValue(v.Value)
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
