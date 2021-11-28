package vapp

import (
	"github.com/lakal3/vge/vge/forward"
	"image"
	"sync"
	"time"

	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
)

const (
	evNil         = 0
	evQuit        = 100
	evCloseWindow = 101
	// evCreateWindow = 102
	evResizeWindow = 103
	evKeyUp        = 200
	evKeyDown      = 201
	evChar         = 202
	evMouseUp      = 300
	evMouseDown    = 301
	evMouseMove    = 302
	evMouseScroll  = 303
)

type Desktop struct {
	// ImageUsage flags for main swapchain images. Default is IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit
	ImageUsage vk.ImageUsageFlags
}

var winCount int

func GetMonitorArea(monitor uint32) (pos vk.WindowPos, exists bool) {
	return appStatic.desktop.GetMonitor(monitor)
}

// NewRenderWindow creates new default size window (1024 x 768)
func NewRenderWindow(title string, renderer vscene.Renderer) *RenderWindow {
	return NewRenderWindowAt(title, vk.WindowPos{Left: -1, Top: -1, Width: 1024, Height: 768}, renderer)
}

// NewRenderWindowAt creates new window with given size, position and state.
// To mimic full screen mode, create a borderless window matching monitor size and location
func NewRenderWindowAt(title string, wp vk.WindowPos, renderer vscene.Renderer) *RenderWindow {
	rw := &RenderWindow{renderer: renderer}
	rw.owner = vk.NewOwner(true)
	rw.owner.AddChild(renderer)
	rw.win = appStatic.desktop.NewWindow(title, wp)
	rw.WindowSize = image.Point{X: int(wp.Width), Y: int(wp.Height)}
	rw.owner.AddChild(rw.win)
	winCount++
	rw.wg = &sync.WaitGroup{}
	rw.state = 1
	rw.wg.Add(1)
	rw.Scene.Init()
	rw.Env = rw.Scene.AddNode(nil, nil)
	rw.Model = rw.Scene.AddNode(nil, nil)
	rw.Ui = rw.Scene.AddNode(nil, nil)
	rw.Camera = vscene.NewPerspectiveCamera(1000)
	RegisterHandler(PRIWindow, rw.eventHandler)
	go rw.renderLoop()
	return rw
}

type RenderWindow struct {
	Scene       vscene.Scene
	Camera      vscene.Camera
	CurrentMods Mods
	Env         *vscene.Node
	Model       *vscene.Node
	Ui          *vscene.Node

	MousePos   image.Point
	WindowSize image.Point
	// OnClose is called when user request closing windows. Default action disposes window
	OnClose func()

	owner      vk.Owner
	renderer   vscene.Renderer
	win        *vk.Window
	caches     []*vk.RenderCache
	wg         *sync.WaitGroup
	lastRender time.Time
	sceneTime  float64
	paused     bool
	state      int
	setup      bool
}

type rawWinEvent struct {
	win *vk.Window
	ev  vk.RawEvent
}

func (r rawWinEvent) Handled() bool {
	return r.win == nil
}

func (d Desktop) InitApp() {
	var wg *sync.WaitGroup
	var shutdown bool
	RegisterHandler(PRILast+1, func(ev Event) (unregister bool) {
		_, ok := ev.(StartupEvent)
		if ok {
			wg = &sync.WaitGroup{}
			wg.Add(1)
			go d.pollDesktopEvents(wg, &shutdown)
		}
		_, ok = ev.(ShutdownEvent)
		if ok {
			if wg != nil {
				shutdown = true
				wg.Wait()
			}
			return true
		}
		return false
	})
	if d.ImageUsage != 0 {
		appStatic.desktop = vk.NewDesktopWithSettings(App, vk.DesktopSettings{ImageUsage: d.ImageUsage})
	} else {
		appStatic.desktop = vk.NewDesktop(App)
	}
}

func (d Desktop) TerminateApp() {

}

func (d *Desktop) pollDesktopEvents(wg *sync.WaitGroup, shutdown *bool) {
	for !*shutdown {
		ev, win := appStatic.desktop.PullEvent()
		if ev.EventType != 0 {
			Post(&rawWinEvent{win: win, ev: ev})
		} else {
			<-time.After(1 * time.Millisecond)
		}
	}
	wg.Done()
}

func (rw *RenderWindow) Closed() bool {
	return rw.state == 2
}

func (rw *RenderWindow) SetPaused(paused bool) {
	rw.paused = paused
}

func (rw *RenderWindow) GetSceneTime() float64 {
	return rw.sceneTime
}

func (rw *RenderWindow) GetRenderer() vscene.Renderer {
	return rw.renderer
}

func (rw *RenderWindow) SetPos(pos vk.WindowPos) {
	rw.win.SetPos(pos)
}

func (rw *RenderWindow) SetClipboard(cpText string) {
	rw.win.SetClipboard(cpText)
}

func (rw *RenderWindow) GetClipboard() string {
	return rw.win.GetClipboard()
}

func (rw *RenderWindow) Dispose() {
	if rw.state < 2 {
		rw.state = 2
		rw.wg.Wait()
		rw.owner.Dispose()
		rw.win = nil
		winCount--
		if winCount <= 0 {
			// Last window closed, terminate application
			go Terminate()
		}
	}
}

