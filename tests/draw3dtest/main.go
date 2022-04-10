package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders/mixshader"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vdraw3d"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"log"
	"math"
	"os"
	"path/filepath"
)

var app struct {
	rw     *vapp.ViewWindow
	nv     *vdraw3d.View
	model  *vmodel.Model
	model2 *vmodel.Model
	model3 *vmodel.Model
	images [3]vmodel.ImageIndex
	nodes  map[string]vdraw3d.FrozenID
}

func main() {
	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	err := vapp.Init("UItest", vapp.Validate{}, vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 1024})
	// err := vapp.Init("UItest", vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 1024})
	if err != nil {
		log.Fatal("App init failed ", err)
	}

	err = vdraw.CompileShapes(vapp.Dev)
	if err != nil {
		log.Fatal("Compile ", err)
	}

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	app.rw = vapp.NewViewWindow("UITest", vk.WindowPos{Left: -1, Top: -1, Width: 1024, Height: 768})
	// Build scene
	err = buildScene()
	if err != nil {
		log.Fatal("Build scene ", err)
	}
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

func buildScene() error {
	err := mintheme.BuildMinTheme()
	if err != nil {
		return err
	}

	err = buildModel()
	if err != nil {
		return err
	}
	sp, err := mixshader.LoadPack()
	if err != nil {
		return err
	}
	sv := vdraw3d.NewCustomView(vapp.Dev, sp, paintStatic, paintScene)
	sv.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		return vdraw.Area{From: mgl32.Vec2{float32(desc.Width) / 3, 0}, To: mgl32.Vec2{float32(desc.Width), float32(desc.Height)}}
	}
	sv.OnEvent = moveMonkey
	c := vapp.NewOrbitControl(0, nil)
	c.ZoomTo(mgl32.Vec3{0, 0, 0}, 10)
	sv.Camera = c

	app.nv = sv
	app.rw.AddView(sv)
	nf := vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, painter)
	nf.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		return vdraw.Area{From: mgl32.Vec2{0, 0}, To: mgl32.Vec2{float32(desc.Width) / 3, float32(desc.Height)}}

	}
	app.rw.AddView(nf)

	return nil
}

func buildModel() (err error) {

	app.model, err = loadModel("gltf/testparts", "simple1.gltf",
		"../../assets/envhdr/kloofendal_48d_partly_cloudy_2k.hdr",
		"../../assets/decals/uc_albedo.png", "../../assets/decals/uc_normal.png")
	if err != nil {
		return err
	}
	app.model2, err = loadModel("gltf/testparts", "testparts_mc.gltf")
	if err != nil {
		return err
	}
	app.model3, err = loadModel("gltf/mechrobot", "mechrobot.gltf")

	return nil
}

