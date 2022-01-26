package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"math"
	"testing"
)

func TestGlyphSet_Build(t *testing.T) {
	pngloader.RegisterPngLoader()
	err := vtestapp.Init("buildglyph", opt{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}

	ft, err := LoadFontFile("../../assets/fonts/RobotoMono_Medium.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}

	gp := &GlyphSet{}
	gp.AddComputed(testCircle)
	for _, r := range testChars {
		p := &Path{}
		err = ft.DrawChar(p, 64, mgl32.Vec2{10, 70}, r)
		if err != nil {
			t.Fatal("Draw char: ", r, " ", err)
		}
		gp.AddPath(p)
	}
	gp.build(vtestapp.TestApp.Dev, image.Pt(64, 64), vk.IMAGELayoutGeneral)
	vtestapp.SaveImage(gp.imGlyph, "glyphset4", vk.IMAGELayoutGeneral)
	vtestapp.Terminate()
}

func testCircle(size image.Point, at image.Point) (depth float32) {
	center := size.Div(2)
	r := float32(center.X) / 1.5
	d := at.Sub(center)
	dist := float32(math.Sqrt(float64(d.X*d.X + d.Y*d.Y)))
	return dist - r
}
