// Robomaze example tests VGE rendering performance and also shows how to make custom animations for different elements.
// Scene is intentionally constructed of repeated rendered small tile. Also some elements like fences are more complex that than needed
//
// Robomaze has special option -oil that will turn on decals rendering in Std shader. You will see small oil stains appearing and disappearing while simulations runs.
// Std shader requires that Vulkan supports VK_EXT_descriptor_indexing feature
// that allows pre rendering phase to attach dynamic array of different images (for example decal images and shadow maps) into frame.
package main

import (
	"flag"
	"fmt"
	"github.com/lakal3/vge/vge/deferred"
	"github.com/lakal3/vge/vge/forward"
	"github.com/lakal3/vge/vge/vscene"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vui/mintheme"
)

var debug bool

var app struct {
	mainWnd     *vapp.RenderWindow
	logoModel   *vmodel.Model
	fenceModel  *vmodel.Model
	robotModel  *vmodel.Model
	envMaze     *env.EquiRectBGNode
	stainSet    *vmodel.Model
	probe       *env.Probe
	theme       *mintheme.Theme
	orbitCamera bool
	predepth    bool
	oil         bool
	deferred    bool
	slowShadows bool
	video       bool
	devIndex    int
	fps         bool
}

func main() {
	flag.BoolVar(&debug, "debug", false, "Use debugging API")
	// Draw predepth pass
	flag.BoolVar(&app.predepth, "predepth", false, "Use predepth pass")
	// Draw oil stains
	flag.BoolVar(&app.oil, "oil", false, "Add oil leak decals")
	flag.IntVar(&app.devIndex, "dev", -1, "Device index")
	flag.BoolVar(&app.video, "video", false, "Set windows to video size 1280 x 768")
	flag.BoolVar(&app.fps, "fps", false, "Add FPS debug control to window")
	flag.BoolVar(&app.deferred, "deferred", false, "Use deferred renderer")
	flag.BoolVar(&app.slowShadows, "Slow shadows", false, "Update shadow maps only ~4 times / second")
	flag.Parse()

	if app.devIndex >= 0 {
		vapp.SelectDevice = selDevice
	}
	if debug {
		vapp.AddOption(vapp.Validate{})
		// vk.VGEDllPath = "vgelibd.dll"
	}
	if app.oil || app.deferred {
		// Add dynamics descriptor support.
		vapp.AddOption(vapp.DynamicDescriptors{MaxDescriptors: 1024})
	}
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	// Initialize forward or deferred rendered
	var rd vscene.Renderer
	if app.deferred {
		rd = deferred.NewRenderer()
	} else {
		rdf := forward.NewRenderer(true)
		if app.predepth {
			// Predepth is useless in deferred shading
			rdf.AddDepthPrePass()
		}
		rd = rdf
	}

	vapp.Init("robomaze", vapp.Desktop{})
	if app.video {
		app.mainWnd = vapp.NewRenderWindowAt("Robo maze", vk.WindowPos{Left: -1, Top: -1, Width: 1246, Height: 730}, rd)
	} else {
		app.mainWnd = vapp.NewRenderWindow("Robo maze", rd)
	}

	err := loadModels1()
	if err != nil {
		log.Fatal("Failed to load models, ", err)
	}
	lw := buildStartScene()
	vapp.AddChild(lw)
	go func() {
		err = loadModels2(lw)
		if err != nil {
			log.Fatal("Failed to load models 2, ", err)
		}
	}()
	vapp.WaitForShutdown()
}

// Select device based on devIndex parameter given in program arguments
func selDevice(devices []vk.DeviceInfo) int32 {
	if len(devices) < app.devIndex {
		log.Fatal("No device ", app.devIndex)
	}
	d := devices[app.devIndex]
	fmt.Println("Selected device ", string(d.Name[:d.NameLen]))
	return int32(app.devIndex)
}

func loadModels1() (err error) {
	// Load logo model
	app.logoModel, err = vapp.LoadModel("gltf/logo/Logo.gltf")
	if err != nil {
		return err
	}
	vapp.AddChild(app.logoModel)
	// Create UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	app.theme = mintheme.NewTheme(vapp.Dev, 15, nil, nil, nil)
	vapp.AddChild(app.theme)
	return nil
}

func loadModels2(lw *logoWindow) (err error) {
	// Load fence model
	app.fenceModel, err = vapp.LoadModel("gltf/fence/Fence.gltf")
	if err != nil {
		return err
	}
	vapp.AddChild(app.fenceModel)
	// Load model or animated robot
	app.robotModel, err = vapp.LoadModel("gltf/mechrobot/mechrobot.gltf")
	if err != nil {
		return err
	}
	vapp.AddChild(app.robotModel)
	app.mainWnd.Scene.Update(func() {
		lw.bRun.Disabled, lw.lStat.Text = false, ""
	})

	rawEnv, err := vapp.AM.Load("envhdr/kloofendal_48d_partly_cloudy_2k.hdr",
		func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Dev, 100, "hdr", content), nil
		})
	if err != nil {
		return err
	}
	app.envMaze = rawEnv.(*env.EquiRectBGNode)
	app.probe = env.NewProbe(vapp.Dev)
	if app.oil {
		b := vmodel.ModelBuilder{}
		rIdx, err := vapp.AM.Load("decals/stain_albedo.png", func(content []byte) (asset interface{}, err error) {
			return b.AddImage("png", content, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit), nil
		})
		if err != nil {
			return err
		}
		props := vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, rIdx.(vmodel.ImageIndex)).
			SetFactor(vmodel.FMetalness, 1).SetFactor(vmodel.FRoughness, 0.2).SetColor(vmodel.CAlbedo, mgl32.Vec4{0.7, 0.7, 0.7, 0.7})
		b.AddDecalMaterial("oil_stain", props)
		app.stainSet, err = b.ToModel(vapp.Dev)
		if err != nil {
			return err
		}
	}
	return nil
}
