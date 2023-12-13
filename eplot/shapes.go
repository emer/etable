// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import "gonum.org/v1/plot/vg/draw"

// Shapes are the different shapes that can be used for plot points.
type Shapes int32 //enums:enum

const (
	// Ring is the outline of a circle
	Ring Shapes = iota

	// Circle is a solid circle
	Circle

	// Square is the outline of a square
	Square

	// Box is a filled square
	Box

	// Triangle is the outline of a triangle
	Triangle

	// Pyramid is a filled triangle
	Pyramid

	// Plus is a plus sign
	Plus

	// Cross is a big X
	Cross
)

// ShapeGlyphs contains the [draw.GlyphDrawer] for each of the [Shapes].
var ShapeGlyphs = map[Shapes]draw.GlyphDrawer{
	Ring:     draw.RingGlyph{},
	Circle:   draw.CircleGlyph{},
	Square:   draw.SquareGlyph{},
	Box:      draw.BoxGlyph{},
	Triangle: draw.TriangleGlyph{},
	Pyramid:  draw.PyramidGlyph{},
	Plus:     draw.PlusGlyph{},
	Cross:    draw.CrossGlyph{},
}

// Glyph returns the [draw.GlyphDrawer] associated with this shape.
func (s Shapes) Glyph() draw.GlyphDrawer {
	return ShapeGlyphs[s]
}
