package vui

import (
	"image"
	"math"

	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
)

const WinPriority = 10
const DialogPriority = 100
const PopupPriority = 200

type Owner interface {
	Theme() Theme
	VisibleSize() image.Point
	SetFocus(ctrl Control)
	GetFocus() (ctrl Control)
}

var FocusView Owner
var ActiveDialog Owner
var ActivePopup Owner

type UIView struct {
	MainCtrl Control
	Focus    Control
	Priority float64
	Area     image.Rectangle
	visible  bool
	disposed bool
	theme    Theme
	fullSize image.Point
	win      *vapp.RenderWindow
	down2Pos image.Point
	down1Pos image.Point
	popup    bool
}

func (uv *UIView) SetFocus(ctrl Control) {
	uv.Focus = ctrl
	FocusView = uv
}

func (uv *UIView) GetFocus() (ctrl Control) {
	if FocusView != uv {
		return nil
	}
	return uv.Focus
}

func (uv *UIView) VisibleSize() image.Point {
	return uv.Area.Size()
}

func (uv *UIView) Process(pi *vscene.ProcessInfo) {
	if !uv.visible {
		return
	}
	uv.fullSize = uv.win.WindowSize
	dp, ok := pi.Phase.(vscene.DrawPhase)
	if ok {
		dc := dp.GetDC(vscene.LAYERUI)
		if dc == nil {
			return
		}
		if uv.MainCtrl != nil {
			pos := vglyph.Position{
				Clip:      uv.Area,
				GlyphArea: uv.Area,
				ImageSize: uv.fullSize,
			}
			// pos.GlyphArea.Max = pos.GlyphArea.Min.Add(uv.MainCtrl.Measure(uv, uv.Area.Size()))
			uv.MainCtrl.Render(uv, dc, pos)
		}
	}
}

func (w *UIView) Hide() {
	w.visible = false
	if w == ActiveDialog {
		ActiveDialog = nil
	}
	if w == ActivePopup {
		ActivePopup = nil
	}
}

func (w *UIView) Theme() Theme {
	return w.theme
}

type FocusNextEvent struct {
	IsHandled bool
}

func (f *FocusNextEvent) Handled() bool {
	return f.IsHandled
}

type FocusPrevEvent struct {
	IsHandled bool
	PrevCtrl  Control
}

func (f *FocusPrevEvent) Handled() bool {
	return f.IsHandled
}

