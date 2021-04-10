package forward

import (
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type FrameLightPhase struct {
	F     *Frame
	Cache *vk.RenderCache
}

func (f FrameLightPhase) GetCache() *vk.RenderCache {
	return f.Cache
}

func (f FrameLightPhase) Begin() (atEnd func()) {
	return func() {
		f.F.writeDynamicFrame(f.Cache)
		f.F.writeFrame(f.Cache)
	}
}

func (f FrameLightPhase) AddLight(standard vscene.Light, shadowMap *vk.ImageView, sampler *vk.Sampler) {
	idx := vmodel.ImageIndex(-1)
	if shadowMap != nil {
		idx = f.F.SetFrameImage(f.Cache, shadowMap, sampler)
	}
	if idx >= 0 {
		standard.Direction = standard.Direction.Vec3().Vec4(float32(idx))
	}
	f.F.AddLight(standard)
}

func (f FrameLightPhase) AddSpecialLight(special interface{}, shadowMap *vk.ImageView, samples *vk.Sampler) bool {
	return false
}