// Add disposable object bound to windows lifetime. Safe for concurrent access.
func (rw *RenderWindow) AddChild(disp vk.Disposable) {
	rw.owner.AddChild(disp)
}

func (rw *RenderWindow) renderLoop() {
	rw.Scene.Init()
	for rw.state == 1 {
		im, imageIndex, submitInfo := rw.win.GetNextFrame(Dev)
		if imageIndex < 0 {
			rw.clearCaches()
			continue
		}
		if len(rw.caches) == 0 {
			rw.caches = make([]*vk.RenderCache, rw.win.GetImageCount())
		}
		if rw.caches[imageIndex] == nil {
			rw.caches[imageIndex] = vk.NewRenderCache(Dev)
			if !rw.setup {
				if rw.renderer == nil {
					rw.renderer = forward.NewRenderer(true)
				}
				rw.setup = true
				rw.renderer.Setup(Dev, rw.win.WindowDesc, rw.win.GetImageCount())
			}
		}
		rc := rw.caches[imageIndex]
		rc.NewFrame()
		rw.Scene.Time = rw.GetSceneTime()
		rw.renderer.Render(rw.Camera, &rw.Scene, rc, im, int(imageIndex), []vk.SubmitInfo{submitInfo})

		// Adjust scene time
		if rw.paused {
			rw.lastRender = time.Time{}
		} else {
			t := time.Now()
			if !rw.lastRender.IsZero() {
				d := t.Sub(rw.lastRender).Seconds()
				rw.sceneTime += d
			}
			rw.lastRender = t
		}
	}
	rw.clearCaches()
	rw.wg.Done()
}

func (rw *RenderWindow) clearCaches() {
	for _, ca := range rw.caches {
		if ca != nil {
			ca.Dispose()
		}
	}
	rw.caches = nil
	rw.setup = false
}

func (rw *RenderWindow) eventHandler(ev Event) (unregister bool) {
	if rw.win == nil {
		return true // Done
	}
	raw, ok := ev.(*rawWinEvent)
	if ok && rw.win == raw.win {
		switch raw.ev.EventType {
		case evResizeWindow:
			rw.WindowSize = image.Point{X: int(raw.ev.Arg1), Y: int(raw.ev.Arg2)}
		case evCloseWindow:
			if rw.OnClose != nil {
				rw.OnClose()
			} else {
				go rw.Dispose()
			}
		case evKeyUp:
			kc := GLFWKeyCode(raw.ev.Arg2)
			m := rw.parseModKey(kc)
			if m != 0 {
				rw.CurrentMods &= ^m
			}
			Post(&KeyUpEvent{KeyCode: kc, ScanCode: uint32(raw.ev.Arg1), UIEvent: rw.uiEvent(), NumLock: rw.hasNumlock(raw.ev.Arg3)})
		case evKeyDown:
			kc := GLFWKeyCode(raw.ev.Arg2)
			m := rw.parseModKey(kc)
			if m != 0 {
				rw.CurrentMods |= m
			}
			Post(&KeyDownEvent{KeyCode: kc, ScanCode: uint32(raw.ev.Arg1), UIEvent: rw.uiEvent(), NumLock: rw.hasNumlock(raw.ev.Arg3)})
		case evChar:
			Post(&CharEvent{Char: rune(raw.ev.Arg1), UIEvent: rw.uiEvent()})
		case evMouseUp:
			m := Mods(MODMouseButton1) << uint32(raw.ev.Arg1)
			rw.CurrentMods &= ^m
			Post(&MouseUpEvent{Button: int(raw.ev.Arg1), UIEvent: rw.uiEvent()})
		case evMouseDown:
			m := Mods(MODMouseButton1) << uint32(raw.ev.Arg1)
			rw.CurrentMods |= m
			Post(&MouseDownEvent{Button: int(raw.ev.Arg1), UIEvent: rw.uiEvent()})
		case evMouseMove:
			rw.MousePos = image.Point{X: int(raw.ev.Arg1), Y: int(raw.ev.Arg2)}
			Post(&MouseMoveEvent{UIEvent: rw.uiEvent()})
		case evMouseScroll:
			Post(&ScrollEvent{Range: image.Point{X: int(raw.ev.Arg1), Y: int(raw.ev.Arg2)}, UIEvent: rw.uiEvent()})
		}
	}
	_, ok = ev.(ShutdownEvent)
	if ok && rw.state == 1 {
		rw.Dispose()
		return true
	}
	return false
}

func (rw *RenderWindow) parseModKey(code GLFWKeyCode) Mods {
	if code >= GLFWKeyLeftShift && code < GLFWKeyLeftShift+4 {
		return Mods(1 << (1 + code - GLFWKeyLeftShift))
	}
	if code >= GLFWKeyRightShift && code < GLFWKeyLeftShift+4 {
		return Mods(1 << (2 + code - GLFWKeyLeftShift))
	}
	return 0
}

func (rw *RenderWindow) uiEvent() UIEvent {
	return UIEvent{Window: rw, CurrentMods: rw.CurrentMods, MousePos: rw.MousePos}
}

func (rw *RenderWindow) hasNumlock(arg3 int32) bool {
	return (arg3 & 0x20) != 0
}
