package vimgui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vdraw"
)

type Painter func(uf *UIFrame)

// VerticalDivider splits DrawArea into to separate areas. User can adjust position of left panel within range from min to max value
// pos is current size of left panel
// VerticalDivider uses DrawArea, not ControlArea to size itself.
func VerticalDivider(uf *UIFrame, minWidth, maxWidth float32, pos *float32, left, right Painter) bool {
	changed := false
	if *pos > maxWidth {
		changed = true
		*pos = maxWidth
	}
	if *pos < minWidth {
		changed = true
		*pos = minWidth
	}

	ss := uf.GetStyles("*divider", "vertical")
	sbs := ss.Get(ScrollBarSize{BarSize: 4, Padding: 3}).(ScrollBarSize)
	da := uf.DrawArea
	uf.ControlArea = vdraw.Area{From: mgl32.Vec2{*pos - sbs.BarSize/2, da.From[1]}, To: mgl32.Vec2{*pos + sbs.BarSize/2, da.To[1]}}
	DrawBorder(uf, uf.GetStyles("*divider", "vertical"))

	// TODO: Move logic

	uf.PushArea(vdraw.Area{From: da.From, To: mgl32.Vec2{*pos - sbs.Padding - sbs.BarSize/2, da.To[1]}}, true)
	uf.ControlArea.From, uf.ControlArea.To = uf.DrawArea.From, uf.DrawArea.From
	if left != nil {
		left(uf)
	}
	uf.Pop()
	uf.PushArea(vdraw.Area{From: mgl32.Vec2{*pos - sbs.Padding - sbs.BarSize/2, da.From[1]}, To: da.To}, true)
	uf.ControlArea.From, uf.ControlArea.To = uf.DrawArea.From, uf.DrawArea.From
	if right != nil {
		right(uf)
	}
	uf.Pop()
	return changed
}

// ScrollArea scrolls contents inside itself. Scroll area content is drawn with content method
// ScrollArea uses DrawArea, not ControlArea to size itself. Size is total size of content painted by content function
func ScrollArea(uf *UIFrame, size mgl32.Vec2, offset *mgl32.Vec2, content Painter) {
	ss := uf.GetStyles("*scrollarea")
	sbs := ss.Get(ScrollBarSize{BarSize: 8, Padding: 3}).(ScrollBarSize)
	da := uf.DrawArea
	ta := da
	ta.To = ta.To.Sub(mgl32.Vec2{sbs.BarSize + sbs.Padding, sbs.BarSize + sbs.Padding})

	if da.Contains(uf.MousePos) {
		if uf.Ev.Scroll.Y != 0 {
			offset[1] -= float32(uf.Ev.Scroll.Y) * ta.Height() / 4
		}
	}
	uf.ControlArea.From = mgl32.Vec2{da.To[0] - sbs.BarSize, da.From[1]}
	uf.ControlArea.To = mgl32.Vec2{da.To[0], da.To[1] - sbs.Padding - sbs.BarSize}
	v := *offset
	DrawSlider(uf, ss, false, 0, size[1], ta.Size()[1], &v[1])
	uf.ControlArea.From = mgl32.Vec2{da.From[0], da.To[1] - sbs.BarSize}
	uf.ControlArea.To = mgl32.Vec2{da.To[0] - sbs.BarSize - sbs.Padding, da.To[1]}
	DrawSlider(uf, ss, true, 0, size[0], ta.Size()[0], &v[0])
	uf.PushArea(ta, true)
	defer uf.Pop()
	*offset = v
	uf.Offset = v.Mul(-1)
	uf.ControlArea.From = ta.From.Add(uf.Offset)
	uf.ControlArea.To = uf.ControlArea.From
	if content != nil {
		content(uf)
	}
}

// Panel draws panel with given title Painter and content Painter.
// Panel adjust drawArea for title and content before calling painter
func Panel(uf *UIFrame, title Painter, content Painter) {
	ss := uf.GetStyles("*panel")
	ps := ss.Get(PanelStyle{}).(PanelStyle)
	br := ss.Get(BorderRadius{}).(BorderRadius)
	inside := DrawBorder(uf, ss)
	dr := uf.ControlArea
	sz := mgl32.Vec2{inside.Size()[0], ps.TitleHeight + ps.Edges.Top - inside.From[1] + dr.From[1]}
	uf.PushControlArea()
	defer uf.Pop()
	th := ps.TitleHeight
	if title == nil {
		th = 0
	}
	if th > 0 {
		var titleBg vdraw.Path
		if br.IsEmpty() {
			titleBg.AddRect(true, inside.From, sz)
		} else {
			c := br.Corners
			c.BottomLeft, c.BottomRight = 0, 0
			titleBg.AddRoundedRect(true, inside.From, sz, c)
		}
		uf.Canvas().Draw(titleBg.Fill(), mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &ps.TitleBg)
		drTitle := ps.Edges.Shrink(dr, 0)
		drTitle.To[1] = drTitle.From[1] + th
		uf.PushArea(drTitle, true)
		uf.PushTags()
		uf.ResetControlArea()
		if title != nil {
			title(uf)
		}
		uf.Pop()
		uf.Pop()
	}
	drContent := ps.Edges.Shrink(dr, 0)
	drContent.From[1] += th
	uf.PushArea(drContent, true)
	uf.PushTags()
	uf.ResetControlArea()
	if content != nil {
		content(uf)
	}
	uf.Pop()
	uf.Pop()
}
