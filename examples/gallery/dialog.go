package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
)

var vDialog *vimgui.View

func newDialog() {
	vDialog = vimgui.NewView(vapp.Dev, vimgui.VMDialog, mintheme.Theme, paintDialog)
	vDialog.OnSize = func(fullArea vdraw.Area) vdraw.Area {
		var a vdraw.Area
		a.From = mgl32.Vec2{fullArea.Width()/2 - 250, 200}
		a.To = mgl32.Vec2{a.From[0] + 500, a.From[1] + 200}
		return a
	}
	app.rw.AddView(vDialog)
}

var kbtnClose = vk.NewHashKey("btnclose")

func paintDialog(fr *vimgui.UIFrame) {
	fr.ControlArea = fr.DrawArea
	vimgui.Panel(fr.WithTags("dialog"), func(uf *vimgui.UIFrame) {
		vimgui.Label(fr.WithTags("h2", "primary"), "Dialog")
	}, func(uf *vimgui.UIFrame) {
		fr.NewLine(-100, 20, 0)
		vimgui.Label(fr.WithTags(""), "Click close to close dialog")
		fr.NewLine(-35, 30, 5)
		fr.NewColumn(120, 10)
		if vimgui.Button(fr, kbtnClose, "Close dialog") {
			app.rw.RemoveView(vDialog)
		}
	})
}
