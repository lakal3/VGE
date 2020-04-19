package vui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
)

type MouseState struct {
	area     image.Rectangle
	State    State
	EventPos image.Point
}

func CheckSetFocus(ctrl Control, owner Owner, ev vapp.Event) bool {
	fne, ok := ev.(*FocusNextEvent)
	if ok {
		fc := owner.GetFocus()
		if fc == nil {
			fne.IsHandled = true
			owner.SetFocus(ctrl)
			return true
		}
		if fc == ctrl {
			owner.SetFocus(nil)
		}
	}
	fpe, ok := ev.(*FocusPrevEvent)
	if ok {
		fc := owner.GetFocus()
		if fc == ctrl {
			owner.SetFocus(fpe.PrevCtrl)
			return true
		} else {
			fpe.PrevCtrl = ctrl
		}
	}
	return false
}

func (ms *MouseState) SetArea(pos vglyph.Position) {
	ms.area = pos.MouseArea()
}

func (ms *MouseState) Event(owner Owner, ev vapp.Event) (clicked bool) {
	hv, ok := ev.(*MouseHoverEvent)
	if ok {
		if hv.At.In(ms.area) {
			ms.State |= STATEHover
			if hv.Pressed {
				ms.State |= STATEPressed
			} else {
				ms.State &= ^STATEPressed
			}
		} else {
			ms.State &= ^(STATEHover | STATEPressed)
		}

	}

	mc, ok := ev.(*MouseClickEvent)
	if ok {
		if mc.At.In(ms.area) {
			ms.EventPos = mc.At
			mc.IsHandled = true
			return true
		}
	}
	return false
}

func (ms *MouseState) DragEvent(owner Owner, ev vapp.Event) (ok bool) {
	mde, ok := ev.(*MouseDragEvent)
	if ok {
		if mde.From.In(ms.area) {
			ms.EventPos = mde.At
			return true
		}
	}
	return false
}

func (ms *MouseState) GetRelXPos() float32 {
	return float32(ms.EventPos.X-ms.area.Min.X) / float32(ms.area.Size().X)
}

func (ms *MouseState) GetRelYPos() float32 {
	return float32(ms.EventPos.Y-ms.area.Min.Y) / float32(ms.area.Size().Y)
}

type Field struct {
	ID       string
	Class    string
	Disabled bool
	Style    Style
}

func (f *Field) GetState(otherStates State) State {
	if f.Disabled {
		return STATEDisabled
	}
	return otherStates
}

type Button struct {
	Field
	Content Control
	OnClick func()
	ms      MouseState
}

type ButtonClickedEvent struct {
	ID      string
	Button  Control
	handled bool
}

func (b *ButtonClickedEvent) Handled() bool {
	return b.handled
}

func (b *ButtonClickedEvent) SetHandled() {
	b.handled = true
}

type ValueChangedEvent struct {
	ID       string
	Source   Control
	NewValue interface{}
	handled  bool
}

func (b *ValueChangedEvent) Handled() bool {
	return b.handled
}

func (b *ValueChangedEvent) SetHandled() {
	b.handled = true
}

func NewButton(minWidth int, text string) *Button {
	b := &Button{}
	b.ID = MakeID()
	b.Content = &Extend{Content: NewLabel(text), MinSize: image.Pt(minWidth, 20), Align: mgl32.Vec2{0.5, 0.25}}
	return b
}

func (b *Button) SetClass(class string) *Button {
	b.Class = class
	return b
}

func (b *Button) SetOnClick(clicked func()) *Button {
	b.OnClick = clicked
	return b
}

func (b *Button) AssignTo(btn **Button) *Button {
	*btn = b
	return b
}

func (b *Button) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if b.Style == nil {
		b.Style = owner.Theme().GetStyle(b, b.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, b.Content, b.Style)
}

func (b *Button) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if b.Style == nil {
		b.Style = owner.Theme().GetStyle(b, b.Class)
	}
	if b.Style != nil {
		b.Style.Draw(owner, b, dc, pos, b.GetState(b.ms.State))
	}
	b.ms.SetArea(pos)
	RenderPaddedContent(owner, dc, pos, b.Content, b.Style)
}

