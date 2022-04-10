package main

import (
	"flag"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/shaders/phongshader"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset/jpgloader"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vk"

	"log"
	"os"
)

var settings struct {
	validate bool
	phong    bool
	bgImage  string
}

var app struct {
	rw           *vapp.ViewWindow
	phongshaders *shaders.Pack
	drives       []string
}

func main() {
	flag.BoolVar(&settings.validate, "validate", false, "Add Vulkan validation layers")
	flag.BoolVar(&settings.phong, "phong", false, "Use phong shader")
	flag.StringVar(&settings.bgImage, "bg", "", "Background image for 3D models")
	flag.Parse()
	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	var options []vapp.ApplicationOption
	options = append(options, vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 2048})
	if settings.validate {
		options = append(options, vapp.Validate{})
	}
	if len(settings.bgImage) > 0 {
		_, err := os.Stat(settings.bgImage)
		if err != nil {
			log.Fatal("Stat background image: ", err)
		}
	}
	err := fillLogicalDriveLetters()
	if err != nil {
		log.Fatal("Get logical driver letters failed", err)
	}
	var dir string
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	} else {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal("Getwd failed", err)
		}
	}
	if settings.phong {
		app.phongshaders, err = phongshader.LoadPack()
		if err != nil {
			log.Fatal("Load phong shaders failed: ", err)
		}
	}
	err = vapp.Init("UItest", options...)
	if err != nil {
		log.Fatal("App init failed ", err)
	}
	pngloader.RegisterPngLoader()
	jpgloader.RegisterJPEGLoader()

	err = vdraw.CompileShapes(vapp.Dev)
	if err != nil {
		log.Fatal("Compile ", err)
	}

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	app.rw = vapp.NewViewWindow("File viewer (alpha)", vk.WindowPos{Left: -1, Top: -1, Width: 1024, Height: 768})
	// Add file tree
	err = addFileTree(dir)
	if err != nil {
		log.Fatal("Add file tree ", err)
	}
	addStatView()
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}
