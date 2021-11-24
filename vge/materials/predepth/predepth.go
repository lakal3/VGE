//

package predepth

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type PreDepthPass struct {
	DC        vmodel.DrawContext
	Cmd       *vk.Command
	BindFrame func() *vk.DescriptorSet
	OnBegin   func()
}

func (p *PreDepthPass) GetCache() *vk.RenderCache {
	return p.DC.Frame.GetCache()
}

func (p *PreDepthPass) Begin() (atEnd func()) {
	if p.OnBegin != nil {
		p.OnBegin()
	}
	return func() {
		if p.DC.List != nil {
			p.Cmd.Draw(p.DC.List)
		}
	}
}

func (p *PreDepthPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader) {
	rc := p.DC.Frame.GetCache()
	gp := p.DC.Pass.Get(kPreDepthPipeline, func() interface{} {
		return p.newPipeline(false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := p.BindFrame()
	uli := rc.GetPerFrame(kPreDepthInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &preDepthInstances{ds: ds, sl: sl}
	}).(*preDepthInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	p.DC.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kPreDepthInstances, nil)
	}
}

func (p *PreDepthPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader, aniMatrix []mgl32.Mat4) {
	rc := p.DC.Frame.GetCache()
	gp := p.DC.Pass.Get(kPreDepthSkinnedPipeline, func() interface{} {
		return p.newPipeline(true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := p.BindFrame()
	uli := rc.GetPerFrame(kPreDepthInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &preDepthInstances{ds: ds, sl: sl}
	}).(*preDepthInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dsMesh, slMesh := uc.Alloc()
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	p.DC.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kPreDepthInstances, nil)
	}
}

func (p *PreDepthPass) newPipeline(skinned bool) *vk.GraphicsPipeline {
	rc := p.DC.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(rc.Device)
	if skinned {
		vmodel.AddInput(gp, vmodel.MESHKindSkinned)
	} else {
		vmodel.AddInput(gp, vmodel.MESHKindNormal)
	}
	la := vscene.GetUniformLayout(rc.Device)
	gp.AddLayout(la)
	gp.AddLayout(la)
	if skinned {
		gp.AddLayout(la)
		gp.AddShader(vk.SHADERStageVertexBit, predepth_vert_skin_spv)
	} else {
		gp.AddShader(vk.SHADERStageVertexBit, predepth_vert_spv)
	}
	gp.AddShader(vk.SHADERStageFragmentBit, predepth_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(p.DC.Pass)
	return gp
}

type preDepthInstances struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kPreDepthPipeline = vk.NewKey()
var kPreDepthSkinnedPipeline = vk.NewKey()
var kPreDepthWorld = vk.NewKey()
var kPreDepthInstances = vk.NewKey()
