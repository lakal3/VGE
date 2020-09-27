package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vapp/vdebug"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"image"
)

type logoWindow struct {
	env         *env.EquiRectBGNode
	probe       *env.Probe
	sl          *shadow.PointLight
	bRun        *vui.Button
	lStat       *vui.Label
	nLogoParent *vscene.Node
	ui          *vui.UIView
	rSize       *vui.RadioGroup
}

func (w *logoWindow) Dispose() {
	if w.env != nil {
		w.env.Dispose()
		w.probe.Dispose()
		w.env = nil
	}
}

type shrinkAmim struct {
	started float64
	m       *maze
	done    bool
}

func (s *shrinkAmim) Process(pi *vscene.ProcessInfo) {
	if s.done {
		pi.Visible = false
		return
	}
	t := pi.Time - s.started
	sf := float32(1 - t)
	pi.World = pi.World.Mul4(mgl32.Scale3D(sf, sf, sf))
	if sf < 0.5 && s.m != nil {
		s.done = true
		s.m.switchScene()
	}
}

func (w *logoWindow) startRun() {
	sa := &shrinkAmim{started: app.mainWnd.GetSceneTime()}
	size := 4
	switch w.rSize.Value {
	case 1:
		size = 8
	case 2:
		size = 12
	}
	go buildMaze(sa, size)
	app.mainWnd.Scene.Update(func() {
		w.ui.Hide()
		app.mainWnd.Ui.Children = nil
		w.nLogoParent.Ctrl = sa

	})
	if app.fps {
		fpsDebug := vdebug.NewFPSTimer(app.mainWnd, app.theme)
		fpsDebug.AddGPUTiming()
	}
}

func buildStartScene() *logoWindow {
	lw := &logoWindow{}
	rw := app.mainWnd
	// Load envhdr/studio.hdr
	lw.env = vapp.MustLoadAsset("envhdr/studio.hdr",
		func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
	// Add loaded background to scene
	rw.Env.Children = append(rw.Env.Children, vscene.NewNode(lw.env))

	nLogo := vscene.NodeFromModel(app.logoModel, 0, true)
	// We will also need a probe to reflect environment to model. Probes reflect everything outside this node inside children of this node.
	// In this case we reflect only background
	lw.probe = env.NewProbe(vapp.Ctx, vapp.Dev)
	// Create a new nodes from model and probe
	lw.nLogoParent = vscene.NewNode(nil, nLogo)
	rw.Model.Children = append(rw.Model.Children, lw.nLogoParent)
	// Assign probe to root model
	rw.Model.Ctrl = lw.probe

	// Attach camera to window (with better location that default one) and orbital control to camera
	c := vscene.NewPerspectiveCamera(1000)
	c.Position = mgl32.Vec3{1, 2, 10}
	c.Target = mgl32.Vec3{5, 0, 0}
	rw.Camera = c

	// Finally create 2 lights before UI
	// Create custom node control to turn light on / off
	nLight := vscene.NewNode(nil)
	rw.Env.Children = append(rw.Env.Children, nLight)
	// Add light
	lw.sl = shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1.4, 1.4, 1.4}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512)

	// Add shadow light to scene on location 1,3,3 and 4,3,3
	nLight.Children = append(nLight.Children,
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(-3, 3, 3)}, vscene.NewNode(lw.sl)),
	)

	var bQuit, bRun *vui.Button
	var rs1, rs2, rs3, cbOrbit *vui.ToggleButton
	ui := vui.NewUIView(app.theme, image.Rect(100, 450, 500, 700), rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("Robot maze demo").SetClass("h2"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewLabel("Maze size"),
			vui.NewRadioButton("Small (4 x 4)").AssignTo(&rs1),
			vui.NewRadioButton("8 x 8").AssignTo(&rs2),
			vui.NewRadioButton("12 x 12").AssignTo(&rs3),
			vui.NewCheckbox("Orbit camera", "").AssignTo(&cbOrbit),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewHStack(10,
				vui.NewButton(100, "Run").SetClass("primary").AssignTo(&bRun),
				vui.NewButton(100, "Quit").SetClass("warning").AssignTo(&bQuit),
			),
			vui.NewLabel("Loading, please wait!").AssignTo(&lw.lStat),
		)).SetClass(""))
	lw.ui = ui
	bRun.Disabled = true
	lw.bRun = bRun
	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}
	cbOrbit.OnChanged = func(checked bool) {
		app.orbitCamera = checked
	}
	lw.rSize = vui.NewRadioGroup(rs1, rs2, rs3)
	// Attach UI to scene and show it. UI panel are by default invisible and must be show
	rw.Ui.Children = append(rw.Ui.Children, vscene.NewNode(ui))
	ui.Show()
	bRun.OnClick = lw.startRun
	return lw
}
