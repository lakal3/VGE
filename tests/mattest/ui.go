package main

import (
	"image"

	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
)

func buildUi() {
	app.theme = mintheme.NewTheme(vapp.Dev, 15, nil, nil, nil)
	vapp.AddChild(app.theme)

	var bQuit *vui.Button
	app.ui = vui.NewUIView(app.theme, image.Rect(20, 50, 300, 400), app.rw).
		SetContent(
			vui.NewPanel(10, vui.NewVStack(5,
				vui.NewLabel("Material tester").SetClass("h2"),
				vui.NewLabel("Select test"),
				vui.NewMenuButton("Fire test").SetOnClick(openFireTest),
				vui.NewMenuButton("Decal test").SetOnClick(openDecalTest),
				&vui.Extend{MinSize: image.Pt(10, 10)},
				vui.NewHStack(5,
					vui.NewButton(100, "Quit").SetClass("warning").AssignTo(&bQuit)))))

	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}

	app.rw.Ui.Children = append(app.rw.Ui.Children, vscene.NewNode(app.ui))
	app.ui.Show()
}