func (w *UIView) handleEvent(ev vapp.Event) (unregister bool) {
	if !w.visible {
		return true
	}
	if w.MainCtrl == nil {
		return false
	}
	if ActivePopup != nil && ActivePopup != w {
		return false
	}
	if ActiveDialog != nil && ActiveDialog != w {
		return false
	}

	mme, ok := ev.(*vapp.MouseMoveEvent)
	if ok {
		if mme.Window != w.win || !mme.MousePos.In(w.Area) {
			return false
		}
		if w.win.CurrentMods < vapp.MODMouseButton1 {
			// No buttons pressed
			he := MouseHoverEvent{At: mme.MousePos}
			w.MainCtrl.Event(w, &he)

		} else if mme.HasMods(vapp.MODMouseButton1) {
			if w.down1Pos.X >= 0 && ptLen(w.down1Pos.Sub(mme.MousePos)) >= 3 {
				md := MouseDragEvent{From: w.down1Pos, At: mme.MousePos}
				w.MainCtrl.Event(w, &md)
			} else {
				he := MouseHoverEvent{At: mme.MousePos, Pressed: true}
				w.MainCtrl.Event(w, &he)
			}
		}
		mme.SetHandled()
		return false
	}
	mde, ok := ev.(*vapp.MouseDownEvent)
	if ok {
		if mde.Window != w.win {
			return false
		}
		if !mde.MousePos.In(w.Area) {
			if w.popup {
				w.Hide()
				mde.SetHandled()
				return true
			}
			return false
		}
		if mde.Button == 0 {
			w.down1Pos = mde.MousePos
			mp := MouseHoverEvent{At: mde.MousePos, Pressed: true}
			w.MainCtrl.Event(w, &mp)
		}
		if mde.Button == 1 {
			w.down2Pos = mde.MousePos
		}
		mde.SetHandled()
	}
	mue, ok := ev.(*vapp.MouseUpEvent)
	if ok {
		if mue.Window != w.win || !mue.MousePos.In(w.Area) {
			return false
		}
		if mue.Button == 0 {
			mp := MouseHoverEvent{At: mue.MousePos, Pressed: false}
			w.MainCtrl.Event(w, &mp)
			if ptLen(w.down1Pos.Sub(mue.MousePos)) < 3 {
				he := MouseClickEvent{At: mue.MousePos}
				w.MainCtrl.Event(w, &he)
			}
			w.down1Pos = image.Pt(-10, -10)
		}
		mue.SetHandled()
	}
	ke, ok := ev.(*vapp.KeyUpEvent)
	if ok {
		if ke.KeyCode == vapp.GLFWKeyTab {
			if w != FocusView {
				return false
			}
			if ke.CurrentMods == 0 {
				w.MainCtrl.Event(w, &FocusNextEvent{})
			}
			if ke.HasMods(vapp.MODShift) {
				w.MainCtrl.Event(w, &FocusPrevEvent{})
			}
		} else {
			if w.IsKeyActive() {
				w.Focus.Event(w, ev)
			}
		}

	}
	kc, ok := ev.(*vapp.CharEvent)
	if ok && w.IsKeyActive() {
		w.Focus.Event(w, kc)
	}
	return false
}

func ptLen(pt image.Point) float64 {
	return math.Sqrt(float64(pt.X*pt.X + pt.Y*pt.Y))
}

func NewUIView(theme Theme, area image.Rectangle, win *vapp.RenderWindow) *UIView {
	w := &UIView{theme: theme, Area: area, win: win}
	return w
}

func (uv *UIView) Show() *UIView {
	uv.visible = true
	vapp.RegisterHandler(WinPriority, uv.handleEvent)
	FocusView = uv
	return uv
}

func (uv *UIView) ShowInactive() *UIView {
	uv.visible = true
	return uv
}

func (uv *UIView) ShowDialog() {
	uv.visible = true
	vapp.RegisterHandler(DialogPriority, uv.handleEvent)
	FocusView = uv
	ActiveDialog = uv
}

func (uv *UIView) ShowPopup() {
	uv.visible = true
	vapp.RegisterHandler(PopupPriority, uv.handleEvent)
	FocusView = uv
	ActivePopup = uv
	uv.popup = true
}

func (uv *UIView) DefaultFrame(content Control) *Panel {
	p := &Panel{Class: "win", Content: &Padding{
		Padding: image.Rect(15, 15, 15, 15), Clip: true,
		Content: content,
	}}
	uv.MainCtrl = p
	return p
}

func (uv *UIView) DrawTo(img *vk.Image) {
	uv.fullSize, uv.visible = image.Pt(int(img.Description.Width), int(img.Description.Height)), true
}

func (uv *UIView) IsKeyActive() bool {
	return uv.Focus != nil
}

func (uv *UIView) SetContent(ctrl Control) *UIView {
	uv.MainCtrl = ctrl
	return uv
}

type MouseHoverEvent struct {
	IsHandled bool
	Pressed   bool
	At        image.Point
}

func (m *MouseHoverEvent) Handled() bool {
	return m.IsHandled
}

type MouseClickEvent struct {
	IsHandled bool
	At        image.Point
}

func (m *MouseClickEvent) Handled() bool {
	return m.IsHandled
}

type MouseDragEvent struct {
	IsHandled bool
	From      image.Point
	At        image.Point
}

func (m *MouseDragEvent) Handled() bool {
	return m.IsHandled
}
