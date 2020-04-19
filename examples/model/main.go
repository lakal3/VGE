package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/pbr"
	"github.com/lakal3/vge/vge/materials/phong"
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
	outPath string
	env     string
	debug   bool
}

type viewerApp struct {
	owner   vk.Owner
	m       *vmodel.Model
	nModel  *vscene.Node
	nLights *vscene.Node
	nEnv    *vscene.Node
}

func (v *viewerApp) Dispose() {
	v.owner.Dispose()
}

var app viewerApp

func main() {
	flag.StringVar(&config.env, "env", "", "Environment HDR image")
	flag.BoolVar(&config.debug, "debug", false, "Use debug layers")
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}
	if config.debug {
		vk.VGEDllPath = "vgelibd.dll"
	}
	app.init()
	defer app.Dispose()
	app.loadModel()

	// Register keyboard handle that quits application when user presses Q
	vapp.RegisterHandler(1000, app.keyHandler)

	// Initialize new window with forward render pass that support depth buffer
	w := vapp.NewRenderWindow("Model viewer", vapp.NewForwardRenderer(true))
	// Orbit control for camera. NewRenderWindow will create default perspective camera that orbit control can manipulate
	oc := vapp.NewOrbitControl(200, w)
	if app.nEnv != nil {
		// Attach environment to camera
		w.Env.Children = append(w.Env.Children, app.nEnv)
	} else {
		w.Env.Children = append(w.Env.Children, vscene.NewNode(env.NewGrayBG()))
	}
	// Probe will create 3D from nodes center point containing everything seen from
	// Probe will not render any child nodes of itself but child nodes can use probe cube image to reflect
	// world around them. PBR shader will not work properly without probe
	nProbe := &vscene.Node{Children: []*vscene.Node{app.nModel}}
	nProbe.Ctrl = env.NewProbe(vapp.Ctx, vapp.Dev)
	w.Scene.Root.Children = append(w.Scene.Root.Children, app.nLights, nProbe)
	oc.Zoom(&w.Scene)
	vapp.WaitForShutdown()
}

func usage() {
	fmt.Println("modelviewer {flags} model")
	os.Exit(1)
}

func (a *viewerApp) init() {
	// Add validation layers
	if config.debug {
		vapp.AddOption(vapp.Validate{})
	}
	vapp.Init("model", vapp.Desktop{})
	// Native image loaders for png and jpeg images
	vasset.RegisterNativeImageLoader(vapp.Ctx, vapp.App)
}

func (v *viewerApp) loadModel() error {
	fName := flag.Arg(0)
	ext := strings.ToLower(filepath.Ext(fName))
	dl := vasset.DirectoryLoader{Directory: filepath.Dir(fName)}
	// Model builder construct actual model
	// File loaders will generate model builder artifacts
	mb := vmodel.ModelBuilder{}
	mb.ShaderFactory = phong.PhongFactory
	// mb.ShaderFactory = unlit.UnlitFactory
	switch ext {
	case ".obj":
		ol := objloader.ObjLoader{Builder: &mb, Loader: dl}
		err := ol.LoadFile(filepath.Base(fName))
		if err != nil {
			return err
		}
	case ".gltf":
		mb.ShaderFactory = pbr.PbrFactory
		ol := gltf2loader.GLTF2Loader{Builder: &mb, Loader: dl}
		err := ol.LoadGltf(filepath.Base(fName))
		if err != nil {
			return err
		}
		// GLTF file can have multiple scenes. We convert each scene individually
		err = ol.Convert(0)
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown model type ", ext)
	}

	v.m = mb.ToModel(vapp.Ctx, vapp.Dev)
	// Record model for dispose at end
	v.owner.AddChild(v.m)
	// Construct actual scene from whole model
	v.nModel = vscene.NodeFromModel(app.m, 0, true)
	v.nLights = &vscene.Node{}
	v.nLights.Children = append(v.nLights.Children,
		&vscene.Node{
			Ctrl: &vscene.DirectionalLight{Intensity: mgl32.Vec3{0.7, 0.7, 0.7}, Direction: mgl32.Vec3{-0.1, -0.8, 0}},
		})
	if len(config.env) > 0 {
		// Add environment (Skybox) to scene
		content, err := ioutil.ReadFile(config.env)
		if err != nil {
			return err
		}
		eq := env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 1000, "hdr", content)
		v.owner.AddChild(eq)
		v.nEnv = &vscene.Node{Ctrl: eq}
	} else {
		v.nLights.Children = append(v.nLights.Children, &vscene.Node{
			Ctrl: &vscene.AmbientLight{Intensity: mgl32.Vec3{0.2, 0.2, 0.2}}})

	}
	return nil
}

func (v *viewerApp) keyHandler(ctx vk.APIContext, ev vapp.Event) (unregister bool) {
	ke, ok := ev.(*vapp.CharEvent)
	if ok && ke.Char == 'Q' {
		// Quit if we press Q
		go vapp.Terminate()
		return true
	}
	return false
}
