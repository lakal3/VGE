// +build test

package vglyph

import "github.com/lakal3/vge/vge/vk"

func (gs *GlyphSet) GetImage() *vk.Image {
	return gs.image
}
