package vimgui

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vk"
	"image"
)

type UIEvent struct {
	MButtonDown int
	MButtonUp   int
	KeyDown     vapp.GLFWKeyCode
	KeyUp       vapp.GLFWKeyCode
	Char        rune
	Scroll      image.Point
}

type UIFrame struct {
	Ev       UIEvent
	Mods     vapp.Mods
	MousePos mgl32.Vec2
	MDownPos mgl32.Vec2

	DrawArea    vdraw.Area
	ControlArea vdraw.Area
	Tags        []string

	TotalTime float64
	DeltaTime float64

	theme         *Theme
	states        map[vk.Key]*vk.State
	focusCtrl     vk.Key
	prevFocusCtrl vk.Key
	popStack      []func()
	winArea       vdraw.Area
	dev           *vk.Device
	cp            *vdraw.CanvasPainter
}

// HasMods check if last event had at least one of given mods and no other mods.
// If mods == 0 HasMods checks that there is no modifiers
func (f *UIFrame) HasMods(mods vapp.Mods) bool {
	if mods == 0 {
		return f.Mods == 0
	}
	if (f.Mods & mods) == 0 {
		return false
	}
	if (f.Mods & ^mods) != 0 {
		return false
	}
	return true
}

func (f *UIFrame) GetStyles(styleNames ...string) StyleSet {
	if len(f.Tags) > 0 {
		tags := append([]string{}, f.Tags...)
		return f.theme.GetStyles(append(tags, styleNames...)...)
	}
	return f.theme.GetStyles(styleNames...)
}

func (f *UIFrame) WithTags(tags ...string) *UIFrame {
	f.Tags = tags
	return f
}

func (f *UIFrame) GetState(ctrlID vk.Key, defaultValue interface{}) interface{} {
	s, ok := f.states[ctrlID]
	if ok {
		return s.Get(defaultValue)
	}
	return defaultValue
}

func (f *UIFrame) SetState(ctrlID vk.Key, values ...interface{}) {
	s, ok := f.states[ctrlID]
	if !ok {
		s = &vk.State{}
		f.states[ctrlID] = s
	}
	s.Set(values...)
}

func (f *UIFrame) Canvas() *vdraw.CanvasPainter {
	return f.cp
}

// ResetControlArea resets control are to upper left corner of draw area. Use NewLine to set size of first control
func (f *UIFrame) ResetControlArea() {
	f.ControlArea.From, f.ControlArea.To = f.DrawArea.From, f.DrawArea.From
}

func (f *UIFrame) NewLine(colWidth float32, lineHeight float32, padding float32) {
	pad := f.ToPixels(padding, 0, false)
	f.ControlArea.From = mgl32.Vec2{f.DrawArea.From[0], f.ControlArea.To[1] + pad}
	f.ControlArea.To = f.ControlArea.From.Add(mgl32.Vec2{f.ToPixels(colWidth, 0, true), f.ToPixels(lineHeight, pad, false)})
}

func (f *UIFrame) NewColumn(colWidth float32, padding float32) {
	pad := f.ToPixels(padding, 0, true)
	f.ControlArea.From[0] = f.ControlArea.To[0] + pad
	f.ControlArea.To[0] = f.ControlArea.From[0] + f.ToPixels(colWidth, pad, true)
}

func (f *UIFrame) IsHidden() bool {
	if f.DrawArea.From[0] > f.ControlArea.To[0] {
		return true
	}
	if f.DrawArea.To[0] < f.ControlArea.From[0] {
		return true
	}
	if f.DrawArea.From[1] > f.ControlArea.To[1] {
		return true
	}
	if f.DrawArea.To[1] < f.ControlArea.From[1] {
		return true
	}
	return false
}

func (f *UIFrame) MouseHover() bool {
	if !f.DrawArea.Contains(f.MousePos) {
		return false
	}
	if f.ControlArea.Contains(f.MousePos) {
		return true
	}
	return false
}

func (f *UIFrame) MouseClick(button int) bool {
	if !f.DrawArea.Contains(f.MousePos) {
		return false
	}
	if f.ControlArea.Contains(f.MousePos) {
		return f.Ev.MButtonUp == button
	}
	return false
}

func (f *UIFrame) MousePress(button int) bool {
	if !f.DrawArea.Contains(f.MousePos) {
		return false
	}
	if f.ControlArea.Contains(f.MousePos) {
		return f.Ev.MButtonDown == button
	}
	return false
}

func (f *UIFrame) MouseDrag(button uint32) bool {
	if !f.DrawArea.Contains(f.MDownPos) || !f.ControlArea.Contains(f.MDownPos) {
		return false
	}
	return f.HasMods(vapp.MODMouseButton1 << (button - 1))
}

