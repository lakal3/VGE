package vui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
)

type CanvasItem struct {
	Content        Control
	Rect           image.Rectangle
	PositionFactor mgl32.Vec2
	SizeFactor     mgl32.Vec2
}

type Canvas struct {
	DesignSize image.Point
	Items      []CanvasItem
}

func (c *Canvas) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if freeWidth > 0 {
		return image.Point{X: freeWidth, Y: c.DesignSize.Y}
	}
	return c.DesignSize
}

func (c *Canvas) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	delta := pos.GlyphArea.Size().Sub(c.DesignSize)
	for _, item := range c.Items {
		gaMin := pos.GlyphArea.Min.Add(item.Rect.Min.Add(c.factorize(delta, item.PositionFactor)))
		gaMax := gaMin.Add(item.Rect.Size().Add(c.factorize(delta, item.SizeFactor)))
		pChild := pos
		pChild.GlyphArea = image.Rectangle{Min: gaMin, Max: gaMax}
		item.Content.Render(owner, dc, pChild)
	}
}

func (c *Canvas) Event(owner Owner, ev vapp.Event) {
	for _, item := range c.Items {
		if ev.Handled() {
			return
		}
		item.Content.Event(owner, ev)
	}
}

func NewCanvas(designSize image.Point) *Canvas {
	return &Canvas{DesignSize: designSize}
}

func (c *Canvas) AddItem(ctrl Control, rect image.Rectangle, posFactor mgl32.Vec2, sizeFactor mgl32.Vec2) *Canvas {
	c.Items = append(c.Items, CanvasItem{Content: ctrl, Rect: rect, PositionFactor: posFactor, SizeFactor: sizeFactor})
	return c
}

func (c *Canvas) AddTopLeft(ctrl Control, rect image.Rectangle) *Canvas {
	return c.AddItem(ctrl, rect, mgl32.Vec2{}, mgl32.Vec2{})
}

func (c *Canvas) AddTopRight(ctrl Control, rect image.Rectangle) *Canvas {
	return c.AddItem(ctrl, rect, mgl32.Vec2{0, 1}, mgl32.Vec2{})
}

func (c *Canvas) AddBottomLeft(ctrl Control, rect image.Rectangle) *Canvas {
	return c.AddItem(ctrl, rect, mgl32.Vec2{1, 0}, mgl32.Vec2{})
}

func (c *Canvas) AddBottomRight(ctrl Control, rect image.Rectangle) *Canvas {
	return c.AddItem(ctrl, rect, mgl32.Vec2{1, 1}, mgl32.Vec2{})
}

func (c *Canvas) AddCenter(ctrl Control, rect image.Rectangle) *Canvas {
	return c.AddItem(ctrl, rect, mgl32.Vec2{}, mgl32.Vec2{1, 1})
}

func (c *Canvas) factorize(point image.Point, factor mgl32.Vec2) image.Point {
	return image.Pt(int(float32(point.X)*factor[0]+0.5), int(float32(point.Y)*factor[1]+0.5))
}
