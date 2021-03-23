package vapp

import (
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
	GLFWKeyLeftShift  GLFWKeyCode = 340
	GLFWKeyRightShift GLFWKeyCode = 344
	GLFWKeyTab        GLFWKeyCode = 258
	GLFWKeyLeft       GLFWKeyCode = 263
	GLFWKeyRight      GLFWKeyCode = 262
	GLFWKeyBackspace  GLFWKeyCode = 259
	GLFWKeyDelete     GLFWKeyCode = 261
	GLFWKeyF1         GLFWKeyCode = 290
)

type UIEvent struct {
	CurrentMods Mods
	MousePos    image.Point
	handled     bool
	Window      *RenderWindow
}

func (u *UIEvent) IsWin(win *RenderWindow) bool {
	return win == u.Window
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
}

type KeyDownEvent struct {
	UIEvent
	KeyCode  GLFWKeyCode
	ScanCode uint32
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
