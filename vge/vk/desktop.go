package vk

import "sync"

type Desktop struct {
	hDesk   hDesktop
	windows *sync.Map
}

type Window struct {
	WindowDesc ImageDescription
	hWin       hWindow
	desktop    *Desktop
	images     []*Image
}

// NewDesktop will initialize application with swap chain.
func NewDesktop(ctx APIContext, app *Application) *Desktop {
	return NewDesktopWithSettings(ctx, app, DesktopSettings{ImageUsage: IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit})
}

type DesktopSettings struct {
	ImageUsage ImageUsageFlags
}

// NewDesktopWithSettings will initialize application with swap chain. You can set requested flags for image usage for main image when
// vgelib constructs new swapchain for Windows. Default settings are IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit
func NewDesktopWithSettings(ctx APIContext, app *Application, settings DesktopSettings) *Desktop {
	if app.hInst != 0 {
		ctx.SetError(ErrInitialized)
	}
	d := &Desktop{windows: &sync.Map{}}
	call_NewDesktop(ctx, app.hApp, settings.ImageUsage, &d.hDesk)
	return d
}

// Pull next desktop event from event queue. If no new event exists, event type will be 0.
func (d *Desktop) PullEvent(ctx APIContext) (ev RawEvent, w *Window) {
	call_Desktop_PullEvent(ctx, d.hDesk, &ev)
	if ev.hWin != 0 && ev.EventType != 0 {
		rw, ok := d.windows.Load(ev.hWin)
		if ok {
			w = rw.(*Window)
		}
	}
	return
}

// NewWindow create a new desktop window at given location and in size. If location is (left, top) is -1, -1, desktop uses default position algorithm for window
func (d *Desktop) NewWindow(ctx APIContext, title string, pos WindowPos) *Window {
	w := &Window{desktop: d}
	call_Desktop_CreateWindow(ctx, d.hDesk, []byte(title), &pos, &w.hWin)
	d.addWindow(w)
	return w
}

func (d *Desktop) GetKeyName(ctx APIContext, keyCode uint32) string {
	var name [256]byte
	var l uint32
	call_Desktop_GetKeyName(ctx, d.hDesk, keyCode, name[:], &l)
	return string(name[:l])
}

func (d *Desktop) IsValid(ctx APIContext) bool {
	if d.hDesk == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

func (d *Desktop) GetMonitor(ctx APIContext, monitor uint32) (pos WindowPos, exists bool) {
	if !d.IsValid(ctx) {
		return
	}
	call_Desktop_GetMonitor(ctx, d.hDesk, monitor, &pos)
	exists = pos.Width > 0
	return
}

func (d *Desktop) addWindow(w *Window) {
	d.windows.Store(w.hWin, w)
}

func (d *Desktop) removeWindow(w *Window) {
	d.windows.Delete(w.hWin)
}

func (w *Window) Dispose() {
	if w.hWin != 0 {
		w.desktop.removeWindow(w)
		call_Disposable_Dispose(hDisposable(w.hWin))
		w.hWin = 0
	}
}

func (w *Window) GetImageCount() int {
	return len(w.images)
}

func (w *Window) GetPos(ctx APIContext) WindowPos {
	var pos WindowPos
	call_Window_GetPos(ctx, w.hWin, &pos)
	return pos
}

func (w *Window) SetPos(ctx APIContext, pos WindowPos) {
	call_Window_SetPos(ctx, w.hWin, &pos)
}

func (w *Window) GetNextFrame(ctx APIContext, dev *Device) (im *Image, imageIndex int32, info SubmitInfo) {
	if len(w.images) == 0 {
		w.prepareSwapchain(ctx, dev)
		return nil, -1, SubmitInfo{}
	}
	var img hImage
	var hInfo hSubmitInfo

	call_Window_GetNextFrame(ctx, w.hWin, &img, &hInfo, &imageIndex)
	if imageIndex >= 0 {
		if w.images[imageIndex] == nil {
			w.images[imageIndex] = &Image{hImage: img, swapbuffer: true, allocated: true, dev: dev,
				Description: w.WindowDesc, Usage: IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit}
		}
	} else {
		w.resetSwapchain()
		return nil, -1, SubmitInfo{}
	}
	return w.images[imageIndex], imageIndex, SubmitInfo{info: hInfo}
}

func (w *Window) prepareSwapchain(ctx APIContext, dev *Device) {
	var imageCount int32
	call_Window_PrepareSwapchain(ctx, w.hWin, dev.hDev, &w.WindowDesc, &imageCount)
	if imageCount > 0 {
		w.images = make([]*Image, imageCount)
	}
}

func (w *Window) resetSwapchain() {
	for _, im := range w.images {
		if im != nil {
			im.Dispose()
		}
	}
	w.images = nil
}
