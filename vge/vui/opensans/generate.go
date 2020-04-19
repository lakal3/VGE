//

//go:generate packspv -m font.manifest -p opensans .
package opensans

import (
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
)

func NewGlyphSet(ctx vk.APIContext, dev *vk.Device, ranges ...vglyph.Range) *vglyph.GlyphSet {
	fl := &vglyph.VectorSetBuilder{}
	fl.AddFont(ctx, opensans_regular_ttf, ranges...)
	return fl.Build(ctx, dev)
}
