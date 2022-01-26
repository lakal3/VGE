package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

// Drawable entity that can a filled path or glyph or shape
type Drawable interface {
	GetDrawData() (areas []DrawArea, segments []DrawSegment)
	GetGlyph() (view *vk.ImageView, layer uint32)
	GetShape() (shape *Shape)
}

// DrawArea for drawable. For glyph and shapes this is actual area to draw with some margins.
// For filled path this is size of each individual areas of path
type DrawArea struct {
	Min  mgl32.Vec2
	Max  mgl32.Vec2
	From uint32
	To   uint32
}

// DrawSegment data that varies by shape type
type DrawSegment struct {
	V1 float32
	V2 float32
	V3 float32
	V4 float32
}
