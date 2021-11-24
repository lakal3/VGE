// Cube example show how to draw a single cube using only vk functions and vmodel to build a cube

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/unlit"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

var config struct {
	debug bool
	dev   int
}

type cubeApp struct {
	owner   vk.Owner
	m       *vmodel.Model
	app     *vk.Application
	dev     *vk.Device
	frp     *vk.ForwardRenderPass
	desktop *vk.Desktop
	yaw     float64
	pitch   float64
	scale   float64
	running bool
}

func (v *cubeApp) Dispose() {
	v.owner.Dispose()
}

var app cubeApp

func main() {
	flag.BoolVar(&config.debug, "debug", false, "Use debug layers")
	flag.IntVar(&config.dev, "dev", 0, "Select device number")
	flag.Parse()

	if config.debug {
		// vk.VGEDllPath = "vgelibd.dll"
	}
	app.init()
	defer app.Dispose()
	err := app.buildModel()
	if err != nil {
		log.Fatal("Build model failed ", err)
	}
	app.pitch = 0.3
	app.yaw = 0.2
	app.scale = 3
	w := app.desktop.NewWindow("Cube", vk.WindowPos{Left: -1, Top: -1, Width: 1024, Height: 768})
	app.owner.AddChild(w)
	go app.renderLoop(w)
	app.running = true
	app.run()
	app.running = false
	<-time.After(100 * time.Millisecond)
	app.Dispose()

}

func (a *cubeApp) init() (err error) {
	// Create a new application
	a.app, err = vk.NewApplication("cube")
	if err != nil {
		return err
	}
	if config.debug {
		// Add validation layer if requested
		a.app.AddValidation()
	}
	// And new desktop handler. This must be created before application is initialized
	// Desktop will register operating system dependent Vulkan extensions that are required to show rendered images
	a.desktop = vk.NewDesktop(a.app)
	a.app.Init()
	// Register application for disposing
	a.owner.AddChild(a.app)
	// Check that device exists.
	// If we don't have device 0 then there is no Vulkan driver available that support all features we requested for
	pds := a.app.GetDevices()
	if len(pds) <= config.dev {
		log.Fatal("No device ", config.dev)
	}
	fmt.Println("Using device ", string(pds[config.dev].Name[:pds[config.dev].NameLen]))
	// Create a new device
	a.dev = a.app.NewDevice(int32(config.dev))
	return nil
}

func (v *cubeApp) buildModel() (err error) {
	mb := vmodel.ModelBuilder{}
	mb.ShaderFactory = unlit.UnlitFactory
	// Create 1x1x1 cube
	meshBuilder := &vmodel.MeshBuilder{}
	meshBuilder.AddCube(mgl32.Ident4())
	// Create blue material
	mat := mb.AddMaterial("blue", vmodel.NewMaterialProperties().SetColor(vmodel.CAlbedo, mgl32.Vec4{0, 0, 1, 1}))
	// Add mesh and node to model builder
	meshIndex := mb.AddMesh(meshBuilder)
	mb.AddNode("cube", nil, mgl32.Ident4()).SetMesh(meshIndex, mat)

	// Create model will prepare model struct. Model will contain a buffer that have all attributes required by VGE standard shaders
	// including position, normal, tangent and uv0. Model will also copy all attributes to device memory.
	v.m, err = mb.ToModel(v.dev)
	if err != nil {
		return err
	}
	// We must dispose model when no longer needed.
	v.owner.AddChild(v.m)
	return nil
}

// Run loop
func (v *cubeApp) run() {
	var prevX, prevY int32
	var mb1 bool
	for true {
		// Pull desktop event
		ev, _ := v.desktop.PullEvent()
		if ev.EventType == 0 {
			<-time.After(time.Millisecond)
			continue
		}
		fmt.Println("Event ", ev)

		// quit if key q was pressed
		if ev.EventType == 202 && ev.Arg1 == int32('q') { // Char event
			return
		}
		if ev.EventType == 101 {
			return
		}

		// Button 1 down
		if ev.EventType == 301 && ev.Arg1 == 0 {
			mb1 = true
			prevX, prevY = 0, 0
		}
		// Button 1 up
		if ev.EventType == 300 && ev.Arg1 == 0 {
			mb1 = false
		}
		if ev.EventType == 302 && mb1 { // Mouse move
			if prevX != 0 {
				dx := ev.Arg1 - prevX
				v.yaw += float64(dx) / 10
			}
			prevX = ev.Arg1
			if prevY != 0 {
				dy := ev.Arg2 - prevY
				v.pitch += float64(dy) / 10
			}
			prevY = ev.Arg2
		}
		// Just write any pressed keys to console. You can check key scan codes from console
		if ev.EventType == 200 {
			fmt.Println("Key ", ev.Arg1, " ", ev.Arg2)
		}
		<-time.After(time.Millisecond)
	}
}

