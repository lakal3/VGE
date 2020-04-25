package fplus

import (
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type ImageFilter interface {
	Filter(dc *vmodel.DrawContext)
}

type CopyFilter struct {
}

func (c CopyFilter) Filter(dc *vmodel.DrawContext) {
	copySrc(dc)
}

var kCopyPl = vk.NewKey()

func copySrc(dc *vmodel.DrawContext) {
	rc := dc.Cache
	pl := getCopyPl(rc.Ctx, rc.Device, dc.Pass)
	ds := GetPhaseDescriptor(rc)
	dc.Draw(pl, 0, 6).AddDescriptors(ds)
}

func getCopyPl(ctx vk.APIContext, dev *vk.Device, rp vk.RenderPass) *vk.GraphicsPipeline {
	return dev.Get(ctx, kCopyPl, func(ctx vk.APIContext) interface{} {
		pl := vk.NewGraphicsPipeline(ctx, dev)
		pl.AddLayout(ctx, GetBindImageLayout(ctx, dev))
		pl.AddShader(ctx, vk.SHADERStageVertexBit, copy_vert_spv)
		pl.AddShader(ctx, vk.SHADERStageFragmentBit, copy_frag_spv)
		pl.Create(ctx, rp)
		return pl
	}).(*vk.GraphicsPipeline)
}
