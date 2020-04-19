package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/vanimation"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"image"
	"log"
)

type Sample struct {
	Name     string
	BasePath string
}

// List of test animations
var SampleList = []*Sample{
	&Sample{Name: "Reset to pose"},
	&Sample{Name: "Simple test animation", BasePath: "bvh/tests/simple1.bvh"},
	&Sample{Name: "Wave", BasePath: "bvh/tests/wave.bvh"},
	&Sample{Name: "Jumping", BasePath: "bvh/tests/test1.bvh"},
	&Sample{Name: "Boxing", BasePath: "bvh/tests/boxing.bvh"},
	&Sample{Name: "Dance", BasePath: "bvh/tests/dance1.bvh"},
	&Sample{Name: "Salsa", BasePath: "bvh/tests/salsa.bvh"},
	&Sample{Name: "Hand", BasePath: "bvh/tests/hand2.bvh"},
	&Sample{Name: "Pick a fruit", BasePath: "bvh/tests/pickfruit.bvh"},
}

func buildSelector() {
	var bQuit *vui.Button
	app.selectorUI = vui.NewUIView(app.theme, image.Rect(20, 50, 300, 700), app.mainWnd).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("GLTF viewer").SetClass("h2"),
			vui.NewCheckbox("Pause", "").SetOnChanged(func(checked bool) {
				setPaused(checked)
			}).SetClass("dark"),
			vui.NewLabel("Select animation"),
			&vui.Sizer{Min: image.Pt(250, 500), Max: image.Pt(250, 500), Content: vui.NewScrollViewer(
				vui.NewVStack(1, listSamples()...))},
			&vui.Extend{MinSize: image.Pt(10, 10)},
			vui.NewHStack(5,
				vui.NewButton(100, "Quit").SetClass("warning").AssignTo(&bQuit)))))
	bQuit.OnClick = func() {
		// Shutdown application
		go vapp.Terminate()
	}
	app.mainWnd.Ui.Children = append(app.mainWnd.Ui.Children, vscene.NewNode(app.selectorUI))
	app.selectorUI.Show()
}

func setPaused(checked bool) {
	app.mainWnd.SetPaused(checked)
}

func buildScene() {
	rw := app.mainWnd
	app.actor = vscene.NodeFromModel(app.actorModel, 0, true)
	// Model sizes don't match so we scale them so that they fit into house and place them on mat
	tr := mgl32.Scale3D(0.07, 0.07, 0.07).Mul4(mgl32.Translate3D(0, 4, -15))
	switch app.elf {
	case 0:
		tr = mgl32.Scale3D(0.1, 0.1, 0.1).Mul4(mgl32.Translate3D(0, 2.6, -10))
	case 1:
		tr = mgl32.Scale3D(0.04, 0.04, 0.04).Mul4(mgl32.Translate3D(0, 8, -10))
	}
	app.actor.Ctrl = &vscene.TransformControl{Transform: tr}
	houseNode := vscene.NodeFromModel(app.houseModel, 0, true)

	// Add two shadow casting light on other end of room
	pl1 := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1.5, 1.5, 1.5}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512)
	pl1.MaxDistance = 10
	pl2 := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1.5, 1.5, 1.5}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512)
	pl2.MaxDistance = 10
	lights := vscene.NewNode(nil,
		vscene.NewNodeAt(mgl32.Translate3D(-1.5, 2.5, 2), pl1),
		vscene.NewNodeAt(mgl32.Translate3D(1.5, 2.5, 2), pl2))

	// Build actual scene
	rw.Env.Children = append(rw.Env.Children, vscene.NewNode(app.bg), houseNode, lights)
	rw.Model.Ctrl = app.probe
	rw.Model.Children = append(rw.Model.Children, app.actor)

	buildSelector()

	// Camera
	pc := vscene.NewPerspectiveCamera(10)
	pc.Position = mgl32.Vec3{0.02, 1.5, 3}
	pc.Target = mgl32.Vec3{0, 1.2, 0}
	app.mainWnd.Camera = pc

	// Orbit camera should have b
	oc := vapp.OrbitControlFrom(-100, app.mainWnd, pc)

	// Add clamp limits to camera, preventing it from running out of scene
	oc.Clamp = checkLimits
}

func checkLimits(point mgl32.Vec3, target bool) mgl32.Vec3 {
	if point[0] < -2 {
		point[0] = -2
	}
	if point[0] > 2 {
		point[0] = 2
	}
	if point[1] < 0.2 {
		point[1] = 0.2
	}
	if point[1] > 2.5 {
		point[1] = 2.5
	}
	if point[2] < -3 {
		point[2] = -3
	}
	if point[2] > 3.5 {
		point[2] = 3.5
	}
	return point
}

