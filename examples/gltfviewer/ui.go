package main

import (
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
	"image"
)

func buildUi() error {
	// Initialize application theme, but lets use Oxanium font this time
	ftRaw, err := vapp.AM.Load("fonts/Oxanium-Medium.ttf", func(content []byte) (asset interface{}, err error) {
		fl := &vglyph.VectorSetBuilder{}
		// Add font will load given range of glyph to Vector set build. Vector set builder will rasterize vectors
		// and convert them to signed depth fields
		// You must also specify character unicode range you are interested in. Some Unicode font can contain quite a large number
		// of characters
		fl.AddFont(vapp.Ctx, content, vglyph.Range{From: 33, To: 255})

		// Build glyphset
		return fl.Build(vapp.Ctx, vapp.Dev), nil
	})
	if err != nil {
		return err
	}
	// Pass loaded font as main font
	app.theme = mintheme.NewTheme(vapp.Ctx, vapp.Dev, 15, nil, ftRaw.(*vglyph.GlyphSet), nil)
	vapp.AddChild(app.theme)
	// Build actual UI
	var bQuit, bQuit2, bReturn *vui.Button
	app.extraUi = &vui.Conditional{}
	app.ui = vui.NewUIView(app.theme, image.Rect(20, 50, 300, 700), app.rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("GLTF viewer").SetClass("h2"),
			vui.NewLabel("Select model"),
			&vui.Sizer{Min: image.Pt(250, 500), Max: image.Pt(250, 500), Content: vui.NewScrollViewer(
				vui.NewVStack(1, listSamples()...))},
			&vui.Extend{MinSize: image.Pt(10, 10)},
			vui.NewHStack(5,
				vui.NewButton(100, "Quit").SetClass("warning").AssignTo(&bQuit)))))
	app.ui2 = vui.NewUIView(app.theme, image.Rect(20, 50, 300, 200), app.rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("GLTF viewer").SetClass("h2"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			app.extraUi,
			vui.NewHStack(5,
				vui.NewButton(100, "Return").SetClass("primary").AssignTo(&bReturn),
				vui.NewButton(100, "Quit").SetClass("warning").AssignTo(&bQuit2)))))
	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}
	bQuit2.OnClick = bQuit.OnClick
	bReturn.OnClick = func() {
		app.rw.Scene.Update(func() {
			app.rw.Model.Children = nil
			app.rw.Env.Children = nil
			app.ui2.Hide()
			app.ui.Show()
			app.extraUi.Visible = false
			go app.atEndScene.Dispose()
		})
	}
	app.rw.Ui.Children = append(app.rw.Ui.Children, vscene.NewNode(app.ui), vscene.NewNode(app.ui2))
	app.ui.Show()
	return nil
}

func listSamples() (choices []vui.Control) {
	for _, sample := range SampleList {
		s := sample
		choices = append(choices, vui.NewMenuButton(sample.Name).SetOnClick(func() {
			app.ui.Hide()
			go loadSample(s)
		}))
	}
	return choices
}
