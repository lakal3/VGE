package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/debugmat"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/materials/unlit"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vmodel/objloader"
	"github.com/lakal3/vge/vge/vscene"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var config struct {
	outPath        string
	env            string
	debug          bool
	devIndex       int
	maxDescriptors int
}

type viewerApp struct {
	owner     vk.Owner
	m         *vmodel.Model
	mLB       *vmodel.Model
	nModel    *vscene.Node
	nLights   *vscene.Node
	nEnv      *vscene.Node
	lightBall vmodel.NodeBuilder
	dSet      *decal.Set
}

func (v *viewerApp) Dispose() {
	v.owner.Dispose()
}

func (v *viewerApp) SetError(err error) {
	log.Fatal("API error ", err)
}

func (v *viewerApp) IsValid() bool {
	return true
}

func (v *viewerApp) Begin(callName string) (atEnd func()) {
	return nil
}

var app viewerApp
var win *vapp.RenderWindow

func main() {
	flag.StringVar(&config.outPath, "out", "", "Image output path")
	flag.StringVar(&config.env, "env", "", "Environment HDR image")
	flag.BoolVar(&config.debug, "debug", false, "Use debug layers")
	flag.IntVar(&config.devIndex, "dev", -1, "Device index")
	flag.IntVar(&config.maxDescriptors, "maxDescriptors", 1024, "Max dynamics descriptors. Set to 0 to disable feature")

	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}
	if config.devIndex >= 0 {
		// Override default
		vapp.SelectDevice = func(devices []vk.DeviceInfo) int32 {
			if len(devices) < config.devIndex {
				log.Fatal("No device ", config.devIndex)
			}
			d := devices[config.devIndex]
			fmt.Println("Selected device ", string(d.Name[:d.NameLen]))
			return int32(config.devIndex)
		}
	}
	if config.debug {
		vk.VGEDllPath = "vgelibd.dll"
	}
	app.init()
	defer app.Dispose()
	app.loadModel()
	if len(config.outPath) > 0 {
		app.renderToFile()
	} else {
		vapp.RegisterHandler(1000, app.keyHandler)
		win = vapp.NewRenderWindow("Model viewer", vapp.NewForwardRenderer(true))
		oc := vapp.NewOrbitControl(200, win)
		if app.nEnv != nil {
			win.Scene.Root.Children = append(win.Scene.Root.Children, app.nEnv)
		}
		nProbe := &vscene.Node{Children: []*vscene.Node{app.nModel}}
		nProbe.Ctrl = env.NewProbe(vapp.Ctx, vapp.Dev)
		win.Scene.Root.Children = append(win.Scene.Root.Children, app.nLights, nProbe)
		oc.Zoom(&win.Scene)
		app.setMode()
		vapp.WaitForShutdown()
	}
}

func usage() {
	fmt.Println("modelviewer {flags} model")
	os.Exit(1)
}

func (a *viewerApp) init() {
	if config.debug {
		vapp.AddOption(vapp.Validate{})
	}
	if len(config.outPath) == 0 {
		vapp.AddOption(vapp.Desktop{})
	}
	if config.maxDescriptors != 0 {
		vapp.AddOption(vapp.DynamicDescriptors{MaxDescriptors: uint32(config.maxDescriptors)})
	}
	vapp.Ctx = a
	vapp.Init("modelview")
}

func (v *viewerApp) loadModel() {
	fName := flag.Arg(0)
	ext := strings.ToLower(filepath.Ext(fName))
	dl := vasset.DirectoryLoader{Directory: filepath.Dir(fName)}
	mb := vmodel.ModelBuilder{}
	// mb.ShaderFactory = phong.PhongFactory
	mb.ShaderFactory = debugmat.DebugMatFactory
	uvr, err := ioutil.ReadFile("assets/tests/uvchecker.dds")
	if err != nil {
		v.SetError(fmt.Errorf("Can load uvchecker image, %v", err))
		return
	}
	mb.AddImage("dds", uvr, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	// mb.ShaderFactory = unlit.UnlitFactory
	switch ext {
	case ".obj":
		ol := objloader.ObjLoader{Builder: &mb, Loader: dl}
		err := ol.LoadFile(filepath.Base(fName))
		if err != nil {
			v.SetError(err)
		}
	case ".gltf":
		ol := gltf2loader.GLTF2Loader{Builder: &mb, Loader: dl}
		err := ol.LoadGltf(filepath.Base(fName))
		if err != nil {
			v.SetError(err)
		}
		err = ol.Convert(0)
		if err != nil {
			v.SetError(err)
		}
	default:
		log.Fatal("Unknown model type ", ext)
	}

	v.m = mb.ToModel(v, vapp.Dev)
	mb2 := &vmodel.ModelBuilder{ShaderFactory: unlit.UnlitFactory}
	mbBall := &vmodel.MeshBuilder{}
	mbBall.AddCube(mgl32.Scale3D(0.1, 0.1, 0.1))
	lightBall := mb2.AddMesh(mbBall)
	ballMat := mb2.AddMaterial("ball", vmodel.NewMaterialProperties().SetColor(vmodel.CAlbedo, mgl32.Vec4{0, 1, 0, 1}))
	ballMat2 := mb2.AddMaterial("ball2", vmodel.NewMaterialProperties().SetColor(vmodel.CAlbedo, mgl32.Vec4{1, 1, 1, 1}))
	mb2.AddNode("lightball", nil, mgl32.Ident4()).SetMesh(lightBall, ballMat)
	mb2.AddNode("lightball2", nil, mgl32.Ident4()).SetMesh(lightBall, ballMat2)
	v.mLB = mb2.ToModel(v, vapp.Dev)
	v.owner.AddChild(v.m)
	v.nModel = vscene.NodeFromModel(app.m, 0, true)
	v.setLights()

	if len(config.env) > 0 {
		content, err := ioutil.ReadFile(config.env)
		if err != nil {
			v.SetError(err)
			return
		}
		eq := env.NewEquiRectBGNode(v, vapp.Dev, 1000, "hdr", content)
		v.owner.AddChild(eq)
		v.nEnv = &vscene.Node{Ctrl: eq}
	}
	// Add decals
	b := &decal.Builder{}
	stAlbedo := v.loadImage(b, "assets/decals/stain_albedo.png")
	props := vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, stAlbedo)
	b.AddDecal("stain", props)
	v.dSet = b.Build(v, vapp.Dev)
	v.owner.AddChild(v.dSet)
	v.nModel.Ctrl = vscene.NewMultiControl(
		v.dSet.NewInstance("stain", mgl32.Ident4()),
		v.dSet.NewInstance("stain", mgl32.Translate3D(2, 0, 2)),
		v.dSet.NewInstance("stain", mgl32.Translate3D(2, 0, -2).Mul4(mgl32.HomogRotate3DX(-1))),
		v.dSet.NewInstance("stain", mgl32.Translate3D(-2, 0, -2).Mul4(mgl32.Scale3D(1.5, 1.5, 1.5))),
	)
}

