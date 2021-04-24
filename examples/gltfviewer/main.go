package main

import (
	"flag"
	"fmt"
	"github.com/lakal3/vge/vge/deferred"
	"github.com/lakal3/vge/vge/forward"
	"log"
	"os"

	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
)

var app struct {
	// Items to dispose when we close scene
	atEndScene vk.Owner
	theme      *mintheme.Theme
	rw         *vapp.RenderWindow
	debug      bool
	ui         *vui.UIView
	ui2        *vui.UIView
	gltfRoot   string
	tools      *vmodel.Model
	probe      *env.Probe
	extraUi    *vui.Conditional
	devIndex   int
	deferred   bool
}

func main() {
	flag.BoolVar(&app.debug, "debug", false, "Use debugging API")
	flag.BoolVar(&app.deferred, "deferred", false, "Use deferred renderer")
	// Use devIndex to select device. If not given, use first device
	flag.IntVar(&app.devIndex, "dev", -1, "Device index")
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

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
		// Uncomment to use C++ debug library
		// vk.VGEDllPath = "VGELibd.dll"
	}
	if app.deferred {
		// Deferred shaders requires dynamic descriptors
		vapp.AddOption(vapp.DynamicDescriptors{MaxDescriptors: 1024})
	}
	app.gltfRoot = flag.Arg(0)
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	var err error

	vapp.Init("gltfViewer", vapp.Desktop{})

	// Load tools model to show some extra elements in some scenes
	app.tools, err = vapp.LoadModel("gltf/tools/Tools.gltf")
	if err != nil {
		log.Fatal("Failed to load initial tools. Ensure that you are in sample directory")
	}

	// Initialize forward rendered
	var rd vscene.Renderer
	if app.deferred {
		rd = deferred.NewRenderer()
	} else {
		rd = forward.NewRenderer(true)
	}
	// Build main window
	app.rw = vapp.NewRenderWindow("GLTF Viewer", rd)

	vapp.RegisterHandler(0, cameraPos)
	// Build UI to select test model from
	err = buildUi()
	if err != nil {
		log.Fatal("Failed to load UI")
	}
	vapp.WaitForShutdown()
}

// Helper to record camera pos when we press F1
func cameraPos(ctx vk.APIContext, ev vapp.Event) (unregister bool) {
	kd, ok := ev.(*vapp.KeyDownEvent)
	if ok && kd.KeyCode == vapp.GLFWKeyF1 {
		c := app.rw.Camera.(*vscene.PerspectiveCamera)
		fmt.Println("At ", c.Position, " Looking to", c.Target)
	}
	return false
}

func usage() {
	fmt.Println("GTFViewer path-to-gltfSamples. See gltfviewer.md for more info")
	os.Exit(1)
}
