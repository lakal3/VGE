package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"log"
	"time"
)

var app struct {
	rw *vapp.ViewWindow
}

var timed bool

func main() {
	flag.BoolVar(&timed, "timed", false, "Time rendering (disables validations)")
	flag.Parse()
	// Initialize application framework. Add validate options to check Vulkan calls and desktop to enable windowing.
	var err error
	if timed {
		err = vapp.Init("UItest", vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 1024})
	} else {
		err = vapp.Init("UItest", vapp.Validate{}, vapp.Desktop{}, vapp.DynamicDescriptors{MaxDescriptors: 1024})
	}
	if err != nil {
		log.Fatal("App init failed ", err)
	}

	err = vdraw.CompileShapes(vapp.Dev)
	if err != nil {
		log.Fatal("Compile ", err)
	}

	// Create a new window. Window will has it's own scene that will be rendered using ForwardRenderer.
	// This first demo is only showing UI so we don't need depth buffer
	app.rw = vapp.NewViewWindow("UITest", vk.WindowPos{Left: -1, Top: -1, Width: 1024, Height: 768})
	if timed {
		app.rw.SetTimedOutput(renderDone)
	}
	// Build scene
	err = buildScene()
	if err != nil {
		log.Fatal("Build scene ", err)
	}
	// Wait until application is shut down (event Queue is stopped)
	vapp.WaitForShutdown()
}

var totalGPU, pendingGPU float64
var frameCount int

func renderDone(started time.Time, times []float64) {
	pendingGPU += (times[len(times)-1] - times[0]) / 1_000_000
	frameCount++
	if frameCount >= 60 {
		totalGPU, pendingGPU, frameCount = pendingGPU/float64(frameCount), 0, 0
	}
}

func buildScene() (err error) {
	err = mintheme.BuildMinTheme()
	if err != nil {
		return err
	}

	err = makeGlyphs()
	if err != nil {
		return err
	}
	nf := vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, painter)
	app.rw.AddView(nf)
	return nil
}

var choice int
var ch1, ch2 bool
var sl1 float32
var page int
var ppos float32 = 200
var name string = "Hello"
var name2 string

func painter(fr *vimgui.UIFrame) {
	fr.ControlArea = fr.DrawArea
	vimgui.Panel(fr, func(uf *vimgui.UIFrame) {
		fr.NewLine(-100, 30, 0)
		vimgui.Label(fr.WithTags("h2", "primary"), "VIMGUI demo")
		fr.Tags = nil
	}, func(uf *vimgui.UIFrame) {
		vimgui.VerticalDivider(fr, 150, 250, &ppos, menu, pages)
	})

}

func menu(fr *vimgui.UIFrame) {
	fr.NewLine(-95, 22, 0)
	vimgui.TabButton(fr, "page1", "Base controls", 0, &page)
	fr.NewLine(-95, 22, 2)
	vimgui.TabButton(fr, "page2", "Scroll area", 1, &page)
	fr.NewLine(-95, 22, 3)
	vimgui.TabButton(fr, "page3", "Shapes", 2, &page)
	fr.NewLine(-95, 3, 3)
	vimgui.Border(fr)
	if timed {
		fr.NewLine(-95, 20, 2)
		vimgui.Label(fr, fmt.Sprintf("Frame time %.3f us", totalGPU))
	}
}

func pages(fr *vimgui.UIFrame) {
	fr.PushPadding(vdraw.Edges{Left: 15}, false)
	defer fr.Pop()
	switch page {
	case 0:
		page1(fr)
	case 1:
		page2(fr)
	case 2:
		page3(fr)
	}
}

func page1(fr *vimgui.UIFrame) {
	fr.NewLine(-100, 25, 0)
	vimgui.Label(fr, "Hello VGE imgui!")
	fr.NewLine(120, 30, 5)
	fr.PushTags("*button")
	vimgui.Border(fr)
	fr.NewColumn(120, 20)
	vimgui.Border(fr)
	fr.Pop()

	fr.NewLine(120, 30, 5)
	fr.PushTags("primary")
	if vimgui.Button(fr, "btn1", "Quit button") {
		fmt.Println("Here")
		go func() {
			vapp.Terminate()
		}()
	}
	fr.Pop()
	fr.NewColumn(120, 10)
	if vimgui.Button(fr, "btnd1", "Open dialog") {
		newDialog()
	}
	fr.NewColumn(120, 10)
	if vimgui.Button(fr, "btnp1", "Open popup") {
		newPopup()
	}
	fr.NewLine(150, 25, 5)
	vimgui.RadioButton(fr, "rdb1", "Choice 1", 0, &choice)
	fr.NewColumn(150, 5)
	vimgui.RadioButton(fr, "rdb2", "Choice 2", 1, &choice)
	fr.NewLine(150, 25, 5)
	vimgui.CheckBox(fr, "cb1", "Check 1", &ch1)
	fr.NewColumn(150, 5)
	vimgui.CheckBox(fr, "cb1", "Check 2", &ch2)
	fr.NewLine(120, 3, 5)
	vimgui.Border(fr)
	fr.NewLine(120, 20, 2)
	vimgui.Label(fr, "Text")
	fr.NewLine(150, 30, 2)
	vimgui.Label(fr, "Name")
	fr.NewColumn(fr.DrawArea.Size()[0]-160, 5)
	vimgui.TextBox(fr, "name", &name)
	fr.NewLine(150, 30, 2)
	vimgui.Label(fr, "Name 2")
	fr.NewColumn(fr.DrawArea.Size()[0]-160, 5)
	vimgui.TextBox(fr, "name2", &name2)
	fr.NewLine(120, 20, 2)
	vimgui.Label(fr, "Sliders")
	fr.Offset = mgl32.Vec2{20}
	fr.NewLine(-60, 10, 5)
	pos := fr.ControlArea.From[1]
	vimgui.HorizontalSlider(fr, "hs1", -100, 200, 10, &sl1)
	fr.NewColumn(-35, 10)
	vimgui.Label(fr, fmt.Sprintf("Value %.1f", sl1))
	fr.Offset = mgl32.Vec2{0}
	fr.NewLine(0, 150, 0)
	fr.PushControlArea()
	fr.ControlArea.From = mgl32.Vec2{fr.ControlArea.From[0], pos}
	fr.ControlArea.To = mgl32.Vec2{fr.ControlArea.From[0] + 10, fr.ControlArea.To[1]}
	vimgui.VerticalSlider(fr, "vs1", -100, 200, 10, &sl1)
	fr.Pop()

}

func makeGlyphs() error {
	gs := &vdraw.GlyphSet{}
	err := gs.AddRunes(mintheme.PrimaryFont, 33, 127)
	if err != nil {
		return err
	}
	gs.Build(vapp.Dev, image.Pt(64, 64))
	vapp.AddChild(gs)
	return nil
}
