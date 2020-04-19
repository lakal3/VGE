package mintheme

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/materialicons"
	"github.com/lakal3/vge/vge/vui/opensans"
	"image"
)

var minPalette *vglyph.Palette

type Theme struct {
	CornerSize int
	P          *vglyph.Palette
}

type Style struct {
	A                vglyph.Appearance
	FontSet          vglyph.GlyphSetIndex
	FontHeight       int
	palette          *vglyph.Palette
	BackgroundName   string
	BackgroundFactor func(state vui.State) float32
	BorderName       string
	BorderFactor     func(state vui.State) float32
	Padding          int
}

func (s Style) Draw(owner vui.Owner, ctrl vui.Control, dc *vmodel.DrawContext, pos vglyph.Position, state vui.State) {
	ap := s.A
	fc := ap.ForeColor

	if state.HasState(vui.STATEDisabled) {
		fc = mgl32.Vec4{0.5, 0.5, 0.5, 0.8}
	}
	if len(s.BackgroundName) > 0 {
		ap.GlyphName = s.BackgroundName
		f := s.BackgroundFactor(state)
		if f > 0 {
			ap.ForeColor = vui.LerpColor(f, ap.BackColor, fc)
			s.palette.Draw(dc, pos, ap)
		}
	}
	if len(s.BorderName) > 0 {
		ap.GlyphName = s.BorderName
		f := s.BorderFactor(state)
		if f > 0 {
			ap.ForeColor = vui.LerpColor(f, ap.BackColor, fc)
			s.palette.Draw(dc, pos, ap)
		}
	}
}

func (s Style) ContentPadding() image.Rectangle {
	return image.Rect(s.Padding, s.Padding, s.Padding, s.Padding)
}

func (s Style) GetFont(owner vui.Owner, ctrl vui.Control, state vui.State) (font *vglyph.GlyphSet, fontHeight int) {
	return s.palette.GetSet(s.FontSet), s.FontHeight
}

func (s Style) DrawString(owner vui.Owner, ctrl vui.Control, dc *vmodel.DrawContext, pos vglyph.Position, st vui.State, text string) {
	ap := s.A
	ap.GlyphSet = s.FontSet
	s.palette.DrawString(dc, s.FontHeight, text, pos, ap)
}

