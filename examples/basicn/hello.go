//go:build examples

package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"log"
)

func main() {

	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	// We must also add DynamicDescriptors and tell how many images we are going to use for one frame
	vapp.Init("hello", vapp.Validate{}, vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 100})

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	rw := vapp.NewViewWindow("VGE Logo", vk.WindowPos{Left: -1, Top: -1, Height: 768, Width: 1024})
	// Build ui view
	buildUi(rw)
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
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
	// Create new UI view
	v := vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, drawUi)
	rw.AddView(v)
}

func drawUi(fr *vimgui.UIFrame) {
	// Draw panel with title and content
	// First we need set control area directly
	fr.ControlArea = vdraw.Area{From: mgl32.Vec2{100, 100}, To: mgl32.Vec2{500, 300}}
	vimgui.Panel(fr, func(uf *vimgui.UIFrame) {
		// We can set control area also using NewLine and NewColumn helpers
		// Next control will have height of 30 and with of 100%. There will be 2 pixes padding to previous line (top)
		uf.NewLine(-100, 30, 2)
		// Set style for title
		uf.WithTags("h2")
		vimgui.Label(uf, "Hello VGE!")
	}, func(uf *vimgui.UIFrame) {
		uf.NewLine(-100, 30, 2)
		vimgui.Label(uf, "Press button to quit")
		// Add new line height 30 pixes and padded 3 pixels from previous line
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
