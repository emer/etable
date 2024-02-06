// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"image/color"
	"strings"

	"cogentcore.org/core/glop/option"
	"cogentcore.org/core/laser"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/minmax"
)

// PlotParams are parameters for overall plot
type PlotParams struct { //gti:add

	// optional title at top of plot
	Title string

	// type of plot to generate.  For a Bar plot, items are plotted ordinally by row and the XAxis is optional
	Type PlotTypes

	// whether to plot lines
	Lines bool

	// whether to plot points with symbols
	Points bool

	// width of lines
	LineWidth float64

	// size of points (default is 3)
	PointSize float64

	// the shape used to draw points
	PointShape Shapes

	// width of bars for bar plot, as fraction of available space -- 1 = no gaps, .8 default
	BarWidth float64 `min:"0.01" max:"1"`

	// draw lines that connect points with a negative X-axis direction -- otherwise these are treated as breaks between repeated series and not drawn
	NegXDraw bool

	// overall scaling factor -- the larger the number, the larger the fonts are relative to the graph
	Scale float64 `default:"2"`

	// what column to use for the common X axis -- if empty or not found, the row number is used.  This optional for Bar plots -- if present and LegendCol is also present, then an extra space will be put between X values.
	XAxisCol string

	// optional column for adding a separate colored / styled line or bar according to this value -- acts just like a separate Y variable, crossed with Y variables
	LegendCol string

	// rotation of the X Axis labels, in degrees
	XAxisRot float64

	// optional label to use for XAxis instead of column name
	XAxisLabel string

	// optional label to use for YAxis -- if empty, first column name is used
	YAxisLabel string

	// our plot, for update method
	Plot *Plot2D `copier:"-" json:"-" xml:"-" view:"-"`
}

