// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/kit"
)

// PlotParams are parameters for overall plot
type PlotParams struct {
	Title      string    `desc:"optional title at top of plot"`
	Type       PlotTypes `desc:"type of plot to generate.  For a Bar plot, items are plotted ordinally by row and the XAxis is optional"`
	Lines      bool      `desc:"plot lines"`
	Points     bool      `desc:"plot points with symbols"`
	LineWidth  float64   `desc:"width of lines"`
	PointSize  float64   `desc:"size of points"`
	BarWidth   float64   `min:"0.01" max:"1" desc:"width of bars for bar plot, as fraction of available space -- 1 = no gaps, .8 default"`
	NegXDraw   bool      `desc:"draw lines that connect points with a negative X-axis direction -- otherwise these are treated as breaks between repeated series and not drawn"`
	Scale      float64   `def:"2" desc:"overall scaling factor -- the larger the number, the larger the fonts are relative to the graph"`
	XAxisCol   string    `desc:"what column to use for the common X axis -- if empty or not found, the row number is used.  This optional for Bar plots -- if present and LegendCol is also present, then an extra space will be put between X values."`
	XAxisLabel string    `desc:"optional label to use for XAxis instead of column name"`
	YAxisLabel string    `desc:"optional label to use for YAxis -- if empty, first column name is used"`
	XAxisRot   float64   `desc:"rotation of the X Axis labels, in degrees"`
	LegendCol  string    `desc:"optional column for adding a separate colored / styled line or bar according to this value -- acts just like a separate Y variable, crossed with Y variables"`
	Plot       *Plot2D   `copy:"-" json:"-" xml:"-" view:"-" desc:"our plot, for update method"`
}

// Defaults sets defaults if nil vals present
func (pp *PlotParams) Defaults() {
	if pp.LineWidth == 0 {
		pp.LineWidth = 1
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
	if op, has := dt.MetaData["lines"]; has {
		if op == "+" || op == "true" {
			pp.Lines = true
		} else {
			pp.Lines = false
		}
	}
	if op, has := dt.MetaData["points"]; has {
		if op == "+" || op == "true" {
			pp.Points = true
		} else {
			pp.Points = false
		}
	}
}

// ColParams are parameters for plotting one column of data
type ColParams struct {
	On        bool           `desc:"plot this column"`
	Col       string         `desc:"name of column we're plotting"`
	Range     minmax.Range64 `desc:"effective range of data to plot -- either end can be fixed"`
	FullRange minmax.F64     `desc:"full actual range of data -- only valid if specifically computed"`
	ColorName gi.ColorName   `desc:"if non-empty, color is set by this name"`
	Color     gi.Color       `desc:"color to use in plotting the line"`
	NTicks    int            `desc:"desired number of ticks"`
	Lbl       string         `desc:"if non-empty, this is an alternative label to use in plotting"`
	TensorIdx int            `desc:"if column has n-dimensional tensor cells in each row, this is the index within each cell to plot -- use -1 to plot *all* indexes as separate lines"`
	ErrCol    string         `desc:"specifies a column containing error bars for this column"`
	IsString  bool           `inactive:"+" desc:"if true this is a string column -- plots as labels"`
	Plot      *Plot2D        `copy:"-" json:"-" xml:"-" view:"-" desc:"our plot, for update method"`
}

// Defaults sets defaults if nil vals present
func (cp *ColParams) Defaults() {
	if cp.NTicks == 0 {
		cp.NTicks = 10
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (cp *ColParams) Update() {
	cp.UpdateVals()
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

// UpdateVals update derived values e.g., color from color name
func (cp *ColParams) UpdateVals() {
	if cp.ColorName != "" {
		clr, err := gi.ColorFromString(string(cp.ColorName), nil)
		if err == nil {
			cp.Color = clr
		}
	}
}

func (cp *ColParams) Label() string {
	if cp.Lbl != "" {
		return cp.Lbl
	}
	return cp.Col
}

// PlotTypes are different types of plots
type PlotTypes int32

//go:generate stringer -type=PlotTypes

var KiT_PlotTypes = kit.Enums.AddEnum(PlotTypesN, kit.NotBitFlag, nil)

func (ev PlotTypes) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *PlotTypes) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	// XY is a standard line / point plot
	XY PlotTypes = iota

	// Bar plots vertical bars
	Bar

	PlotTypesN
)
