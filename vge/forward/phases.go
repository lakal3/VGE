package forward

import (
	"github.com/lakal3/vge/vge/vk"
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
	return nil
}

func (f FrameLightPhase) AddLight(standard vscene.Light, special interface{}) {
	f.F.AddLight(standard)
}
