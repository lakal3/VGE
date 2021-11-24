// Theme3D is simple theme build with bitmaps
package theme3d

import (
	"image"
	"io/ioutil"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vui"
	"github.com/lakal3/vge/vge/vui/materialicons"
	"github.com/lakal3/vge/vge/vui/opensans"
)

// Actual theme
type Theme struct {
	P         *vglyph.Palette
	imagePath string
	err       error
}

// Style for control. Themes will construct styles based on control type and classes. You can have several different styles in one theme but this example
// uses only single theme
type Style struct {
	A                 vglyph.Appearance
	FontSet           vglyph.GlyphSetIndex
	FontHeight        int
	palette           *vglyph.Palette
	BackgroundName    string
	BackgroundFactor  func(state vui.State) float32
	CheckedBackground string
	Padding           int
	FocusBackground   string
	HoverBackground   string
	ContentBackground string
}

// Draw will draw control. In VGE styles are responsible for drawing controls. As themes will create styles, Theme can actually decide how each control looks
func (s Style) Draw(owner vui.Owner, ctrl vui.Control, dc *vmodel.DrawContext, pos vglyph.Position, state vui.State) {
	ap := s.A
	fc := ap.ForeColor

	// We change forecolor for each disabled control
	if state.HasState(vui.STATEDisabled) {
		fc = mgl32.Vec4{0.5, 0.5, 0.5, 0.8}
	}
	bg := s.BackgroundName
	// Use given function to calculate foreground / background ratio. This function is used typically to dim control when no active or hovered over.
	f := s.BackgroundFactor(state)

	// Logic to select different glyph for control base on it's state. Background values are initialize in theme
	if (state.HasState(vui.STATEChecked) || state.HasState(vui.STATEPressed)) && len(s.CheckedBackground) > 0 {
		bg = s.CheckedBackground
	}
	if state.HasState(vui.STATEHover) && len(s.HoverBackground) > 0 {
		bg = s.HoverBackground
		f = 1
	}
	if state.HasState(vui.STATEFocus) && len(s.FocusBackground) > 0 {
		bg = s.FocusBackground
	}
	if state.HasState(vui.STATEContent) && len(s.ContentBackground) > 0 {
		bg = s.ContentBackground
	}
	if len(bg) > 0 {
		ap.GlyphName = bg
		if f > 0 {
			// Adjust for color and draw glyph
			ap.ForeColor = vui.LerpColor(f, ap.BackColor, fc)
			s.palette.Draw(dc, pos, ap)
		}
	}
}

// Some controls with border like buttons need some padding to work. Style can provide default padding for control content.
func (s Style) ContentPadding() image.Rectangle {
	return image.Rect(s.Padding, s.Padding, s.Padding, s.Padding)
}

// Get font and font height. Used in controls that draw text.
func (s Style) GetFont(owner vui.Owner, ctrl vui.Control, state vui.State) (font *vglyph.GlyphSet, fontHeight int) {
	return s.palette.GetSet(s.FontSet), s.FontHeight
}

// DrawString draws a text
func (s Style) DrawString(owner vui.Owner, ctrl vui.Control, dc *vmodel.DrawContext, pos vglyph.Position, st vui.State, text string) {
	ap := s.A
	ap.GlyphSet = s.FontSet
	s.palette.DrawString(dc, s.FontHeight, text, pos, ap)
}