func (t *Theme) GetStyle(ctrl vui.Control, class string) vui.Style {
	ap := vglyph.Appearance{BackColor: mgl32.Vec4{0, 0, 0, 0}, ForeColor: mgl32.Vec4{1, 1, 1, 1},
		Edges: image.Rect(8, 8, 8, 8)}
	st := Style{A: ap, FontHeight: 14, palette: t.P, Padding: 2, FontSet: 1, BorderFactor: one, BackgroundFactor: one}
	classList := vui.SplitClass(class)

	switch ct := ctrl.(type) {
	case *vui.Panel:
		st.BackgroundName = "solid_bg"
		st.BorderName = "solid_border"
		if !vui.HasClass("solid", classList) {
			st.BackgroundFactor = Dim(0.3)
		}
		st.A.Edges = image.Rect(15, 15, 15, 15)
		_ = ct
	case *vui.Button:
		st.BackgroundName = "solid_bg"
		st.BorderName = "solid_border"
		if !vui.HasClass("solid", classList) {
			st.BackgroundFactor = bgButton
		}
		st.Padding = 8
	case *vui.TextBox:
		st.BackgroundName = "solid_filled"
		st.BorderName = "solid_border_line"
		st.BackgroundFactor = defHover
		st.BorderFactor = tbBorder
	case *vui.MenuButton:
		st.BackgroundName = "solid_filled"
		st.BackgroundFactor = defHover
	case *vui.ToggleButton:
		st.BackgroundName = "solid_filled"

		st.BackgroundFactor = defHover
		if vui.HasClass("underline", classList) {
			st.BorderName = "solid_border_line"
			st.BorderFactor = toggleBorder
		}

	case *vui.Caret:
		st.A.Edges = image.Rect(0, 10, 0, 10)
		st.BackgroundName = "solid_vline"
		// st.ForeColor = vui.InvertColor(st.ForeColor)
	case *vui.VSlider:
		st.BackgroundName = "solid_vline"
		st.BorderName = "solid_vline_border"
		st.BorderFactor = sliderBorder
		st.BackgroundFactor = sliderBackground
		st.Padding = 8
	case *vui.HSlider:
		st.BackgroundName = "solid_hline"
		st.BorderName = "solid_hline_border"
		st.BorderFactor = sliderBorder
		st.BackgroundFactor = sliderBackground
		st.Padding = 8
	}
	for _, cl := range classList {
		switch cl {
		case "primary":
			st.A.ForeColor = mgl32.Vec4{0, 0.5, 1, 1}
		case "warning":
			st.A.ForeColor = mgl32.Vec4{1, 0.75, 0.05, 1}
		case "danger":
			st.A.ForeColor = mgl32.Vec4{0.86, 0.21, 0.27, 1}
		case "success":
			st.A.ForeColor = mgl32.Vec4{0.16, 0.65, 0.27, 1}
		case "info":
			st.A.ForeColor = mgl32.Vec4{0.09, 0.63, 0.72, 1}
		case "h1":
			st.FontHeight = 28
		case "h2":
			st.FontHeight = 20
		case "dark":
			st.A.ForeColor = mgl32.Vec4{0.2, 0.2, 0.2, 1}
		case "light":
			st.A.ForeColor = mgl32.Vec4{0.96, 0.96, 0.96, 1}
		case "white":
			st.A.ForeColor = mgl32.Vec4{1, 1, 1, 1}
		case "icon":
			st.FontSet = 2
			st.FontHeight = st.FontHeight * 4 / 3
		default:
			vui.ApplyComputedSyles(cl, &st.A)
		}

	}

	return st
}

func one(st vui.State) float32 {
	return 1
}

func defHover(st vui.State) float32 {
	if st.HasState(vui.STATEHover) {
		return 0.4
	}
	return 0
}

func tbBorder(st vui.State) float32 {
	if st.HasState(vui.STATEFocus) {
		return 1
	}
	return 0.5
}

func sliderBorder(st vui.State) float32 {
	if st.HasState(vui.STATEContent) {
		return 0
	}
	if st.HasState(vui.STATEHover) {
		return 0.8
	}
	return 0.5
}

func sliderBackground(st vui.State) float32 {
	if !st.HasState(vui.STATEContent) {
		return 0
	}
	return 1
}

func bgButton(st vui.State) float32 {
	if st.HasState(vui.STATEPressed) {
		return 1
	}
	if st.HasState(vui.STATEHover) {
		return 0.85
	}
	return 0.6
}

func toggleBorder(st vui.State) float32 {
	if st.HasState(vui.STATEChecked) {
		return 1
	}
	return 0
}

func Dim(f float32) func(st vui.State) float32 {
	return func(st vui.State) float32 {
		return f
	}
}

func (t *Theme) Dispose() {
	if t.P != nil {
		t.P.Dispose()
	}
}

func (t *Theme) Palette() *vglyph.Palette {
	return t.P
}

// Constructs new minimal theme using given palette, mainFont and glyph font. palette, mainFont and iconFont can be nil if defaults are fine.
func NewTheme(ctx vk.APIContext, dev *vk.Device, cornerSize int, palette *vglyph.Palette, mainFont *vglyph.GlyphSet,
	iconFont *vglyph.GlyphSet, sets ...*vglyph.GlyphSet) *Theme {
	if mainFont == nil {
		mainFont = opensans.NewGlyphSet(ctx, dev, vglyph.Range{From: 33, To: 256})
	}
	if iconFont == nil {
		iconFont = materialicons.NewDefaultGlyphSet(ctx, dev)
	}
	if palette == nil {
		palette = vglyph.NewPalette(ctx, dev, 2, 128)
	}
	th := &Theme{P: palette, CornerSize: cornerSize}
	buildPalette(ctx, dev, th, mainFont, iconFont, sets)
	return th
}

