// Copyright (c) 2019, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"bytes"
	"log"
	"log/slog"
	"os"

	"cogentcore.org/core/gi"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
)

// PlotViewSVG shows the given gonum Plot in given Cogent Core svg editor widget.
// The scale rescales the default font sizes -- 2-4 recommended.
// This call must generally be enclosed within an UpdateStart / End
// as part of the overall update routine using it.
// if called from a different goroutine, it is essential to
// surround with BlockUpdates on Viewport as this does full
// damage to the tree.
func PlotViewSVG(plt *plot.Plot, svge *gi.SVG, scale float64) {
	sz := svge.Geom.ContentBBox.Size()
	if sz.X < 10 || sz.Y < 10 || scale == 0 {
		return
	}

	w := float64(sz.X-4) / scale
	h := float64(sz.Y-4) / scale

	// Create a Canvas for writing SVG images.
	c := vgsvg.New(vg.Length(w), vg.Length(h))

	// Draw to the Canvas.
	plt.Draw(draw.New(c))

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		slog.Error(err.Error())
	} else {
		err := svge.SVG.ReadXML(&buf)
		if err != nil {
			slog.Error("eplot: svg render errors", "err", err)
		}
		svge.SVG.Fill = true
		svge.SetNeedsRender(true)
	}
}

// SaveSVGView saves the given gonum Plot exactly as it is rendered given Cogent Core svg editor widget.
// The scale rescales the default font sizes -- 2-4 recommended.
func SaveSVGView(fname string, plt *plot.Plot, svge *gi.SVG, scale float64) error {
	sz := svge.Geom.ContentBBox.Size()
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

// StringViewSVG shows the given svg string in given Cogent Core svg editor widget
// Scale to fit your window -- e.g., 2-3 depending on sizes
func StringViewSVG(svgstr string, svge *gi.SVG, scale float64) {
	updt := svge.UpdateStart()
	defer svge.UpdateEndRender(updt)

	var buf bytes.Buffer
	buf.Write([]byte(svgstr))
	svge.SVG.ReadXML(&buf)

	svge.SVG.Fill = true
}
