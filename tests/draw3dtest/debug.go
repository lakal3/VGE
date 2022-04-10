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
var debugType int

var debugKeys = vk.NewKeys(10)

func paintDebug(fr *vimgui.UIFrame) {
	fr.ControlArea = fr.DrawArea
	fr.WithTags("dialog")

	fr.PushArea(vdraw.Area{From: fr.DrawArea.From.Add(mgl32.Vec2{100, 50}), To: fr.DrawArea.To.Sub(mgl32.Vec2{100, 50})}, true)
	fr.ControlArea = fr.DrawArea
	vimgui.Panel(fr, func(uf *vimgui.UIFrame) {
		fr.WithTags("")
		fr.NewLine(-100, 30, 0)
		vimgui.Label(fr.WithTags("h2", "primary"), "Debug settings")
		fr.Tags = nil
	}, func(uf *vimgui.UIFrame) {
		fr.NewLine(120, 25, 5)
		if vimgui.RadioButton(fr, debugKeys, "Full color", 0, &debugType) {
			app.nv.SetDebug(uint32(debugType))
			app.rw.RemoveView(vDialog)
		}
		fr.NewColumn(120, 10)
		if vimgui.RadioButton(fr, debugKeys+1, "Albedo", 1, &debugType) {
			app.nv.SetDebug(uint32(debugType))
			app.rw.RemoveView(vDialog)
		}
		fr.NewColumn(120, 10)
		if vimgui.RadioButton(fr, debugKeys+2, "Normal", 3, &debugType) {
			app.nv.SetDebug(uint32(debugType))
			app.rw.RemoveView(vDialog)
		}
		fr.NewLine(-100, 5, 2)
		vimgui.Border(fr)
		fr.NewLine(120, 25, 5)
		if vimgui.RadioButton(fr, debugKeys+3, "Direct light", 11, &debugType) {
			app.nv.SetDebug(uint32(debugType))
			app.rw.RemoveView(vDialog)
		}
		fr.NewColumn(120, 10)
		if vimgui.RadioButton(fr, debugKeys+4, "Indirect light", 12, &debugType) {
			app.nv.SetDebug(uint32(debugType))
			app.rw.RemoveView(vDialog)
		}
		fr.NewLine(-100, 5, 2)
		vimgui.Border(fr)
		fr.NewLine(120, 25, 3)
		fr.NewColumn(120, 5)
		if vimgui.Button(fr, debugKeys+5, "Close") {
			app.rw.RemoveView(vDialog)
		}
	})
	fr.Pop()
}

func showDebug() {
	vDialog = vimgui.NewView(vapp.Dev, vimgui.VMDialog, mintheme.Theme, paintDebug)
	app.rw.AddView(vDialog)
}
