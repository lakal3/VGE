# VGE (Vulkan Graphics Engine)

VGE is graphics engine for [Vulkan](https://www.khronos.org/vulkan/). It is mostly written in Go language.


## VGE code example (logo.go)

Simple example showing simple user interface and 3D model on screen. See full sample with import in examples/basic/logo.go

```go
func main() {
	// Set loader for assets (images, models). This assume that current directory is same where hello.go is!
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	vapp.Init("hello", vapp.Validate{}, vapp.Desktop{})

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	rw := vapp.NewRenderWindow("hello", vapp.NewForwardRenderer(true))
	// Build scene
	buildScene(rw)
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

func buildScene(rw *vapp.RenderWindow) {
	// Load envhdr/studio.hdr and create background "skybox" from it. VGE can create whole 360 background
	// from full 360 / 180 equirectangular image without needing 6 images for full cube
	// MustLoadAsset will handle loading loading actual asset using vasset.DefaultLoader set in start of program
	// MustLoadAsset will also handle ownership of asset (if will be disposed with device)
	eq := vapp.MustLoadAsset("envhdr/studio.hdr",
		func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
	// Add loaded background to scene
	rw.Env.Children = append(rw.Env.Children, vscene.NewNode(eq))

	// Load actual model
	model, err := vapp.LoadModel("gltf/logo/Logo.gltf")
	if err != nil {
		log.Fatal("Failed to load gltf/logo/Logo.gltf")
	}

	// Again, register model ownership to window
	rw.AddChild(model)
	// Create a new nodes from model
	rw.Model.Children = append(rw.Model.Children, vscene.NodeFromModel(model, 0, true))
	// We will also need a probe to reflect environment to model. Probes reflect everything outside this node inside children of this node.
	// In this case we reflect only background
	p := env.NewProbe(vapp.Ctx, vapp.Dev)
	rw.AddChild(p) // Remember to dispose probe
	// Assign probe to root model
	rw.Model.Ctrl = p

	// Attach camera to window (with better location that default one) and orbital control to camera
	c := vscene.NewPerspectiveCamera(1000)
	c.Position = mgl32.Vec3{1, 2, 10}
	c.Target = mgl32.Vec3{5, 0, 0}
	rw.Camera = c
	// Add orbital controls to camera. If priority > 0 panning and scrolling will work event if mouse is on UI. UI default show priority is 0
	vapp.OrbitControlFrom(-10, rw, c)

	// Finally create 2 lights before UI
	// Create custom node control to turn light on / off
	visible := &nodeVisible{}
	nLight := vscene.NewNode(visible)
	rw.Env.Children = append(rw.Env.Children, nLight)
	// First light won't cast shadows, second will
	l1 := &vscene.PointLight{Intensity: mgl32.Vec3{1.4, 1.4, 1.4}, Attenuation: mgl32.Vec3{0, 0, 0.3}}
	l2 := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{0, 1.4, 1.4}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512)

	// Add shadow light to scene on location 1,3,3 and 4,3,3
	nLight.Children = append(nLight.Children,
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(1, 3, 3)}, vscene.NewNode(l1)),
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(6, 3, 3)}, vscene.NewNode(l2)))
	// Create UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	th := mintheme.NewTheme(vapp.Ctx, vapp.Dev, 15, nil, nil, nil)
	// Add theme to RenderWindow dispose list. In real app we might use theme multiple times on multiple windows and should handling disposing it
	// as part of disposing device.
	rw.AddChild(th)
	var bQuit *vui.Button
	ui := vui.NewUIView(th, image.Rect(100, 500, 500, 700), rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("Hello VGE!").SetClass("h2"),
			vui.NewLabel("Use mouse with left button down to rotate view").SetClass("info"),
			vui.NewLabel("Use mouse with right button down to pan view").SetClass("info"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewCheckbox("Lights", "").SetOnChanged(func(checked bool) {
				visible.visible = checked
			}).SetClass("dark"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewButton(120, "Quit").SetClass("warning").AssignTo(&bQuit),
		)).SetClass(""))
	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}
	// Attach UI to scene and show it. UI panel are by default invisible and must be show
	rw.Ui.Children = append(rw.Ui.Children, vscene.NewNode(ui))
	ui.Show()
}

type nodeVisible struct {
	visible bool
}

func (n *nodeVisible) Process(pi *vscene.ProcessInfo) {
	pi.Visible = n.visible
}
```

Running logo.go with `go run logo.go` should produce something like

![alt text](docs/logo_example.png "Logo example")

## Sample videos

Few sample videos recorded from VGE example projects:
- [glTFViewer](https://youtu.be/MAgn8qudW-w) browser for glTF 2.0 samples.
- [Robomaze](https://youtu.be/RQ3mJl3lQ0Y) performance test animation.
- [Animation](https://youtu.be/FTCOo1gcA8I) BVH animation support (Experimental)

You can also install VGE and run same examples your self!

## Installation

First: you need Go compiler. VGE has been mostly tested with Go1.13 and Go1.14. 
Older version may work but I recommend using latest released version.

Install VGE like any go package:
`go get github.com/lakal3/VGE`

Note **Only 64bit (amd64) Windows (only tested on Windows 10) or Linux (experimental) is supported**. You must also have **Uptodate** Vulkan drivers.

### Additional steps on Windows

Some lower level functions are implemented in C++. See [VGE architecture](docs/architecture.md) for more description of why and how VGE is implemented.
 
On Windows, you don't have to install C/C++ compiler. Instead you can use prebuilt C++ VGELib.dll. Windows implementation don't use CGO at all! 
 
Copy prebuilt/win_amd64/VGELib.dll to some directory in your search PATH or update searchpath to include prebuilt directory.
Alternatively you can [build](docs/build_vgelib.md) VGELib.dll yourself.

If you plan to do any more thing on Vulkan or VGE, I also recommend to installing [Vulkan SDK](https://www.lunarg.com/vulkan-sdk/). 
Vulkan SDK contains SPIR-V compiler that you will need when developing you own shaders. 
Vulkan SDK also contains Vulkan validation layers that are really helpful in pinpointing possible errors in you API calls. 
VGE supports a simple way of enabling Vulkan validation layers.

### [Installation on Linux](docs/linux_install.md) (Experimental)
 
## Learn to use VGE

VGE engine is really tiny compared to most of graphics frameworks that support Vulkan. 
However there is still quite a lot of code and browsing API help document isn't that fruitful. 

Maybe best approach to learn VGE is browse through examples included in project. 
These examples try to pinpoint most important aspects of VGE and it's features.

- Basic - Most basic samples that you should be able to `go run` if everything is installed ok.
- Cube - Simple cube using lower level API to draw cube on screen
- Model - Additional samples on how to load models and manipulate scene.
- WebView - Render images background and serve them to web client 
(*Demo will show that Vulkan can do multithreaded rendering and you can easilly use Vulkan to render images background*)
- glTFViewer - Tool to browse some of glTF sample images.  
- Robomaze - Performance test tool / example that support some advanced features like decals.
- Animate (Experimental) - Support reading animations from BVH (Biovision Hierarchy) files and apply them on rigged models.

VGE documentation don't go into details of Vulkan API. 
To lean core Vulkan features, here are some nice web articles about it. 
- [https://gpuopen.com/understanding-vulkan-objects/] - Good overview of Vulkan object model
- [https://vulkan-tutorial.com/] - (Be prepared to write >600 LoC C/C++ to draw a triangle on screen)
- [https://software.intel.com/en-us/articles/api-without-secrets-introduction-to-vulkan-part-1]

## Features of VGE

This is short list of features (existing). [Roadmap](docs/roadmap.md) have a list of planned features.

- Handling VGE [core objects](docs/vk.md)
- [Image](docs/vimage.md) loading and handling
- [Model](docs/vmodel.md) building and loading. VGE currently support OBJ and glTF 2.0 formats.
- Basic materials: Unlit, Phong and PBR [Model](docs/vmodel.md)
- Lights: Point and directional, including shadows for point lights 
- Scene building and animation [Scene](docs/vscene.md)
- Glyph rendering including support for TTF fonts. Mainly used to support [VGE UI](docs/vui.md).
- Background environment using (HDR) equirectangular images [Scene](docs/vscene.md)
- Environmental probe to support reflections of PBR materials. Probe will also generate spherical harmonics to approximate ambient lightning.  [Scene](docs/vscene.md)
- Easily extendable game UI support [VGE UI](docs/vui.md).
- UI theming [VGE UI](docs/vui.md).
- Event queue to handle system level events [VGE App](docs/vapp.md)
- Window, mouse, keyboard etc management  (Uses GLFW C library for multi platform support)
- High level, easier to use setup for basic applications [VGE App](docs/vapp.md)
- Multi threaded [Architecture](docs/architecture.md)
- Rendering without UI (for example to quickly render images in a web server!)
- Calculating algorithms in GPU. VGE supports Vulkan compute shaders. (See for example vglyph/vdepth shader)
- (Experimental) Decal support [Scene](docs/vscene.md)
- (Experimental) Apply BVH animations to existing rigged models


## Status

VGE started when I wanted to learn Vulkan programming. 
I gradually added new features to handle different aspects of Graphics programming.

Now VGE has all the basic elements to make Vulkan based 3D programs using Go. 
Examples give quite nice overview of existing features in VGE.

Features not marked as preview or experimental should be considered fairly stable and breaking API changes will happen only if really necessary.

**Note that number of different devices VGE has been tested on is limit and 
Vulkan drivers have bugs and inconsistencies. Please report device incompatibilities you find out.**

[Roadmap](docs/roadmap.md) lists more planned features. [Changes](changes.md) will contain all important changes between versions.

## [MIT License](license.md)

## [Credits](docs/credits.md) 

