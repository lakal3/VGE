// Animation example shows how to load external animation from BVH file and apply it to rigged models
// Currently VGE support only BVH (Biovision Hierarchy) format for animations.
//
// VGE has three example models, one from makehuman and other download from opengameart.com. See readme in assets for more details.
// Use command line switch -elf to change model.

package main

import (
	"flag"
	"fmt"
	"github.com/lakal3/vge/vge/forward"
	"log"

	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
)

var app struct {
	debug      bool
	elf        int
	mainWnd    *vapp.RenderWindow
	houseModel *vmodel.Model
	actorModel *vmodel.Model
	actor      *vscene.Node
	probe      *env.Probe
	theme      *mintheme.Theme
	bg         *env.EquiRectBGNode
	selectorUI *vui.UIView
	devIndex   int
}

func main() {
	flag.BoolVar(&app.debug, "debug", false, "Use debugging API")
	flag.IntVar(&app.devIndex, "dev", -1, "Device index")
	flag.IntVar(&app.elf, "elf", 0, "Elf model 1 or 2")
	flag.Parse()

	if app.devIndex >= 0 {
		vapp.SelectDevice = selDevice
	}
	if app.debug {
		vapp.AddOption(vapp.Validate{})
		// Uncommment next line to run with debug version on vgelib. NOTE! Debug versions are not included in prebuilt. You must build one yourself
		// vk.VGEDllPath = "vgelibd.dll"
	}
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	// Use standard forward renderer
	rd := forward.NewRenderer(true)
	vapp.Init("Elf", vapp.Desktop{})
	app.mainWnd = vapp.NewRenderWindow("Animation demo", rd)
	err := loadModels()
	if err != nil {
		log.Fatal("Failed to load models, ", err)
	}
	buildScene()
	vapp.WaitForShutdown()
}

func loadModels() (err error) {
	// Load logo model based on command line switch
	switch app.elf {
	case 0:
		app.actorModel, err = vapp.LoadModel("gltf/female1/Female1.gltf")
		if err != nil {
			return err
		}
		boneMap = boneMap2
	case 1:
		app.actorModel, err = vapp.LoadModel("gltf/elf/elf.gltf")
		if err != nil {
			return err
		}
	case 2:
		app.actorModel, err = vapp.LoadModel("gltf/elf/n_elf.gltf")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("-elf switch must be 0, 1 or 2")
	}
	vapp.AddChild(app.actorModel)
	// Load background image (old house)
	app.houseModel, err = vapp.LoadModel("gltf/Cabin/cabin_hp.gltf")
	if err != nil {
		return err
	}
	vapp.AddChild(app.houseModel)

	// Initialize application with Oxanium font
	ftRaw, err := vapp.AM.Load("fonts/Oxanium-Medium.ttf", func(content []byte) (asset interface{}, err error) {
		fl := &vglyph.VectorSetBuilder{}
		// Add font will load given range of glyph to Vector set build. Vector set builder will rasterize vectors
		// and convert them to signed depth fields
		// You must also specify character unicode range you are interested in. Some Unicode font can contain quite a large number
		// of characters
		err = fl.AddFont(content, vglyph.Range{From: 33, To: 255})
		if err != nil {
			return nil, err
		}

		// Build glyphset
		return fl.Build(vapp.Dev), nil
	})
	if err != nil {
		return err
	}
	// Create UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	app.theme = mintheme.NewTheme(vapp.Dev, 15, nil, ftRaw.(*vglyph.GlyphSet), nil)
	vapp.AddChild(app.theme)

	app.probe = env.NewProbe(vapp.Dev)
	vapp.AddChild(app.probe)

	// Night time background image
	app.bg = vapp.MustLoadAsset("envhdr/preller_drive_2k.hdr", func(content []byte) (asset interface{}, err error) {
		return env.NewEquiRectBGNode(vapp.Dev, 100, "hdr", content), nil
	}).(*env.EquiRectBGNode)
	return nil
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
