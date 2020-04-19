package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/noise"
	"github.com/lakal3/vge/vge/materials/pbr"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

type Sample struct {
	Name     string
	BasePath string
	Target   mgl32.Vec3
	Position mgl32.Vec3
	Lights   int
	Scene    int
	Env      int
	// High of model base, need for floor
	Base float32
}

var SampleList = []*Sample{
	&Sample{Name: "Metal rouhgness test", BasePath: "2.0/MetalRoughSpheres/glTF/MetalRoughSpheres.gltf",
		Position: mgl32.Vec3{2, 3, 10}, Lights: 1, Env: 1},
	&Sample{Name: "Metal rouhgness (studio)", BasePath: "2.0/MetalRoughSpheres/glTF/MetalRoughSpheres.gltf",
		Position: mgl32.Vec3{2, 3, 10}, Lights: 1, Env: 0},
	&Sample{Name: "Antique camera ", BasePath: "2.0/AntiqueCamera/glTF/AntiqueCamera.gltf",
		Position: mgl32.Vec3{2, 10, 12}, Target: mgl32.Vec3{0, 10, 0}, Lights: 2, Env: 0, Scene: 1, Base: 0},
	&Sample{Name: "Boom box", BasePath: "2.0/BoomBoxWithAxes/glTF/BoomBoxWithAxes.gltf",
		Position: mgl32.Vec3{0.02, 0.03, 0.2}, Target: mgl32.Vec3{0, 0, 0}, Lights: 0, Env: 0, Scene: 1, Base: -0.1},
	&Sample{Name: "DamagedHelmet", BasePath: "2.0/DamagedHelmet/glTF/DamagedHelmet.gltf",
		Position: mgl32.Vec3{0.2, 1, 3}, Target: mgl32.Vec3{0, 1.5, 0}, Lights: 2, Env: 1, Scene: 0, Base: 0},
	&Sample{Name: "Lantern", BasePath: "2.0/Lantern/glTF/Lantern.gltf",
		Position: mgl32.Vec3{5, 40, 30}, Target: mgl32.Vec3{0, 25, 0}, Lights: 2, Env: 0, Scene: 1, Base: 0},
	&Sample{Name: "Sponza", BasePath: "2.0/Sponza/glTF/Sponza.gltf",
		Position: mgl32.Vec3{-7, 1, -0.25}, Target: mgl32.Vec3{0, 1, -0.25}, Lights: 4, Env: 2, Scene: 2, Base: 0},
	&Sample{Name: "Fox", BasePath: "2.0/Fox/glTF/Fox.gltf",
		Position: mgl32.Vec3{6, 30, 200}, Target: mgl32.Vec3{0, 50, 0}, Lights: 2, Env: 0, Scene: 1, Base: -0.1},
	&Sample{Name: "BrainStem", BasePath: "2.0/BrainStem/glTF/BrainStem.gltf",
		Position: mgl32.Vec3{0.2, 1, 4}, Target: mgl32.Vec3{0, 1.5, 0}, Lights: 2, Env: 0, Scene: 3, Base: -0.1},
}

func loadSample(sample *Sample) {
	cl := customLoader{dir: filepath.Join(app.gltfRoot, filepath.Dir(sample.BasePath))}
	// User directly gltf loader
	// First we will need a model builder. All model loader use model builder to load model geometry and images
	mb := &vmodel.ModelBuilder{}

	// Then we need a factory to build material. pbr factory suits find with PBR model. Of cause we can make own
	// factory that can use multiple shaders based on material info from model
	mb.ShaderFactory = pbr.PbrFactory

	// Generate mipmaps for sponza
	if sample.Scene == 2 {
		mb.MipLevels = 6
	}
	gl := gltf2loader.GLTF2Loader{Builder: mb, Loader: cl}
	err := gl.LoadGltf(filepath.Base(sample.BasePath))
	if err != nil {
		log.Fatal("Failed to load gltf file, ", sample.BasePath, ": ", err)
	}
	// GLTF can have multiple scenes. We must choose one to convert. You can convert scenes to single models
	// We just pick first
	err = gl.Convert(0)
	if err != nil {
		log.Fatal("Failed to build model, ", sample.BasePath, ": ", err)
	}
	// No we have model builder ready
	m := mb.ToModel(vapp.Ctx, vapp.Dev)
	// Let's reqister model so we remember to dispose it when we close scene
	app.atEndScene.AddChild(m)

	// Make a scene node (and most likely subnodes) from model. Node 0 always root node of a model.
	// See robomaze demo on how to import individual nodes or node sets from model
	nModel := vscene.NodeFromModel(m, 0, true)

	buildScene(sample, nModel)
}

