package vk

import "sync"

type Desktop struct {
	hDesk   hDesktop
	windows *sync.Map
	app     *Application
}

type Window struct {
	WindowDesc ImageDescription
	hWin       hWindow
	desktop    *Desktop
	images     []*Image
}

// NewDesktop will initialize application with swap chain.
func NewDesktop(app *Application) *Desktop {
	return NewDesktopWithSettings(app, DesktopSettings{ImageUsage: IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit})
}

type DesktopSettings struct {
	ImageUsage ImageUsageFlags
}

// NewDesktopWithSettings will initialize application with swap chain. You can set requested flags for image usage for main image when
// vgelib constructs new swapchain for Windows. Default settings are IMAGEUsageColorAttachmentBit | IMAGEUsageTransferSrcBit
func NewDesktopWithSettings(app *Application, settings DesktopSettings) *Desktop {
	if !app.isValid() {
		return nil
	}
	d := &Desktop{windows: &sync.Map{}, app: app}
	call_NewDesktop(app, app.hApp, settings.ImageUsage, &d.hDesk)
	return d
}

// Pull next desktop event from event queue. If no new event exists, event type will be 0.
func (d *Desktop) PullEvent() (ev RawEvent, w *Window) {
	call_Desktop_PullEvent(d.app, d.hDesk, &ev)
	if ev.hWin != 0 && ev.EventType != 0 {
		rw, ok := d.windows.Load(ev.hWin)
		if ok {
			w = rw.(*Window)
		}
	}
	return
}

// NewWindow create a new desktop window at given location and in size. If location is (left, top) is -1, -1, desktop uses default position algorithm for window
func (d *Desktop) NewWindow(title string, pos WindowPos) *Window {
	w := &Window{desktop: d}
	call_Desktop_CreateWindow(d.app, d.hDesk, []byte(title), &pos, &w.hWin)
	d.addWindow(w)
	return w
}

func (d *Desktop) GetKeyName(keyCode uint32) string {
	var name [256]byte
	var l uint32
	call_Desktop_GetKeyName(d.app, d.hDesk, keyCode, name[:], &l)
	return string(name[:l])
}

func (d *Desktop) isValid() bool {
	if d.hDesk == 0 {
		d.app.setError(ErrDisposed)
		return false
	}
	return true
}

func (d *Desktop) GetMonitor(monitor uint32) (pos WindowPos, exists bool) {
	if !d.isValid() {
		return
	}
	call_Desktop_GetMonitor(d.app, d.hDesk, monitor, &pos)
	exists = pos.Width > 0
	return
}

func (d *Desktop) addWindow(w *Window) {
	d.windows.Store(w.hWin, w)
}

func (d *Desktop) removeWindow(w *Window) {
	d.windows.Delete(w.hWin)
}

func (d *Desktop) GetClipboard() string {
	var buf []byte
	var cpLen uint64
	call_Desktop_GetClipboard(d.app, d.hDesk, &cpLen, buf)
	if cpLen > 0 {
		buf = make([]byte, cpLen)
		// call_Window_GetClipboard(w.desktop.app, w.hWin, &cpLen, buf)
		return string(buf[:])
	}
	return ""
}

func (d *Desktop) SetClipboard(newContent string) {
	var bytes = []byte(newContent)
	call_Desktop_SetClipboard(d.app, d.hDesk, bytes)
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

func (w *Window) GetPos() WindowPos {
	var pos WindowPos
	call_Window_GetPos(w.desktop.app, w.hWin, &pos)
	return pos
}

func (w *Window) SetPos(pos WindowPos) {
	call_Window_SetPos(w.desktop.app, w.hWin, &pos)
}

func (w *Window) GetClipboard() string {
	return w.desktop.GetClipboard()
}

func (w *Window) SetClipboard(newContent string) {
	w.desktop.SetClipboard(newContent)
}

func (w *Window) GetNextFrame(dev *Device) (im *Image, imageIndex int32, info SubmitInfo) {
	if len(w.images) == 0 {
		w.prepareSwapchain(dev)
		return nil, -1, SubmitInfo{}
	}
	var img hImage
	var hInfo hSubmitInfo

	call_Window_GetNextFrame(dev, w.hWin, &img, &hInfo, &imageIndex)
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

func (w *Window) prepareSwapchain(dev *Device) {
	var imageCount int32
	call_Window_PrepareSwapchain(dev, w.hWin, dev.hDev, &w.WindowDesc, &imageCount)
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
