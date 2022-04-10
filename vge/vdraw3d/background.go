package vdraw3d

import (
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"unsafe"
)

func DrawBackground(fl *FreezeList, model *vmodel.Model, image vmodel.ImageIndex) {
	fl.Add(&bgDrawer{model: model, image: image})
}

type bgDrawer struct {
	model      *vmodel.Model
	image      vmodel.ImageIndex
	boundIndex float32
}

func (b *bgDrawer) Clone() Frozen {
	return &bgDrawer{model: b.model, image: b.image}
}

var kBgPipeline = vk.NewKey()

func (b *bgDrawer) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	return storageOffset
}

func (b *bgDrawer) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(UpdateFrame)
	if ok && b.model != nil {
		return true
	}
	_, ok = phase.(RenderProbe)
	if ok {
		return true
	}
	_, ok = phase.(RenderColor)
	return ok
}

func (b *bgDrawer) Render(fi *vk.FrameInstance, phase Phase) {
	uf, ok := phase.(UpdateFrame)
	if ok {
		b.boundIndex = uf.AddView(b.model.GetImageView(b.image))
	}
	rd, ok := phase.(RenderColor)
	if ok {
		pl := b.getPipeline(fi.Device(), rd.Pass, rd.Shaders)
		ptr, offset := rd.DL.AllocPushConstants(4 * 9)
		pc := unsafe.Slice((*float32)(ptr), 9)
		pc[8] = b.boundIndex
		rd.DL.Draw(pl, 0, 12).AddDescriptor(0, rd.DSFrame).AddPushConstants(4*9, offset)
	}
	rp, ok := phase.(RenderProbe)
	if ok {
		pl := b.getProbePipeline(fi.Device(), rp.Pass, rp.Shaders)
		ptr, offset := rp.DL.AllocPushConstants(4 * 9)
		pc := unsafe.Slice((*float32)(ptr), 9)
		pc[8] = b.boundIndex
		rp.DL.Draw(pl, 0, 12).AddDescriptors(rp.DSFrame, rp.DSProbeFrame).AddPushConstants(4*9, offset)
	}
}

func (b *bgDrawer) getPipeline(dev *vk.Device, pass *vk.GeneralRenderPass, st *shaders.Pack) *vk.GraphicsPipeline {
	return pass.Get(kBgPipeline, func() interface{} {
		pl := vk.NewGraphicsPipeline(dev)
		pl.AddLayout(GetFrameLayout(dev))
		pl.AddPushConstants(vk.SHADERStageAll, 4*9)
		code := st.MustGet(dev, "background")
		pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
		pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
		pl.Create(pass)
		return pl
	}).(*vk.GraphicsPipeline)
}

func (b *bgDrawer) getProbePipeline(dev *vk.Device, pass *vk.GeneralRenderPass, st *shaders.Pack) *vk.GraphicsPipeline {
	return pass.Get(kBgPipeline, func() interface{} {
		pl := vk.NewGraphicsPipeline(dev)
		pl.AddLayout(GetFrameLayout(dev))
		pl.AddLayout(GetShadowFrameLayout(dev))
		pl.AddPushConstants(vk.SHADERStageAll, 4*9)
		code := st.MustGet(dev, "probe_background")
		pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
		pl.AddShader(vk.SHADERStageGeometryBit, code.Geometry)
		pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
		pl.Create(pass)
		return pl
	}).(*vk.GraphicsPipeline)
}
