package vglyph

import (
	"image"
	"math"

	"github.com/lakal3/vge/vge/vk"
)

type SetKind int

const (
	SETDepthField = SetKind(0)
	SETGrayScale  = SetKind(1)
	SETRGBA       = SetKind(2)
)

// Glyph is individual glyph in glyph set.
type Glyph struct {
	Name       string
	Location   image.Rectangle
	CharOffset image.Point
	// Edges from 3 or 9 part glyph. left, top, right, bottom
	Edges image.Rectangle
}

// GlyphSet is set of prerendered glyph. GlyphSet can be either
// SETDepthField - All glyphs are rendered as signed depth fields. These allows more accurate sampling of glyph edges
// when sizing them. This is ideal for font's and other single colored glyphs
// SETGrayScale - Grays scale glyph mixes blending between font color and back color based on image grayness. Alpha channel is
// used to control glyphs alpha factor.
type GlyphSet struct {
	Desc    vk.ImageDescription
	Advance func(height int, from, to rune) float32

	kind   SetKind
	glyphs map[string]Glyph
	pool   *vk.MemoryPool
	image  *vk.Image
}

func (set *GlyphSet) Dispose() {
	if set.pool != nil {
		set.pool.Dispose()
		set.pool, set.image = nil, nil
	}
}

// GetImage retrieves image associated to GlyphSet
func (set *GlyphSet) GetImage() *vk.Image {
	return set.image
}

// Kind retrieves layout of image (SDF, GrayScale, RGBA)
func (set *GlyphSet) Kind() SetKind {
	return set.kind
}

// Get glyph from set
func (set *GlyphSet) Get(name string) Glyph {
	return set.glyphs[name]
}

// If glyph set is made from font, measure string will calculate length of text using given font height
func (gs *GlyphSet) MeasureString(text string, fontHeight int) int {
	pos := float32(0)
	prevChar := rune(0)
	w := float32(0)
	for _, ch := range text {
		if prevChar != 0 {
			pos += gs.Advance(fontHeight, prevChar, ch)
		}
		gl := gs.Get(string(ch))
		if len(gl.Name) > 0 {
			sz := gl.Location.Size()
			w = float32(fontHeight*sz.X) / NOMINALFontSize
			w += float32(fontHeight*gl.CharOffset.X) / NOMINALFontSize
		}
		prevChar = ch
		pos += w
	}
	return int(math.Ceil(float64(pos)))
}

func newGlyphSet(ctx vk.APIContext, dev *vk.Device, glyphs int, w int, h int, kind SetKind) *GlyphSet {
	gs := &GlyphSet{glyphs: make(map[string]Glyph, glyphs), kind: kind}
	gs.pool = vk.NewMemoryPool(dev)
	f := vk.FORMATR8Unorm
	if kind == SETGrayScale {
		f = vk.FORMATR8g8Unorm
	}
	if kind == SETRGBA {
		f = vk.FORMATR8g8b8a8Unorm
	}
	gs.Desc = vk.ImageDescription{Width: uint32(w), Height: uint32(h), Depth: 1,
		Format: f, MipLevels: 1, Layers: 1}
	gs.image = gs.pool.ReserveImage(ctx, gs.Desc, vk.IMAGEUsageSampledBit|vk.IMAGEUsageStorageBit|vk.IMAGEUsageTransferSrcBit)
	gs.pool.Allocate(ctx)
	return gs
}

func DefaultAdvance(height int, from, to rune) float32 {
	return float32(height) / 12
}