// Themes main function that convert control and class to style. Class may contains several style classes separated by space line html class.
// In VGE there is no hierarchy with control class. You cannot use controls parent or siblings class(es) in GetStyle
func (t *Theme) GetStyle(ctrl vui.Control, class string) vui.Style {
	ap := vglyph.Appearance{BackColor: mgl32.Vec4{0, 0, 0, 0}, ForeColor: mgl32.Vec4{1, 1, 1, 1},
		Edges: image.Rect(8, 8, 8, 8)}
	st := Style{A: ap, FontHeight: 14, palette: t.P, Padding: 2, FontSet: 1, BackgroundFactor: one}
	classList := vui.SplitClass(class)

	// Set proper style values for each control
	switch ct := ctrl.(type) {
	case *vui.Panel:
		st.BackgroundName = "panel"
		if !vui.HasClass("solid", classList) {
			st.BackgroundFactor = dim(0.3)
		}
		st.A.Edges = image.Rect(15, 15, 15, 15)
		_ = ct
	case *vui.Button:
		st.BackgroundName = "btn_up"
		if !vui.HasClass("solid", classList) {
			st.BackgroundFactor = hoverState
		}
		st.CheckedBackground = "btn_down"
		st.Padding = 10
	case *vui.TextBox:
		st.BackgroundName = "tb_normal"
		st.FocusBackground = "tb_edit"
		st.BackgroundFactor = tbBorder
		st.Padding = 8
	case *vui.MenuButton:
		st.HoverBackground = "highlight"
		st.Padding = 4
	case *vui.ToggleButton:
		st.HoverBackground = "highlight"
		if vui.HasClass("underline", classList) {
			st.BackgroundName = "line"
			st.BackgroundFactor = toggleBorder
		}
		st.Padding = 4
	case *vui.Caret:
		st.A.Edges = image.Rect(0, 10, 0, 10)
		st.BackgroundName = "caret"
		// st.ForeColor = vui.InvertColor(st.ForeColor)
	case *vui.VSlider:
		st.BackgroundName = "vslider_base"
		st.BackgroundFactor = hoverState
		st.ContentBackground = "vslider"
		st.A.Edges = image.Rect(0, 10, 0, 10)
		st.Padding = 8
	case *vui.HSlider:
		st.BackgroundName = "slider_base"
		st.BackgroundFactor = hoverState
		st.ContentBackground = "slider"
		st.A.Edges = image.Rect(10, 0, 10, 0)
		st.Padding = 8
	}
	// Default style overrides controls.
	// If you need more classes you can easily override existing theme and implement your own theme that adds more classes to controls
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

// Helper function to set background factor
func one(st vui.State) float32 {
	return 1
}

// Helper function to set background factor. Show control more visible when it has focus
func tbBorder(st vui.State) float32 {
	if st.HasState(vui.STATEFocus) {
		return 1
	}
	return 0.5
}

// Helper function to set background factor. Show control more visible when it is hovered over
func hoverState(st vui.State) float32 {
	if st.HasState(vui.STATEHover) {
		return 1
	}
	return 0.6
}

func toggleBorder(st vui.State) float32 {
	if st.HasState(vui.STATEChecked) {
		return 1
	}
	return 0
}

// Constant dimming for background
func dim(f float32) func(st vui.State) float32 {
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

// Constructs new 3D theme. You must supply path for assets/glyphs/3dui so that theme can locate bitmaps used to build up this Theme
func NewTheme(dev *vk.Device, imagePath string) *Theme {
	mainFont := opensans.NewGlyphSet(dev, vglyph.Range{From: 33, To: 256})
	iconFont := materialicons.NewDefaultGlyphSet(dev)
	palette := vglyph.NewPalette(dev, 0, 0)
	th := &Theme{P: palette, imagePath: imagePath}
	th.buildPalette(dev, mainFont, iconFont)
	return th
}

// Build actual palette containing main theme, font and glyph font
func (th *Theme) buildPalette(dev *vk.Device, mainFont *vglyph.GlyphSet, iconFont *vglyph.GlyphSet) {
	mainSet := th.buildMainSet(dev)
	pl := th.P
	pl.AddGlyphSet(mainSet)
	pl.AddGlyphSet(mainFont)
	pl.AddGlyphSet(iconFont)
}

// Build glyphs for main glyph set.
func (th *Theme) buildMainSet(dev *vk.Device) *vglyph.GlyphSet {
	b := vglyph.NewSetBuilder(vglyph.SETGrayScale)
	th.addGlyph(b, "panel")
	th.addGlyph(b, "btn_up")
	th.addGlyph(b, "btn_down")
	th.addGlyph(b, "tb_normal")
	th.addGlyph(b, "tb_edit")
	th.addGlyph(b, "caret")
	th.addGlyph(b, "highlight")
	th.addGlyph(b, "line")
	th.addGlyph(b, "slider")
	th.addGlyph(b, "slider_base")
	th.addGlyph(b, "vslider")
	th.addGlyph(b, "vslider_base")
	return b.Build(dev)
}

func (th *Theme) addGlyph(sb *vglyph.SetBuilder, glyphName string) {
	edges := image.Rect(20, 20, 20, 20)
	fName := glyphName + ".png"
	// Override default edges for some controls
	switch glyphName {
	case "panel":
		edges = image.Rect(35, 35, 35, 35)
	case "caret":
		edges = image.Rect(0, 16, 0, 16)
	case "slider":
		// Only left/right edge
		edges = image.Rect(35, 0, 35, 0)
	case "slider_base":
		// Only left/right edge
		edges = image.Rect(35, 0, 35, 0)
	case "vslider":
		// Only top/bottom edge
		edges = image.Rect(0, 35, 0, 35)
	case "vslider_base":
		// Only top/bottom edge
		edges = image.Rect(0, 35, 0, 35)
	}
	content, err := ioutil.ReadFile(filepath.Join(th.imagePath, fName))
	if err != nil {
		th.setError(err)
	}
	// Add glyph to glyph set builder. RED_GREENA indicates that RED color tells foreground intesity.
	// GREEN color if > 0.5 will tell that pixel is transparent (alpha = 0). This is bit similar things used in film (with blue color).
	// This is done because it is easier to render GREEN than transparent background. Three3D images are rendered bitmaps from 3D looking controls.
	// Typically, if you draw controls using some 2D drawing program, you use directly RED channel for foreground / background ratio and alpha channel for alpha.
	sb.AddEdgedGlyph(glyphName, vglyph.RED_GREENA, "png", content, edges)
}

func (t *Theme) setError(err error) {
	if t.err == nil {
		t.err = err
	}
}
