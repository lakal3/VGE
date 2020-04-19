package vui

import (
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
)

type VStack struct {
	Children []Control
	Padding  int
}

func NewVStack(padding int, ctrls ...Control) *VStack {
	return &VStack{Padding: padding, Children: ctrls}
}

func (v *VStack) Measure(owner Owner, freeWidth int) image.Point {
	y := 0
	x := 0
	for idx, ch := range v.Children {
		chSize := ch.Measure(owner, freeWidth)
		y += chSize.Y
		if idx > 0 {
			y += v.Padding
		}
		if chSize.X > x {
			x = chSize.X
		}
	}
	return image.Pt(x, y)
}

func (v *VStack) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	fw := pos.GlyphArea.Size().X
	y := pos.GlyphArea.Min.Y
	maxy := pos.GlyphArea.Max.Y
	for _, ch := range v.Children {
		chSize := ch.Measure(owner, fw)
		if y > maxy {
			return
		}
		pos.GlyphArea.Min.Y = y
		pos.GlyphArea.Max.Y = y + chSize.Y
		if pos.GlyphArea.Max.Y > maxy {
			pos.GlyphArea.Max.Y = maxy
		}
		ch.Render(owner, dc, pos)
		y += chSize.Y + v.Padding
	}
}

func (v *VStack) Event(owner Owner, ev vapp.Event) {
	for _, ch := range v.Children {
		if ev.Handled() {
			return
		}
		ch.Event(owner, ev)
	}
}

type HStack struct {
	Children []Control
	Padding  int
}

func NewHStack(padding int, ctrls ...Control) *HStack {
	return &HStack{Padding: padding, Children: ctrls}
}

func (h *HStack) Measure(owner Owner, freeWidth int) image.Point {
	y := 0
	x := 0
	for idx, ch := range h.Children {
		chSize := ch.Measure(owner, freeWidth)
		x += chSize.X
		if freeWidth > 0 {
			freeWidth = max(0, freeWidth-chSize.X)
		}
		if idx > 0 {
			x += h.Padding
			if freeWidth > 0 {
				freeWidth = max(0, freeWidth-h.Padding)
			}
		}
		if chSize.Y > y {
			y = chSize.Y
		}
	}
	return image.Pt(x, y)
}

func (h *HStack) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	fw := pos.GlyphArea.Size().X
	x := pos.GlyphArea.Min.X
	for _, ch := range h.Children {
		chSize := ch.Measure(owner, fw)
		pos.GlyphArea.Min.X = x
		pos.GlyphArea.Max.X = x + chSize.X
		ch.Render(owner, dc, pos)
		x += chSize.X + h.Padding
	}
}

func (h *HStack) Event(owner Owner, ev vapp.Event) {
	for _, ch := range h.Children {
		if ev.Handled() {
			return
		}
		ch.Event(owner, ev)
	}
}

type Conditional struct {
	Visible bool
	Content Control
}

func NewConditional(visible bool, content Control) *Conditional {
	return &Conditional{Visible: visible, Content: content}
}

func (c *Conditional) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if !c.Visible || c.Content == nil {
		return image.Pt(0, 0)
	}
	return c.Content.Measure(owner, freeWidth)
}

func (c *Conditional) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if !c.Visible || c.Content == nil {
		return
	}
	c.Content.Render(owner, dc, pos)
}

func (c *Conditional) Event(owner Owner, ev vapp.Event) {
	if !c.Visible || c.Content == nil {
		return
	}
	c.Content.Event(owner, ev)
}

type ScrollViewer struct {
	Content Control
	Offset  image.Point
	vs      VSlider
	hs      HSlider
}

func NewScrollViewer(content Control) *ScrollViewer {
	sv := &ScrollViewer{Content: content}
	sv.vs.OnChanged = sv.vScroll
	sv.hs.OnChanged = sv.hScroll
	return sv
}

func (s *ScrollViewer) vScroll(newPos int) {
	s.Offset.Y = newPos
}

func (s *ScrollViewer) hScroll(newPos int) {
	s.Offset.X = newPos
}

func (s *ScrollViewer) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	var chMain image.Point
	hsSize := s.hs.Measure(owner, freeWidth)
	vsSize := s.vs.Measure(owner, freeWidth)
	if freeWidth > hsSize.X {
		freeWidth -= hsSize.X
	}
	if s.Content != nil {
		chMain = s.Content.Measure(owner, freeWidth)
	}
	chMain.X += hsSize.X
	chMain.Y += vsSize.Y
	return chMain
}

func (s *ScrollViewer) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	var chSize image.Point
	fw := pos.GlyphArea.Size().X
	hsSize := s.hs.Measure(owner, fw)
	fw -= hsSize.X
	vsSize := s.vs.Measure(owner, fw)
	if s.Content != nil {
		chSize = s.Content.Measure(owner, fw)
	}
	s.vs.Maximum, s.hs.Maximum = chSize.Y, chSize.X
	s.hs.Visible = pos.GlyphArea.Size().X - hsSize.X
	s.vs.Visible = pos.GlyphArea.Size().Y - vsSize.Y
	if s.Offset.Y+s.vs.Visible > s.vs.Maximum {
		s.Offset.Y = s.vs.Maximum - s.vs.Visible
	}
	if s.Offset.X+s.hs.Visible > s.hs.Maximum {
		s.Offset.X = s.hs.Maximum - s.hs.Visible
	}
	if s.Offset.Y < 0 {
		s.Offset.Y = 0
	}
	if s.Offset.X < 0 {
		s.Offset.X = 0
	}
	s.vs.Current, s.hs.Current = s.Offset.Y, s.Offset.X
	pos2 := pos
	pos2.GlyphArea.Min.X = pos.GlyphArea.Max.X - hsSize.X
	pos2.GlyphArea.Max.Y -= vsSize.Y
	s.vs.Render(owner, dc, pos2)
	pos2 = pos
	pos2.GlyphArea.Min.Y = pos.GlyphArea.Max.Y - vsSize.Y
	s.hs.Render(owner, dc, pos2)
	pos.GlyphArea.Max = pos.GlyphArea.Max.Sub(image.Pt(hsSize.X, vsSize.Y))
	pos = pos.AddClip(pos.GlyphArea)
	pos.GlyphArea.Min = pos.GlyphArea.Min.Sub(s.Offset)
	pos.GlyphArea.Max = pos.GlyphArea.Min.Add(chSize)
	if s.Content != nil {
		s.Content.Render(owner, dc, pos)
	}

}

func (s *ScrollViewer) Event(owner Owner, ev vapp.Event) {
	s.vs.Event(owner, ev)
	s.hs.Event(owner, ev)
	if s.Content != nil && !ev.Handled() {
		s.Content.Event(owner, ev)
	}
}
