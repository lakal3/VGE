//

package vglyph

import (
	"errors"
	"image"

	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
)

type ColorIndex uint32

const (
	// Foreground / Background ratio in RED channel. Alpha channel is alpha
	RED = ColorIndex(0)
	// Foreground / Background ratio in GREEN channel. Alpha channel is alpha
	GREEN = ColorIndex(1)
	// Foreground / Background ratio in BLUE channel. Alpha channel is alpha
	BLUE = ColorIndex(2)
	// Foreground / Background ratio in RED channel. RED channel is also alpha.
	GRAY = ColorIndex(3)
	// Foreground / Background ratio in RED channel. GREEN channel controls alpha. If GREEN > 0.5 then alpha = 0 else alpha = 1
	RED_GREENA = ColorIndex(4)
)
const MAXImageWidth = 32768

type GlyphBuilder struct {
	name       string
	desc       vk.ImageDescription
	kind       string
	content    []byte
	mainColor  ColorIndex
	offset     image.Point
	charOffset image.Point
	edges      image.Rectangle
	margin     int
}

type SetBuilder struct {
	images         []*GlyphBuilder
	kind           SetKind
	baselineOffset int
}

type buildInfo struct {
	gs        *GlyphSet
	owner     vk.Owner
	buffers   []*vk.Buffer
	sUniforms []*vk.Slice
	dsIn      []*vk.DescriptorSet
	dsOut     *vk.DescriptorSet
	pl        *vk.ComputePipeline
}

// NewSetBuilder initialize new build that can build glyphset using images. Images can have foreground / background ratio (SETGrayScale) or full rgba image (SETRGBA)
// All images in one glyph set must have same kind
func NewSetBuilder(kind SetKind) *SetBuilder {
	return &SetBuilder{kind: kind}
}

// Add glyph to glyph set
func (sb *SetBuilder) AddGlyph(name string, mainColor ColorIndex, kind string, content []byte) {
	sb.AddEdgedGlyph(name, mainColor, kind, content, image.Rectangle{})
}

// Add edged glyph. See docs/vui.md) for more info
func (sb *SetBuilder) AddEdgedGlyph(name string, mainColor ColorIndex, kind string, content []byte, edges image.Rectangle) *GlyphBuilder {
	gb := &GlyphBuilder{name: name, kind: kind, content: content, mainColor: mainColor, edges: edges}
	vasset.DescribeImage(kind, &gb.desc, content)
	sb.images = append(sb.images, gb)
	return gb
}

// Add computed glyph. Function will panic it set is not SETGrayScale
func (sb *SetBuilder) AddComputedGray(name string, size image.Point, edges image.Rectangle,
	intensity func(x, y int) (color float32, alpha float32)) {
	if sb.kind != SETGrayScale {
		panic("AddComputedGray can be used only with SETGrayScale builder")
	}
	content := make([]byte, size.X*size.Y*4)
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			offset := 4 * (x + y*size.X)
			c, a := intensity(x, y)
			content[offset] = byte(255 * c)
			content[offset+3] = byte(255 * a)
		}
	}
	gb := &GlyphBuilder{kind: "font", name: name, edges: edges, mainColor: RED,
		desc: vk.ImageDescription{Width: uint32(size.X), Height: uint32(size.Y), Depth: 1,
			MipLevels: 1, Layers: 1, Format: vk.FORMATR8g8b8a8Unorm}}
	gb.content = content

	sb.images = append(sb.images, gb)
}

// Build will create actual GlyphSet and load glyph images to GPU
func (sb *SetBuilder) Build(dev *vk.Device) *GlyphSet {
	w := 256
	h := 0
	for h == 0 || h > w {
		w = 2 * w
		h = sb.calcSize(w)
		if w > MAXImageWidth {
			dev.ReportError(errors.New("Glyph set too large"))
			return nil
		}
	}
	bi := &buildInfo{}
	defer bi.owner.Dispose()
	bi.gs = newGlyphSet(dev, len(sb.images), w, h, sb.kind)
	bi.prepare(dev, sb.images)
	bi.render(dev, sb.images)
	sb.addGlyphs(bi.gs)
	bi.gs.Advance = DefaultAdvance
	return bi.gs
}

func (sb *SetBuilder) calcSize(w int) int {
	h := 0
	var offset image.Point
	var lHeight int
	for _, im := range sb.images {
		wTmp := int(im.desc.Width) + im.margin*2
		if wTmp > w {
			return 0 // Won't fit
		}
		if wTmp+offset.X > w {
			h += lHeight
			offset = image.Point{0, h}
			lHeight = 0
		}
		im.offset = offset
		if lHeight < int(im.desc.Height)+2*im.margin {
			lHeight = int(im.desc.Height) + 2*im.margin
		}
		offset.X = offset.X + wTmp
	}
	return h + lHeight
}