func (f *UIFrame) SetFocus(ctrlId vk.Key) {
	f.focusCtrl = ctrlId
}

func (f *UIFrame) HasFocus(ctrlId vk.Key) bool {
	return f.focusCtrl == ctrlId
}

func (f *UIFrame) MoveFocus(ctrlId vk.Key) bool {
	if f.focusCtrl == 0 {
		f.focusCtrl = ctrlId
	}
	if f.focusCtrl == ctrlId {
		if f.Ev.KeyDown == vapp.GLFWKeyTab {
			if f.HasMods(0) {
				f.focusCtrl = 0
				f.Ev.KeyDown = 0
				return false
			}
			if f.HasMods(vapp.MODShift) {
				f.focusCtrl = f.prevFocusCtrl
				f.Ev.KeyDown = 0
				return false
			}
		}
		return true
	}
	f.prevFocusCtrl = ctrlId
	return false
}

func (f *UIFrame) ToPixels(val float32, pad float32, horizontal bool) float32 {
	if val >= 0 {
		return val
	}
	if horizontal {
		return -val*f.DrawArea.Width()/100 - pad
	}
	return -val*f.DrawArea.Height()/100 - pad
}

func (f *UIFrame) Pop() {
	l := len(f.popStack)
	if l == 0 {
		f.dev.ReportError(errors.New("Attempt to pop empty stack"))
		return
	}
	pp := f.popStack[l-1]
	pp()
	f.popStack = f.popStack[:l-1]
}

func (f *UIFrame) PushControlArea() {
	cp := f.ControlArea
	f.popStack = append(f.popStack, func() {
		f.ControlArea = cp
	})
}

func (f *UIFrame) PushArea(newArea vdraw.Area, clip bool) {
	oldArea := f.DrawArea
	oldClip := f.cp.Clip
	f.DrawArea = newArea
	if clip {
		f.cp.Clip = newArea
	}
	f.popStack = append(f.popStack, func() {
		f.DrawArea = oldArea
		if clip {
			f.cp.Clip = oldClip
		}
	})
}

func (f *UIFrame) PushPadding(padding vdraw.Edges, clip bool) {
	a := padding.Shrink(f.DrawArea, 0)
	f.PushArea(a, clip)
}

func (f *UIFrame) PushTags(tags ...string) {
	oTags := f.Tags
	f.Tags = tags
	f.popStack = append(f.popStack, func() {
		f.Tags = oTags
	})
}

func (f *UIFrame) handleEvent(ev vapp.Event) {
	from := f.winArea.From
	mme, ok := ev.(*vapp.MouseMoveEvent)
	if ok {
		pos := mgl32.Vec2{float32(mme.MousePos.X), float32(mme.MousePos.Y)}
		pos = pos.Sub(from)
		f.MousePos = pos
		f.Mods = mme.CurrentMods
	}
	mde, ok := ev.(*vapp.MouseDownEvent)
	if ok {
		pos := mgl32.Vec2{float32(mde.MousePos.X), float32(mde.MousePos.Y)}
		pos = pos.Sub(from)
		f.MousePos = pos
		f.MDownPos = pos
		f.Mods = mde.CurrentMods
		f.Ev.MButtonDown = mde.Button + 1
	}
	mdu, ok := ev.(*vapp.MouseUpEvent)
	if ok {
		pos := mgl32.Vec2{float32(mdu.MousePos.X), float32(mdu.MousePos.Y)}
		pos = pos.Sub(from)
		f.MousePos = pos
		f.Ev.MButtonUp = mdu.Button + 1
		f.Mods = mdu.CurrentMods
	}
	ku, ok := ev.(*vapp.KeyUpEvent)
	if ok {
		f.Ev.KeyUp = ku.KeyCode
		f.Mods = ku.CurrentMods
	}
	kd, ok := ev.(*vapp.KeyDownEvent)
	if ok {
		f.Ev.KeyDown = kd.KeyCode
		f.Mods = kd.CurrentMods
	}
	kc, ok := ev.(*vapp.CharEvent)
	if ok {
		f.Ev.Char = kc.Char
		f.Mods = kc.CurrentMods
	}
	sc, ok := ev.(*vapp.ScrollEvent)
	if ok {
		f.Ev.Scroll = f.Ev.Scroll.Add(sc.Range)
	}
}

func (f *UIFrame) Pad(edges vdraw.Edges) {
	f.DrawArea.From = f.DrawArea.From.Add(mgl32.Vec2{edges.Left, edges.Top})
	f.DrawArea.To = f.DrawArea.To.Sub(mgl32.Vec2{edges.Right, edges.Bottom})
}
