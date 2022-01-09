package vglyph

import (
	"image"
	"io/ioutil"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"golang.org/x/image/font/sfnt"
)

func TestVectorSetBuilder(t *testing.T) {
	err := vtestapp.Init("vglyphbuilder", vtestapp.UnitTest{T: t})
	if err != nil {
		t.Fatal("Init app ", err)
	}
	vb := testVectorSet1()
	gs := vb.Build(vtestapp.TestApp.Dev)
	vtestapp.SaveImage(gs.image, "vglyphbuilder", vk.IMAGELayoutShaderReadOnlyOptimal)
	gs.Dispose()
	vtestapp.Terminate()
}

func testVectorSet1() *VectorSetBuilder {
	vb := &VectorSetBuilder{}
	vb.AddEdgedGlyph("rect_out", 2, image.Rect(20, 20, 20, 20)).
		AddRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 60}).
		AddRect(false, mgl32.Vec2{4, 4}, mgl32.Vec2{52, 52})
	vb.AddEdgedGlyph("rrect_out", 6, image.Rect(20, 20, 20, 20)).
		AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{52, 52}, mgl32.Vec4{10, 10, 10, 10})
	vb.AddEdgedGlyph("rect_in", 6, image.Rect(20, 20, 20, 20)).
		AddRect(true, mgl32.Vec2{}, mgl32.Vec2{52, 52})
	vb.AddEdgedGlyph("crect_in", 6, image.Rect(20, 20, 20, 20)).
		AddCornerRect(true, mgl32.Vec2{}, mgl32.Vec2{52, 52}, mgl32.Vec4{8, 8, 8, 8})
	vb.AddEdgedGlyph("crect_out", 2, image.Rect(20, 20, 20, 20)).
		AddCornerRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 60}, mgl32.Vec4{4, 8, 4, 8}).
		AddCornerRect(false, mgl32.Vec2{4, 4}, mgl32.Vec2{52, 52}, mgl32.Vec4{4, 8, 4, 8})
	vb.AddEdgedGlyph("rrect_in", 2, image.Rect(20, 20, 20, 20)).
		AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 60}, mgl32.Vec4{10, 10, 10, 10}).
		AddRoundedRect(false, mgl32.Vec2{4, 4}, mgl32.Vec2{52, 52}, mgl32.Vec4{10, 10, 10, 10})
	return vb
}

