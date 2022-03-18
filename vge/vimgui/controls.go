package vimgui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vk"
)

func Label(uf *UIFrame, text string) {
	s := uf.GetStyles("*label")
	DrawLabel(uf, text, s)
}

func DrawLabel(uf *UIFrame, text string, ss StyleSet) {
	if uf.IsHidden() {
		return
	}
	ft := ss.Get(FontStyle{}).(FontStyle)
	if ft.Font == nil {
		return
	}
	fc := ss.Get(ForeColor{}).(ForeColor)
	uf.Canvas().DrawText(ft.Font, ft.Size, uf.ControlArea.From, &fc.Brush, text)
}

func Border(uf *UIFrame) {
	s := uf.GetStyles("*border")
	DrawBorder(uf, s)
}

func DrawBorder(uf *UIFrame, ss StyleSet) (inside vdraw.Area) {
	if uf.IsHidden() {
		return
	}
	area := uf.ControlArea
	inside = area
	bc := ss.Get(BorderColor{}).(BorderColor)
	if bc.Brush.IsNil() {
		return
	}
	bt := ss.Get(BorderThickness{}).(BorderThickness)
	br := ss.Get(BorderRadius{}).(BorderRadius)
	if !bt.IsEmpty() {
		inside = bt.Shrink(area, 0)
		if !inside.IsNil() {
			if br.IsEmpty() {
				uf.Canvas().Draw(vdraw.Border{Area: area, Edges: bt.Edges}, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &bc.Brush)
			} else {
				uf.Canvas().Draw(vdraw.RoundedBorder{Area: area, Edges: bt.Edges, Corners: br.Corners},
					mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &bc.Brush)
			}
		}
	} else {
		ut := ss.Get(UnderlineThickness{}).(UnderlineThickness)
		if ut.Thickness > 0 {
			ulCenter := area.To[1] - ut.Thickness*2/3
			uf.Canvas().Draw(vdraw.Line{Thickness: ut.Thickness, From: mgl32.Vec2{area.From[0], ulCenter}, To: mgl32.Vec2{area.To[0], ulCenter}},
				mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &bc.Brush)
		}
		inside.To[1] -= ut.Thickness
	}

	var p vdraw.Path
	bg := ss.Get(BackgroudColor{}).(BackgroudColor)
	if bg.Brush.IsNil() {
		return
	}
	p.Clear()
	if br.IsEmpty() {
		uf.Canvas().Draw(vdraw.Rect{Area: inside}, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &bg.Brush)
	} else {
		uf.Canvas().Draw(vdraw.RoudedRect{Area: area, Corners: br.Corners}, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &bg.Brush)
	}
	return
}

func Button(uf *UIFrame, id vk.Key, title string) bool {
	if uf.IsHidden() {
		return false
	}
	var styles = []string{"*button"}
	if uf.MouseHover() {
		styles = []string{":hover", "*button"}
	}
	s := uf.GetStyles(styles...)
	DrawBorder(uf, s)
	uf.PushControlArea()
	uf.ControlArea = vdraw.UniformEdge(5).Shrink(uf.ControlArea, 0)
	DrawLabel(uf, title, s)
	uf.Pop()
	return uf.MouseClick(1)
}

func ToggleButton(uf *UIFrame, id vk.Key, kind, title string, myValue int, value *int) (changed bool) {
	if uf.IsHidden() {
		return
	}
	var styles = []string{"*togglebutton", kind}
	if uf.MouseHover() {
		styles = append(styles, ":hover")
	}
	if myValue == *value {
		styles = append(styles, ":checked")
	}
	s := uf.GetStyles(styles...)
	DrawBorder(uf, s)
	pi := s.Get(PrefixIcons{}).(PrefixIcons)
	fc := s.Get(ForeColor{}).(ForeColor)
	uf.PushControlArea()
	if pi.Font != nil && len(pi.Icons) >= 2 && !fc.Brush.IsNil() {
		ir := pi.Icons[0]
		if *value == myValue {
			ir = pi.Icons[1]
		}
		uf.Canvas().DrawText(pi.Font, pi.Size, uf.ControlArea.From, &fc.Brush, string(ir))
		uf.ControlArea.From[0] += pi.Size + pi.Padding
	}
	if len(title) > 0 {
		DrawLabel(uf, title, s)
	}
	uf.Pop()
	if uf.MouseClick(1) {
		*value = myValue
		return true
	}
	return false
}

func RadioButton(uf *UIFrame, id vk.Key, title string, myValue int, value *int) (changed bool) {
	return ToggleButton(uf, id, "radiobutton", title, myValue, value)
}

func TabButton(uf *UIFrame, id vk.Key, title string, myValue int, value *int) (changed bool) {
	return ToggleButton(uf, id, "tab", title, myValue, value)
}