func (v *viewerApp) loadImage(sb *decal.Builder, imageName string) vmodel.ImageIndex {
	content, err := ioutil.ReadFile(imageName)
	if err != nil {
		v.SetError(err)
		return 0
	}
	return sb.AddImage("png", content, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
}

var layers = mgl32.Vec3{1, 1, 1}
var mode float32
var imageIndex float32 = 2
var paused = false

func (v *viewerApp) keyHandler(ctx vk.APIContext, ev vapp.Event) (unregister bool) {
	ke, ok := ev.(*vapp.CharEvent)
	if ok {
		switch ke.Char {
		case ' ':
			paused = !paused
			win.SetPaused(paused)
		case 'Q':
			go vapp.Terminate()
			return true
		case '0':
			mode = 0
			v.setMode()
		case 'x':
			layers = mgl32.Vec3{1, 0, 0}
			v.setMode()
		case 'y':
			layers = mgl32.Vec3{0, 1, 0}
			v.setMode()
		case 'z':
			layers = mgl32.Vec3{0, 0, 1}
			v.setMode()
		case 'a':
			layers = mgl32.Vec3{1, 1, 1}
			v.setMode()
		case 'f':
			mode = 0
			v.setMode()
		case 'n':
			mode = 2
			v.setMode()
		case 't':
			mode = 4
			v.setMode()
		case 'u':
			mode = 1
			v.setMode()
		case 'e':
			mode = 5
			v.setMode()
		case 's':
			mode = 6
			v.setMode()
		case 'R':
			mode = 3
			v.setMode()
		}
	}
	return false
}

func (v *viewerApp) setMode() {
	switch mode {
	case 1:
		debugmat.DebugModes = mgl32.Vec4{imageIndex, 0, 0, 1}
		return
	}
	debugmat.DebugModes = layers.Vec4(mode)
}

func (v *viewerApp) setLights() {
	v.nLights = &vscene.Node{}
	nPoint := vscene.NewNode(vscene.NewMultiControl(
		&vscene.RotateAnimate{Speed: 1.2, Axis: mgl32.Vec3{0, 1, 0}},
		&vscene.TransformControl{mgl32.Translate3D(4.3, 3.5, 0)}),
		&vscene.Node{
			Ctrl: shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{0.2, 1.8, 0.2}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512),
		}, vscene.NodeFromModel(app.mLB, 1, false))
	nPoint2 := vscene.NewNode(vscene.NewMultiControl(
		&vscene.RotateAnimate{Speed: 0.7, Axis: mgl32.Vec3{0, 1, 0}},
		&vscene.TransformControl{mgl32.Translate3D(6, 1.5, 0)}),
		&vscene.Node{
			Ctrl: shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1.2, 1.2, 1.2}, Attenuation: mgl32.Vec3{0, 0, 0.32}}, 512),
		}, vscene.NodeFromModel(app.mLB, 2, false))
	dl := &vscene.Node{
		Ctrl: &vscene.DirectionalLight{Intensity: mgl32.Vec3{0.3, 0.3, 0.3}, Direction: mgl32.Vec3{-0.1, -0.8, 0}},
	}
	_ = dl
	v.nLights.Children = append(v.nLights.Children, nPoint, nPoint2, dl)
	if len(config.env) == 0 {
		v.nLights.Children = append(v.nLights.Children, &vscene.Node{
			Ctrl: &vscene.AmbientLight{Intensity: mgl32.Vec3{0.2, 0.2, 0.2}}})

	}
}
