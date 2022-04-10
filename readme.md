# VGE (Vulkan Graphics Engine)

The VGE is a graphics engine for the Go language that uses the [Vulkan](https://www.khronos.org/vulkan/) API.


## VGE code example (logo.go)

A simple example showing a simple user interface and a 3D model on screen.
See full sample with imports in examples/basic/logo.go

```go
var app struct {
	model   *vmodel.Model
	bgImage vmodel.ImageIndex
	probe   vdraw3d.FrozenID
	lights  bool
}

func main() {

	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	// We must also add DynamicDescriptors and tell how many images we are going to use for one frame
	vapp.Init("logo", vapp.Validate{}, vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 100})

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	rw := vapp.NewViewWindow("VGE Logo", vk.WindowPos{Left: -1, Top: -1, Height: 768, Width: 1024})
	buildScene(rw)
	// Build ui view
	buildUi(rw)
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

func buildScene(rw *vapp.ViewWindow) {
	// Load envhdr/studio.hdr and create background "skybox" from it. VGE can create whole 360 background
	// from full 360 / 180 equirectangular image without needing 6 images for full cube
	studio, err := os.ReadFile("../../assets/envhdr/studio.hdr")
	if err != nil {
		log.Fatalln("Failed to load studio.hdr: ", err)
	}
	// Init model builder to build assets need for 3D rendering
	b := &vmodel.ModelBuilder{}
	// Add studio image to model. In Vulkan we must specify how we are going to use each image.
	// In this case we copy content to image (vk.IMAGEUsageTransferDstBit) and then sample from it in shaders(vk.IMAGEUsageSampledBit)
	app.bgImage = b.AddImage("hdr", studio, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
	// Initialize loader to load gltf image
	gl := gltf2loader.GLTF2Loader{Builder: b, Loader: vasset.DirectoryLoader{Directory: "../../assets/gltf/logo"}}
	err = gl.LoadGltf("logo.gltf")
	if err != nil {
		log.Fatalln("Failed to load logo.gltf: ", err)
	}
	// Convert first scene (the only one in this gltf) to ModelBuilder
	err = gl.Convert(0)
	err = gl.LoadGltf("logo.gltf")
	if err != nil {
		log.Fatalln("Failed to convert logo.gltf: ", err)
	}
	// Prepare and load model to GPU
	app.model, err = b.ToModel(vapp.Dev)
	if err != nil {
		log.Fatalln("Failed to load logo.gltf: ", err)
	}
	// Register model for disposal when application terminates
	vapp.AddChild(app.model)

	v := vdraw3d.NewView(vapp.Dev, drawStatic, drawDynamic)
	c := vscene.NewPerspectiveCamera(1000)
	c.Position = mgl32.Vec3{1, 2, 10}
	c.Target = mgl32.Vec3{5, 0, 0}
	v.Camera = vapp.OrbitControlFrom(0, nil, c)
	rw.AddView(v)
}

func buildUi(rw *vapp.ViewWindow) {

	// Create hello UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	err := mintheme.BuildMinTheme()
	if err != nil {
		log.Fatalln("Error loading theme: ", err)
	}
	// We must compile all ui shapes. Shapes are not precompiled because you can add custom shape primitives before compiling them
	err = vdraw.CompileShapes(vapp.Dev)
	if err != nil {
		log.Fatalln("Error compiling shapes: ", err)
	}
	// Add custom style info
	mintheme.Theme.Add(20, vimgui.Tags("info"), vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{0, 0.75, 1, 1})})
	// Create new UI view
	v := vimgui.NewView(vapp.Dev, vimgui.VMTransparent, mintheme.Theme, drawUi)
	rw.AddView(v)
}

func drawUi(fr *vimgui.UIFrame) {
	// Draw panel with title and content
	// First we need set control area directly
	fr.ControlArea = vdraw.Area{From: mgl32.Vec2{100, fr.DrawArea.To[1] - 300}, To: mgl32.Vec2{500, fr.DrawArea.To[1] - 100}}
	vimgui.Panel(fr, func(uf *vimgui.UIFrame) {
		// We can set control area also using NewLine and NewColumn helpers
		// Next control will have height of 30 and with of 100%. There will be 2 pixes padding to previous line (top)
		uf.NewLine(-100, 30, 2)
		// Set style for title
		uf.WithTags("h2")
		vimgui.Label(uf, "Hello VGE!")
	}, func(uf *vimgui.UIFrame) {
		uf.NewLine(-100, 20, 5)
		uf.WithTags("info")
		vimgui.Label(uf, "Use mouse with left button down to rotate view")
		uf.NewLine(-100, 20, 5)
		vimgui.Label(uf, "Use mouse with right button down to pan view")
		// Add new line height 30 pixes and padded 3 pixels from previous line
		uf.NewLine(120, 30, 3)
		vimgui.CheckBox(uf.WithTags(), vk.NewHashKey("cbLights"), "Lights", &app.lights)
		uf.NewLine(120, 30, 3)
		// Add new column with width 120 pixes
		uf.NewColumn(120, 0)
		// Add button. Button function will return true when button is clicked
		if vimgui.Button(uf.WithTags("primary"), vk.NewHashKey("bQuit"), "Quit") {
			// Terminate application in go routine. Calling terminate must be done in go routine to prevent deadlock
			go func() {
				vapp.Terminate()
			}()
		}
	})
}

var kShadow = vk.NewKey()
var kProbe = vk.NewKey()

func drawStatic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	// Draw background image
	vdraw3d.DrawBackground(dl, app.model, app.bgImage)
	app.probe = vdraw3d.DrawProbe(dl, kProbe, mgl32.Vec3{0, 0, 0})
	// Don't draw anything but background to probe
	dl.Exclude(app.probe, app.probe)

	// Draw all nodes starting from root (node == 0)
	vdraw3d.DrawNodes(dl, app.model, 0, mgl32.Ident4())
}

func drawDynamic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	// Don't draw anything but background to probe
	dl.Exclude(app.probe, app.probe)
	if app.lights {
		// Set properties for point light
		props := vmodel.NewMaterialProperties()
		props.SetColor(vmodel.CIntensity, mgl32.Vec4{1.4, 1.4, 1.4, 1})
		props.SetFactor(vmodel.FLightAttenuation2, 0.3)
		vdraw3d.DrawPointLight(dl, 0, mgl32.Vec3{1, 3, 4}, props)
		props.SetColor(vmodel.CIntensity, mgl32.Vec4{0, 0.4, 1.4, 1})
		vdraw3d.DrawPointLight(dl, kShadow, mgl32.Vec3{5, 4, 4}, props)
	}
}
```

Running logo.go with `go run logo.go` should produce something like. 

![alt text](docs/logo_example.png "Logo example")

## Sample videos

Some sample videos recorded from VGE example projects:
- [glTFViewer](https://youtu.be/MAgn8qudW-w) browser for glTF 2.0 samples.
- [Robomaze](https://youtu.be/RQ3mJl3lQ0Y) performance test animation.
- [Animation](https://youtu.be/FTCOo1gcA8I) BVH animation support (Experimental)

You can also install the VGE and run the same examples yourself!

## Installation

First: you need a Go compiler. Go version 1.17 or later is required to build VGE.
Some modules use new go:embed directive to import shaders and other assets to modules.


Install VGE like any go package:
`go get github.com/lakal3/VGE`

Note **Only 64bit (amd64) Windows (only tested on Windows 10) or Linux (experimental) is supported**. You must also have **updated** Vulkan drivers.

### Additional steps on Windows

Some lower level functions are implemented in C++. See [VGE architecture](docs/architecture.md) for more description of why and how VGE is implemented.

On Windows, you do not have to install the C/C++ compiler. You can use a prebuilt C++ vgelib.dll. Windows implementation does not use the CGO at all!

Copy prebuilt/win_amd64/vgelib.dll to some directory in your search PATH or update searchpath to include prebuilt directory.
Alternatively you can [build vgelib](docs/build_vgelib.md) vgelib.dll yourself.

If you plan to do any additional projects with Vulkan or the VGE, I also recommend you to install [Vulkan SDK](https://www.lunarg.com/vulkan-sdk/).
Vulkan SDK also contains the SPIR-V compiler that you will need when developing your own shaders.
Vulkan SDK also contains Vulkan validation layers that are really helpful in pinpointing possible errors in your API calls.
The VGE supports a simple way of enabling Vulkan validation layers.

### [Installation on Linux](docs/linux_install.md) (Experimental)

## Learn to use VGE

The VGE engine is really tiny compared to most of the graphics frameworks that support Vulkan.
However, there is still quite a lot of code and browsing the API help document is not very enlightening.

Perhaps the best approach for learning VGE is to browse through the examples included in the project.
These examples try to pinpoint most important aspects of VGE and it's features.

- Basic - The most basic sample that you should be able to `go run` if everything is installed ok.
- Basicn - The most basic sample that you should be able to `go run`. These are using new vimgui/vdraw3d modules
- Cube - A simple cube using the lower level API to draw cube on screen
- Model - Additional samples on how to load models and manipulate a scene.
- WebView - Render images in the background and serve them to the web client
(*The demo will show that Vulkan can do multithreaded rendering, and you can easily use Vulkan to render images in the background*)
- glTFViewer - A tool to browse some of the glTF sample scenes.
- Robomaze - A performance test tool / example that supports some advanced features like decals.
- Animate (Experimental) - Support for reading animations from the BVH (Biovision Hierarchy) files and apply them on rigged models.
- Gallery (New) - Demonstrates how to use vdraw drawing library and vimgui immediate mode user interface
- Fileviewer (New) - First somewhat usable application that can browse file system and
  view different kind of files like images (jpeg, png, hdr) and 3D models (obj, glTF)
  - Fileviewer also implement custom View to render images for new ViewWindow 

The VGE documentation does not go into the details of the Vulkan API.
To lean the core Vulkan features, below I have listed some nice web articles about it.
- [https://gpuopen.com/understanding-vulkan-objects/] - Good overview of Vulkan object model
- [https://vulkan-tutorial.com/] - (Be prepared to write >600 LoC C/C++ to draw a triangle on screen)
- [https://software.intel.com/en-us/articles/api-without-secrets-introduction-to-vulkan-part-1]

## Features of the VGE

This is a short list of the features (existing). [Roadmap](docs/roadmap.md) has a list of planned features.

- Handling VGE [core objects](docs/vk.md)
- [Image](docs/vimage.md) loading and handling
- [Model](docs/vmodel.md) building and loading. VGE currently supports OBJ and glTF 2.0 formats.
- Basic materials: Unlit, Phong and PBR [Model](docs/vmodel.md)
- Lights: Point and directional, including shadows for point lights
- Scene building and animation [Scene](docs/vscene.md)
- Glyph rendering including support for TTF fonts in [Vector drawing library](docs/vdraw.md)
- Background environment using (HDR) equirectangular images [Scene](docs/vscene.md)
- An environmental probe to support reflections of PBR materials.
Probe will also generate spherical harmonics to approximate ambient lightning.  [Scene](docs/vscene.md)
- Easily extendable game UI support [Immediate mode UI](docs/vimgui.md) (<s>[VGE UI](docs/vui.md)</s>).
- UI theming [Immediate mode UI](docs/vimgui.md) (<s>[VGE UI](docs/vui.md)</s>).
- Event queue to handle system level events [VGE App](docs/vapp.md)
- Window, mouse, keyboard etc management  (Uses GLFW C library for multi platform support)
- High level, easier to use setup for basic applications [VGE App](docs/vapp.md)
- Multi threaded [Architecture](docs/architecture.md)
- Rendering without UI (for example to quickly render images in a web server!)
- Calculating algorithms in GPU. VGE supports Vulkan compute shaders. (See for example vglyph/vdepth shader)
- (Experimental) Decal support [Scene](docs/vscene.md)
- (Experimental) Apply BVH animations to existing rigged models
- Integrated glsl -> SPIR-V compiler
- (New) [Vector drawing library](docs/vdraw.md)
- (New) [Immediate mode UI](docs/vimgui.md)
- (New) MultiView support (mix several UI/3D/Custom views in one window)
- (New) [3D Drawing library](docs/vscene.md)
- (New) Tool to compile several glsl fragments into shader packs [vgecompile](docs/vgecompile.md)



## Status

VGE started when I wanted to learn Vulkan programming.
I gradually added new features to handle different aspects of Graphics programming.

Now VGE has all the basic elements to make Vulkan based 3D programs using Go.
The examples give a quite nice overview of the existing features in VGE.

Features not marked as preview or experimental should be considered fairly stable. 
However, breaking API changes are still likely need to support new features.

If you like to use VGE for more serious work, I recommend that you make own fork of project. 

**Note that the number of different devices VGE has been tested on is limited and
Vulkan drivers have bugs and inconsistencies. Please report any device incompatibilities you find.**

[Roadmap](docs/roadmap.md) lists more planned features. [Changes](changes.md) will contain all important changes between versions.

## [MIT License](license.md)

## [Credits](docs/credits.md)

