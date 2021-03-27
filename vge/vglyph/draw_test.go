package vglyph

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/lakal3/vge/vge/vapp/vtestapp"
)

func TestDrawInfo_Draw(t *testing.T) {
	ctx := vtestapp.TestContext{T: t}
	vtestapp.Init(ctx, "drawtest")
	// vasset.RegisterNativeImageLoader(ctx, vtestapp.TestApp.App)
	pngloader.RegisterPngLoader()
	theme := testBuildPalette(ctx)
	mm := vtestapp.NewMainImage()
	vtestapp.AddChild(mm)
	mm.ForwardRender(false, func(cmd *vk.Command, dc *vmodel.DrawContext) {
		s := Position{ImageSize: image.Pt(int(mm.Desc.Width), int(mm.Desc.Height)),
			Clip: image.Rect(100, 100, 500, 780)}
		s.GlyphArea = image.Rect(150, 100, 300, 200)
		ap := Appearance{ForeColor: mgl32.Vec4{0, 1, 0, 1}, BackColor: mgl32.Vec4{0, 0, 0, 0.2}}
		ap.GlyphName = "btn_focus"
		ap.Edges = image.Rect(30, 30, 30, 30)
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(150, 220, 300, 260)
		ap.GlyphName = "white"
		ap.Edges = image.Rect(1, 1, 1, 1)
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(150, 310, 300, 360)
		ap.GlyphName = "stripe"
		ap.Edges = image.Rect(1, 1, 1, 1)
		theme.Draw(dc, s, ap)
		ap.ForeColor = mgl32.Vec4{1, 1, 1, 1}
		ap.Edges = image.Rect(10, 10, 10, 10)
		s.GlyphArea = image.Rect(114, 370, 300, 410)
		ap.GlyphName = "rrect_in"
		ap.GlyphSet = 1
		ap.FgMask = 1
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(114, 420, 300, 460)
		ap.GlyphName = "rrect_out"
		ap.GlyphSet = 1
		ap.FgMask = 2
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(150, 480, 300, 560)
		ap.GlyphName = "white"
		ap.GlyphSet = 0
		ap.Edges = image.Rect(1, 1, 1, 1)
		ap.FgMask = 1
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(150, 580, 300, 660)
		ap.GlyphName = "btn_next"
		ap.GlyphSet = 2
		ap.FgMask = 0
		ap.Edges = image.Rect(20, 20, 20, 20)
		theme.Draw(dc, s, ap)
	})

	mm.Save("drawtest", vk.IMAGELayoutTransferSrcOptimal)
	vtestapp.Terminate()
}

func testBuildPalette(ctx vtestapp.TestContext) *Palette {
	gb := NewSetBuilder(ctx, SETGrayScale)
	tl := vtestapp.TestLoader{Path: "glyphs/test"}
	testLoadImage(ctx, gb, "btn_focus", tl, "button_focus.png", RED_GREENA, image.Rect(40, 40, 50, 50))
	gb.AddComputedGray("white", image.Pt(64, 64), image.Rect(16, 16, 16, 16),
		func(x, y int) (float32, float32) {
			return 1, 1
		})
	gb.AddComputedGray("stripe", image.Pt(64, 64), image.Rect(16, 16, 16, 16),
		func(x, y int) (float32, float32) {
			if (x/6)%2 == 0 {
				return 1, 1
			}
			if (x/12)%2 == 0 {
				return 0, 0
			}
			return 0, 1
		})
	gs := gb.Build(vtestapp.TestApp.Dev)
	vtestapp.AddChild(gs)
	gb = NewSetBuilder(ctx, SETRGBA)
	testLoadImage(ctx, gb, "btn_next", tl, "next_button.png", 0, image.Rect(40, 40, 40, 40))
	gs3 := gb.Build(vtestapp.TestApp.Dev)
	vtestapp.AddChild(gs3)

	vb := testVectorSet1()
	gs2 := vb.Build(ctx, vtestapp.TestApp.Dev)
	vtestapp.AddChild(gs2)
	th := NewPalette(ctx, vtestapp.TestApp.Dev, 4, 128)
	th.ComputeMask(ctx, vtestapp.TestApp.Dev, func(x, y, maskSize int) color.RGBA {
		return color.RGBA{A: 255, B: 0,
			R: byte(127 * (1 + math.Sin(float64(x)*math.Pi*2/float64(maskSize)))),
			G: byte(127 * (1 + math.Sin(float64(y)*math.Pi*2/float64(maskSize)))),
		}
	})
	th.ComputeMask(ctx, vtestapp.TestApp.Dev, func(x, y, maskSize int) color.RGBA {
		c := color.RGBA{A: 255}
		c.B = byte(x)
		if x&0x10 == 0 {
			c.R = 255
		}
		return c
	})
	th.AddGlyphSet(ctx, gs)
	th.AddGlyphSet(ctx, gs2)
	th.AddGlyphSet(ctx, gs3)
	vtestapp.AddChild(th)
	return th
}