func loadModel(subpath string, modelName string, imagePaths ...string) (*vmodel.Model, error) {
	mb := vmodel.ModelBuilder{ShaderFactory: func(dev *vk.Device, propSet vmodel.MaterialProperties) (sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
		return nil, nil, nil, nil
	}}

	gl := gltf2loader.GLTF2Loader{Builder: &mb, Loader: vasset.DirectoryLoader{Directory: "../../assets/" + subpath}}
	err := gl.LoadGltf(filepath.Base(modelName))
	if err != nil {
		log.Fatal("Failed to load gltf file, simple1.gltf: ", err)
	}
	// GLTF can have multiple scenes. We must choose one to convert. You can convert scenes to single models
	// We just pick first
	err = gl.Convert(0)
	if err != nil {
		log.Fatal("Failed to build model simple1.gltf: ", err)
	}
	for idx, imPath := range imagePaths {
		var imageBytes []byte
		imageBytes, err = os.ReadFile(imPath)
		if err != nil {
			return nil, err
		}
		if filepath.Ext(imPath) == ".hdr" {
			app.images[idx] = mb.AddImage("hdr", imageBytes, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
		} else {
			app.images[idx] = mb.AddImage("png", imageBytes, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
		}
	}
	return mb.ToModel(vapp.Dev)
}

var modelType int

var uiKeys = vk.NewKeys(10)

func painter(fr *vimgui.UIFrame) {
	fr.NewLine(-100, 25, 5)
	vimgui.RadioButton(fr, uiKeys, "Suzie", 0, &modelType)
	fr.NewLine(-100, 25, 2)
	vimgui.RadioButton(fr, uiKeys+1, "Robot", 1, &modelType)
	fr.NewLine(120, 30, 5)
	if vimgui.Button(fr, uiKeys+2, "Debug settings") {
		showDebug()
	}
}

var sceneKeys = vk.NewKeys(3)

func paintStatic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	vdraw3d.DrawBackground(dl, app.model, app.images[0])
	pl2 := vmodel.NewMaterialProperties()
	pl2.SetColor(vmodel.CIntensity, mgl32.Vec4{10, 2.5, 0, 1})
	vdraw3d.DrawProbe(dl, sceneKeys, mgl32.Vec3{0, 1, -1})
	vdraw3d.DrawDirectionalLight(dl, mgl32.Vec3{0.1, -1, 0.1}, nil)
	vdraw3d.DrawPointLight(dl, sceneKeys+1, mgl32.Vec3{0, 5, -1}, pl2)
	app.model2.GetNode(0).Enum(mgl32.Ident4(), func(local mgl32.Mat4, n vmodel.Node) {
		if n.Name == "Cube_Grass_1" || n.Name == "Cube_Rock_1" {
			return
		}
		if n.Mesh > 0 {
			if n.Name == "Ground_1" {
				paintGround(dl, n, local)
			} else {
				vdraw3d.DrawMesh(dl, app.model2.GetMesh(n.Mesh), local, app.model2.GetMaterial(n.Material).Props)
			}
		}
	})

}

func paintGround(dl *vdraw3d.FreezeList, n vmodel.Node, local mgl32.Mat4) {
	m := app.model2.GetMaterial(n.Material)
	props := m.Props
	mg := app.model2.FindMaterial("Grass")
	mr := app.model2.FindMaterial("Rock")
	if mg < 0 || mr < 0 {
		return
	}
	t1 := app.model2.GetMaterial(mg).Props.GetImage(vmodel.TxAlbedo)
	t2 := app.model2.GetMaterial(mr).Props.GetImage(vmodel.TxAlbedo)
	props.SetImage(vmodel.TxCustom1, t1)
	props.SetImage(vmodel.TxCustom2, t2)
	vdraw3d.DrawMesh(dl, app.model2.GetMesh(n.Mesh), local, props, vdraw3d.ColorShader{
		Shader: "mix_color",
	})
}

var monkeyPos mgl32.Mat4 = mgl32.Translate3D(0, 1.2, 0)
var monkeyMove mgl32.Vec3
var eMonkey float64

func moveMonkey(ev vapp.Event) {
	kd, ok := ev.(*vapp.KeyDownEvent)
	if ok {
		if kd.KeyCode == 'A' {
			monkeyMove[0] = -2
			kd.SetHandled()
		}
		if kd.KeyCode == 'D' {
			monkeyMove[0] = 2
			kd.SetHandled()
		}
		if kd.KeyCode == 'W' {
			monkeyMove[2] = -2
			kd.SetHandled()
		}
		if kd.KeyCode == 'S' {
			monkeyMove[2] = 2
			kd.SetHandled()
		}
		if kd.KeyCode == 'E' {
			monkeyMove[1] = -2
			kd.SetHandled()
		}
		if kd.KeyCode == 'Q' {
			monkeyMove[1] = 2
			kd.SetHandled()
		}
	}
	ku, ok := ev.(*vapp.KeyUpEvent)
	if ok {
		if ku.KeyCode == 'A' || ku.KeyCode == 'D' {
			monkeyMove[0] = 0
		}
		if ku.KeyCode == 'Q' || ku.KeyCode == 'E' {
			monkeyMove[1] = 0
		}
		if ku.KeyCode == 'W' || ku.KeyCode == 'S' {
			monkeyMove[2] = 0
		}
	}
}

func paintScene(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	if eMonkey == 0 {
		eMonkey = v.Elapsed
	} else {
		delta := v.Elapsed - eMonkey
		eMonkey = v.Elapsed
		mp := monkeyMove.Mul(float32(delta))
		monkeyPos = monkeyPos.Mul4(mgl32.Translate3D(mp[0], mp[1], mp[2]))
	}
	if modelType == 0 {
		n := app.model.GetNode(2)
		if n.Mesh >= 0 {
			props := app.model.GetMaterial(n.Material).Props
			props.SetColor(vmodel.CAlbedo, mgl32.Vec4{0.2, 0.8, 0.2, 0.2})
			vdraw3d.DrawMesh(dl, app.model.GetMesh(n.Mesh), monkeyPos, props, vdraw3d.Transparent{})
		}
	} else {
		n := app.model3.GetNode(5)
		if n.Mesh >= 0 {
			props := app.model3.GetMaterial(n.Material).Props
			p := monkeyPos.Mul4(mgl32.Scale3D(0.01, 0.01, 0.01)).Mul4(mgl32.HomogRotate3DX(math.Pi / 2))
			// vimscene.DrawMesh(dl, app.model3.GetMesh(n.Mesh), p, props)
			sk := app.model3.GetSkin(n.Skin)
			vdraw3d.DrawAnimated(dl, app.model3.GetMesh(n.Mesh), sk, sk.Animations[0], v.Elapsed, p, props)
		}

	}
	dp := vmodel.NewMaterialProperties()
	dp.SetColor(vmodel.CAlbedo, mgl32.Vec4{1, 1, 1, 1})
	dp.SetImage(vmodel.TxAlbedo, app.images[1])
	dp.SetImage(vmodel.TxBump, app.images[2])
	vdraw3d.DrawDecal(dl, app.model, mgl32.Scale3D(2, 2, 2), 0, 0, dp)
	f := app.nodes["Solid1.003_1"]
	vdraw3d.DrawDecal(dl, app.model, mgl32.Scale3D(2, 4, 2).Mul4(mgl32.Translate3D(-2, 0, 1)),
		f, f+1, dp)
}
