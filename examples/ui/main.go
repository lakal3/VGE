package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/examples/ui/theme3d"
	"github.com/lakal3/vge/vge/materials/env"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/mintheme"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var config struct {
	env   string
	debug bool
}

type viewerApp struct {
	owner   vk.Owner
	theme   vui.Theme
	theme3D vui.Theme
	mv      *vui.UIView
	win     *vapp.RenderWindow
	chWin   *vapp.RenderWindow

	mv2 *vui.UIView
	mv3 *vui.UIView
	mv4 *vui.UIView
}

func (v *viewerApp) Dispose() {
	v.owner.Dispose()
}

func (v *viewerApp) SetError(err error) {
	if strings.Index(err.Error(), "VK_IMAGE_VIEW_TYPE_CUBE") > 0 {
		return
	}
	log.Fatal("API error ", err)
}

func (v *viewerApp) IsValid() bool {
	return true
}

func (v *viewerApp) Begin(callName string) (atEnd func()) {
	return nil
}

var app viewerApp

func main() {
	flag.StringVar(&config.env, "env", "", "Environment HDR image")
	flag.BoolVar(&config.debug, "debug", false, "Use debug layers")
	flag.Parse()
	if config.debug {
		// Comment out if you wan't to test app with debug library
		//vk.VGEDllPath = "VGELibd.dll"
	}
	app.init()
	defer app.Dispose()

	vapp.RegisterHandler(1000, app.keyHandler)
	app.win = vapp.NewRenderWindow("Model viewer", vapp.NewForwardRenderer(true))
	_ = vapp.NewOrbitControl(-50, app.win)
	// Load themes
	app.loadThemes()
	app.loadEnv()
	app.switchTheme(false)

	vapp.WaitForShutdown()
}

func (v *viewerApp) loadThemes() {
	app.theme = mintheme.NewTheme(vapp.Ctx, vapp.Dev, 15, nil, nil, nil)
	app.theme3D = theme3d.NewTheme(vapp.Ctx, vapp.Dev, "../../assets/glyphs/3dui")
}

func usage() {
	fmt.Println("ui {flags} ")
	os.Exit(1)
}

func (a *viewerApp) init() {
	// Add validation layers
	if config.debug {
		vapp.AddOption(vapp.Validate{})
	}
	vapp.Init("ui", vapp.Desktop{})
}

func (v *viewerApp) keyHandler(ctx vk.APIContext, ev vapp.Event) (unregister bool) {
	ke, ok := ev.(*vapp.CharEvent)
	if ok && ke.Char == 'Q' {
		go vapp.Terminate()
		return true
	}
	bb, ok := ev.(*vui.ButtonClickedEvent)
	if ok {
		if bb.ID == "QUIT" {
			go vapp.Terminate()
			return true
		}
		if bb.ID == "CLOSE" {
			if v.chWin != nil {
				wTmp := v.chWin
				v.chWin = nil
				go wTmp.Dispose()
			}
			v.mv3.Hide()
		}
	}
	return false
}

func (v *viewerApp) genUI(newTheme bool) {
	// v.theme = squaretheme.NewTheme(vapp.Ctx, vapp.Dev, nil, nil, nil)
	th := v.theme
	if newTheme {
		th = v.theme3D
	}
	var tbTheme *vui.ToggleButton

	vbDis := vui.NewButton(100, "Disabled")
	vbDis.Disabled = true
	pageSel1 := vui.NewToggleButton("TB1", "Page 1", "Page 1").SetClass("underline")
	pageSel2 := vui.NewToggleButton("TB2", "Page 2", "Page 2").SetClass("underline")
	pageSel3 := vui.NewToggleButton("TB3", "Page 3", "Page 3").SetClass("underline")
	page1 := vui.NewConditional(true, vui.NewVStack(5,
		vui.NewCheckbox("3D Theme", "").SetOnChanged(v.switchTheme).AssignTo(&tbTheme),
		vui.NewLabel("Test page 1").SetClass("h2"),
		vui.NewTextBox("TB1", 20, "Enter some text here"),
		vui.NewTextBox("TB2", 20, "Or here"),
		vui.NewHStack(10,
			vui.NewButton(100, "Test btn").SetClass("primary"),
			vui.NewButton(100, "Quit!").SetClass("warning").SetID("QUIT"),
			vbDis),
		buildTestIcons(),
		vui.NewHSlider(20, 20, 100),
		vui.NewHStack(5, vui.NewSizer(
			vui.NewVSlider(30, 15, 100), image.Pt(20, 150), image.Pt(40, 150)),
			vui.NewLabel("HSlider"))))
	page2 := vui.NewConditional(false, vui.NewVStack(5,
		vui.NewLabel("Test page 2").SetClass("h2"),
		vui.NewHStack(5,
			vui.NewMenuButton("Menu 1").SetOnClick(v.showPopup),
			vui.NewMenuButton("Open file").SetOnClick(v.showDialog),
			vui.NewMenuButton("Open window").SetOnClick(v.openWindow)),
		vui.NewCheckbox("Select me", ""),
		vui.NewRadioButton("Select me")))
	page3 := vui.NewConditional(false,
		v.buildFileCanvas())

	vs := vui.NewVStack(8,
		vui.NewLabel("Test set 1").SetClass("h1"),
		vui.NewHStack(5, pageSel1, pageSel2, pageSel3),
		page1, page2, page3)

	rg := vui.NewRadioGroup(pageSel1, pageSel2, pageSel3)
	rg.OnChanged = func(value int) {
		page1.Visible = value == 0
		page2.Visible = value == 1
		page3.Visible = value == 2
	}
	v.mv = vui.NewUIView(th, image.Rect(100, 100, 500, 700), v.win)
	v.mv.DefaultFrame(vs)
	v.mv2 = vui.NewUIView(th, image.Rect(600, 100, 1000, 700), v.win)
	v.mv2.DefaultFrame(loadSource()).SetClass("solid")
	v.mv3 = vui.NewUIView(th, image.Rect(200, 200, 700, 500), v.win)
	v.mv3.DefaultFrame(vui.NewVStack(5,
		v.buildFileCanvas()))
	v.mv4 = vui.NewUIView(th, image.Rect(300, 300, 480, 360), v.win)
	v.mv4.DefaultFrame(vui.NewLabel("Test popup"))
	tbTheme.Checked = newTheme
}

