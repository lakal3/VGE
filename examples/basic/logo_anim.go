//+build examples

package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vapp/vdebug"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
)

func main() {
	var devIndex int
	// Use devIndex to select device. If not given, use first device
	flag.IntVar(&devIndex, "dev", 0, "Device index")
	flag.Parse()

	// Override default
	vapp.SelectDevice = func(devices []vk.DeviceInfo) int32 {
		if len(devices) < devIndex {
			log.Fatal("No device ", devIndex)
		}
		fmt.Println("Selected device ", string(devices[devIndex].Name[:devices[devIndex].NameLen]))
		return int32(devIndex)
	}

	// Set loader for assets (images, models). This assume that current directory is same where hello.go is!
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	vapp.Init("hello", vapp.Validate{}, vapp.Desktop{})

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	rw := vapp.NewRenderWindow("hello", forward.NewRenderer(true))
	// Build scene
	buildScene(rw)
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

func buildScene(rw *vapp.RenderWindow) {
	// Load envhdr/studio.hdr and create background from it.
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

	// Create new root node for logo
	nm := vscene.NewNode(rotateCtrl{speed: 0.5},
		// Move logos center to origo
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(-5, 0, 0)},
			// Create a new nodes from model
			vscene.NodeFromModel(model, 0, true)))

	// Append nodes to scene
	rw.Model.Children = append(rw.Model.Children, nm)
	// We will also need a probe to reflect environment to model. Probes reflect everything outside this node inside children of this node.
	// In this case we reflect only background
	p := env.NewProbe(vapp.Ctx, vapp.Dev)
	rw.AddChild(p) // Remember to dispose probe
	// Assign probe to root model
	rw.Model.Ctrl = p

	// Attach camera to window (with better location that default one) and orbital control to camera
	c := vscene.NewPerspectiveCamera(1000)
	c.Position = mgl32.Vec3{1, 2, 12}
	c.Target = mgl32.Vec3{0, 0, 0}
	rw.Camera = c
	// Add orbital controls to camera. If priority > 0 panning and scrolling will work event if mouse is on UI. UI default show priority is 0
	vapp.OrbitControlFrom(-10, rw, c)

	// Finally create 2 lights before UI
	// Create custom node control to turn light on / off
	visible := &nodeVisible{}
	nLight := vscene.NewNode(vscene.NewMultiControl(visible, rotateCtrl{-1}))
	rw.Env.Children = append(rw.Env.Children, nLight)
	// First light won't cast shadows, second will
	l1 := &vscene.PointLight{Intensity: mgl32.Vec3{1.4, 1.4, 1.4}, Attenuation: mgl32.Vec3{0, 0, 0.3}}
	l2 := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{0, 1.4, 1.4}, Attenuation: mgl32.Vec3{0, 0, 0.2}}, 512)

	// Add shadow light to scene on location -3,3,3 and 3,3,3
	nLight.Children = append(nLight.Children,
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(-3, 3, 3)}, vscene.NewNode(l1)),
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(3, 3, 3)}, vscene.NewNode(l2)))
	// Create UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	th := mintheme.NewTheme(vapp.Ctx, vapp.Dev, 15, nil, nil, nil)
	// Add theme to RenderWindow dispose list. In real app we might use theme multiple times on multiple windows and should handling disposing it
	// as part of disposing device.
	rw.AddChild(th)
	var bQuit *vui.Button
	var lTime *vui.Label
	ui := vui.NewUIView(th, image.Rect(100, 500, 500, 720), rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("Hello VGE!").SetClass("h2"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewLabel("Scene time").AssignTo(&lTime),
			vui.NewCheckbox("Lights", "").SetOnChanged(func(checked bool) {
				visible.visible = checked
			}).SetClass("dark"),
			vui.NewCheckbox("Pause", "").SetOnChanged(func(checked bool) {
				rw.SetPaused(checked)
			}).SetClass("dark"),
			&vui.Extend{MinSize: image.Pt(10, 10)}, // Some spacing
			vui.NewButton(120, "Quit").SetClass("warning").AssignTo(&bQuit),
		)).SetClass(""))
	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}
	go updateTime(rw, lTime)
	// Attach UI to scene and show it. UI panel are by default invisible and must be show
	rw.Ui.Children = append(rw.Ui.Children, vscene.NewNode(ui))
	ui.Show()

	// Add FPS debug UI
	vdebug.NewFPSTimer(rw, th)
}

func updateTime(rw *vapp.RenderWindow, tl *vui.Label) {
	// Run go thread until window is closed
	for !rw.Closed() {
		<-time.After(100 * time.Millisecond)
		// Always update live scene using scene.Update. If will schedule update to proper time if scene is currently in
		// read only mode. Visible UI:s are part of scene
		rw.Scene.Update(func() {
			tl.Text = fmt.Sprintf("Scene time %.2f s", rw.GetSceneTime())
		})
	}
}

type nodeVisible struct {
	visible bool
}

func (n *nodeVisible) Process(pi *vscene.ProcessInfo) {
	pi.Visible = n.visible
}

type rotateCtrl struct {
	speed float64
}

func (r rotateCtrl) Process(pi *vscene.ProcessInfo) {
	// Rotate object around Y axis with given speed. Not that rotate angles are in radians
	pi.World = pi.World.Mul4(mgl32.HomogRotate3DY(float32(pi.Time * r.speed)))
}
