// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"bytes"
	"log"
	"os"

	"github.com/goki/gi/svg"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
)

// PlotViewSVG shows the given gonum Plot in given GoGi svg editor widget.
// The scale rescales the default font sizes -- 2-4 recommended.
func PlotViewSVG(plt *plot.Plot, svge *svg.Editor, scale float64) {
	updt := svge.UpdateStart()
	defer svge.UpdateEnd(updt)
	svge.SetFullReRender()

	sz := svge.BBox.Size()
	w := (float64(sz.X) * 72.0) / (scale * 96.0)
	h := (float64(sz.Y) * 72.0) / (scale * 96.0)

	// Create a Canvas for writing SVG images.
	c := vgsvg.New(vg.Length(w), vg.Length(h))

	// Draw to the Canvas.
	plt.Draw(draw.New(c))

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		log.Println(err)
		return
	}

	svge.ReadXML(&buf)

	svge.SetNormXForm()
	svge.Scale = float32(scale)
	svge.SetTransform()

	svge.FullInit2DTree() // critical to enable immediate rendering
}

// SaveSVGView saves the given gonum Plot exactly as it is rendered given GoGi svg editor widget.
// The scale rescales the default font sizes -- 2-4 recommended.
func SaveSVGView(fname string, plt *plot.Plot, svge *svg.Editor, scale float64) error {
	sz := svge.BBox.Size()
	w := (float64(sz.X) * 72.0) / (scale * 96.0)
	h := (float64(sz.Y) * 72.0) / (scale * 96.0)

	// Create a Canvas for writing SVG images.
	c := vgsvg.New(vg.Length(w), vg.Length(h))

	// Draw to the Canvas.
	plt.Draw(draw.New(c))

	f, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	if _, err = c.WriteTo(f); err != nil {
		log.Println(err)
	}
	return err
}

// StringViewSVG shows the given svg string in given GoGi svg editor widget
// Scale to fit your window -- e.g., 2-3 depending on sizes
func StringViewSVG(svgstr string, svge *svg.Editor, scale float64) {
	updt := svge.UpdateStart()
	defer svge.UpdateEnd(updt)
	svge.SetFullReRender()

	var buf bytes.Buffer
	buf.Write([]byte(svgstr))
	svge.ReadXML(&buf)

	svge.SetNormXForm()
	svge.Scale = float32(scale) * (svge.Viewport.Win.LogicalDPI() / 96.0)
	svge.SetTransform()
}