func (v *viewerApp) buildFileCanvas() *vui.Canvas {
	return vui.NewCanvas(image.Point{500, 500}).
		AddBottomRight(vui.NewButton(100, "Close").SetClass("primary").SetID("CLOSE"), image.Rect(400, 460, 495, 495)).
		AddTopLeft(vui.NewLabel("File:"), image.Rect(10, 10, 100, 40)).
		AddItem(vui.NewTextBox("FILENAME", 10, "test"), image.Rect(120, 10, 490, 40),
			mgl32.Vec2{0, 0}, mgl32.Vec2{1, 0}).
		AddItem(vui.NewLabel("Status area"), image.Rect(10, 430, 490, 450), mgl32.Vec2{0, 1}, mgl32.Vec2{0, 1}).
		AddCenter(vui.NewScrollViewer(vui.NewLabel("Scroll content here!")), image.Rect(10, 50, 490, 420))
}

func (v *viewerApp) showDialog() {
	v.mv3.ShowDialog()
}

func (v *viewerApp) showPopup() {
	v.mv4.ShowPopup()
}

func (v *viewerApp) openWindow() {
	if v.chWin != nil && !v.chWin.Closed() {
		return
	}
	// Place window on monitor 1, if available

	ms, exists := vapp.GetMonitorArea(1)
	if !exists {
		ms, _ = vapp.GetMonitorArea(0)
	}
	// Place window at center of monitor
	w := ms.Width / 4
	h := ms.Height / 4
	wp := vk.WindowPos{Left: ms.Left + w, Top: ms.Top + h, Width: 2 * w, Height: 2 * h, State: vk.WINDOWStateBorderless}
	v.chWin = vapp.NewRenderWindowAt("Open file", wp, vapp.NewForwardRenderer(false))
	r := image.Rect(100, 100, int(2*w-100), int(2*h-100))
	uv := vui.NewUIView(v.theme, r, v.chWin)
	uv.DefaultFrame(v.buildFileCanvas())
	v.chWin.Scene.AddNode(nil, uv)
	uv.Show()
}

func loadSource() vui.Control {
	vSrc := vui.NewVStack(2)
	content, err := ioutil.ReadFile("main.go")
	if err != nil {
		vapp.Ctx.SetError(err)
		return nil
	}
	for _, line := range strings.Split(string(content), "\n") {
		vSrc.Children = append(vSrc.Children, vui.NewLabel(line).SetClass("dark"))
	}
	vs := vui.NewVStack(8,
		vui.NewLabel("Source").SetClass("h1 dark"),
		vui.NewSizer(vui.NewScrollViewer(vSrc), image.Pt(0, 0), image.Pt(320, 500)))
	return vs
}

func buildTestIcons() vui.Control {
	return vui.NewHStack(5,
		vui.NewLabel(string(rune(0xe002))).SetClass("icon"),
		vui.NewLabel(string(rune(0xe5ca))).SetClass("icon"),
		vui.NewLabel(string(rune(0xe834))).SetClass("icon"),
	)
}

func (v *viewerApp) loadEnv() {
	if len(config.env) > 0 {
		// Add environment (Skybox) to scene
		content, err := ioutil.ReadFile(config.env)
		if err != nil {
			v.SetError(err)
			return
		}
		eq := env.NewEquiRectBGNode(v, vapp.Dev, 1000, "hdr", content)
		v.owner.AddChild(eq)
		v.win.Env.Children = append(v.win.Env.Children, vscene.NewNode(eq))
	} else {
		bg := vscene.NewNode(env.NewGrayBG())
		v.win.Env.Children = append(v.win.Env.Children, bg)
	}
}

func (v *viewerApp) switchTheme(checked bool) {
	if v.mv != nil {
		v.mv.Hide()
		v.mv2.Hide()
	}
	v.genUI(checked)
	v.win.Scene.Update(func() {
		v.win.Ui.Children = []*vscene.Node{vscene.NewNode(v.mv), vscene.NewNode(v.mv2), vscene.NewNode(v.mv3), vscene.NewNode(v.mv4)}
		v.mv.Show()
		v.mv2.Show()
	})
}