// Render images to window
func (v *cubeApp) renderLoop(w *vk.Window) {
	var caches []*vk.RenderCache
	<-time.After(100 * time.Millisecond)
	for v.running {
		// Aquire next image
		im, imageIndex, submitInfo := w.GetNextFrame(v.dev)
		// If image index < 0, there was some error acquiring image. Most likely window was resized which in many cases mean that we
		// must destroy existing rendering assets like depth buffers etc and recreate them. Next call to GetNextFrame should again succeed
		if imageIndex < 0 {
			// Reset all render caches
			for _, ca := range caches {
				if ca != nil {
					ca.Dispose()
				}
			}
			caches = nil
			continue
		}
		if len(caches) == 0 {
			// Recreate render caches. Render caches stores important assets that are valid as long as render image is valid.
			// For example same depth buffer can be used as long as image is same
			// Render cache also support per frame assets than will be disposed after frame has been rendered
			//
			// We must create individual render cache for each image. Typically we have 2 - 4 images per window that Vulkan swapchain
			// api will cycle during rendering process
			caches = make([]*vk.RenderCache, w.GetImageCount())
		}
		if caches[imageIndex] == nil {
			caches[imageIndex] = vk.NewRenderCache(v.dev)
			if v.frp == nil {
				// Also create a render pass. We can't create render pass before we get an actual image as render pass
				// must match window image format
				v.frp = vk.NewForwardRenderPass(v.dev, w.WindowDesc.Format, vk.IMAGELayoutPresentSrcKhr, vk.FORMATUndefined)
				v.owner.AddChild(v.frp)
			}

		}
		// Get proper cache
		rc := caches[imageIndex]
		// Start new frame
		rc.NewFrame()
		frame := v.setCamera(rc, w.WindowDesc)
		// 1.20< dc := &vmodel.DrawContext{Cache: rc, Pass: v.frp}
		dc := &vmodel.DrawContext{Frame: frame, Pass: v.frp}
		// Retrieve view for image. If cache already has this value it is just returned. Otherwise we construct it using
		// given constructor function. This is standard way to create assets in VGE. Get api will handle disposing created items
		// when render cache is disposed. There is separate struct Owner in vk module if you want to implement same pattern in your
		// own classes
		iv := rc.Get(kImageView, func() interface{} {
			// Construct default view of image
			return im.NewView(0, 0)
		}).(*vk.ImageView)
		// Render image
		cmd := v.render(dc, iv)
		// Submit rendering. In case of swap chain images, we must pass submitinfo from GetNextFrame to Submit. This information
		// instructs VGE to present image after it has been rendered
		cmd.Submit(submitInfo)
		// Just wait that command in completed. This is the point where rendering actually happened. In more advances application you
		// could delay waiting and do other task while GPU is processing commands.
		cmd.Wait()
	}
	for _, ca := range caches {
		if ca != nil {
			ca.Dispose()
		}
	}
}

var kFp = vk.NewKey()
var kCmd = vk.NewKey()
var kImageView = vk.NewKeys(5)

func (v *cubeApp) render(dc *vmodel.DrawContext, mainView *vk.ImageView) *vk.Command {
	rc := dc.Frame.GetCache()
	// Get or create frame buffer from main image view
	fb := rc.Get(kFp, func() interface{} {
		return vk.NewFramebuffer(dc.Pass, []*vk.ImageView{mainView})
	}).(*vk.Framebuffer)
	// Create command suitable for graphics (render pass)
	cmd := rc.Get(kCmd, func() interface{} {
		return vk.NewCommand(rc.Device, vk.QUEUEGraphicsBit, false)
	}).(*vk.Command)
	// Start command, bust be always first call to command buffer
	cmd.Begin()
	// Start render pass
	cmd.BeginRenderPass(v.frp, fb)
	dc.List = &vk.DrawList{}
	root := v.m.GetNode(0)
	// Record all rendering command to draw list
	v.renderNodes(root, dc, mgl32.Ident4())
	// VGE uses a list of draw commands that are passed to C++/VGE in single call. This allows VGE to reduce somewhat
	// slower call from golang to C(++). Typically single drawlist can draw nearly whole scene.
	cmd.Draw(dc.List)
	// End render pass. We don't call end to cmd. This is done by cmd.Submit automatically
	cmd.EndRenderPass()
	return cmd
}

// Just recursively loop through model nodes and draw them
func (v *cubeApp) renderNodes(node vmodel.Node, dc *vmodel.DrawContext, world mgl32.Mat4) {
	world = world.Mul4(node.Transform)
	if node.Mesh >= 0 {
		mat := node.Model.GetMaterial(node.Material)
		mat.Shader.Draw(dc, node.Model.GetMesh(node.Mesh), world, nil)
	}
	for _, ch := range node.Children {
		v.renderNodes(node.Model.GetNode(ch), dc, world)
	}
}

// Transform from opengl coordinates to Vulkan one
var VulkanProj = mgl32.Mat4{1, 0, 0, 0, 0, -1, 0, 0, 0, 0, 0.5, 0.5, 0, 0, 0, 1}

func (v *cubeApp) setCamera(rc *vk.RenderCache, mainImage vk.ImageDescription) *vscene.SimpleFrame {
	// Frame object is stored in render cache. VGE Shaders can access this information from cache.
	// Frame must be typically at start of rendering. See unlit implementation for more details
	f := &vscene.SimpleFrame{Cache: rc}
	// Calculate camera view and projection. Math left as an exercise to reader
	aspect := float32(mainImage.Width) / float32(mainImage.Height)
	eyePos := mgl32.Vec3{
		float32(app.scale * math.Sin(v.yaw) * math.Cos(v.pitch)),
		float32(app.scale * math.Cos(v.yaw) * math.Cos(v.pitch)),
		float32(app.scale * math.Sin(v.pitch)),
	}.Vec4(1)
	proj := mgl32.Perspective(1, aspect, 0.01, 100)
	f.SSF.Projection = VulkanProj.Mul4(proj)
	f.SSF.View = mgl32.LookAtV(eyePos.Vec3(), mgl32.Vec3{}, mgl32.Vec3{0, 1, 0})
	// f.Far = 100
	return f
}
