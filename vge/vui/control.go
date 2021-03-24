package vui

import (
	"fmt"
	"image"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
)

type Control interface {
	Measure(owner Owner, freeWidth int) (optimalSize image.Point)
	Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position)
	Event(owner Owner, ev vapp.Event)
}

var lastId int64

func MakeID() string {
	return fmt.Sprintf("ID_%d", atomic.AddInt64(&lastId, 1))
}

type Panel struct {
	Class     string
	Content   Control
	Style     Style
	mouseArea image.Rectangle
}

// Create a panel with padding and add content inside it
func NewPanel(padding int, content Control) *Panel {
	p := &Panel{Content: &Padding{Content: content, Padding: image.Rect(padding, padding, padding, padding), Clip: true}}
	return p
}

func MeasurePaddedContent(owner Owner, freeWidth int, content Control, st Style) (size image.Point) {
	if content != nil {
		size = content.Measure(owner, freeWidth)
	}
	if st != nil {
		cp := st.ContentPadding()
		size = size.Add(image.Pt(cp.Min.X+cp.Max.X, cp.Min.Y+cp.Max.Y))
	}
	return size
}

func RenderPaddedContent(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position, content Control, style Style) {
	if content != nil {
		if style != nil {
			pad := style.ContentPadding()
			pos.GlyphArea.Min = pos.GlyphArea.Min.Add(pad.Min)
			pos.GlyphArea.Max = pos.GlyphArea.Max.Sub(pad.Max)
		}
		content.Render(owner, dc, pos)
	}
}

func (p *Panel) Measure(owner Owner, freeWidth int) image.Point {
	if p.Style == nil {
		p.Style = owner.Theme().GetStyle(p, p.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, p.Content, p.Style)
}

func (p *Panel) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	p.mouseArea = pos.MouseArea()
	if p.Style == nil {
		p.Style = owner.Theme().GetStyle(p, p.Class)
	}
	if p.Style != nil {
		p.Style.Draw(owner, p, dc, pos, 0)
	}
	RenderPaddedContent(owner, dc, pos, p.Content, p.Style)
}

func (p *Panel) Event(owner Owner, ev vapp.Event) {
	if p.Content != nil && !ev.Handled() {
		p.Content.Event(owner, ev)
	}
	hv, ok := ev.(*MouseHoverEvent)
	if ok && hv.At.In(p.mouseArea) {
		hv.IsHandled = true
	}

	mc, ok := ev.(*MouseClickEvent)
	if ok && mc.At.In(p.mouseArea) {
		mc.IsHandled = true
	}
}

func (p *Panel) SetClass(class string) *Panel {
	p.Class = class
	return p
}

type Padding struct {
	Padding image.Rectangle
	Clip    bool
	Content Control
}

func (p *Padding) Measure(owner Owner, freeWidth int) image.Point {
	var chSize image.Point
	if p.Content != nil {
		chSize = p.Content.Measure(owner, freeWidth)
	}
	return chSize.Add(image.Pt(p.Padding.Min.X+p.Padding.Max.X, p.Padding.Min.Y+p.Padding.Max.Y))
}

func (p *Padding) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if p.Content != nil {
		pos.GlyphArea.Min = pos.GlyphArea.Min.Add(p.Padding.Min)
		pos.GlyphArea.Max = pos.GlyphArea.Max.Sub(p.Padding.Max)
		if p.Clip {
			pos = pos.AddClip(pos.GlyphArea)
		}
		p.Content.Render(owner, dc, pos)
	}
}

func (p *Padding) Event(owner Owner, ev vapp.Event) {
	if p.Content != nil && !ev.Handled() {
		p.Content.Event(owner, ev)
	}
}

type Extend struct {
	MinSize image.Point
	Align   mgl32.Vec2
	Content Control
}

func max(v1, v2 int) int {
	if v1 > v2 {
		return v1
	}
	return v2
}

func (e *Extend) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	var chSize image.Point
	if e.Content != nil {
		chSize = e.Content.Measure(owner, freeWidth)
	}
	return image.Pt(max(chSize.X, e.MinSize.X), max(chSize.Y, e.MinSize.Y))
}

func (e *Extend) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if e.Content != nil {
		s := pos.GlyphArea.Size()
		chSize := e.Content.Measure(owner, s.X)
		if chSize.X < s.X {
			offsetX := int(float32(s.X-chSize.X) * e.Align.X())
			pos.GlyphArea.Min.X += offsetX
			pos.GlyphArea.Max.X -= s.X - chSize.X - offsetX
		}
		if chSize.Y < s.Y {
			offsetY := int(float32(s.Y-chSize.Y) * e.Align.Y())
			pos.GlyphArea.Min.Y += offsetY
			pos.GlyphArea.Max.Y -= s.Y - chSize.Y - offsetY
		}
		e.Content.Render(owner, dc, pos)
	}
}

func (e *Extend) Event(owner Owner, ev vapp.Event) {
	if e.Content != nil && !ev.Handled() {
		e.Content.Event(owner, ev)
	}
}

func NewSizer(content Control, min image.Point, max image.Point) *Sizer {
	s := &Sizer{Content: content, Min: min, Max: max}
	if max.X < min.X && max.Y < min.Y {
		s.Max = s.Min
	}
	return s
}

type Sizer struct {
	Min     image.Point
	Max     image.Point
	Content Control
}

func (s *Sizer) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if s.Min.X == s.Max.X && s.Min.Y == s.Max.Y {
		return s.Min
	}
	chSize := s.Min
	if s.Content != nil {
		if freeWidth >= 0 {
			if freeWidth < s.Min.X {
				freeWidth = s.Min.X
			}
			if freeWidth > s.Max.X {
				freeWidth = s.Max.X
			}
		}
		chSize = s.Content.Measure(owner, freeWidth)
	}
	if chSize.X > s.Max.X {
		chSize.X = s.Max.X
	}
	if chSize.Y > s.Max.Y {
		chSize.Y = s.Max.Y
	}
	if chSize.X < s.Min.X {
		chSize.X = s.Min.X
	}
	if chSize.Y < s.Min.Y {
		chSize.Y = s.Min.Y
	}
	return chSize
}

func (s *Sizer) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	// TODO: Force
	if s.Content != nil {
		s.Content.Render(owner, dc, pos)
	}
}

func (s *Sizer) Event(owner Owner, ev vapp.Event) {
	if s.Content != nil {
		s.Content.Event(owner, ev)
	}
}