func TestFontVBuild(t *testing.T) {
	fl, err := testLoadGoFont("MaterialIcons_Regular.ttf")
	if err != nil {
		t.Fatal("Load font failed ", err)
	}
	err = vtestapp.Init("vectorfont", vtestapp.UnitTest{T: t})
	if err != nil {
		t.Fatal("Init app ", err)
	}
	gb := VectorSetBuilder{}
	for r := rune(0xe000); r < 0xefff; r++ {
		gb.AddChar(fl, NOMINALFontSize, r)
	}
	gs := gb.Build(vtestapp.TestApp.Dev)
	vtestapp.SaveImage(gs.image, "vfontbuilder", vk.IMAGELayoutShaderReadOnlyOptimal)
	gs.Dispose()
	vtestapp.Terminate()
}
func TestVectorPalette(t *testing.T) {
	err := vtestapp.Init("vdrawtest")
	if err != nil {
		t.Fatal("Load font failed ", err)
	}
	theme, err := testBuildVPalette()
	if err != nil {
		t.Fatal("Build V palette", err)
	}
	mm := vtestapp.NewMainImage()
	vtestapp.AddChild(mm)
	mm.ForwardRender(false, func(cmd *vk.Command, dc *vmodel.DrawContext) {
		s := Position{ImageSize: image.Pt(int(mm.Desc.Width), int(mm.Desc.Height)),
			Clip: image.Rect(100, 100, 500, 780)}
		s.GlyphArea = image.Rect(150, 150, 300, 250)
		ap := Appearance{ForeColor: mgl32.Vec4{0, 1, 0, 1}, BackColor: mgl32.Vec4{0, 0, 0, 0.2}}
		ap.Edges = image.Rect(10, 10, 10, 10)
		ap.GlyphName = "rect_out"
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(150, 300, 300, 350)
		ap.GlyphName = "rect_in"
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(350, 400, 550, 450)
		ap.GlyphName = "crect_in"
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(114, 410, 300, 450)
		ap.GlyphName = "rrect_out"
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(114, 460, 300, 500)
		ap.GlyphName = "rrect_in"
		theme.Draw(dc, s, ap)
		s.Rotate = 0.3
		theme.Draw(dc, s, ap)
		s.Rotate = 0
		s.GlyphArea = image.Rect(150, 700, 400, 780)
		theme.Draw(dc, s, ap)
		s.Clip = image.Rect(600, 100, 1000, 700)
		s.GlyphArea = image.Rect(650, 150, 700, 200)
		ap.GlyphSet = 1
		ap.GlyphName = "A"
		theme.Draw(dc, s, ap)
		s.GlyphArea = image.Rect(650, 250, 700, 280)
		theme.DrawString(dc, 28, "Hello VULKAN!", s, ap)
		s.GlyphArea = image.Rect(650, 300, 700, 316)
		theme.DrawString(dc, 16, "The quick brown fox jumps over a lazy dog", s, ap)
		s.GlyphArea = image.Rect(650, 325, 700, 345)
		theme.DrawString(dc, 20, "The quick brown fox jumps over a lazy dog", s, ap)
		s.GlyphArea = image.Rect(600, 375, 1000, 500)
		ap2 := ap
		ap2.Edges = image.Rect(2, 2, 2, 2)
		ap2.GlyphName = "rect_in"
		ap2.ForeColor = mgl32.Vec4{1, 1, 1, 1}
		ap2.GlyphSet = 0
		theme.Draw(dc, s, ap2)
		ap.BackColor = ap2.ForeColor
		ap.ForeColor = mgl32.Vec4{0, 0, 0, 1}
		s.GlyphArea = image.Rect(650, 400, 700, 425)
		theme.DrawString(dc, 24, "The quick brown fox jumps over a lazy dog", s, ap)
		s.GlyphArea = image.Rect(650, 430, 700, 455)
		// ap.GlyphSet = 3
		theme.DrawString(dc, 18, "The quick brown fox jumps over a lazy dog", s, ap)
		s.GlyphArea = image.Rect(650, 460, 700, 475)
		// ap.GlyphSet = 3
		theme.DrawString(dc, 14, "The quick brown fox jumps over a lazy dog", s, ap)
	})

	mm.Save("vdrawtest", vk.IMAGELayoutTransferSrcOptimal)
	vtestapp.Terminate()
}

func testBuildVPalette() (*Palette, error) {
	vb := testVectorSet1()
	gs := vb.Build(vtestapp.TestApp.Dev)
	vtestapp.AddChild(gs)
	th := NewPalette(vtestapp.TestApp.Dev, 4, 128)
	vtestapp.AddChild(th)
	th.AddGlyphSet(gs)
	fl, err := testLoadGoFont("OpenSans_Regular.ttf")
	if err != nil {
		return nil, err
	}
	vb = &VectorSetBuilder{}
	for r := rune(33); r < 256; r++ {
		vb.AddChar(fl, 32, r)
	}
	gs = vb.Build(vtestapp.TestApp.Dev)
	th.AddGlyphSet(gs)
	vtestapp.AddChild(gs)
	return th, nil
}

func testLoadGoFont(fontFile string) (*sfnt.Font, error) {
	tl := vtestapp.TestLoader{Path: "fonts"}
	rd, err := tl.Open(fontFile)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	content, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	font, err := sfnt.Parse(content)
	if err != nil {
		return nil, err
	}
	return font, nil
}
