package vui

import (
	"image"
	"io/ioutil"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"golang.org/x/image/font/sfnt"
)

func TestMain(m *testing.M) {
	// vk.VGEDllPath = "vgelibd.dll"
	m.Run()
}

func TestNewUIView(t *testing.T) {
	err := vtestapp.Init("drawtest", vtestapp.UnitTest{T: t})
	if err != nil {
		t.Fatal("Init app ", err)
	}
	vasset.RegisterNativeImageLoader(vtestapp.TestApp.App)
	theme, err := testBuildTheme()
	if err != nil {
		t.Fatal("Build theme ", err)
	}
	mm := vtestapp.NewMainImage()
	rwDummu := &vapp.RenderWindow{WindowSize: image.Pt(int(mm.Desc.Width), int(mm.Desc.Height))}
	mv := NewUIView(theme, image.Rect(100, 100, 500, 700), rwDummu)
	mv.visible = true
	mv.DefaultFrame(buildTestView1())
	mm.Root.AddNode(nil, mv)
	mm.RenderScene(0, false)
	mm.Save("uiview", vk.IMAGELayoutTransferSrcOptimal)
}

func buildTestView1() Control {
	vs := NewVStack(0,
		NewLabel("Test set 1").SetClass("h1"),
		NewHStack(10, NewLabel("Button"),
			NewButton(120, "Test btn").SetClass("primary"),
			NewButton(120, "btn 2").SetClass("warning")),
	)
	return vs
}

type testTheme struct {
	palette *vglyph.Palette
}

type testStyle struct {
	palette    *vglyph.Palette
	ap         vglyph.Appearance
	fontHeight int
	padding    int
	bgName     string
	borderName string
	bgFactor   float32
}

func (t testStyle) DrawString(owner Owner, ctrl Control, dc *vmodel.DrawContext, pos vglyph.Position, st State, text string) {
	ap := t.ap
	ap.GlyphSet = 1
	ap.ForeColor = InvertColor(ap.ForeColor)
	t.palette.DrawString(dc, t.fontHeight, text, pos, ap)

}

func (t testStyle) Draw(owner Owner, ctrl Control, dc *vmodel.DrawContext, pos vglyph.Position, state State) {
	ap := t.ap
	if len(t.bgName) > 0 {
		ap.GlyphName = t.bgName
		ap.ForeColor = LerpColor(t.bgFactor, ap.BackColor, ap.ForeColor)
		t.palette.Draw(dc, pos, ap)
	}
	if len(t.borderName) > 0 {
		ap.ForeColor = t.ap.ForeColor
		ap.GlyphName = t.borderName
		t.palette.Draw(dc, pos, ap)
	}
}

func (t testStyle) ContentPadding() image.Rectangle {
	return image.Rect(t.padding, t.padding, t.padding, t.padding)
}

func (t testStyle) GetFont(owner Owner, ctrl Control, state State) (font *vglyph.GlyphSet, fontHeight int) {
	return t.palette.GetSet(1), t.fontHeight
}

func (t testTheme) GetStyle(ctrl Control, class string) Style {
	ap := vglyph.Appearance{BackColor: mgl32.Vec4{0, 0, 0, 0}, ForeColor: mgl32.Vec4{1, 1, 1, 1},
		Edges: image.Rect(8, 8, 8, 8)}
	st := testStyle{ap: ap, fontHeight: 22, palette: t.palette}
	st.bgFactor = 0.7
	switch ct := ctrl.(type) {
	case *Panel:
		st.bgName = "panel_bg"
		st.borderName = "panel_border"
		st.padding = 15
		st.ap.Edges = image.Rect(15, 15, 15, 15)
		_ = ct
	case *Button:
		st.bgName = "btn_bg"
		st.borderName = "btn_border"
		st.padding = 10
	}
	for _, cl := range SplitClass(class) {
		switch cl {
		case "primary":
			st.ap.ForeColor = mgl32.Vec4{0, 0, 1, 1}
		case "warning":
			st.ap.ForeColor = mgl32.Vec4{0.8, 0.8, 0, 1}
		case "h1":
			st.fontHeight = 36
		case "h2":
			st.fontHeight = 28
		}
	}

	return st
}

func testBuildTheme() (testTheme, error) {
	p, err := testBuildPalette()
	tt := testTheme{palette: p}
	return tt, err
}

func testBuildPalette() (*vglyph.Palette, error) {
	tl := vtestapp.TestLoader{Path: "glyphs/basicui"}
	gb := vglyph.NewSetBuilder(vglyph.SETGrayScale)
	err := testLoadImage(gb, "btn_border", tl, "btn.png", vglyph.RED, image.Rect(20, 20, 20, 20))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "btn_bg", tl, "btn.png", vglyph.GREEN, image.Rect(20, 20, 20, 20))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "vslider_border", tl, "vslider.png", vglyph.RED, image.Rect(0, 20, 0, 20))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "vslider_bg", tl, "vslider.png", vglyph.GREEN, image.Rect(0, 20, 0, 20))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "hslider_border", tl, "hslider.png", vglyph.RED, image.Rect(20, 0, 20, 0))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "hslider_bg", tl, "hslider.png", vglyph.GREEN, image.Rect(20, 0, 20, 0))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "panel_border", tl, "panel.png", vglyph.RED, image.Rect(20, 20, 20, 20))
	if err != nil {
		return nil, err
	}
	err = testLoadImage(gb, "panel_bg", tl, "panel.png", vglyph.GREEN, image.Rect(20, 20, 20, 20))
	if err != nil {
		return nil, err
	}
	gb.AddComputedGray("white", image.Pt(64, 64), image.Rect(16, 16, 16, 16), func(x, y int) (float32, float32) {
		return 1, 1
	})
	gs := gb.Build(vtestapp.TestApp.Dev)
	th := vglyph.NewPalette(vtestapp.TestApp.Dev, 4, 128)
	th.AddGlyphSet(gs)
	fl, err := testLoadGoFont("OpenSans_Regular.ttf")
	if err != nil {
		return nil, err
	}
	vb := &vglyph.VectorSetBuilder{}
	for r := rune(33); r < 256; r++ {
		vb.AddChar(fl, vglyph.NOMINALFontSize, r)
	}
	gs = vb.Build(vtestapp.TestApp.Dev)
	th.AddGlyphSet(gs)
	return th, nil
}

func testLoadImage(gb *vglyph.SetBuilder, name string, tl vtestapp.TestLoader, image string, color vglyph.ColorIndex,
	edges image.Rectangle) error {
	rd, err := tl.Open(image)
	if err != nil {
		return err
	}
	defer rd.Close()
	context, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}
	gb.AddEdgedGlyph(name, color, "png", context, edges)
	return nil
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
