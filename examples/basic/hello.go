package main

import (
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
	"image"
)

func main() {
	// Set loader for assets (images, models). This assume that current directory is same where hello.go is!
	vasset.DefaultLoader = vasset.DirectoryLoader{Directory: "../../assets"}

	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	vapp.Init("logo", vapp.Validate{}, vapp.Desktop{})

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	rw := vapp.NewRenderWindow("VGE Logo", vapp.NewForwardRenderer(false))
	// Build scene
	buildScene(rw)
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

func buildScene(rw *vapp.RenderWindow) {
	// Load envhdr/kloofendal_48d_partly_cloudy_2k.hdr and create background "skybox" from it. VGE can create whole 360 background
	// from full 360 / 180 equirectangular image without needing 6 images for full cube
	// MustLoadAsset will handle loading loading actual asset using vasset.DefaultLoader set in start of program
	// MustLoadAsset will also handle ownership of asset (if will be disposed with device)
	eq := vapp.MustLoadAsset("envhdr/kloofendal_48d_partly_cloudy_2k.hdr",
		func(content []byte) (asset interface{}, err error) {
			return env.NewEquiRectBGNode(vapp.Ctx, vapp.Dev, 100, "hdr", content), nil
		}).(*env.EquiRectBGNode)
	// Add loaded background to scene
	rw.Env.Children = append(rw.Env.Children, vscene.NewNode(eq))

	// Create hello UI. First we must create a theme.
	// There is builtin minimal theme we can use here. It will use OpenSans font on material icons font if none other given.
	th := mintheme.NewTheme(vapp.Ctx, vapp.Dev, 15, nil, nil, nil)
	// Add theme to RenderWindow dispose list. In real app we might use theme multiple times on multiple windows and should handling disposing it
	// as part of disposing device.
	rw.AddChild(th)
	var bQuit *vui.Button
	ui := vui.NewUIView(th, image.Rect(100, 100, 400, 250), rw).
		SetContent(vui.NewPanel(10, vui.NewVStack(5,
			vui.NewLabel("Hello VGE!").SetClass("h2 dark"),
			vui.NewLabel("Press button to quit"),
			&vui.Extend{MinSize: image.Pt(10, 20)}, // Some spacing
			vui.NewButton(120, "Quit").SetClass("warning").AssignTo(&bQuit),
		)))
	bQuit.OnClick = func() {
		// Terminate application. We should run it like most UI events on separate go routine. Otherwise we have a change to deadlock engine
		go vapp.Terminate()
	}
	// Attach UI to scene and show it. UI panel are by default invisible and must be show
	rw.Ui.Children = append(rw.Ui.Children, vscene.NewNode(ui))
	ui.Show()
}
