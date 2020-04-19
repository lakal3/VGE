package vglyph

import (
	"errors"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"image/color"
)

const (
	MAXGlyphSets = 16
)

type GlyphSetIndex int
type MaskIndex int

const (
	NullSet = GlyphSetIndex(-1)
)

// Palette is set of glyph sets and masks. Palette, when built maps to single descriptor that can be bound to glyph shader.
type Palette struct {
	noMasks   MaskIndex
	usedMask  MaskIndex
	maskSize  int
	pool      *vk.MemoryPool
	dsPool    *vk.DescriptorPool
	ds        *vk.DescriptorSet
	maskImage *vk.Image
	glyphSets []*GlyphSet
	sampler   *vk.Sampler
}

func (pl *Palette) Dispose() {
	if pl.dsPool != nil {
		pl.dsPool.Dispose()
		pl.dsPool, pl.ds = nil, nil
		pl.glyphSets, pl.sampler = nil, nil
	}
	if pl.pool != nil {
		pl.pool.Dispose()
		pl.pool, pl.maskImage = nil, nil

	}
}

var kThemeLayout = vk.NewKeys(2)

// NewPalette initialized new palette and allocates room for mask image. Mask image is single multilayered image
// that glyph shader can use to amplify forecolor or backcolor. Masks can be used to create gradient effects to
// rendered glyph
func NewPalette(ctx vk.APIContext, dev *vk.Device, noMasks int, maskSize int) *Palette {

	laTheme := getThemeLayout(ctx, dev)
	if maskSize == 0 {
		maskSize = 128
	}
	desc := vk.ImageDescription{Layers: uint32(noMasks + 1), MipLevels: 1, Width: uint32(maskSize), Height: uint32(maskSize),
		Depth: 1, Format: vk.FORMATR8g8b8a8Unorm}
	th := &Palette{noMasks: MaskIndex(noMasks + 1), maskSize: maskSize}
	th.pool = vk.NewMemoryPool(dev)
	th.maskImage = th.pool.ReserveImage(ctx, desc, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	th.pool.Allocate(ctx)
	th.dsPool = vk.NewDescriptorPool(ctx, laTheme, 1)
	th.ds = th.dsPool.Alloc(ctx)
	th.ComputeMask(ctx, dev, func(x, y, maskSize int) color.RGBA {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	})
	return th
}

// ComputeMask fills one mask using given function.
func (th *Palette) ComputeMask(ctx vk.APIContext, dev *vk.Device, compute func(x, y, maskSize int) color.RGBA) MaskIndex {
	cp := vmodel.NewCopier(ctx, dev)
	defer cp.Dispose()
	tmp := make([]byte, th.maskSize*th.maskSize*4)
	for y := 0; y < th.maskSize; y++ {
		for x := 0; x < th.maskSize; x++ {
			c := compute(x, y, th.maskSize)
			idx := (x + y*th.maskSize) * 4
			tmp[idx], tmp[idx+1], tmp[idx+2], tmp[idx+3] = c.R, c.G, c.B, c.A
		}
	}
	mi := th.usedMask
	rg := th.maskImage.FullRange()
	if th.usedMask == 0 {
		cp.SetLayout(th.maskImage, rg, vk.IMAGELayoutShaderReadOnlyOptimal)
	}
	rg.FirstLayer, rg.LayerCount = uint32(th.usedMask), 1
	cp.CopyToImage(th.maskImage, "raw", tmp, rg, vk.IMAGELayoutShaderReadOnlyOptimal)
	th.sampler = GetPalletteSampler(ctx, dev)
	th.ds.WriteImage(ctx, 1, 0, th.maskImage.DefaultView(ctx), th.sampler)

	th.usedMask++
	return mi
}

var kPalletteSampler = vk.NewKey()

// Sampler used in gyph shader. Clamps samping to edge
func GetPalletteSampler(ctx vk.APIContext, dev *vk.Device) *vk.Sampler {
	return dev.Get(ctx, kPalletteSampler, func(ctx vk.APIContext) interface{} {
		return vk.NewSampler(ctx, dev, vk.SAMPLERAddressModeClampToEdge)
	}).(*vk.Sampler)
}

// Add glyph set to palette. You should add all glyph sets to palette before using it.
func (th *Palette) AddGlyphSet(ctx vk.APIContext, gs *GlyphSet) GlyphSetIndex {
	at := GlyphSetIndex(len(th.glyphSets))
	if at >= MAXGlyphSets {
		ctx.SetError(errors.New("Too many glyph sets"))
		return 0
	}
	view := gs.image.DefaultView(ctx)
	th.ds.WriteImage(ctx, 0, uint32(at), view, th.sampler)
	if at == 0 {
		for idx := uint32(1); idx < MAXGlyphSets; idx++ {
			th.ds.WriteImage(ctx, 0, idx, view, th.sampler)
		}
	}
	th.glyphSets = append(th.glyphSets, gs)
	return at
}

// GetSet retrieves glyph set from palette.
func (th *Palette) GetSet(index GlyphSetIndex) *GlyphSet {
	if index < 0 || index >= GlyphSetIndex(len(th.glyphSets)) {
		return nil
	}
	return th.glyphSets[index]
}

func getThemeLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	la1 := dev.Get(ctx, kThemeLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, MAXGlyphSets)
	}).(*vk.DescriptorLayout)
	la2 := dev.Get(ctx, kThemeLayout+1, func(ctx vk.APIContext) interface{} {
		return la1.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 1)
	}).(*vk.DescriptorLayout)
	return la2
}
