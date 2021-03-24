//

package opensans

import (
	_ "embed"

	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
)

//go:embed OpenSans_Regular.ttf
var opensans_regular_ttf []byte

func NewGlyphSet(ctx vk.APIContext, dev *vk.Device, ranges ...vglyph.Range) *vglyph.GlyphSet {
	fl := &vglyph.VectorSetBuilder{}
	fl.AddFont(ctx, opensans_regular_ttf, ranges...)
	return fl.Build(ctx, dev)
}
