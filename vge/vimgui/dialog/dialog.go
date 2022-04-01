package dialog

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vk"
)

type Dialog struct {
	RelativeWidth float32
	Height        float32
	Title         string
	Kind          string
	win           *vapp.ViewWindow
	view          *vimgui.View
	painter       func(dlg *Dialog, uf *vimgui.UIFrame)
}

func NewDialog(win *vapp.ViewWindow, th *vimgui.Theme, title string, painter func(dlg *Dialog, uf *vimgui.UIFrame)) *Dialog {
	d := &Dialog{win: win, RelativeWidth: 50, Height: 100, Title: title, painter: painter, Kind: "info"}
	d.view = vimgui.NewView(vapp.Dev, vimgui.VMDialog, th, d.paint)
	d.view.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		w := float32(desc.Width) * d.RelativeWidth / 100
		left := (float32(desc.Width) - w) / 2
		top := (float32(desc.Height) - d.Height) / 2.0
		if top < 0 {
			top = 0
		}
		return vdraw.Area{From: mgl32.Vec2{left, top}, To: mgl32.Vec2{left + w, top + d.Height}}
	}
	win.AddView(d.view)
	return d
}

func (d *Dialog) paint(uf *vimgui.UIFrame) {
	uf.WithTags("dialog", d.Kind)

	uf.ControlArea = uf.DrawArea
	vimgui.Panel(uf, func(uf *vimgui.UIFrame) {
		uf.WithTags("h2")
		vimgui.Label(uf, d.Title)
	}, func(uf *vimgui.UIFrame) {
		uf.WithTags("")
		d.painter(d, uf)
		h := uf.ControlArea.To[1] - uf.DrawArea.From[1]
		if h+40 > d.Height {
			d.Height = h + 40
		}
	})
}

func (d *Dialog) Close() {
	d.win.RemoveView(d.view)
}

var kDlgCtrls = vk.NewKeys(10)

func Alert(win *vapp.ViewWindow, th *vimgui.Theme, title, text string, onclose func()) {
	_ = NewDialog(win, th, title, func(dlg *Dialog, uf *vimgui.UIFrame) {
		uf.NewLine(-100, 20, 10)
		vimgui.Text(uf, text)
		w := (uf.ControlArea.Width() - 120) / 2
		uf.NewLine(0, 30, 10)
		uf.NewColumn(120, w)
		if vimgui.Button(uf, kDlgCtrls+0, "OK") {
			dlg.Close()
			if onclose != nil {
				onclose()
			}
		}
		uf.NewLine(100, 5, 0)

	})
}

func Query(win *vapp.ViewWindow, th *vimgui.Theme, title, text string, onclose func(yes bool)) {
	_ = NewDialog(win, th, title, func(dlg *Dialog, uf *vimgui.UIFrame) {
		uf.NewLine(-100, 20, 10)
		vimgui.Text(uf, text)
		w := (uf.ControlArea.Width() - 250) / 2
		uf.NewLine(0, 30, 10)
		uf.NewColumn(120, w)
		if vimgui.Button(uf, kDlgCtrls+0, "Yes") {
			dlg.Close()
			if onclose != nil {
				onclose(true)
			}
		}
		uf.NewColumn(120, 10)
		if vimgui.Button(uf, kDlgCtrls+0, "No") {
			dlg.Close()
			if onclose != nil {
				onclose(false)
			}
		}
		uf.NewLine(100, 5, 0)

	})
}
