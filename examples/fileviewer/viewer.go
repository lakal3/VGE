package main

import (
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"os"
)

type FileViewer interface {
	Close()
	Stat(fr *vimgui.UIFrame)
}

func SetViewer(viewer FileViewer, v vapp.View) (remover func()) {
	current.mx.Lock()
	old := current.viewer
	current.viewer = viewer
	current.mx.Unlock()
	if old != nil {
		old.Close()
	}
	if v != nil {
		app.rw.AddView(v)
		return func() {
			app.rw.RemoveView(v)
		}
	}
	return func() {

	}
}

func DrawFileInfo(fr *vimgui.UIFrame, path string, info os.FileInfo) {
	fr.NewLine(80, 20, 5)
	vimgui.Label(fr, "Path:")
	fr.NewColumn(-100, 5)
	vimgui.Label(fr, path)
}

var current struct {
	viewer FileViewer
	mx     *vk.SpinLock
}

func init() {
	current.mx = &vk.SpinLock{}
}

const StatHeight = 100

func addStatView() {
	sv := vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, drawStat)
	sv.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		return vdraw.Area{From: mgl32.Vec2{float32(desc.Width)/4 + 1}, To: mgl32.Vec2{float32(desc.Width), StatHeight}}
	}
	app.rw.AddView(sv)
}

func drawStat(fr *vimgui.UIFrame) {
	current.mx.Lock()
	cv := current.viewer
	current.mx.Unlock()
	if cv == nil {
		fr.NewLine(-100, 20, 5)
		vimgui.Label(fr, "Select a file to show file stats!")
	} else {
		cv.Stat(fr)
	}
	fr.ControlArea = fr.DrawArea
	fr.ControlArea.From[1] = fr.DrawArea.To[1] - 2
	vimgui.Border(fr)
}

type loadView struct {
	path   string
	err    error
	fi     os.FileInfo
	closed bool
}

func (l *loadView) Stat(fr *vimgui.UIFrame) {
	if l.err != nil {
		fr.NewLine(-100, 20, 5)
		fr.WithTags("error")
		vimgui.Label(fr, "Failed to load file: "+l.err.Error())
		return
	}
	DrawFileInfo(fr, l.path, l.fi)
	fr.NewLine(-100, 20, 5)
	vimgui.Label(fr, "Loading....!")
}

func (l *loadView) Load(rm func()) {
	l.fi, l.err = os.Stat(l.path)
	if l.err != nil {
		return
	}
	if l.fi.Size() > 128*1024*1024 {
		l.err = fmt.Errorf("File too large (%d bytes)", l.fi.Size())
		return
	}
	content, err := os.ReadFile(l.path)
	if err != nil {
		l.err = err
		return
	}
	isModel := newModelViewer(l.path, l.fi, content)
	if isModel {
		rm()
		return
	}
	isImage := newImageViewer(l.path, l.fi, content)
	if isImage {
		rm()
		return
	}
	isBinary := parseText(l.path, l.fi, content)
	if !isBinary {
		rm()
	} else {
		l.err = errors.New("File is binary")
	}
}

func (l *loadView) View() vapp.View {
	return nil
}

func (l *loadView) Close() {

}

func viewFile(path string) {
	l := &loadView{path: path}
	rm := SetViewer(l, l.View())
	go l.Load(rm)
}
