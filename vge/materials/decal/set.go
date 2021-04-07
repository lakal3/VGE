package decal

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/forward"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type Decal struct {
	Name              string
	AlbedoFactor      mgl32.Vec4
	MetalnessFactor   float32
	RoughnessFactor   float32
	NormalAttenuation float32
	At                mgl32.Mat4
	txAlbedo          vmodel.ImageIndex
	txNormal          vmodel.ImageIndex
	txMetalRoughness  vmodel.ImageIndex
	set               *Set
}

type Set struct {
	dev      *vk.Device
	pool     *vk.MemoryPool
	images   []*vk.Image
	decals   []Decal
	imageKey vk.Key
}

func (s *Set) Dispose() {
	if s.pool != nil {
		s.pool.Dispose()
		s.pool, s.images = nil, nil
	}
}

func (s *Set) NewInstance(name string, at mgl32.Mat4) *Decal {
	for _, dc := range s.decals {
		if dc.Name == name {
			instance := dc
			instance.At = at
			return &instance
		}
	}
	return nil
}

func (s *Set) addImage(f *forward.Frame, rc *vk.RenderCache, setIndex vmodel.ImageIndex) (frameIndex vmodel.ImageIndex) {
	if setIndex == 0 {
		return 0
	}
	sampler := vmodel.GetDefaultSampler(rc.Ctx, rc.Device)
	frameIndex = rc.GetPerFrame(s.imageKey+vk.Key(setIndex), func(ctx vk.APIContext) interface{} {
		return f.SetFrameImage(rc, s.images[setIndex].DefaultView(rc.Ctx), sampler)
	}).(vmodel.ImageIndex)
	return
}

func (s *Set) getImage(rc *vk.RenderCache, setIndex vmodel.ImageIndex) (frameIndex vmodel.ImageIndex) {
	if setIndex == 0 {
		return 0
	}
	frameIndex = rc.GetPerFrame(s.imageKey+vk.Key(setIndex), func(ctx vk.APIContext) interface{} {
		return vmodel.ImageIndex(-1)
	}).(vmodel.ImageIndex)
	return
}
