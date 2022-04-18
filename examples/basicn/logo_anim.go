//go:build examples

package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vdraw3d"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vscene"
	"log"
	"math"
	"os"
)

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

	// ps, _ := phongshader.LoadPack()
	// v := vdraw3d.NewCustomView(vapp.Dev, ps, drawStatic, drawDynamic)
	v := vdraw3d.NewView(vapp.Dev, drawStatic, drawDynamic)
	c := vscene.NewPerspectiveCamera(1000)
	c.Position = mgl32.Vec3{1, 2, 10}
	c.Target = mgl32.Vec3{5, 0, 0}
	v.Camera = vapp.OrbitControlFrom(c)
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

}

func drawDynamic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	dl.Exclude(app.probe, app.probe)
	w := mgl32.Translate3D(4, 0, 0).Mul4(mgl32.HomogRotate3DY(float32(v.Elapsed / 2))).Mul4(mgl32.Translate3D(-4, 0, 0))
	if app.lights {
		// Set properties for point light
		props := vmodel.NewMaterialProperties()
		props.SetColor(vmodel.CIntensity, mgl32.Vec4{1.4, 1.4, 1.4, 1})
		props.SetFactor(vmodel.FLightAttenuation2, float32(1+math.Sin(v.Elapsed*3.2)*0.7))
		vdraw3d.DrawPointLight(dl, 0, mgl32.Vec3{1, 3, 3}, props)
		props.SetColor(vmodel.CIntensity, mgl32.Vec4{0, 3, 3, 1})
		at := mgl32.Vec3{float32(4 + 2*math.Sin(v.Elapsed/1.73)), float32(2 + math.Sin(v.Elapsed/3.42)), 3}
		// vdraw3d.DrawPointLight(dl, kShadow, w.Mul4x1(at.Vec4(1)).Vec3(), props)
		vdraw3d.DrawPointLight(dl, kShadow, at, props)
	}
	// Draw all nodes starting from root (node == 0)
	vdraw3d.DrawNodes(dl, app.model, 0, w)
}
