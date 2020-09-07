// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is copied and modified directly from gonum to add better error-bar
// plotting for bar plots, along with multiple groups.

// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// A ErrBarChart presents ordinally-organized data with rectangular bars
// with lengths proportional to the data values, and an optional
// error bar ("handle") at the top of the bar using given error value
// (single value, like a standard deviation etc, not drawn below the bar).
//
// Bars are plotted centered at integer multiples of Stride plus Start offset.
// Full data range also includes Pad value to extend range beyond edge bar centers.
// Bar Width is in data units, e.g., should be <= Stride.
// Defaults provide a unit-spaced plot.
type ErrBarChart struct {
	// Values are the plotted values
	Values plotter.Values

	// YErrors is a copy of the Y errors for each point.
	Errors plotter.Values

	// Start is starting offset -- first bar is centered at this point.
	// Defaults to 1.
	Start float64

	// Stride is distance between bars. Defaults to 1.
	Stride float64

	// Width is the width of the bars in source data units.  Defaults to .8
	Width float64

	// Pad is additional space at start / end of data range, to keep bars from
	// overflowing ends.  This amount is subtracted from Start
	// and added to (len(Values)-1)*Stride -- no other accommodation for bar
	// width is provided, so that should be built into this value as well.
	Pad float64

	// Color is the fill color of the bars.
	Color color.Color

	// LineStyle is the style of the outline of the bars.
	draw.LineStyle

	// Offset is added to the X location of each bar.
	// When the Offset is zero, the bars are drawn
	// centered at their X location.
	Offset vg.Length

	// Horizontal dictates whether the bars should be in the vertical
	// (default) or horizontal direction. If Horizontal is true, all
	// X locations and distances referred to here will actually be Y
	// locations and distances.
	Horizontal bool

	// stackedOn is the bar chart upon which
	// this bar chart is stacked.
	stackedOn *ErrBarChart
}

// NewErrBarChart returns a new bar chart with a single bar for each value.
// The bars heights correspond to the values and their x locations correspond
// to the index of their value in the Valuer.  Optional error-bar values can be
// provided.
func NewErrBarChart(vs, ers plotter.Valuer) (*ErrBarChart, error) {
	values, err := plotter.CopyValues(vs)
	if err != nil {
		return nil, err
	}
	var errs plotter.Values
	if ers != nil {
		errs, err = plotter.CopyValues(ers)
		if err != nil {
			return nil, err
		}
	}
	b := &ErrBarChart{
		Values: values,
		Errors: errs,
	}
	b.Defaults()
	return b, nil
}

func (b *ErrBarChart) Defaults() {
	b.Start = 1
	b.Stride = 1
	b.Width = .8
	b.Pad = 1
	b.Color = color.Black
	b.LineStyle = plotter.DefaultLineStyle
}

// BarHeight returns the maximum y value of the
// ith bar, taking into account any bars upon
// which it is stacked.
func (b *ErrBarChart) BarHeight(i int) float64 {
	ht := 0.0
	if b == nil {
		return 0
	}
	if i >= 0 && i < len(b.Values) {
		ht += b.Values[i]
	}
	if b.stackedOn != nil {
		ht += b.stackedOn.BarHeight(i)
	}
	return ht
}

// StackOn stacks a bar chart on top of another,
// and sets the bar positioning params to that of the
// chart upon which it is being stacked.
func (b *ErrBarChart) StackOn(on *ErrBarChart) {
	b.Start = on.Start
	b.Stride = on.Stride
	b.Pad = on.Pad
	b.stackedOn = on
}

