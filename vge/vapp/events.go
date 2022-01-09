package vapp

import (
	"github.com/lakal3/vge/vge/vk"
	"image"
	"sync"
)

type ShutdownEvent struct {
	wg *sync.WaitGroup
}

func (s ShutdownEvent) Done() {
	s.wg.Done()
}

func (s ShutdownEvent) Handled() bool {
	return false
}

type StartupEvent struct {
}

func (s StartupEvent) Handled() bool {
	return false
}

type Mods uint32

const (
	MODLeftShift    = Mods(1)
	MODRightShift   = Mods(2)
	MODShift        = MODLeftShift + MODRightShift
	MODLeftCtrl     = Mods(4)
	MODRightCtrl    = Mods(8)
	MODCtrl         = MODLeftCtrl + MODRightCtrl
	MODLeftAlt      = Mods(16)
	MODRightAlt     = Mods(32)
	MODAlt          = MODLeftAlt + MODRightAlt
	MODMouseButton1 = Mods(0x1000)
	MODMouseButton2 = Mods(0x2000)
	MODMouseButton3 = Mods(0x4000)
)

const (
	// Windows event priority
	PRIWindow = 100
	// Lowest priority
	PRILast = -1000
)

type GLFWKeyCode uint32

const (
	// Shirt, Ctrl (+1), Alt (+2), System (+3)
	GLFWKeyLeftShift      GLFWKeyCode = 340
	GLFWKeyRightShift     GLFWKeyCode = 344
	GLFWKeyEscape         GLFWKeyCode = 256
	GLFWKeyTab            GLFWKeyCode = 258
	GLFWKeyLeft           GLFWKeyCode = 263
	GLFWKeyRight          GLFWKeyCode = 262
	GLFWKeyBackspace      GLFWKeyCode = 259
	GLFWKeyDelete         GLFWKeyCode = 261
	GLFWKeyHome           GLFWKeyCode = 268
	GLFWKeyEnd            GLFWKeyCode = 269
	GLFWKeyNumlock        GLFWKeyCode = 282
	GLFWKeyF1             GLFWKeyCode = 290
	GLFWKeyNumpad0        GLFWKeyCode = 320
	GLWDKeyNumpadDecimal  GLFWKeyCode = 330
	GLWDKeyNumpadDivide   GLFWKeyCode = 331
	GLWDKeyNumpadMultiply GLFWKeyCode = 332
	GLWDKeyNumpadSub      GLFWKeyCode = 333
	GLWDKeyNumpadAdd      GLFWKeyCode = 334
	GLWDKeyNumpadEnter    GLFWKeyCode = 335
)

type EventSource interface {
}

type UIState struct {
	CurrentMods Mods
	MousePos    image.Point
}

func (us *UIState) MakeUIEvent(raw *RawWinEvent, es EventSource) {
	switch raw.Ev.EventType {
	case evKeyUp:
		kc := GLFWKeyCode(raw.Ev.Arg2)
		m := us.parseModKey(kc)
		if m != 0 {
			us.CurrentMods &= ^m
		}
		Post(&KeyUpEvent{KeyCode: kc, ScanCode: uint32(raw.Ev.Arg1), UIEvent: us.uiEvent(es), NumLock: us.hasNumlock(raw.Ev.Arg3)})
	case evKeyDown:
		kc := GLFWKeyCode(raw.Ev.Arg2)
		m := us.parseModKey(kc)
		if m != 0 {
			us.CurrentMods |= m
		}
		Post(&KeyDownEvent{KeyCode: kc, ScanCode: uint32(raw.Ev.Arg1), UIEvent: us.uiEvent(es), NumLock: us.hasNumlock(raw.Ev.Arg3)})
	case evChar:
		Post(&CharEvent{Char: rune(raw.Ev.Arg1), UIEvent: us.uiEvent(es)})
	case evMouseUp:
		m := Mods(MODMouseButton1) << uint32(raw.Ev.Arg1)
		us.CurrentMods &= ^m
		Post(&MouseUpEvent{Button: int(raw.Ev.Arg1), UIEvent: us.uiEvent(es)})
	case evMouseDown:
		m := Mods(MODMouseButton1) << uint32(raw.Ev.Arg1)
		us.CurrentMods |= m
		Post(&MouseDownEvent{Button: int(raw.Ev.Arg1), UIEvent: us.uiEvent(es)})
	case evMouseMove:
		us.MousePos = image.Point{X: int(raw.Ev.Arg1), Y: int(raw.Ev.Arg2)}
		Post(&MouseMoveEvent{UIEvent: us.uiEvent(es)})
	case evMouseScroll:
		Post(&ScrollEvent{Range: image.Point{X: int(raw.Ev.Arg1), Y: int(raw.Ev.Arg2)}, UIEvent: us.uiEvent(es)})
	}
}

func (us *UIState) parseModKey(code GLFWKeyCode) Mods {
	if code >= GLFWKeyLeftShift && code < GLFWKeyLeftShift+4 {
		return Mods(1 << (1 + code - GLFWKeyLeftShift))
	}
	if code >= GLFWKeyRightShift && code < GLFWKeyLeftShift+4 {
		return Mods(1 << (2 + code - GLFWKeyLeftShift))
	}
	return 0
}

func (us *UIState) uiEvent(es EventSource) UIEvent {
	return UIEvent{Source: es, CurrentMods: us.CurrentMods, MousePos: us.MousePos}
}

func (us *UIState) hasNumlock(arg3 int32) bool {
	return (arg3 & 0x20) != 0
}

type UIEvent struct {
	CurrentMods Mods
	MousePos    image.Point
	Source      EventSource
	handled     bool
}

// Deprecated: Use IsSource
func (u *UIEvent) IsWin(rw *RenderWindow) bool {
	return u.Source == rw
}

func (u *UIEvent) IsSource(es EventSource) bool {
	return u.Source == es
}

func (u *UIEvent) HasMods(mods Mods) bool {
	if (mods & MODShift) == MODShift {
		return u.HasMods(mods & ^MODRightShift) || u.HasMods(mods & ^MODLeftShift)
	}
	if (mods & MODCtrl) == MODCtrl {
		return u.HasMods(mods & ^MODRightCtrl) || u.HasMods(mods & ^MODLeftCtrl)
	}
	if (mods & MODAlt) == MODAlt {
		return u.HasMods(mods & ^MODRightAlt) || u.HasMods(mods & ^MODLeftAlt)
	}
	return (u.CurrentMods & mods) == mods
}

func (u *UIEvent) Handled() bool {
	return u.handled
}

func (u *UIEvent) SetHandled() {
	u.handled = true
}

type KeyUpEvent struct {
	UIEvent
	KeyCode  GLFWKeyCode
	ScanCode uint32
	NumLock  bool
}

type KeyDownEvent struct {
	UIEvent
	KeyCode  GLFWKeyCode
	ScanCode uint32
	NumLock  bool
}

type CharEvent struct {
	UIEvent
	Char rune
}

type ScrollEvent struct {
	UIEvent
	Range image.Point
}

type MouseMoveEvent struct {
	UIEvent
}

type MouseUpEvent struct {
	// first button = 0
	Button int
	UIEvent
}

type MouseDownEvent struct {
	// first button = 0
	Button int
	UIEvent
}

type RawWinEvent struct {
	Win *vk.Window
	Ev  vk.RawEvent
}

func (r RawWinEvent) Handled() bool {
	return r.Win == nil
}