// Build final scene
func buildScene(sample *Sample, model *vscene.Node) {
	// Create
	if app.probe == nil {
		// We need probe to sample environment image
		app.probe = env.NewProbe(vapp.Ctx, vapp.Dev)
		vapp.AddChild(app.probe)
	} else {
		app.probe.Update()
	}
	// Load environment
	var bg *env.EquiRectBGNode
	nLights := vscene.NewNode(nil)
	// Put model inside modelroot so it won't effect rendering probe
	switch sample.Env {
	case 0:
		// We don't have to dispose loaded assets. This function will handle it. Also we don't have to recall already
		// loaded images. If we load same asset twice, only one instance of it is actually created
		// First parameter kind allows us to use same image for different purposes as this call will recall constructed
		// object EquiRectBGNode
		bg = vapp.MustLoadAsset("envhdr/studio.hdr", func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
	case 1:
		bg = vapp.MustLoadAsset("envhdr/kloofendal_48d_partly_cloudy_2k.hdr", func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
	case 2:
		bg = vapp.MustLoadAsset("envhdr/moonless_golf_1k.hdr", func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
		// assets/envhdr/moonless_golf_1k.hdr
	}

	// Create default camera and camera controls
	far := sample.Position.Sub(sample.Target).Len() * 10
	pc := vscene.NewPerspectiveCamera(far)
	pc.Position = sample.Position

	var vc *vapp.WalkControl
	if sample.Scene == 2 {
		vc = vapp.WalkControlFrom(-100, app.rw, pc)
	} else {
		vapp.OrbitControlFrom(-100, app.rw, pc)
	}
	nModelRool := vscene.NewNode(app.probe, model)
	switch sample.Scene {
	case 1: // Sample floor
		nModelRool.Children = append(nModelRool.Children, vscene.NewNode(floorPos(sample),
			vscene.NodeFromModel(app.tools, app.tools.FindNode("Dullbg"), true)))
	case 3: // Sample floor
		model = vscene.NewNode(&vscene.TransformControl{Transform: mgl32.HomogRotate3DX(math.Pi / 2)}, model)
		nModelRool.Children = []*vscene.Node{model, vscene.NewNode(floorPos(sample),
			vscene.NodeFromModel(app.tools, app.tools.FindNode("Dullbg"), true))}
	}

	switch sample.Lights {
	case 1:
		// One directional light from ~top
		nLights.Children = append(nLights.Children, vscene.NewNode(
			&vscene.DirectionalLight{Intensity: mgl32.Vec3{0.7, 0.7, 0.7}, Direction: mgl32.Vec3{0.1, -1, -0.2}.Normalize()}))
	case 2: // Two rotating point lights with shadows
		nLights.Children = append(nLights.Children, newRotLight(sample, 0.7, mgl32.Vec3{1, 1, 1}),
			newRotLight(sample, 1.2, mgl32.Vec3{1, 1, 1}))
	case 3: // Two rotating point lights with shadows
		nLights.Children = append(nLights.Children, newRotLight(sample, 0.7, mgl32.Vec3{3, 3, 3}),
			newRotLight(sample, 1.2, mgl32.Vec3{3, 3, 3}))
	case 4:
		nLights.Children = append(nLights.Children,
			newPlacedLight(sample, mgl32.Vec3{-5, 1.14, 1.05}),
			newPlacedLight(sample, mgl32.Vec3{-5, 1.14, -1.7}),
			newPlacedLight(sample, mgl32.Vec3{3.85, 1.14, 1.05}),
			newPlacedLight(sample, mgl32.Vec3{3.85, 1.14, -1.7}),
			newPlacedLight(sample, mgl32.Vec3{-5.5, 5.8, -2.1}),
			newPlacedLight(sample, mgl32.Vec3{-4.2, 5.8, -2.1}),
			newPlacedLight(sample, mgl32.Vec3{1.8, 5.8, -2.1}),
			newPlacedLight(sample, mgl32.Vec3{-5.5, 5.8, 1.6}),
		)
	}
	checkAnim(nModelRool)
	app.rw.Scene.Update(func() {
		app.rw.Camera = pc
		app.rw.Env.Children = []*vscene.Node{vscene.NewNode(bg), nLights}
		if vc != nil {
			app.rw.Env.Children = append(app.rw.Env.Children, vscene.NewNode(vc))
		}
		app.rw.Model.Children = []*vscene.Node{nModelRool}
		app.ui2.Show()
	})
}

// Check if we have animates nodes
func checkAnim(modelRool *vscene.Node) {
	for _, n := range modelRool.Children {
		an, ok := n.Ctrl.(*vscene.AnimatedNodeControl)
		if ok && len(an.Skin.Animations) > 1 {
			ai := 0
			app.extraUi.Visible = true
			app.extraUi.Content = vui.NewMenuButton("Toggle animation").SetOnClick(func() {
				ai++
				if ai >= len(an.Skin.Animations) {
					ai = 0
				}
				an.SetAnimationIndex(app.rw.GetSceneTime(), ai)
			})
		}
		checkAnim(n)
	}

}

func newPlacedLight(s *Sample, position mgl32.Vec3) *vscene.Node {
	dl := float32(3)
	l := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1, 0.8, 0.4}, Attenuation: mgl32.Vec3{0, 0, 1 / (dl * dl)}}, 512)
	l.MaxDistance = dl * 4
	f := vscene.NewNode(vscene.NewMultiControl(&vscene.TransformControl{Transform: mgl32.Translate3D(0, -0.15, 0)},
		noise.NewFire(0.4, 0.8)))
	return vscene.NewNode(
		vscene.NewMultiControl(&vscene.TransformControl{Transform: mgl32.Translate3D(position[0], position[1], position[2])},
			l), f,
	)
}

