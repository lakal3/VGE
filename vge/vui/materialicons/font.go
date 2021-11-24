//

package materialicons

import (
	_ "embed"

	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vk"
)

//go:embed MaterialIcons_Regular.ttf
var materialicons_regular_ttf []byte

func NewGlyphSet(dev *vk.Device, ranges ...vglyph.Range) *vglyph.GlyphSet {
	fl := &vglyph.VectorSetBuilder{}
	err := fl.AddFont(materialicons_regular_ttf, ranges...)
	if err != nil {
		dev.ReportError(err)
		return nil
	}
	return fl.Build(dev)

}

func NewDefaultGlyphSet(dev *vk.Device) *vglyph.GlyphSet {
	return NewGlyphSet(dev, vglyph.Range{From: rune(0xe000), To: rune(0xea00)})
}