func listSamples() (choices []vui.Control) {
	for _, sample := range SampleList {
		s := sample
		choices = append(choices, vui.NewMenuButton(sample.Name).SetOnClick(func() {
			go loadAnimation(s)
		}))
	}
	return choices
}

func loadAnimation(s *Sample) {
	if len(s.BasePath) == 0 {
		// If base path is empty, reset animation
		app.mainWnd.Scene.Update(func() {
			for _, ch := range app.actor.Children {
				resetAnimation(ch)
			}
		})
		return
	}
	// First load animation
	bvh, err := vanimation.LoadBVH(vasset.DefaultLoader, s.BasePath)
	if err != nil {
		log.Fatal("Can't load bvh ", s.BasePath, ": ", err)
	}
	// Then update scene safely
	app.mainWnd.Scene.Update(func() {
		for _, ch := range app.actor.Children {
			updateAnimation(ch, bvh)
		}
	})
}

func resetAnimation(node *vscene.Node) {
	an, ok := node.Ctrl.(*vscene.AnimatedNodeControl)
	if ok {
		// Reset will copy initial rotate state /and translate for root bone) to current joints.
		// Animation is set to empty (no channels)
		an.Animation = vmodel.Animation{}
		for jIdx, j := range an.Skin.Joints {
			an.Joints[jIdx].Rotate = j.Rotate
			if jIdx == 0 {
				an.Joints[jIdx].Translate = j.Translate
			}
		}
		an.StartTime = 0
	}
	for _, ch := range node.Children {
		// Apply to child nodes also
		resetAnimation(ch)
	}
}

func updateAnimation(node *vscene.Node, bvh *vanimation.BVHAnimation) {
	an, ok := node.Ctrl.(*vscene.AnimatedNodeControl)
	if ok {
		// Set properer bone map. Original elf model have different bone names
		bMap := mixamoBones
		if app.elf == 1 {
			bMap = elfBones
		}
		// Use loaded bvh animation to build animation matched to current skin
		an.Animation = bvh.BuildAnimation(an.Skin, bMap)
		an.StartTime = 0
	}
	for _, ch := range node.Children {
		updateAnimation(ch, bvh)
	}
}

var boneMap1 = map[string]string{
	"Hips":          "bacino",
	"Spine":         "addome",
	"Spine1":        "schienasu",
	"RightUpLeg":    "coscia.L",
	"RightLeg":      "polpaccio.L",
	"LeftUpLeg":     "coscia.R",
	"LeftLeg":       "polpaccio.R",
	"RightShoulder": "clavicola.L",
	"RightArm":      "bracciosu.L",
	"RightForeArm":  "bracciogiu.L",
	"RightHand":     "mano.L",
	"LeftShoulder":  "clavicola.R",
	"LeftArm":       "bracciosu.R",
	"LeftForeArm":   "bracciogiu.R",
	"LeftHand":      "mano.R",
	"Neck":          "collogiu",
}

var boneMap2 = map[string]string{
	"Hips":          "pelvis",
	"Spine":         "spine_1",
	"Spine1":        "spine_2",
	"RightUpLeg":    "thigh_r",
	"RightLeg":      "calf_r",
	"RightToeBase":  "foot_r",
	"LeftUpLeg":     "thigh_l",
	"LeftLeg":       "calf_l",
	"LeftToeBase":   "foot_l",
	"RightShoulder": "clavicle_r",
	"RightArm":      "upperarm_r",
	"RightForeArm":  "lowerarm_r",
	"RightHand":     "hand_r",
	"LeftShoulder":  "clavicle_l",
	"LeftArm":       "upperarm_l",
	"LeftForeArm":   "lowerarm_l",
	"LeftHand":      "hand_l",
	"Neck":          "neck_01",
}

var boneMap map[string]string

func mixamoBones(sk *vmodel.Skin, name string) (jIndex int) {
	en := "mixamorig:" + name
	for idx, j := range sk.Joints {
		if j.Name == en {
			return idx
		}
	}
	return -1
}

func elfBones(sk *vmodel.Skin, name string) (jIndex int) {
	en, ok := boneMap1[name]
	if !ok {
		return -1
	}
	for idx, j := range sk.Joints {
		if j.Name == en {
			return idx
		}
	}
	return -1
}
