//

//go:generate packspv -m font.manifest -p materialicons .
package materialicons

import (
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
)

func NewGlyphSet(ctx vk.APIContext, dev *vk.Device, ranges ...vglyph.Range) *vglyph.GlyphSet {
	fl := &vglyph.VectorSetBuilder{}
	fl.AddFont(ctx, materialicons_regular_ttf, ranges...)
	return fl.Build(ctx, dev)

}

func NewDefaultGlyphSet(ctx vk.APIContext, dev *vk.Device) *vglyph.GlyphSet {
	return NewGlyphSet(ctx, dev, vglyph.Range{From: rune(0xe000), To: rune(0xea00)})
}
