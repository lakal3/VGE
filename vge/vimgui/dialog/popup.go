package dialog

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vk"
)

type Popup struct {
	At      mgl32.Vec2
	Width   float32
	Height  float32
	Title   string
	Kind    string
	win     *vapp.ViewWindow
	view    *vimgui.View
	painter func(popup *Popup, uf *vimgui.UIFrame)
}

func NewPopup(win *vapp.ViewWindow, th *vimgui.Theme, at mgl32.Vec2, width float32, painter func(popup *Popup, uf *vimgui.UIFrame)) *Popup {
	d := &Popup{win: win, At: at, Width: width, Height: 100, painter: painter, Kind: "info"}
	d.view = vimgui.NewView(vapp.Dev, vimgui.VMPopup, th, d.paint)

	d.view.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		return vdraw.Area{From: d.At, To: d.At.Add(mgl32.Vec2{d.Width, d.Height})}
	}
	d.view.OnClose = func() {
		win.RemoveView(d.view)
	}
	win.AddView(d.view)
	return d
}

func (popup *Popup) paint(uf *vimgui.UIFrame) {
	uf.WithTags("dialog", popup.Kind)

	uf.ControlArea = uf.DrawArea
	vimgui.Panel(uf, nil, func(uf *vimgui.UIFrame) {
		uf.WithTags("")
		popup.painter(popup, uf)
		h := uf.ControlArea.To[1] - uf.DrawArea.From[1]
		if h+10 > popup.Height {
			popup.Height = h + 10
		}
	})
}

func (popup *Popup) Close() {
	popup.win.RemoveView(popup.view)
}

func PopupFor(win *vapp.ViewWindow, theme *vimgui.Theme, parent *vimgui.UIFrame, content func(popup *Popup, uf *vimgui.UIFrame)) *Popup {
	at := mgl32.Vec2{parent.ControlArea.From[0], parent.ControlArea.To[1]}
	w := parent.ControlArea.Width()
	return NewPopup(win, theme, at, w, func(popup *Popup, uf *vimgui.UIFrame) {
		content(popup, uf)
	})
}
