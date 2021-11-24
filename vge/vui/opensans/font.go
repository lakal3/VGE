//

package opensans

import (
	_ "embed"

	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
)

//go:embed OpenSans_Regular.ttf
var opensans_regular_ttf []byte

func NewGlyphSet(dev *vk.Device, ranges ...vglyph.Range) *vglyph.GlyphSet {
	fl := &vglyph.VectorSetBuilder{}
	err := fl.AddFont(opensans_regular_ttf, ranges...)
	if err != nil {
		dev.ReportError(err)
		return nil
	}
	return fl.Build(dev)
}
