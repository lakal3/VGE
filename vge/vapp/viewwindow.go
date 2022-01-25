package vapp

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"sync"
)

type Completed func() []vk.SubmitInfo

type View interface {
	Allocate(fi *vk.FrameInstance)
	PreRender(fi *vk.FrameInstance) (done Completed)
	Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass)
}

type EventView interface {
	View
	HandleEvent(event Event)
}

type ViewWindow struct {
	OnClose func()

	eventState UIState
	fc         *vk.FrameCache
	views      []View
	sp         *vk.SpinLock
	win        *vk.Window
	wg         *sync.WaitGroup
	state      int
}

func NewViewWindow(title string, wp vk.WindowPos, views ...View) *ViewWindow {
	w := &ViewWindow{sp: &vk.SpinLock{}}
	w.win = appStatic.desktop.NewWindow(title, wp)
	RegisterHandler(PRIWindow, w.eventHandler)
	winCount++
	go w.renderLoop()
	return w
}

func (w *ViewWindow) Dispose() {
	if w.state == 1 {
		w.state = 2
		w.wg.Wait()
		if w.fc != nil {
			w.fc.Dispose()
			w.fc = nil
		}
		if w.win != nil {
			w.win.Dispose()
			w.win = nil
		}
		winCount--
		if winCount <= 0 {
			// Last window closed, terminate application
			go Terminate()
		}

	}
}

func (w *ViewWindow) Views() []View {
	w.sp.Lock()
	defer w.sp.Unlock()
	if len(w.views) == 0 {
		return nil
	}
	views := make([]View, len(w.views))
	copy(views, w.views)
	return views
}

func (w *ViewWindow) SetViews(views ...View) {
	w.sp.Lock()
	w.views = views
	w.sp.Unlock()
}

func (w *ViewWindow) AddView(view View) {
	w.sp.Lock()
	w.views = append(w.views, view)
	w.sp.Unlock()
}

func (w *ViewWindow) RemoveView(view View) {
	w.sp.Lock()
	defer w.sp.Unlock()
	for idx, v := range w.views {
		if v == view {
			w.views = append(w.views[:idx], w.views[idx+1:]...)
			return
		}
	}
}

func (w *ViewWindow) eventHandler(ev Event) (unregister bool) {
	if w.win == nil {
		return true // Done
	}
	raw, ok := ev.(*RawWinEvent)
	if ok && w.win == raw.Win {
		switch raw.Ev.EventType {
		case evResizeWindow:
			// w.WindowSize = image.Point{X: int(raw.Ev.Arg1), Y: int(raw.Ev.Arg2)}
		case evCloseWindow:
			if w.OnClose != nil {
				w.OnClose()
			} else {
				go w.Dispose()
			}
		default:
			w.eventState.MakeUIEvent(raw, w)

		}
	}
	_, ok = ev.(ShutdownEvent)
	if ok {
		w.Dispose()
		return true
	}
	w.handleMyEvent(ev)
	return false
}

func (w *ViewWindow) renderLoop() {
	w.wg = &sync.WaitGroup{}
	w.wg.Add(1)
	w.state = 1
	for w.state == 1 {
		im, imageIndex, submitInfo := w.win.GetNextFrame(Dev)
		if imageIndex < 0 {
			continue
		}
		if w.fc == nil {
			w.fc = vk.NewFrameCache(Dev, w.win.GetImageCount())
		}
		fi := w.fc.Instances[imageIndex]
		fi.Output = im
		fi.BeginFrame()
		views := w.Views()
		submits := []vk.SubmitInfo{submitInfo}
		var completed []Completed
		for _, v := range views {
			v.Allocate(fi)
		}
		fi.Commit()
		for _, v := range views {
			c := v.PreRender(fi)
			if c != nil {
				completed = append(completed, c)
			}
		}
		rp, fb := w.getRenderPass(fi, im.DefaultView())
		cmd := w.newCommand(fi)
		cmd.Begin()
		cmd.BeginRenderPass(rp, fb)
		for _, v := range views {
			v.Render(fi, cmd, rp)
		}
		cmd.EndRenderPass()
		for _, c := range completed {
			submits = append(submits, c()...)
		}
		fi.Freeze()
		cmd.Submit(submits...)
		cmd.Wait()
	}
	w.wg.Done()
}

var kRenderPass = vk.NewKey()
var kFrameBuffer = vk.NewKey()
var kCommand = vk.NewKey()

func (w *ViewWindow) getRenderPass(fi *vk.FrameInstance, imView vk.VImageView) (rp *vk.GeneralRenderPass, fb *vk.Framebuffer) {
	rp = fi.GetShared(kRenderPass, func() interface{} {
		desc := fi.Output.Describe()
		return vk.NewGeneralRenderPass(Dev, false, []vk.AttachmentInfo{{Format: desc.Format,
			InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutPresentSrcKhr,
			ClearColor: mgl32.Vec4{0.2, 0.2, 0.2, 1},
		}})
	}).(*vk.GeneralRenderPass)
	fb = fi.Get(kFrameBuffer, func() interface{} {
		return vk.NewFramebuffer2(rp, imView)
	}).(*vk.Framebuffer)
	return
}

func (w *ViewWindow) newCommand(fi *vk.FrameInstance) *vk.Command {
	return fi.Get(kCommand, func() interface{} {
		return vk.NewCommand(Dev, vk.QUEUEGraphicsBit, true)
	}).(*vk.Command)
}

func (w *ViewWindow) handleMyEvent(ev Event) {
	se, ok := ev.(SourcedEvent)
	if !ok || !se.IsSource(w) {
		return
	}
	views := w.Views()
	for idx := len(views) - 1; idx >= 0; idx-- {
		if ev.Handled() {
			return
		}
		ve, ok := views[idx].(EventView)
		if ok {
			ve.HandleEvent(ev)
		}
	}
}