func buildPalette(ctx vk.APIContext, dev *vk.Device, th *Theme, mainFont *vglyph.GlyphSet, iconFont *vglyph.GlyphSet, sets []*vglyph.GlyphSet) {
	mainSet := buildMainSet(ctx, dev, th)
	pl := th.P
	pl.AddGlyphSet(ctx, mainSet)
	pl.AddGlyphSet(ctx, mainFont)
	pl.AddGlyphSet(ctx, iconFont)
	for _, set := range sets {
		pl.AddGlyphSet(ctx, set)
	}
}

func buildMainSet(ctx vk.APIContext, dev *vk.Device, th *Theme) *vglyph.GlyphSet {
	b := &vglyph.VectorSetBuilder{}
	DefaultMainGlyph(b, "solid_bg", th)
	DefaultMainGlyph(b, "solid_border", th)
	DefaultMainGlyph(b, "solid_border_line", th)
	DefaultMainGlyph(b, "solid_vline", th)
	DefaultMainGlyph(b, "solid_hline", th)
	DefaultMainGlyph(b, "solid_vline_border", th)
	DefaultMainGlyph(b, "solid_hline_border", th)
	DefaultMainGlyph(b, "solid_filled", th)
	return b.Build(ctx, dev)
}

var DefaultMainGlyph = func(sb *vglyph.VectorSetBuilder, glyphName string, th *Theme) {
	edges := image.Rect(20, 20, 20, 20)
	roundning := mgl32.Vec4{float32(th.CornerSize), float32(th.CornerSize), float32(th.CornerSize), float32(th.CornerSize)}
	switch glyphName {
	case "solid_bg":
		sb.AddEdgedGlyph(glyphName, 8, edges).AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{48, 48},
			roundning)

	case "solid_border":
		sb.AddEdgedGlyph(glyphName, 2, edges).
			AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 60}, roundning).
			AddRoundedRect(false, mgl32.Vec2{6, 6}, mgl32.Vec2{48, 48}, roundning)

	case "solid_filled":
		sb.AddEdgedGlyph(glyphName, 2, edges).
			AddRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 60})
	case "solid_border_line":
		sb.AddEdgedGlyph(glyphName, 2, edges).AddPoint(mgl32.Vec2{}).
			AddRect(true, mgl32.Vec2{0, 56}, mgl32.Vec2{60, 4})
	case "solid_hline":
		sb.AddEdgedGlyph(glyphName, 2, image.Rect(20, 0, 20, 0)).
			AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{60, 28}, mgl32.Vec4{13, 13, 13, 13})
	case "solid_vline":
		sb.AddEdgedGlyph(glyphName, 2, image.Rect(0, 20, 0, 20)).
			AddRoundedRect(true, mgl32.Vec2{}, mgl32.Vec2{28, 60}, mgl32.Vec4{13, 13, 13, 13})
	case "solid_hline_border":
		sb.AddEdgedGlyph(glyphName, 2, image.Rect(20, 0, 20, 0)).
			AddPoint(mgl32.Vec2{}).AddPoint(mgl32.Vec2{0, 28}).
			AddRoundedRect(true, mgl32.Vec2{0, 8}, mgl32.Vec2{60, 12}, mgl32.Vec4{5, 5, 5, 5})

	case "solid_vline_border":
		sb.AddEdgedGlyph(glyphName, 2, image.Rect(0, 20, 0, 20)).
			AddPoint(mgl32.Vec2{}).AddPoint(mgl32.Vec2{28, 0}).
			AddRoundedRect(true, mgl32.Vec2{8, 0}, mgl32.Vec2{12, 60}, mgl32.Vec4{5, 5, 5, 5})
	}
}