// Plot implements the plot.Plotter interface.
func (b *ErrBarChart) Plot(c draw.Canvas, plt *plot.Plot) {
	trCat, trVal := plt.Transforms(&c)
	if b.Horizontal {
		trCat, trVal = trVal, trCat
	}

	for i, ht := range b.Values {
		cat := b.Start + float64(i)*b.Stride
		catVal := trCat(cat)
		if !b.Horizontal {
			if !c.ContainsX(catVal) {
				continue
			}
		} else {
			if !c.ContainsY(catVal) {
				continue
			}
		}
		catMin := trCat(cat - b.Width/2)
		catMax := trCat(cat + b.Width/2)
		bottom := b.stackedOn.BarHeight(i) // nil safe
		valMin := trVal(bottom)
		valMax := trVal(bottom + ht)

		var pts []vg.Point
		var poly []vg.Point
		if !b.Horizontal {
			pts = []vg.Point{
				{catMin, valMin},
				{catMin, valMax},
				{catMax, valMax},
				{catMax, valMin},
			}
			poly = c.ClipPolygonY(pts)
		} else {
			pts = []vg.Point{
				{valMin, catMin},
				{valMin, catMax},
				{valMax, catMax},
				{valMax, catMin},
			}
			poly = c.ClipPolygonX(pts)
		}
		c.FillPolygon(b.Color, poly)

		var outline [][]vg.Point
		if !b.Horizontal {
			pts = append(pts, vg.Point{X: catMin, Y: valMin})
			outline = c.ClipLinesY(pts)
		} else {
			pts = append(pts, vg.Point{X: valMin, Y: catMin})
			outline = c.ClipLinesX(pts)
		}
		c.StrokeLines(b.LineStyle, outline...)

		if i < len(b.Errors) {
			errval := b.Errors[i]
			eVal := trVal(bottom + ht + math.Abs(errval))
			if !b.Horizontal {
				bar := c.ClipLinesY([]vg.Point{{catVal, valMax}, {catVal, eVal}})
				c.StrokeLines(b.LineStyle, bar...)
				c.StrokeLine2(b.LineStyle, trCat(cat-b.Width/3), eVal, trCat(cat+b.Width/3), eVal)
			} else {
				bar := c.ClipLinesY([]vg.Point{{valMax, catVal}, {eVal, catVal}})
				c.StrokeLines(b.LineStyle, bar...)
				c.StrokeLine2(b.LineStyle, eVal, trCat(cat-b.Width/3), eVal, trCat(cat+b.Width/3))
			}
		}
	}
}

// DataRange implements the plot.DataRanger interface.
func (b *ErrBarChart) DataRange() (xmin, xmax, ymin, ymax float64) {
	catMin := b.Start - b.Pad
	catMax := b.Start + float64(len(b.Values)-1)*b.Stride + b.Pad

	valMin := math.Inf(1)
	valMax := math.Inf(-1)
	for i, val := range b.Values {
		valBot := b.stackedOn.BarHeight(i)
		valTop := valBot + val
		if i < len(b.Errors) {
			valTop += math.Abs(b.Errors[i])
		}
		valMin = math.Min(valMin, math.Min(valBot, valTop))
		valMax = math.Max(valMax, math.Max(valBot, valTop))
	}
	if !b.Horizontal {
		return catMin, catMax, valMin, valMax
	}
	return valMin, valMax, catMin, catMax
}

// GlyphBoxes implements the GlyphBoxer interface.
func (b *ErrBarChart) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	boxes := make([]plot.GlyphBox, len(b.Values))
	for i := range b.Values {
		cat := b.Start + float64(i)*b.Stride
		if !b.Horizontal {
			boxes[i].X = plt.X.Norm(cat)
			xr := plt.X.Max - plt.X.Min
			wd := vg.Length((b.Width / 2) / xr)
			boxes[i].Rectangle = vg.Rectangle{
				Min: vg.Point{X: -wd},
				Max: vg.Point{X: wd},
			}
		} else {
			boxes[i].Y = plt.Y.Norm(cat)
			xr := plt.Y.Max - plt.Y.Min
			wd := vg.Length((b.Width / 2) / xr)
			boxes[i].Rectangle = vg.Rectangle{
				Min: vg.Point{Y: -wd},
				Max: vg.Point{Y: wd},
			}
		}
	}
	return boxes
}

// Thumbnail fulfills the plot.Thumbnailer interface.
func (b *ErrBarChart) Thumbnail(c *draw.Canvas) {
	pts := []vg.Point{
		{c.Min.X, c.Min.Y},
		{c.Min.X, c.Max.Y},
		{c.Max.X, c.Max.Y},
		{c.Max.X, c.Min.Y},
	}
	poly := c.ClipPolygonY(pts)
	c.FillPolygon(b.Color, poly)

	pts = append(pts, vg.Point{X: c.Min.X, Y: c.Min.Y})
	outline := c.ClipLinesY(pts)
	c.StrokeLines(b.LineStyle, outline...)
}
