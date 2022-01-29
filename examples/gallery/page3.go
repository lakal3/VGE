package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"image"
)

var gsPage3 *vdraw.GlyphSet
var gRect, gRr, gCircle *vdraw.Glyph
var p3shapes int
var brGreenBlue = vdraw.Brush{Color: mgl32.Vec4{0, 1, 0, 1}, Color2: mgl32.Vec4{0, 0, 1, 1}, UVTransform: mgl32.Scale2D(0.01, 0.01)}
var uvScale float32 = 100

func page3(fr *vimgui.UIFrame) {
	if gsPage3 == nil {
		buildGlyphs()
	}
	fr.NewLine(-33, 26, 0)
	vimgui.TabButton(fr, "p3tab1", "Path shapes", 0, &p3shapes)
	fr.NewColumn(-33, 10)
	vimgui.TabButton(fr, "p3tab2", "Glyph shapes", 1, &p3shapes)
	fr.NewColumn(-33, 10)
	vimgui.TabButton(fr, "p3tab2", "Shader shapes", 2, &p3shapes)
	fr.NewLine(-100, 3, 3)
	vimgui.Border(fr)
	fr.NewLine(100, 20, 2)
	vimgui.Label(fr, "UV scale")
	fr.NewColumn(-50, 5)
	if vimgui.HorizontalSlider(fr, "uvscale", 10, 300, 1, &uvScale) {
		brGreenBlue.UVTransform = mgl32.Scale2D(1/uvScale, 1/uvScale)
	}
	fr.NewColumn(40, 5)
	vimgui.Label(fr, fmt.Sprintf("%.1f", uvScale))
	fr.NewLine(-100, 3, 3)
	// vimgui.Border(fr)

	switch p3shapes {
	case 0:
		page3path(fr)
	case 1:
		page3glyph(fr)
	case 2:
		page3shader(fr)
	}
}

func page3path(fr *vimgui.UIFrame) {
	var p1 vdraw.Path
	fr.NewLine(-100, 120, 0)
	p1.AddRect(true, mgl32.Vec2{10, 10}, mgl32.Vec2{120, 60})
	f := p1.Fill()
	fr.Canvas().Draw(f, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(f, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
	fr.NewLine(-100, 120, 0)
	p1.Clear()
	p1.AddRoundedRect(true, mgl32.Vec2{10, 10}, mgl32.Vec2{120, 60}, vdraw.UniformCorners(15))
	f = p1.Fill()
	fr.Canvas().Draw(f, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(f, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
	fr.NewLine(-100, 130, 10)
	p1.Clear()
	p1.AddRoundedRect(true, mgl32.Vec2{0, 0}, mgl32.Vec2{100, 100}, vdraw.UniformCorners(49))
	f = p1.Fill()
	fr.Canvas().Draw(f, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(f, fr.ControlArea.From.Add(mgl32.Vec2{200, 10}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
}

func page3glyph(fr *vimgui.UIFrame) {
	fr.NewLine(-100, 120, 0)
	fr.Canvas().Draw(gRect, fr.ControlArea.From, mgl32.Vec2{1, 0.5}, &brGreenBlue)
	fr.Canvas().Draw(gRect, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 0.75}, &brGreenBlue)
	fr.NewLine(-100, 120, 0)
	fr.Canvas().Draw(gRr, fr.ControlArea.From, mgl32.Vec2{1, 0.5}, &brGreenBlue)
	fr.Canvas().Draw(gRr, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 0.75}, &brGreenBlue)
	fr.NewLine(-100, 130, 10)
	fr.Canvas().Draw(gCircle, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(gCircle, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
}

func page3shader(fr *vimgui.UIFrame) {
	fr.NewLine(-100, 120, 0)
	sRect := vdraw.Rect{Area: vdraw.Area{From: mgl32.Vec2{10, 10}}}
	sRect.Area.To = sRect.Area.From.Add(mgl32.Vec2{120, 60})
	fr.Canvas().Draw(sRect, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(sRect, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
	sRRect := vdraw.RoudedRect{Area: sRect.Area, Corners: vdraw.Corners{TopLeft: 5, TopRight: 10, BottomLeft: 20}}
	fr.NewLine(-100, 120, 0)
	fr.Canvas().Draw(sRRect, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(sRRect, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
	fr.NewLine(-100, 130, 0)
	sRRect2 := vdraw.RoudedRect{Area: vdraw.Area{To: mgl32.Vec2{100, 100}}, Corners: vdraw.UniformCorners(50)}
	fr.Canvas().Draw(sRRect2, fr.ControlArea.From, mgl32.Vec2{1, 1}, &brGreenBlue)
	fr.Canvas().Draw(sRRect2, fr.ControlArea.From.Add(mgl32.Vec2{200, 0}), mgl32.Vec2{1.5, 1.5}, &brGreenBlue)
}

func buildGlyphs() {
	gsPage3 = &vdraw.GlyphSet{}
	var p vdraw.Path
	p.AddRect(true, mgl32.Vec2{5, 5}, mgl32.Vec2{115, 115})
	gRect = gsPage3.AddPath(&p)
	var p2 vdraw.Path
	p2.AddRoundedRect(true, mgl32.Vec2{5, 5}, mgl32.Vec2{115, 115}, vdraw.Corners{TopLeft: 3, TopRight: 6, BottomLeft: 10})
	gRr = gsPage3.AddPath(&p2)
	gCircle = gsPage3.AddComputed(func(size image.Point, at image.Point) (depth float32) {
		c := size.Div(2)
		return mgl32.Vec2{float32(c.X), float32(c.Y)}.Sub(mgl32.Vec2{float32(at.X), float32(at.Y)}).Len() - 49
	})
	gsPage3.Build(vapp.Dev, image.Pt(120, 120))
}
