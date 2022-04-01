package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

// Brush define how to paint a shape
type Brush struct {
	// Color is primary color of shape
	Color mgl32.Vec4
	// Color2 is secondary color of shape. Final color of shape is color + u * Color2
	Color2 mgl32.Vec4
	// Image used for painting
	Image vk.VImageView
	// Sampler for image. Only used with Image
	Sampler *vk.Sampler

	// Transform from shape coordinates to UV coordinates for Color2 and Image sampling
	// If share size is for example 120 x 60 and you want uv coordinates from 0,0 to 1,1 (draw image once) use
	// a transformation that scales coordinates by (1.0 / 120.0, 1.0 / 60.0)
	UVTransform mgl32.Mat3
}

// SolidColor builds a brush with just one color
func SolidColor(color mgl32.Vec4) Brush {
	return Brush{Color: color}
}

// IsNil return true is brush is nil (=not set or fully transparent)
func (br *Brush) IsNil() bool {
	if br == nil {
		return true
	}
	if br.Color[3] == 0 {
		return true
	}
	return false
}
