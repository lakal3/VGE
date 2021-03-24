package mintheme

import (
	"image"
	"testing"

	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vui"
)

func TestMain(m *testing.M) {
	vk.VGEDllPath = "vgelibd.dll"
	m.Run()
}

func TestNewTheme(t *testing.T) {
	ctx := vtestapp.TestContext{T: t}
	vtestapp.Init(ctx, "mintheme_test")
	vasset.RegisterNativeImageLoader(ctx, vtestapp.TestApp.App)
	theme := NewTheme(ctx, vtestapp.TestApp.Dev, 0, nil, nil, nil)
	mm := vtestapp.NewMainImage()
	rwDummy := &vapp.RenderWindow{WindowSize: image.Pt(int(mm.Desc.Width), int(mm.Desc.Height))}
	mv := vui.NewUIView(theme, image.Rect(100, 100, 500, 700), rwDummy)
	mv.DrawTo(mm.Image)
	mv.DefaultFrame(buildTestView1())
	mm.Root.AddNode(nil, mv)
	mm.RenderScene(0, false)
	mm.Save("mintheme", vk.IMAGELayoutTransferSrcOptimal)
	gs := theme.Palette().GetSet(0)
	vtestapp.SaveImage(gs.GetImage(), "mintheme_image", vk.IMAGELayoutShaderReadOnlyOptimal)
	vtestapp.Terminate()
}

func buildTestView1() vui.Control {

	vs := vui.NewVStack(0,
		vui.NewLabel("Test set 1").SetClass("h1"),
		vui.NewHStack(10, vui.NewLabel("Button"),
			vui.NewButton(120, "Test btn").SetClass("primary"),
			vui.NewButton(120, "btn 2").SetClass("warning")),
		vui.NewVSlider(50, 20, 100),
	)
	return vs
}
