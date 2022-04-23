package vapp

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"sync"
	"time"
)

type Completed func() []vk.SubmitInfo

type View interface {
	// Reserve is called at start of frame. View must reserve all resources it needs to render a frame
	Reserve(fi *vk.FrameInstance)

	// PreRender is called after reservation. This allows view to prepare final assets that are need to render actual view output
	// For example 3D views will use several passes to render final 3D image into a separate image. This image is bitblitted to main image on Render phase
	// Done function, if not nil, will be called before final Render phase starts. Done functions can add one or more submit infos will be waited
	// before final render phase is started
	PreRender(fi *vk.FrameInstance) (done Completed)

	// Render is final phase of view where is should copy views output to main image. RenderPass will contain one output (main image)
	// Some views (like UI) can completed all rendering in this phase. Some more complex views must use PreRender to prepare final image
	Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass)

	// PostRender is called after rendering is completed (in GPU side)
	PostRender(fi *vk.FrameInstance)
}

// EventView is view that also handles mouse/keyboard events targeted to this window
type EventView interface {
	View

	// HandleEvent is called to handle mouse/keyboard events targeted to this window
	HandleEvent(event Event)
}

// ViewWindow is new rendering window that can contain several views
type ViewWindow struct {
	OnClose func()

	eventState UIState
	fc         *vk.FrameCache
	views      []View
	sp         *vk.SpinLock
	win        *vk.Window
	wg         *sync.WaitGroup
	timed      func(started time.Time, gpuTimes []float64)
	state      int
}

// SetTimedOutput sets funtion to time rendering of views
func (w *ViewWindow) SetTimedOutput(output func(started time.Time, gpuTimes []float64)) {
	w.timed = output
}

// NewViewWindow creates new view windows and displays it
func NewViewWindow(title string, wp vk.WindowPos, views ...View) *ViewWindow {
	w := &ViewWindow{sp: &vk.SpinLock{}}
	w.win = appStatic.desktop.NewWindow(title, wp)
	RegisterHandler(PRIWindow, w.eventHandler)
	winCount++
	go w.renderLoop()
	return w
}

// Dispose ViewWindows
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

// Views retrieve list of active views. This method is thread safe
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

// SetViews sets list of active views. This method is thread safe
func (w *ViewWindow) SetViews(views ...View) {
	w.sp.Lock()
	w.views = views
	w.sp.Unlock()
}

// AddView adds new topmost view to window. This method is thread safe
func (w *ViewWindow) AddView(view View) {
	w.sp.Lock()
	w.views = append(w.views, view)
	w.sp.Unlock()
}

// RemoveView removes one view from window. This method is thread safe
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

var kTimer = vk.NewKeys(2)

func (w *ViewWindow) renderLoop() {
	w.wg = &sync.WaitGroup{}
	w.wg.Add(1)
	w.state = 1
	for w.state == 1 {
		start := time.Now()
		im, imageIndex, submitInfo := w.win.GetNextFrame(Dev)
		if imageIndex < 0 {
			continue
		}
		if w.fc == nil {
			w.fc = vk.NewFrameCache(Dev, w.win.GetImageCount())
		}
		tp := w.timed
		fi := w.fc.Instances[imageIndex]
		fi.Output = im
		fi.BeginFrame()
		views := w.Views()
		submits := []vk.SubmitInfo{submitInfo}
		var completed []Completed
		for _, v := range views {
			v.Reserve(fi)
		}
		fi.Commit()
		var timer *vk.TimerPool
		if tp != nil {
			timer = vk.NewTimerPool(Dev, 3)
			timeCmd := vk.NewCommand(Dev, vk.QUEUEComputeBit, true)
			timeCmd.Begin()
			timeCmd.WriteTimer(timer, 0, vk.PIPELINEStageTopOfPipeBit)
			fi.Set(kTimer, timer)
			fi.Set(kTimer+1, timeCmd)
			submits = append(submits, timeCmd.SubmitForWait(0, vk.PIPELINEStageTopOfPipeBit))
		}
		for _, v := range views {
			c := v.PreRender(fi)
			if c != nil {
				completed = append(completed, c)
			}
		}
		rp, fb := w.getRenderPass(fi, im.DefaultView())
		cmd := w.newCommand(fi)
		cmd.Begin()
		if tp != nil {
			cmd.WriteTimer(timer, 1, vk.PIPELINEStageTopOfPipeBit)
		}
		cmd.BeginRenderPass(rp, fb)
		for _, v := range views {
			v.Render(fi, cmd, rp)
		}
		cmd.EndRenderPass()
		for _, c := range completed {
			submits = append(submits, c()...)
		}
		fi.Freeze()
		if tp != nil {
			cmd.WriteTimer(timer, 2, vk.PIPELINEStageAllCommandsBit)
		}
		cmd.Submit(submits...)
		cmd.Wait()
		for _, v := range views {
			v.PostRender(fi)
		}

		if tp != nil {
			tp(start, timer.Get())
		}
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
