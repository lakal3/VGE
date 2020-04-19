package vui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"strconv"
	"strings"
)

type State uint32

func (state State) HasState(check State) bool {
	if state&check != 0 {
		return true
	}
	return false
}

// type Phase uint32

const (
	STATEDisabled = State(0x0001)
	STATEHover    = State(0x0002)
	STATEFocus    = State(0x0004)
	STATEPressed  = State(0x0008)
	STATEChecked  = State(0x0010)
	// Draw inner content of control (for sliders)
	STATEContent = State(0x0020)

	//PHASEContent    = Phase(0)
	//PHASEBorder     = Phase(1)
	//PHASEBackground = Phase(2)
	//PHASEShadow     = Phase(3)
)

type Style interface {
	Draw(owner Owner, ctrl Control, dc *vmodel.DrawContext, pos vglyph.Position, state State)
	ContentPadding() image.Rectangle
	GetFont(owner Owner, ctrl Control, state State) (font *vglyph.GlyphSet, fontHeight int)
	DrawString(owner Owner, ctrl Control, context *vmodel.DrawContext, position vglyph.Position, st State, text string)
}

func SplitClass(class string) []string {
	if len(class) == 0 {
		return nil
	}
	return strings.Split(class, " ")
}

func HasClass(class string, classes []string) bool {
	for _, cl := range classes {
		if cl == class {
			return true
		}
	}
	return false
}

func ApplyComputedSyles(class string, ap *vglyph.Appearance) bool {
	idx := strings.IndexRune(class, ':')
	if idx < 0 {
		return false
	}
	rest := class[idx+1:]
	switch class[:idx] {
	case "fc":
		return parseColor(rest, &ap.ForeColor)
	case "bc":
		return parseColor(rest, &ap.BackColor)
	}
	return false
}

func parseColor(colStr string, color *mgl32.Vec4) bool {
	col, err := strconv.ParseInt(colStr, 16, 64)
	if err != nil {
		return false
	}
	switch len(colStr) {
	case 6:
		*color = mgl32.Vec4{float32((col>>16)&255) / 255.0, float32((col>>8)&255) / 255.0, float32((col)&255) / 255.0, 1}
		return true
	case 8:
		*color = mgl32.Vec4{float32((col>>24)&255) / 255.0, float32((col>>16)&255) / 255.0, float32((col>>8)&255) / 255.0, float32((col)&255) / 255.0}
		return true
	}
	return false
}

func Lerpf32(ratio float32, f1, f2 float32) float32 {
	return f1*(1-ratio) + f2*ratio
}

func LerpColor(ratio float32, c1, c2 mgl32.Vec4) mgl32.Vec4 {
	return mgl32.Vec4{
		Lerpf32(ratio, c1[0], c2[0]),
		Lerpf32(ratio, c1[1], c2[1]),
		Lerpf32(ratio, c1[2], c2[2]),
		Lerpf32(ratio, c1[3], c2[3]),
	}
}

// InvertColor return black or white depending on color intensity
func InvertColor(c mgl32.Vec4) mgl32.Vec4 {
	if c[0]+c[1]+c[2] > 1.5 {
		return mgl32.Vec4{0, 0, 0, 1}
	}
	return mgl32.Vec4{1, 1, 1, 1}
}

type Theme interface {
	GetStyle(ctrl Control, class string) Style
}