func (sb *SetBuilder) addGlyphs(gs *GlyphSet) {
	for _, gb := range sb.images {
		max := gb.offset.Add(image.Pt(int(gb.desc.Width)+2*gb.margin-2, int(gb.desc.Height)+2*gb.margin-2))
		gs.glyphs[gb.name] = Glyph{Name: gb.name, CharOffset: gb.charOffset,
			Location: image.Rectangle{Min: gb.offset.Add(image.Pt(1, 1)), Max: max}, Edges: gb.edges}
	}
}

var kInputLayout = vk.NewKeys(2)
var kOutputLayout = vk.NewKey()
var kCopyPipeline = vk.NewKey()
var kRGBCopyPipeline = vk.NewKey()

func (bi *buildInfo) prepare(dev *vk.Device, images []*GlyphBuilder) {
	pool := vk.NewMemoryPool(dev)
	bi.owner.AddChild(pool)
	bi.buffers = make([]*vk.Buffer, len(images))
	for idx, im := range images {
		bi.buffers[idx] = pool.ReserveBuffer(im.desc.ImageSize(), true, vk.BUFFERUsageUniformTexelBufferBit)
	}
	bUniforms := pool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment*uint64(len(images)), true,
		vk.BUFFERUsageUniformBufferBit)
	pool.Allocate()
	laIn1 := dev.Get(kInputLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	laIn := dev.Get(kInputLayout+1, func() interface{} {
		return laIn1.AddBinding(vk.DESCRIPTORTypeUniformTexelBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	laOut := dev.Get(kOutputLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)

	if bi.gs.kind == SETRGBA {
		bi.pl = dev.Get(kRGBCopyPipeline, func() interface{} {
			cp := vk.NewComputePipeline(dev)
			cp.AddShader(copy_rgb_comp_spv)
			cp.AddLayout(laIn)
			cp.AddLayout(laOut)
			cp.Create()
			return cp
		}).(*vk.ComputePipeline)
	} else {
		bi.pl = dev.Get(kCopyPipeline, func() interface{} {
			cp := vk.NewComputePipeline(dev)
			cp.AddShader(copy_comp_spv)
			cp.AddLayout(laIn)
			cp.AddLayout(laOut)
			cp.Create()
			return cp
		}).(*vk.ComputePipeline)
	}
	dpOut := vk.NewDescriptorPool(laOut, 1)
	bi.owner.AddChild(dpOut)
	bi.dsOut = dpOut.Alloc()
	dpIn := vk.NewDescriptorPool(laIn, len(images))
	bi.owner.AddChild(dpIn)
	bi.dsIn = make([]*vk.DescriptorSet, len(images))
	bi.sUniforms = make([]*vk.Slice, len(images))
	for idx, im := range images {
		ds := dpIn.Alloc()
		sl := bUniforms.Slice(uint64(idx)*vk.MinUniformBufferOffsetAlignment, uint64(idx+1)*vk.MinUniformBufferOffsetAlignment)
		ds.WriteSlice(0, 0, sl)
		view := vk.NewBufferView(bi.buffers[idx], im.desc.Format)
		bi.owner.AddChild(view)
		ds.WriteBufferView(1, 0, view)
		bi.sUniforms[idx] = sl
		bi.dsIn[idx] = ds
		if im.kind == "font" {
			copy(bi.buffers[idx].Bytes(), im.content)
		} else {
			vasset.LoadImage(im.kind, im.content, bi.buffers[idx])
		}
	}
	fr := bi.gs.image.FullRange()
	fr.Layout = vk.IMAGELayoutGeneral
	view := vk.NewImageView(bi.gs.image, &fr)
	bi.owner.AddChild(view)
	bi.dsOut.WriteImage(0, 0, view, nil)
}

func (bi *buildInfo) render(dev *vk.Device, images []*GlyphBuilder) {
	cmd := vk.NewCommand(dev, vk.QUEUEComputeBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	ir := bi.gs.image.FullRange()
	cmd.SetLayout(bi.gs.image, &ir, vk.IMAGELayoutGeneral)
	for idx, im := range images {
		ubfs := []float32{float32(im.desc.Width) + float32(2*im.margin), float32(im.desc.Height) + float32(2*im.margin),
			float32(im.offset.X), float32(im.offset.Y), float32(im.mainColor), float32(im.margin)}
		copy(bi.sUniforms[idx].Content, vk.Float32ToBytes(ubfs))
		cmd.Compute(bi.pl, uint32((int(im.desc.Width)+2*im.margin)/16+1),
			uint32((int(im.desc.Height)+2*im.margin)/16+1), 1, bi.dsIn[idx], bi.dsOut)
	}
	cmd.SetLayout(bi.gs.image, &ir, vk.IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
}
