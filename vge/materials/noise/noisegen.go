package noise

import (
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
)

type PerlinNoise struct {
	// Image size
	size int

	// Content, one float per pixel
	content []float32
}

func NewPerlinNoise(size int) *PerlinNoise {
	pn := &PerlinNoise{size: size}
	pn.content = make([]float32, size*size)
	return pn
}

func (pn *PerlinNoise) Add(factor float32, gridSize float32) {
	cn := int(float32(pn.size)/gridSize + 2)
	corners := make([]mgl32.Vec2, cn*cn)
	for idx, _ := range corners {
		v := mgl32.Vec2{rand.Float32()*2 - 1, rand.Float32()*2 - 1}.Normalize()
		corners[idx] = v
	}
	for y := 0; y < pn.size; y++ {
		for x := 0; x < pn.size; x++ {
			at := y*pn.size + x
			pn.content[at] += pn.calcAt(corners, x, y, cn, gridSize) * factor
		}
	}
}

func (pn *PerlinNoise) ToBytes() (content []byte) {
	var min, max float32
	for _, v := range pn.content {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	d := max - min
	result := make([]byte, len(pn.content))
	for idx, v := range pn.content {
		result[idx] = byte(255 * (v - min) / d)
	}
	return result
}

func (pn *PerlinNoise) AddToModel(mb *vmodel.ModelBuilder, usage vk.ImageUsageFlags) vmodel.ImageIndex {
	idx := mb.AddImage("raw", pn.ToBytes(), usage)
	mb.Images[idx].Desc = vk.ImageDescription{
		Width:     uint32(pn.size),
		Height:    uint32(pn.size),
		Depth:     1,
		Format:    vk.FORMATR8Unorm,
		Layers:    1,
		MipLevels: 1,
	}
	return idx
}

func (pn *PerlinNoise) calcAt(corners []mgl32.Vec2, x int, y int, cn int, gridSize float32) float32 {
	y0, x0 := int(float32(y)/gridSize), int(float32(x)/gridSize)
	y1, x1 := y0+1, x0+1
	dx0 := float32(x)/gridSize - float32(x0)
	dy0 := float32(y)/gridSize - float32(y0)
	dx1 := float32(x)/gridSize - float32(x1)
	dy1 := float32(y)/gridSize - float32(y1)
	c0 := pn.fromPoint(corners, x0, y0, dx0, dy0, cn)
	c1 := pn.fromPoint(corners, x1, y0, dx1, dy0, cn)
	l1 := lerp(c0, c1, curve(dx0))
	c0 = pn.fromPoint(corners, x0, y1, dx0, dy1, cn)
	c1 = pn.fromPoint(corners, x1, y1, dx1, dy1, cn)
	l2 := lerp(c0, c1, curve(dx0))
	return lerp(l1, l2, curve(dy0))
}

func curve(dx float32) float32 {
	// See: https://mrl.nyu.edu/~perlin/paper445.pdf
	return dx * dx * (3 - 2*dx)
}

func (pn *PerlinNoise) fromPoint(corners []mgl32.Vec2, cx int, cy int, dx float32, dy float32, cn int) float32 {
	v1 := corners[cx+cy*cn]
	v2 := mgl32.Vec2{dx, dy}
	return v1.Dot(v2)
}

func lerp(a0, a1, w float32) float32 {
	return (1-w)*a0 + w*a1
}