func CheckBox(uf *UIFrame, id vk.Key, title string, value *bool) (changed bool) {
	v := 0
	if *value {
		v = 1
	}
	if ToggleButton(uf, id, "checkbox", title, 1, &v) {
		*value = !*value
		return true
	}
	return false
}

func HorizontalSlider(uf *UIFrame, id vk.Key, min float32, max float32, visible float32, value *float32) bool {
	var styles = []string{"*slider", "horizontal"}
	if uf.MouseHover() {
		styles = []string{":hover", "*slider", "horizontal"}
	}
	return DrawSlider(uf, uf.GetStyles(styles...), true, min, max, visible, value)
}

func VerticalSlider(uf *UIFrame, id vk.Key, min float32, max float32, visible float32, value *float32) bool {
	var styles = []string{"*slider", "vertical"}
	if uf.MouseHover() {
		styles = []string{":hover", "*slider", "vertical"}
	}
	return DrawSlider(uf, uf.GetStyles(styles...), false, min, max, visible, value)
}

func DrawSlider(uf *UIFrame, ss StyleSet, horizontal bool, min float32, max float32, visible float32, value *float32) (changed bool) {
	if uf.IsHidden() {
		return
	}
	w := max - min
	if w <= 0 {
		return false
	}
	bc := ss.Get(BorderColor{}).(BorderColor)
	bg := ss.Get(BackgroudColor{}).(BackgroudColor)
	fc := ss.Get(ForeColor{}).(ForeColor)
	ts := ss.Get(ThumbSize{}).(ThumbSize)
	if *value > max-visible {
		*value = max - visible
		changed = true
	}
	if *value < min {
		*value = min
		changed = true
	}
	var pt, pb vdraw.Path
	if horizontal {
		var hb, hb0, ht0 float32
		ht := uf.ControlArea.Height()
		if ts.ThumbSize < 0 {
			hb, ht, ht0 = ht, ht+ts.ThumbSize, -ts.ThumbSize
		} else {
			hb, hb0 = ht-ts.ThumbSize, ts.ThumbSize/2
		}
		pb.AddRoundedRect(true, mgl32.Vec2{0, hb0}, mgl32.Vec2{uf.ControlArea.Width(), hb}, vdraw.UniformCorners(hb/2))
		lt := (*value - min) / w * uf.ControlArea.Width()
		wt := visible / w * uf.ControlArea.Width()
		if wt < ht {
			wt = ht
		}
		pt.AddRoundedRect(true, mgl32.Vec2{lt, ht0}, mgl32.Vec2{wt, ht}, vdraw.UniformCorners(ht/2))
	} else {
		var vb, vb0, vt0 float32
		vt := uf.ControlArea.Width()
		if ts.ThumbSize < 0 {
			vb, vt, vt0 = vt, vt+ts.ThumbSize, -ts.ThumbSize
		} else {
			vb, vb0 = vt-ts.ThumbSize, ts.ThumbSize/2
		}
		pb.AddRoundedRect(true, mgl32.Vec2{vb0, 0}, mgl32.Vec2{vb, uf.ControlArea.Height()}, vdraw.UniformCorners(vb/2))
		lt := (*value - min) / w * uf.ControlArea.Height()
		wt := visible / w * uf.ControlArea.Height()
		if wt < vt {
			wt = vt
		}
		pt.AddRoundedRect(true, mgl32.Vec2{vt0, lt}, mgl32.Vec2{vt, wt}, vdraw.UniformCorners(vt/2))

	}
	c := uf.Canvas()
	if !bg.Brush.IsNil() {
		var p vdraw.Path
		p.AddRect(true, mgl32.Vec2{}, uf.ControlArea.Size())
		c.Draw(p.Fill(), uf.ControlArea.From, mgl32.Vec2{1, 1}, &bg.Brush)
	}
	c.Draw(pb.Fill(), uf.ControlArea.From, mgl32.Vec2{1, 1}, &bc.Brush)
	c.Draw(pt.Fill(), uf.ControlArea.From, mgl32.Vec2{1, 1}, &fc.Brush)
	if uf.MouseDrag(1) {
		changed = true
		var newValue float32
		if horizontal {
			newValue = (uf.MousePos[0]-uf.ControlArea.From[0])/uf.ControlArea.Width()*(max-min) + min - visible/2
		} else {
			newValue = (uf.MousePos[1]-uf.ControlArea.From[1])/uf.ControlArea.Height()*(max-min) + min - visible/2
		}
		if newValue > max-visible {
			*value = max - visible
		} else if newValue < min {
			*value = min
		} else {
			*value = newValue
		}
	}
	return changed
}