type rotLight struct {
	offset float64
	speed  float64
}

func (r *rotLight) Process(pi *vscene.ProcessInfo) {
	// Rotate light round y (up) axis with given speed (radians / s)
	pi.World = pi.World.Mul4(mgl32.HomogRotate3DY(float32(
		(pi.Time - r.offset) * r.speed)))
}

// Create new shadow light node
func newRotLight(s *Sample, speed float64, intesity mgl32.Vec3) *vscene.Node {
	// Put light a target height and about half distance from initial camera distance
	dl := s.Position.Sub(s.Target).Len()
	l := shadow.NewPointLight(vscene.PointLight{Intensity: intesity, Attenuation: mgl32.Vec3{0, 0, 1 / (dl * dl)}}, 512)
	l.MaxDistance = dl * 1.5
	return vscene.NewNode(&rotLight{offset: rand.Float64() * 3, speed: speed},
		vscene.NewNode(
			vscene.NewMultiControl(&vscene.TransformControl{Transform: mgl32.Translate3D(dl/2, s.Target.Y(), 0)}, l),
		))
}

func floorPos(sample *Sample) vscene.NodeControl {
	return floorPosControl{s: sample}
}

type floorPosControl struct {
	s *Sample
}

func (f floorPosControl) Process(pi *vscene.ProcessInfo) {
	dir := f.s.Position.Sub(f.s.Target)
	scale := dir.Len()
	pi.World = pi.World.Mul4(mgl32.Translate3D(-dir.X()*2, f.s.Base, -dir.Z()*2))
	pi.World = pi.World.Mul4(mgl32.Scale3D(scale*2, scale, scale*4))
}

// Custom loader where we can change directory based on final model path
type customLoader struct {
	dir string
}

func (c customLoader) Open(filename string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(c.dir, filename))
}
