package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"math"
	"os"
)

type Font struct {
	sf     *sfnt.Font
	buf    *sfnt.Buffer
	cache  map[rune]*Filled
	glyphs map[rune]*Glyph
}

const FontStrokeSize = 64

// LoadFontFile loads truetype font from a file.
func LoadFontFile(filePath string) (*Font, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return LoadFont(content)
}

// LoadFont loads truetype font from binary array
func LoadFont(content []byte) (*Font, error) {
	sf, err := sfnt.Parse(content)
	if err != nil {
		return nil, err
	}
	return &Font{sf: sf, buf: &sfnt.Buffer{}, cache: make(map[rune]*Filled), glyphs: make(map[rune]*Glyph)}, nil
}

// DrawChar draws one character. DrawChar will not stroke recorded path. You must call fill actually render recorded character path
// Note! Although it is possible to record several character path before calling a fill this will have performance hit
// (severe one if you record lost of characters before fill).
func (f *Font) DrawChar(p *Path, size int, at mgl32.Vec2, ch rune) error {
	idx, err := f.sf.GlyphIndex(f.buf, ch)
	if err != nil || idx == 0 {
		return nil
	}
	segments, err := f.sf.LoadGlyph(f.buf, idx, fixed.I(size), nil)
	if err != nil {
		return nil
	}
	for _, sg := range segments {
		switch sg.Op {
		case sfnt.SegmentOpMoveTo:
			p.MoveTo(toVector(at, sg, 0))
		case sfnt.SegmentOpLineTo:
			p.LineTo(toVector(at, sg, 0))
		case sfnt.SegmentOpQuadTo:
			mid := toVector(at, sg, 0)
			pos := toVector(at, sg, 1)
			p.BezierTo(mid, pos)
		default:
			panic("Invalid OP")
		}
	}
	return nil
}

// GetFilled retrieves filled version of rune. GetFilled will create new Filled shape if non existed before
func (f *Font) GetFilled(ch rune) (*Filled, error) {
	dr, ok := f.cache[ch]
	if ok {
		return dr, nil
	}
	p := Path{}
	err := f.DrawChar(&p, FontStrokeSize, mgl32.Vec2{0, FontStrokeSize}, ch)
	if err != nil {
		return nil, err
	}
	if len(p.segments) == 0 {
		f.cache[ch] = nil
		return nil, err
	}
	d := p.Fill()
	f.cache[ch] = d
	return d, nil
}

// AddToSet add rune to glyph set. GlyphSet creates a signed depth field of character that can be used to draw character as a glyph.
// You must ensure that GlyphSet is not disposed as long as this font in use.
// Once added there is no way to remove character from GlyphSet
// You can only add some characters from a font to a GlyphSet and you can add characters from multiple fonts to one GlyphSet
func (f *Font) AddToSet(ch rune, gs *GlyphSet) error {
	p := Path{}
	err := f.DrawChar(&p, FontStrokeSize, mgl32.Vec2{0, FontStrokeSize}, ch)
	if err != nil {
		return err
	}
	if len(p.segments) == 0 {
		return nil
	}
	gl := gs.AddPath(&p)
	f.glyphs[ch] = gl
	return nil
}

// GetGlyph returns glyph for rune. This may be nil if you have not added character to GlyphSet
func (f *Font) GetGlyph(ch rune) *Glyph {
	return f.glyphs[ch]
}

// MeasureText measures length and height of text using standard advance and kerning
func (f *Font) MeasureText(size float32, text string) (mgl32.Vec2, error) {
	return f.MeasureTextWith(size, text, func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2) {
		return at.Add(mgl32.Vec2{advance, 0})
	})
}

// MeasureTextWith measures length and height of text using custom advance and kerning
// Advance function must move at to nextAt using advance. If kern is true this advance is due to kerning and may be negative
// If kern is false, advance is size of current character
func (f *Font) MeasureTextWith(size float32, text string,
	advance func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2)) (mgl32.Vec2, error) {
	var iPrev sfnt.GlyphIndex

	var sz mgl32.Vec2
	var at mgl32.Vec2
	for idx, r := range text {

		iGl, err := f.sf.GlyphIndex(f.buf, r)
		if err != nil {
			return sz, err
		}
		if iGl == 0 {
			at = advance(idx, at, size/4, false)
			iPrev = 0
			continue
		}
		var defKern fixed.Int26_6
		if iPrev > 0 {
			defKern, err = f.sf.Kern(f.buf, iPrev, iGl, toFixed(size), font.HintingFull)
			if err == nil {
				at = advance(idx, at, float32(defKern)/FontStrokeSize, true)
			}
		}
		defAdvance, err := f.sf.GlyphAdvance(f.buf, iGl, toFixed(size), font.HintingFull)
		if err != nil {
			return sz, err
		}
		at = advance(idx, at, float32(math.Ceil(float64(defAdvance)/FontStrokeSize)), false)
		iPrev = iGl
	}
	sz[0] = at[0]
	sz[1] = size
	return sz, nil
}

func toFixed(size float32) fixed.Int26_6 {
	return fixed.I(int(math.Ceil(float64(size))))
}

func toVector(at mgl32.Vec2, sg sfnt.Segment, pos int) mgl32.Vec2 {
	p := sg.Args[pos]
	return mgl32.Vec2{float32(p.X), float32(p.Y)}.Mul(1.0 / 64.0).Add(at)
}
