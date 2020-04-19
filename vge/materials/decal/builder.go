package decal

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type Builder struct {
	images []*ImageBuilder
	decals []decalInfo
}

type DecalIndex int

type ImageBuilder struct {
	kind    string
	content []byte
	usage   vk.ImageUsageFlags
	index   vmodel.ImageIndex
	desc    vk.ImageDescription
}

// Attach image to decal set.
func (b *Builder) AddImage(kind string, content []byte, usage vk.ImageUsageFlags) vmodel.ImageIndex {
	im := &ImageBuilder{kind: kind, content: content, index: vmodel.ImageIndex(len(b.images) + 1), usage: usage}
	b.images = append(b.images, im)
	return im.index
}

type decalInfo struct {
	props vmodel.MaterialProperties
	name  string
}

// Add new material to model builders. Model builders ShaderFactory is finally used to convert material properties
// to shader. Most of shader also build a descriptor set that links all static assets like color and textures
// of a material to a single Vulkan descriptor set.
func (b *Builder) AddDecal(name string, props vmodel.MaterialProperties) DecalIndex {
	mi := DecalIndex(len(b.decals))
	b.decals = append(b.decals, decalInfo{props: props, name: name})
	return mi
}

func (b *Builder) Build(ctx vk.APIContext, dev *vk.Device) *Set {
	s := &Set{dev: dev}
	s.pool = vk.NewMemoryPool(dev)
	s.images = make([]*vk.Image, len(b.images)+1)
	var imMaxLen uint64
	for idx, ib := range b.images {
		vasset.DescribeImage(ctx, ib.kind, &ib.desc, ib.content)
		imLen := ib.desc.ImageSize()
		if imLen > imMaxLen {
			imMaxLen = imLen
		}
		desc := ib.desc
		/* Mips later
		ib.orignalMips = desc.MipLevels
		if desc.MipLevels < mb.MipLevels && mb.canDoMips(desc) {
			desc.MipLevels = mb.MipLevels
			ib.desc.MipLevels = mb.MipLevels
			ib.usage |= vk.IMAGEUsageStorageBit
		}

		*/
		img := s.pool.ReserveImage(ctx, desc, ib.usage)
		s.images[idx+1] = img
	}
	s.imageKey = vk.NewKeys(uint64(len(s.images) + 1))
	s.pool.Allocate(ctx)
	for _, dl := range b.decals {
		s.decals = append(s.decals, Decal{
			Name: dl.name, AlbedoFactor: dl.props.GetColor(vmodel.CAlbedo, mgl32.Vec4{1, 1, 1, 1}),
			txAlbedo: dl.props.GetImage(vmodel.TxAlbedo), txNormal: dl.props.GetImage(vmodel.TxBump),
			txMetalRoughness:  dl.props.GetImage(vmodel.TxMetallicRoughness),
			MetalnessFactor:   dl.props.GetFactor(vmodel.FMetalness, 1),
			RoughnessFactor:   dl.props.GetFactor(vmodel.FRoughness, 1),
			NormalAttenuation: dl.props.GetFactor(vmodel.FNormalAttenuation, 1),
			set:               s,
		})
	}
	cp := vmodel.NewCopier(ctx, dev)
	defer cp.Dispose()
	for idx, ib := range b.images {
		cp.CopyToImage(s.images[idx+1], ib.kind, ib.content, ib.desc.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	}
	return s
}