func (b *Button) Event(owner Owner, ev vapp.Event) {
	if b.Disabled {
		return
	}
	if b.ms.Event(owner, ev) {
		if b.OnClick != nil {
			b.OnClick()
		} else {
			vapp.Post(&ButtonClickedEvent{ID: b.ID, Button: b})
		}
	}
	if !ev.Handled() && b.Content != nil {
		b.Content.Event(owner, ev)
	}
}

func (b *Button) SetID(id string) *Button {
	b.ID = id
	return b
}

type VSlider struct {
	Field
	OnChanged func(newPos int)
	ms        MouseState
	Maximum   int
	Current   int
	Visible   int
}

func NewVSlider(current int, visible int, max int) *VSlider {
	return &VSlider{Maximum: max, Current: current, Visible: visible}
}

func (v *VSlider) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if v.Style == nil {
		v.Style = owner.Theme().GetStyle(v, v.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, nil, v.Style)
}

func (v *VSlider) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	v.ms.SetArea(pos)
	if v.Style == nil {
		v.Style = owner.Theme().GetStyle(v, v.Class)
	}
	if v.Style != nil {
		v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State))
		if v.Maximum <= v.Visible {
			v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State)|STATEContent)
			return
		}
		sz := pos.GlyphArea.Size().Y - 8
		pos.GlyphArea.Min.Y += 4 + sz*v.Current/(v.Maximum)
		pos.GlyphArea.Max.Y = pos.GlyphArea.Min.Y + sz*v.Visible/(v.Maximum)

		v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State)|STATEContent)
	}
}

func (v *VSlider) Event(owner Owner, ev vapp.Event) {
	if v.ms.Event(owner, ev) || v.ms.DragEvent(owner, ev) {
		v.Current = int(float32(v.Maximum-v.Visible) * (v.ms.GetRelYPos()*1.2 - 0.1))
		if v.Current > v.Maximum-v.Visible {
			v.Current = v.Maximum - v.Visible
		}
		if v.Current < 0 {
			v.Current = 0
		}
		if v.OnChanged != nil {
			v.OnChanged(v.Current)
		} else {
			vapp.Post(&ValueChangedEvent{ID: v.ID, Source: v, NewValue: v.Current})
		}
	}

}

type HSlider struct {
	Field
	OnChanged func(newPos int)
	ms        MouseState
	Maximum   int
	Current   int
	Visible   int
}

func NewHSlider(current int, visible int, max int) *HSlider {
	return &HSlider{Maximum: max, Current: current, Visible: visible}
}

func (v *HSlider) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if v.Style == nil {
		v.Style = owner.Theme().GetStyle(v, v.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, nil, v.Style)
}

func (v *HSlider) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	v.ms.SetArea(pos)
	if v.Style == nil {
		v.Style = owner.Theme().GetStyle(v, v.Class)
	}
	if v.Style != nil {
		v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State))
		if v.Maximum <= v.Visible {
			v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State)|STATEContent)
			return
		}
		sz := pos.GlyphArea.Size().X - 8
		pos.GlyphArea.Min.X += 4 + sz*v.Current/(v.Maximum)
		pos.GlyphArea.Max.X = pos.GlyphArea.Min.X + sz*v.Visible/(v.Maximum)
		v.Style.Draw(owner, v, dc, pos, v.GetState(v.ms.State)|STATEContent)
	}
}

func (v *HSlider) Event(owner Owner, ev vapp.Event) {
	if v.ms.Event(owner, ev) || v.ms.DragEvent(owner, ev) {
		v.Current = int(float32(v.Maximum-v.Visible) * (v.ms.GetRelXPos()*1.2 - 0.1))
		if v.Current > v.Maximum-v.Visible {
			v.Current = v.Maximum - v.Visible
		}
		if v.Current < 0 {
			v.Current = 0
		}
		if v.OnChanged != nil {
			v.OnChanged(v.Current)
		} else {
			vapp.Post(&ValueChangedEvent{ID: v.ID, Source: v, NewValue: v.Current})
		}
	}

}

