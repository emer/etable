// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/gi"
)

// PlotParams are parameters for overall plot
type PlotParams struct {
	Title      string  `desc:"optional title at top of plot"`
	Lines      bool    `desc:"plot lines"`
	Points     bool    `desc:"plot points with symbols"`
	LineWidth  float64 `desc:"width of lines"`
	PointSize  float64 `desc:"size of points"`
	NegXDraw   bool    `desc:"draw lines that connect points with a negative X-axis direction -- otherwise these are treated as breaks between repeated series and not drawn"`
	Scale      float64 `def:"2" desc:"overall scaling factor -- the larger the number, the larger the fonts are relative to the graph"`
	XAxisCol   string  `desc:"what column to use for the common x axis -- if empty or not found, the row number is used"`
	XAxisLabel string  `desc:"optional label to use for XAxis instead of column name"`
	YAxisLabel string  `desc:"optional label to use for YAxis -- if empty, first column name is used"`
	Plot       *Plot2D `copy:"-" json:"-" xml:"-" view:"-" desc:"our plot, for update method"`
}

// Defaults sets defaults if nil vals present
func (pp *PlotParams) Defaults() {
	if pp.LineWidth == 0 {
		pp.LineWidth = 1
		pp.Lines = true
		pp.Points = false
		pp.PointSize = 3
	}
	if pp.Scale == 0 {
		pp.Scale = 2
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (pp *PlotParams) Update() {
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

// ColParams are parameters for plotting one column of data
type ColParams struct {
	On        bool           `desc:"plot this column"`
	Col       string         `desc:"name of column we're plotting"`
	IsString  bool           `desc:"if true this is a string column -- plots as labels"`
	Range     minmax.Range64 `desc:"effective range of data to plot -- either end can be fixed"`
	FullRange minmax.F64     `desc:"full actual range of data -- only valid if specifically computed"`
	ColorName gi.ColorName   `desc:"if non-empty, color is set by this name"`
	Color     gi.Color       `desc:"color to use in plotting the line"`
	NTicks    int            `desc:"desired number of ticks"`
	Lbl       string         `desc:"if non-empty, this is an alternative label to use in plotting"`
	TensorIdx int            `desc:"if column has n-dimensional tensor cells in each row, this is the index within each cell to plot"`
	ErrCol    string         `desc:"specifies a column containing error bars for this column"`
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
