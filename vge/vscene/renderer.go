package vscene

import (
	"github.com/lakal3/vge/vge/vk"
)

type Renderer interface {
	vk.Disposable
	Setup(ctx vk.APIContext, dev *vk.Device, mainImage vk.ImageDescription, images int)
	Render(camera Camera, sc *Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo)
}