// Defaults sets defaults if nil vals present
func (pp *PlotParams) Defaults() {
	if pp.LineWidth == 0 {
		pp.LineWidth = 2
		pp.Lines = true
		pp.Points = false
		pp.PointSize = 3
		pp.BarWidth = .8
	}
	if pp.Scale == 0 {
		pp.Scale = 2
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (pp *PlotParams) Update() {
	if pp.BarWidth > 1 {
		pp.BarWidth = .8
	}
	if pp.Plot != nil {
		pp.Plot.Update()
	}
}

// CopyFrom copies from other col params
func (pp *PlotParams) CopyFrom(fr *PlotParams) {
	pl := pp.Plot
	*pp = *fr
	pp.Plot = pl
}

// FmMeta sets plot params from meta data
func (pp *PlotParams) FmMeta(dt *etable.Table) {
	pp.FmMetaMap(dt.MetaData)
}

// MetaMapLower tries meta data access by lower-case version of key too
func MetaMapLower(meta map[string]string, key string) (string, bool) {
	vl, has := meta[key]
	if has {
		return vl, has
	}
	vl, has = meta[strings.ToLower(key)]
	return vl, has
}

// FmMetaMap sets plot params from meta data map
func (pp *PlotParams) FmMetaMap(meta map[string]string) {
	if typ, has := MetaMapLower(meta, "Type"); has {
		pp.Type.SetString(typ)
	}
	if op, has := MetaMapLower(meta, "Lines"); has {
		if op == "+" || op == "true" {
			pp.Lines = true
		} else {
			pp.Lines = false
		}
	}
	if op, has := MetaMapLower(meta, "Points"); has {
		if op == "+" || op == "true" {
			pp.Points = true
		} else {
			pp.Points = false
		}
	}
	if lw, has := MetaMapLower(meta, "LineWidth"); has {
		pp.LineWidth, _ = laser.ToFloat(lw)
	}
	if ps, has := MetaMapLower(meta, "PointSize"); has {
		pp.PointSize, _ = laser.ToFloat(ps)
	}
	if bw, has := MetaMapLower(meta, "BarWidth"); has {
		pp.BarWidth, _ = laser.ToFloat(bw)
	}
	if op, has := MetaMapLower(meta, "NegXDraw"); has {
		if op == "+" || op == "true" {
			pp.NegXDraw = true
		} else {
			pp.NegXDraw = false
		}
	}
	if scl, has := MetaMapLower(meta, "Scale"); has {
		pp.Scale, _ = laser.ToFloat(scl)
	}
	if xc, has := MetaMapLower(meta, "XAxisCol"); has {
		pp.XAxisCol = xc
	}
	if lc, has := MetaMapLower(meta, "LegendCol"); has {
		pp.LegendCol = lc
	}
	if xrot, has := MetaMapLower(meta, "XAxisRot"); has {
		pp.XAxisRot, _ = laser.ToFloat(xrot)
	}
	if lb, has := MetaMapLower(meta, "XAxisLabel"); has {
		pp.XAxisLabel = lb
	}
	if lb, has := MetaMapLower(meta, "YAxisLabel"); has {
		pp.YAxisLabel = lb
	}
}

// ColParams are parameters for plotting one column of data
type ColParams struct { //gti:add

	// whether to plot this column
	On bool

	// name of column we're plotting
	Col string `label:"Column"`

	// whether to plot lines; uses the overall plot option if unset
	Lines option.Option[bool]

	// whether to plot points with symbols; uses the overall plot option if unset
	Points option.Option[bool]

	// the width of lines; uses the overall plot option if unset
	LineWidth option.Option[float64]

	// the size of points; uses the overall plot option if unset
	PointSize option.Option[float64]

	// the shape used to draw points; uses the overall plot option if unset
	PointShape option.Option[Shapes]

	// effective range of data to plot -- either end can be fixed
	Range minmax.Range64

	// full actual range of data -- only valid if specifically computed
	FullRange minmax.F64

	// color to use when plotting the line / column
	Color color.Color

	// desired number of ticks
	NTicks int

	// if non-empty, this is an alternative label to use in plotting
	Lbl string `label:"Label"`

	// if column has n-dimensional tensor cells in each row, this is the index within each cell to plot -- use -1 to plot *all* indexes as separate lines
	TensorIdx int

	// specifies a column containing error bars for this column
	ErrCol string

	// if true this is a string column -- plots as labels
	IsString bool `edit:"-"`

	// our plot, for update method
	Plot *Plot2D `copier:"-" json:"-" xml:"-" view:"-"`
}

// Defaults sets defaults if nil vals present
func (cp *ColParams) Defaults() {
	if cp.NTicks == 0 {
		cp.NTicks = 10
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (cp *ColParams) Update() {
	if cp.Plot != nil {
		cp.Plot.Update()
	}
}

// CopyFrom copies from other col params
func (cp *ColParams) CopyFrom(fr *ColParams) {
	pl := cp.Plot
	*cp = *fr
	cp.Plot = pl
}

func (cp *ColParams) Label() string {
	if cp.Lbl != "" {
		return cp.Lbl
	}
	return cp.Col
}

// FmMetaMap sets plot params from meta data map
func (cp *ColParams) FmMetaMap(meta map[string]string) {
	if op, has := MetaMapLower(meta, cp.Col+":On"); has {
		if op == "+" || op == "true" || op == "" {
			cp.On = true
		} else {
			cp.On = false
		}
	}
	if op, has := MetaMapLower(meta, cp.Col+":Off"); has {
		if op == "+" || op == "true" || op == "" {
			cp.On = false
		} else {
			cp.On = true
		}
	}
	if op, has := MetaMapLower(meta, cp.Col+":FixMin"); has {
		if op == "+" || op == "true" {
			cp.Range.FixMin = true
		} else {
			cp.Range.FixMin = false
		}
	}
	if op, has := MetaMapLower(meta, cp.Col+":FixMax"); has {
		if op == "+" || op == "true" {
			cp.Range.FixMax = true
		} else {
			cp.Range.FixMax = false
		}
	}
	if op, has := MetaMapLower(meta, cp.Col+":FloatMin"); has {
		if op == "+" || op == "true" {
			cp.Range.FixMin = false
		} else {
			cp.Range.FixMin = true
		}
	}
	if op, has := MetaMapLower(meta, cp.Col+":FloatMax"); has {
		if op == "+" || op == "true" {
			cp.Range.FixMax = false
		} else {
			cp.Range.FixMax = true
		}
	}
	if vl, has := MetaMapLower(meta, cp.Col+":Max"); has {
		cp.Range.Max, _ = laser.ToFloat(vl)
	}
	if vl, has := MetaMapLower(meta, cp.Col+":Min"); has {
		cp.Range.Min, _ = laser.ToFloat(vl)
	}
	if lb, has := MetaMapLower(meta, cp.Col+":Label"); has {
		cp.Lbl = lb
	}
	if lb, has := MetaMapLower(meta, cp.Col+":ErrCol"); has {
		cp.ErrCol = lb
	}
	if vl, has := MetaMapLower(meta, cp.Col+":TensorIdx"); has {
		iv, _ := laser.ToInt(vl)
		cp.TensorIdx = int(iv)
	}
}

// PlotTypes are different types of plots
type PlotTypes int32 //enums:enum

const (
	// XY is a standard line / point plot
	XY PlotTypes = iota

	// Bar plots vertical bars
	Bar
)