type MenuButton struct {
	Field
	Content Control
	OnClick func()
	ms      MouseState
}

func NewMenuButton(text string) *MenuButton {
	mb := &MenuButton{Field: Field{ID: MakeID()}}
	mb.Content = NewLabel(text)
	return mb
}

func (m *MenuButton) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if m.Style == nil {
		m.Style = owner.Theme().GetStyle(m, m.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, m.Content, m.Style)
}

func (m *MenuButton) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	m.ms.SetArea(pos)
	if m.Style == nil {
		m.Style = owner.Theme().GetStyle(m, m.Class)
	}
	if m.Style != nil {
		m.Style.Draw(owner, m, dc, pos, m.GetState(m.ms.State))
	}
	m.ms.SetArea(pos)
	RenderPaddedContent(owner, dc, pos, m.Content, m.Style)
}

func (m *MenuButton) Event(owner Owner, ev vapp.Event) {
	if m.Disabled {
		return
	}
	if m.ms.Event(owner, ev) {
		if m.OnClick != nil {
			m.OnClick()
		} else {
			vapp.Post(&ButtonClickedEvent{ID: m.ID, Button: m})
		}
	}
	if !ev.Handled() && m.Content != nil {
		m.Content.Event(owner, ev)
	}
}

func (m *MenuButton) SetOnClick(action func()) *MenuButton {
	m.OnClick = action
	return m
}

type ToggleButton struct {
	Field
	Checked        bool
	Content        Control
	CheckedContent Control
	OnChanged      func(checked bool)
	ms             MouseState
}

func NewToggleButton(id string, text1 string, text2 string) *ToggleButton {
	mb := &ToggleButton{Field: Field{ID: id}}
	mb.Content = NewLabel(text1)
	mb.CheckedContent = NewLabel(text2)
	return mb
}

func (tb *ToggleButton) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if tb.Style == nil {
		tb.Style = owner.Theme().GetStyle(tb, tb.Class)
	}
	return MeasurePaddedContent(owner, freeWidth, tb.Content, tb.Style)
}

func (tb *ToggleButton) getContent() Control {
	if tb.Checked {
		return tb.CheckedContent
	}
	return tb.Content
}

func (tb *ToggleButton) GetState() State {
	s := tb.ms.State
	if tb.Checked {
		s |= STATEChecked
	}
	return tb.Field.GetState(s)
}

func (tb *ToggleButton) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	tb.ms.SetArea(pos)
	if tb.Style == nil {
		tb.Style = owner.Theme().GetStyle(tb, tb.Class)
	}
	if tb.Style != nil {
		tb.Style.Draw(owner, tb, dc, pos, tb.GetState())
	}
	tb.ms.SetArea(pos)
	RenderPaddedContent(owner, dc, pos, tb.getContent(), tb.Style)
}

func (tb *ToggleButton) Event(owner Owner, ev vapp.Event) {
	if tb.Disabled {
		return
	}
	if tb.ms.Event(owner, ev) {
		tb.Checked = !tb.Checked
		if tb.OnChanged != nil {
			tb.OnChanged(tb.Checked)
		} else {
			vapp.Post(&ValueChangedEvent{ID: tb.ID, Source: tb, NewValue: tb.Checked})
		}
	}
	if !ev.Handled() {
		ct := tb.getContent()
		if ct != nil {
			ct.Event(owner, ev)
		}
	}
}

func (tb *ToggleButton) SetOnChanged(action func(checked bool)) *ToggleButton {
	tb.OnChanged = action
	return tb
}

func (tb *ToggleButton) SetClass(class string) *ToggleButton {
	tb.Class = class
	return tb
}

func (tb *ToggleButton) AssignTo(to **ToggleButton) *ToggleButton {
	*to = tb
	return tb
}
