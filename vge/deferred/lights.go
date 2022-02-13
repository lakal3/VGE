package deferred

import (
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
)

type DrawLights struct {
	ds       *vk.DescriptorSet
	fb       *vk.Framebuffer
	rp       *vk.GeneralRenderPass
	cmd      *vk.Command
	dl       *vk.DrawList
	cache    *vk.RenderCache
	pipeline *vk.GraphicsPipeline
	frame    *DeferredFrame
}

func (dl *DrawLights) AddLight(standard vscene.Light, special interface{}) {
	if dl.frame.lightUsed >= MAX_LIGHTS {
		return
	}
	dl.frame.LightsFrame.Lights[dl.frame.lightUsed] = standard
	dl.frame.lightUsed++
}

func (dl *DrawLights) GetCache() *vk.RenderCache {
	return dl.cache
}

func (dl *DrawLights) Begin() (atEnd func()) {
	dl.cmd.BeginRenderPass(dl.rp, dl.fb)
	dl.dl = &vk.DrawList{}
	return dl.atEnd
}

func (dl *DrawLights) atEnd() {
	dl.frame.writeLightsFrame()
	dl.dl.Draw(dl.pipeline, 0, 3).AddDescriptor(0, dl.frame.dsLight)
	dl.cmd.Draw(dl.dl)
}

func (r *Renderer) newLightsPipeline(dev *vk.Device) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(dev)
	gp.AddLayout(r.laLights)
	gp.AddShader(vk.SHADERStageVertexBit, lights_vert_spv)
	gp.AddShader(vk.SHADERStageFragmentBit, lights_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(r.rpFinal)
	return gp
}
