package main

import (
	"fmt"
	"github.com/lakal3/vge/vge/forward"
	"image"
	"log"
	"math"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/pbr"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vscene"
)

func initApp(ctx initContext) error {
	wwApp.app = vk.NewApplication(ctx, "WWW render")
	if wwApp.debug {
		// Add validation layer if requested
		wwApp.app.AddValidation(ctx)
	}
	// Initialize application and register it for disposal at end of application
	wwApp.app.Init(ctx)
	wwApp.owner.AddChild(wwApp.dev)

	// Needed to png images
	pngloader.RegisterPngLoader()

	// Check that device exists.
	// If we don't have device 0 then there is no Vulkan driver available that support all features we requested for
	pds := wwApp.app.GetDevices(ctx)
	if len(pds) <= wwApp.pdIndex {
		log.Fatal("No device ", wwApp.pdIndex)
	}
	fmt.Println("Using device ", string(pds[wwApp.pdIndex].Name[:pds[wwApp.pdIndex].NameLen]))

	// Create a new device and register it for disposal at end of application
	wwApp.dev = wwApp.app.NewDevice(ctx, int32(wwApp.pdIndex))
	wwApp.owner.AddChild(wwApp.dev)

	// Create render pass with depth buffer support. Final image will be copied so final layout is IMAGELayoutTransferSrcOptimal
	wwApp.rp = vk.NewForwardRenderPass(ctx, wwApp.dev, vk.FORMATR8g8b8a8Unorm, vk.IMAGELayoutTransferSrcOptimal, vk.FORMATD32Sfloat)

	// Load model to render
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets/gltf/elf"}
	// Initialize model builder with pbr shader. Pbr or std shader is best suited to visualize glTF models
	// Std shader requires dynamics descriptors extension that we don't want to request in this example app
	mb := vmodel.ModelBuilder{ShaderFactory: pbr.PbrFactory}
	ol := gltf2loader.GLTF2Loader{Builder: &mb, Loader: vasset.DefaultLoader}
	err := ol.LoadGltf(filepath.Base("elf.gltf"))
	if err != nil {
		return err
	}

	// GLTF file can have multiple scenes. We convert each scene individually
	err = ol.Convert(0)
	if err != nil {
		return err
	}

	// Create model
	wwApp.model = mb.ToModel(ctx, wwApp.dev)
	wwApp.owner.AddChild(wwApp.model)
	return nil
}

// Camera orbit is dependant of rendered image size. Elf model is about 25 units high
const cameraOrbit = 50.0

type imageRenderer struct {
	rc *vk.RenderCache
}

func (i imageRenderer) GetPerRenderer(key vk.Key, ctor func(ctx vk.APIContext) interface{}) interface{} {
	return i.rc.Get(key, ctor)
}

func renderImage(ctx vk.APIContext, angle float64, imageSize image.Point) (pngImage []byte) {
	// Initialize render cache to handle frame related object
	rc := vk.NewRenderCache(ctx, wwApp.dev)
	defer rc.Dispose()
	// Pool for images
	pool := vk.NewMemoryPool(wwApp.dev)
	ddMain := vk.ImageDescription{Width: uint32(imageSize.X), Height: uint32(imageSize.Y),
		Depth: 1, MipLevels: 1, Layers: 1, Format: vk.FORMATR8g8b8a8Unorm}
	ddDepth := ddMain
	ddDepth.Format = vk.FORMATD32Sfloat
	// Main image is used for rendering (IMAGEUsageInputAttachmentBit) and then we copy image (IMAGEUsageTransferSrcBit)
	mainImg := pool.ReserveImage(ctx, ddMain, vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageColorAttachmentBit)
	depthImg := pool.ReserveImage(ctx, ddDepth, vk.IMAGEUsageDepthStencilAttachmentBit)
	// Copy buffer where we can copy final image
	imLen := imageSize.Y * imageSize.X * 4
	copyBuffer := pool.ReserveBuffer(ctx, uint64(imLen), true, vk.BUFFERUsageTransferDstBit)
	// Allocate reserved elements
	pool.Allocate(ctx)

	rc.DisposePerFrame(pool)

	// Build scene
	sc := vscene.Scene{}
	sc.Init()

	// Ambient light
	al := vscene.NewNode(&env.AmbientLight{Intensity: mgl32.Vec3{0.7, 0.7, 0.7}})

	// Gray background
	bgNode := vscene.NewNode(env.NewGrayBG())

	// Copy whole model (node 0 == root) to node
	elfNode := vscene.NodeFromModel(wwApp.model, 0, true)

	// Build scene
	sc.Root.Children = append(sc.Root.Children, al, bgNode, elfNode)

	// Allocate frame buffer that attached depth and main images to rendering
	fb := vk.NewFramebuffer(ctx, wwApp.rp, []*vk.ImageView{mainImg.DefaultView(ctx), depthImg.DefaultView(ctx)})
	rc.DisposePerFrame(fb)

	// Allocate command buffer to
	cmd := vk.NewCommand(ctx, wwApp.dev, vk.QUEUEGraphicsBit, true)

	// Render scene
	cmd.Begin()

	pc := vscene.NewPerspectiveCamera(1000)
	pc.Target = mgl32.Vec3{0, cameraOrbit / 2, 0}
	pc.Position = mgl32.Vec3{float32(math.Sin(angle) * cameraOrbit), cameraOrbit / 2, float32(cameraOrbit * math.Cos(angle))}
	// Update camera projection and view matrix to current frame
	frame := forward.NewFrame(rc, imageRenderer{rc: rc})
	frame.SF.Projection, frame.SF.View = pc.CameraProjection(imageSize)
	frame.SF.EyePos = frame.SF.View.Inv().Col(3)

	// We only need predraw phase, background phase and 3d phase
	bg := vscene.NewDrawPhase(frame, wwApp.rp, vscene.LAYERBackground, cmd, func() {
		cmd.BeginRenderPass(wwApp.rp, fb)
	}, nil)
	dp := vscene.NewDrawPhase(frame, wwApp.rp, vscene.LAYER3D, cmd, nil, func() {
		cmd.EndRenderPass()
	})

	// Predraw phase draws shadow maps etc.
	ppPhase := &vscene.PredrawPhase{Scene: &sc, Cmd: cmd}
	sc.Process(sc.Time, frame, &vscene.AnimatePhase{Cache: rc}, ppPhase, &forward.FrameLightPhase{F: frame, Cache: rc}, bg, dp)

	// Process pending request from predraw phase
	for _, pd := range ppPhase.Pending {
		pd()
	}
	r := mainImg.FullRange()
	r.Layout = vk.IMAGELayoutTransferSrcOptimal
	// Finally copy image to buffer
	cmd.CopyImageToBuffer(copyBuffer, mainImg, &r)
	// Pre render phases bay queue rendering of shadow maps etc. ppPhase need must be completed before we render main image
	cmd.Submit(ppPhase.Needeed...)
	cmd.Wait()
	// Save buffer to png image
	return vasset.SaveImage(ctx, "png", mainImg.Description, copyBuffer)
}
