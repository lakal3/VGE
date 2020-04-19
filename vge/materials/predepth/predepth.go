//

//go:generate glslangValidator.exe -V predepth.vert.glsl -o predepth.vert.spv
//go:generate glslangValidator.exe -V -DSKINNED=1 predepth.vert.glsl -o predepth.vert_skin.spv
//go:generate glslangValidator.exe -V predepth.frag.glsl -o predepth.frag.spv
//go:generate packspv -p predepth .

package predepth

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type PreDepthPass struct {
	DC      vmodel.DrawContext
	Cmd     *vk.Command
	OnBegin func()
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

func (p *PreDepthPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex) {
	rc := p.DC.Cache
	gp := p.DC.Pass.Get(rc.Ctx, kPreDepthPipeline, func(ctx vk.APIContext) interface{} {
		return p.newPipeline(ctx, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := rc.GetPerFrame(kPreDepthWorld, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		f := vscene.GetFrame(rc)
		f.CopyTo(sl)
		return ds
	}).(*vk.DescriptorSet)
	uli := rc.GetPerFrame(kPreDepthInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
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

func (p *PreDepthPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex, aniMatrix []mgl32.Mat4) {
	rc := p.DC.Cache
	gp := p.DC.Pass.Get(rc.Ctx, kPreDepthSkinnedPipeline, func(ctx vk.APIContext) interface{} {
		return p.newPipeline(ctx, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := rc.GetPerFrame(kPreDepthWorld, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		f := vscene.GetFrame(rc)
		f.CopyTo(sl)
		return ds
	}).(*vk.DescriptorSet)
	uli := rc.GetPerFrame(kPreDepthInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &preDepthInstances{ds: ds, sl: sl}
	}).(*preDepthInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dsMesh, slMesh := uc.Alloc(p.DC.Cache.Ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	p.DC.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kPreDepthInstances, nil)
	}
}

func (p *PreDepthPass) newPipeline(ctx vk.APIContext, skinned bool) *vk.GraphicsPipeline {
	rc := p.DC.Cache
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	if skinned {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
	} else {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
	}
	la := vscene.GetUniformLayout(ctx, rc.Device)
	gp.AddLayout(ctx, la)
	gp.AddLayout(ctx, la)
	if skinned {
		gp.AddLayout(ctx, la)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, predepth_vert_skin_spv)
	} else {
		gp.AddShader(ctx, vk.SHADERStageVertexBit, predepth_vert_spv)
	}
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, predepth_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, p.DC.Pass)
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
