// Test application for different kind of material
//

package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
	"log"
)

var app struct {
	debug          bool
	devIndex       int
	maxDescriptors int
	tools          *vmodel.Model
	rw             *vapp.RenderWindow
	theme          *mintheme.Theme
	ui             *vui.UIView
	nLights        *vscene.Node
	cam            *vscene.PerspectiveCamera
}

func main() {
	flag.BoolVar(&app.debug, "debug", false, "Use debugging API")
	flag.IntVar(&app.devIndex, "dev", -1, "Device index")
	flag.IntVar(&app.maxDescriptors, "maxDescriptors", 1024, "Max dynamics descriptors. Set to 0 to disable feature")
	flag.Parse()

	if app.devIndex >= 0 {
		// Override default
		vapp.SelectDevice = func(devices []vk.DeviceInfo) int32 {
			if len(devices) < app.devIndex {
				log.Fatal("No device ", app.devIndex)
			}
			d := devices[app.devIndex]
			fmt.Println("Selected device ", string(d.Name[:d.NameLen]))
			return int32(app.devIndex)
		}
	}
	if app.debug {
		vapp.AddOption(vapp.Validate{})
		vk.VGEDllPath = "VGELibd.dll"
	}

	var err error
	if app.maxDescriptors != 0 {
		vapp.AddOption(vapp.DynamicDescriptors{MaxDescriptors: uint32(app.maxDescriptors)})
	}
	vapp.Init("material tester", vapp.Desktop{})

	// Load tools model to show some extra elements in some scenes
	app.tools, err = vapp.LoadModel("assets/gltf/tools/Tools.gltf")
	if err != nil {
		log.Fatal("Failed to load initial tools. Ensure that you are in sample directory")
	}

	// Initialize forward rendered
	rd := vapp.NewForwardRenderer(true)
	// Build main window
	app.rw = vapp.NewRenderWindow("Material tester", rd)

	initScene()
	// Build UI to select test model from
	buildUi()

	vapp.WaitForShutdown()
}

func initScene() {
	bg := vapp.MustLoadAsset("assets/envhdr/studio.hdr", func(content []byte) (asset interface{}, err error) {
		return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
	}).(*env.EquiRectBGNode)
	app.rw.Model.Ctrl = env.NewProbe(vapp.Ctx, vapp.Dev)
	app.nLights = vscene.NewNode(nil)
	app.rw.Env.Children = []*vscene.Node{vscene.NewNode(bg), app.nLights}
	app.cam = vscene.NewPerspectiveCamera(1000)
	app.cam.Position = mgl32.Vec3{0.5, 1, 2.5}
	app.rw.Camera = app.cam
	vapp.OrbitControlFrom(-100, app.rw, app.cam)
}
