package vglyph

import (
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"io/ioutil"
	"testing"
)

func TestMain(m *testing.M) {
	vk.VGEDllPath = "vgelibd.dll"
	m.Run()
}

func TestNewGlyphBuilder(t *testing.T) {
	tl := vtestapp.TestLoader{Path: "glyphs/test"}
	ctx := vtestapp.TestContext{T: t}
	vtestapp.Init(ctx, "glyphbuilder")
	vasset.RegisterNativeImageLoader(ctx, vtestapp.TestApp.App)
	gb := NewSetBuilder(ctx, SETGrayScale)
	testLoadImage(ctx, gb, "btn_focus", tl, "button_focus.png", RED, image.Rect(40, 40, 50, 50))
	gb.AddComputedGray("white", image.Pt(64, 64), image.Rect(16, 16, 16, 16),
		func(x, y int) (float32, float32) {
			return 1, 1
		})
	gb.AddComputedGray("stripe", image.Pt(64, 64), image.Rect(16, 16, 16, 16),
		func(x, y int) (float32, float32) {
			if (x/6)%2 == 0 {
				return 1, 1
			}
			return 0, 1
		})
	gs := gb.Build(vtestapp.TestApp.Dev)
	vtestapp.SaveImage(gs.image, "glyphbuilder", vk.IMAGELayoutShaderReadOnlyOptimal)
	gs.Dispose()
	vtestapp.Terminate()
}

func testLoadImage(ctx vk.APIContext, gb *SetBuilder, name string, tl vtestapp.TestLoader, image string, color ColorIndex,
	edges image.Rectangle) {
	rd, err := tl.Open(image)
	if err != nil {
		ctx.SetError(err)
		return
	}
	defer rd.Close()
	context, err := ioutil.ReadAll(rd)
	if err != nil {
		ctx.SetError(err)
		return
	}
	gb.AddEdgedGlyph(name, color, "png", context, edges)
}
