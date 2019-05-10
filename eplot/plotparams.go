// Copyright (c) 2019, The Emergent Authors. All rights reserved.
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
	LineWidth  float64 `desc:"width of lines"`
	XAxisCol   string  `desc:"what column to use for the common x axis -- if empty or not found, the row number is used"`
	XAxisLabel string  `desc:"optional label to use for XAxis instead of column name"`
	YAxisLabel string  `desc:"optional label to use for YAxis -- if empty, first column name is used"`
}

// Defaults sets defaults if nil vals present
func (pp *PlotParams) Defaults() {
	if pp.LineWidth == 0 {
		pp.LineWidth = 1
	}
}

// ColParams are parameters for plotting one column of data
type ColParams struct {
	On        bool       `desc:"plot this column"`
	Column    string     `desc:"name of column we're plotting"`
	Range     Range      `desc:"effective range of data to plot -- either end can be fixed"`
	FullRange minmax.F64 `desc:"full actual range of data"`
	ColorName string     `desc:"if non-empty, color is set by this name"`
	Color     gi.Color   `desc:"color to use in plotting the line"`
	NTicks    int        `desc:"desired number of ticks"`
	Lbl       string     `desc:"if non-empty, this is an alternative label to use in plotting"`
}

// Defaults sets defaults if nil vals present
func (cp *ColParams) Defaults() {
	if cp.NTicks == 0 {
		cp.NTicks = 10
	}
}

// Update updates e.g., color from color name
func (cp *ColParams) Update() {
	if cp.ColorName != "" {
		clr, err := gi.ColorFromString(cp.ColorName, nil)
		if err == nil {
			cp.Color = clr
		}
	}
}

func (cp *ColParams) Label() string {
	if cp.Lbl != "" {
		return cp.Lbl
	}
	return cp.Column
}
