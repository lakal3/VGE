package vdraw

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"unsafe"
)

// Glyph is drawable shape. You generate Glyphs with GlyphSet
type Glyph struct {
	min   mgl32.Vec2
	max   mgl32.Vec2
	set   *GlyphSet
	index uint32
}

func (gl *Glyph) GetShape() (shape *Shape) {
	return nil
}

func (gl *Glyph) GetGlyph() (view *vk.ImageView, layer uint32) {
	return gl.set.vGlyph, gl.index
}

func (gl *Glyph) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	area := DrawArea{Min: gl.min, Max: gl.max, From: 0, To: 2}
	return []DrawArea{area}, []DrawSegment{
		DrawSegment{V1: gl.min[0], V2: gl.max[0], V3: gl.min[1]},
		DrawSegment{V1: gl.min[0], V2: gl.max[0], V3: gl.max[1]},
	}
}

func (gl Glyph) Bounds() (min mgl32.Vec2, max mgl32.Vec2) {
	return gl.min, gl.max
}

// GlyphSet is used to build Glyph. Glyphs are signed depth field recorded from Paths or using a custom function
type GlyphSet struct {
	builders []builder
	mImage   *vk.MemoryPool
	imGlyph  *vk.Image
	vGlyph   *vk.ImageView
}

func (gs *GlyphSet) Dispose() {
	if gs.mImage != nil {
		gs.imGlyph.Dispose()
		gs.mImage.Dispose()
		gs.mImage, gs.imGlyph, gs.vGlyph = nil, nil, nil
	}
}

type DepthFunc func(size image.Point, at image.Point) (depth float32)

const (
	glMargin = 3
)

type builder struct {
	p  *Path
	fn DepthFunc
	g  *Glyph
}

// AddPath creates a new glyph using Path
func (gs *GlyphSet) AddPath(p *Path) *Glyph {
	g := &Glyph{set: gs, index: uint32(len(gs.builders))}
	gs.builders = append(gs.builders, builder{p: p, g: g})
	return g
}

// AddComputed create a glyph using given DepthFunc function.
func (gs *GlyphSet) AddComputed(df DepthFunc) *Glyph {
	g := &Glyph{set: gs, index: uint32(len(gs.builders))}
	gs.builders = append(gs.builders, builder{fn: df, g: g})
	return g
}

// AddRunes will add a range of runes from a font to GlyphSet
func (gs *GlyphSet) AddRunes(f *Font, from, to rune) error {
	for r := from; r <= to; r++ {
		err := f.AddToSet(r, gs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Build will build actual multilayer image that contains signed depth field for each added Glyph. Each images will have same size
// After GlyphSet is built, you may no longer add shapes to it.
func (gs *GlyphSet) Build(dev *vk.Device, size image.Point) {
	if gs.mImage != nil {
		dev.ReportError(errors.New("already built"))
		return
	}
	gs.build(dev, size, vk.IMAGELayoutShaderReadOnlyOptimal)
}

func (gs *GlyphSet) build(dev *vk.Device, size image.Point, layout vk.ImageLayout) {
	desc := vk.ImageDescription{Layers: uint32(len(gs.builders)), Width: uint32(size.X), Height: uint32(size.Y), Depth: 1,
		MipLevels: 1, Format: vk.FORMATR8Snorm}
	gs.mImage = vk.NewMemoryPool(dev)
	gs.imGlyph = gs.mImage.ReserveImage(desc, vk.IMAGEUsageSampledBit|vk.IMAGEUsageStorageBit|vk.IMAGEUsageTransferSrcBit)
	gs.mImage.Allocate()

	mTemp := vk.NewMemoryPool(dev)
	defer mTemp.Dispose()
	buffers := make([]*vk.Buffer, len(gs.builders))
	for idx, b := range gs.builders {
		if b.fn != nil {
			buffers[idx] = mTemp.ReserveBuffer(uint64(size.X*size.Y*4+16), true, vk.BUFFERUsageStorageBufferBit)
		} else {
			buffers[idx] = mTemp.ReserveBuffer(
				uint64(len(b.p.segments)+1)*uint64(unsafe.Sizeof(segment{})), true, vk.BUFFERUsageStorageBufferBit)
		}
	}
	mTemp.Allocate()
	for idx, b := range gs.builders {
		if b.fn != nil {
			gs.computeBuffer(buffers[idx], b.fn, size, b.g)
		} else {
			gs.copySegments(buffers[idx], b.p, size, b.g)
		}
	}

	cp := gs.getCopyPipeline(dev)
	pp := gs.getPathPipeline(dev)
	imLayout := gs.getImageLayout(dev)
	bufLayout := gs.getBufferLayout(dev)
	dpImage := vk.NewDescriptorPool(imLayout, len(buffers))
	defer dpImage.Dispose()
	dpBuffer := vk.NewDescriptorPool(bufLayout, len(buffers))
	defer dpBuffer.Dispose()
	views := make([]*vk.ImageView, len(buffers))
	defer func() {
		for _, v := range views {
			v.Dispose()
		}
	}()
	cmd := vk.NewCommand(dev, vk.QUEUEComputeBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	irFull := gs.imGlyph.FullRange()
	cmd.SetLayout(gs.imGlyph, &irFull, vk.IMAGELayoutGeneral)
	for idx, b := range gs.builders {
		ir := vk.ImageRange{FirstLayer: uint32(idx), LayerCount: 1, FirstMipLevel: 0, LevelCount: 1, Layout: vk.IMAGELayoutGeneral}
		v := vk.NewImageView(gs.imGlyph, &ir)
		views[idx] = v
		dsFrom, dsTo := dpBuffer.Alloc(), dpImage.Alloc()
		dsFrom.WriteBuffer(0, 0, buffers[idx])
		dsTo.WriteImage(0, 0, v, nil)
		if b.fn != nil {
			cmd.Compute(cp, uint32(size.X+15)/16, uint32(size.Y+15)/16, 1, dsFrom, dsTo)
		} else {
			cmd.Compute(pp, uint32(size.X+15)/16, uint32(size.Y+15)/16, 1, dsFrom, dsTo)
		}
	}
	cmd.SetLayout(gs.imGlyph, &irFull, layout)
	cmd.Submit()
	cmd.Wait()
	gs.vGlyph = gs.imGlyph.DefaultView()
}

var kCopyPipeline = vk.NewKey()
var kPathPipeline = vk.NewKey()
var kImageLayout = vk.NewKey()
var kStorageLayout = vk.NewKey()

func (gs *GlyphSet) computeBuffer(buffer *vk.Buffer, fn DepthFunc, size image.Point, gl *Glyph) {
	b := buffer.Bytes()
	gl.max = mgl32.Vec2{float32(size.X), float32(size.Y)}
	f := unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), size.X*size.Y+4)
	f[0] = float32(size.X)
	f[1] = float32(size.Y)
	s2 := size.Sub(image.Pt(glMargin*2, glMargin*2))
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			d := fn(s2, image.Pt(x-glMargin, y-glMargin))
			f[x+y*size.X+4] = d
		}
	}
}

func (gs *GlyphSet) copySegments(buffer *vk.Buffer, p *Path, size image.Point, gl *Glyph) {
	b := buffer.Bytes()
	area := p.Bounds()
	segments := unsafe.Slice((*segment)(unsafe.Pointer(&b[0])), len(p.segments)+1)
	segments[0].id = uint32(len(p.segments))
	dx, dy := area.To[0]-area.From[0], area.To[1]-area.From[1]
	d := dx
	if dy > dx {
		d = dy
	}
	dmx := glMargin / float32(size.X-2*glMargin)
	dmy := glMargin / float32(size.Y-2*glMargin)
	segments[0].from = mgl32.Vec2{area.From[0] - dmx, area.From[1] - dmy}

	segments[0].to = mgl32.Vec2{
		d * float32(size.X) / float32(size.X-2*glMargin),
		d * float32(size.Y) / float32(size.Y-2*glMargin),
	}
	gl.min = segments[0].from
	gl.max = gl.min.Add(mgl32.Vec2{d + 2*dmx, d + 2*d*dmy})
	// gl.max = area.To
	segments[0].mid = mgl32.Vec2{float32(size.X), float32(size.Y)}
	copy(segments[1:], p.segments)
}

func (gs *GlyphSet) getCopyPipeline(dev *vk.Device) *vk.ComputePipeline {
	imLayout := gs.getImageLayout(dev)
	bufLayout := gs.getBufferLayout(dev)
	return dev.Get(kCopyPipeline, func() interface{} {
		cp := vk.NewComputePipeline(dev)
		cp.AddShader(copy_comp_spv)
		cp.AddLayout(bufLayout)
		cp.AddLayout(imLayout)
		cp.Create()
		return cp
	}).(*vk.ComputePipeline)

}

func (gs *GlyphSet) getPathPipeline(dev *vk.Device) *vk.ComputePipeline {
	imLayout := gs.getImageLayout(dev)
	bufLayout := gs.getBufferLayout(dev)
	return dev.Get(kPathPipeline, func() interface{} {
		cp := vk.NewComputePipeline(dev)
		cp.AddShader(path_comp_spv)
		cp.AddLayout(bufLayout)
		cp.AddLayout(imLayout)
		cp.Create()
		return cp
	}).(*vk.ComputePipeline)

}

func (gs *GlyphSet) getBufferLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kStorageLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
}

func (gs *GlyphSet) getImageLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kImageLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
}
