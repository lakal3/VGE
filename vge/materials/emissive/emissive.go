//

package emissive

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const MaxInstances = 200

type EmissiveMaterial struct {
	// Static color used to render mesh
	Color mgl32.Vec4
}

func (e *EmissiveMaterial) SetDescriptor(dsMat *vk.DescriptorSet) {
}

func (e *EmissiveMaterial) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4) {
	scf := vscene.GetSimpleFrame(dc.Frame)
	if scf == nil {
		return // Not supported
	}
	rc := scf.GetCache()
	gp := dc.Pass.Get(rc.Ctx, kEmissivePipeline, func(ctx vk.APIContext) interface{} {
		return e.newPipeline(ctx, dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kEmissiveInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &emissiveInstances{ds: ds, sl: sl}
	}).(*emissiveInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	copy(uli.sl.Content[MaxInstances*64+uli.count*16:MaxInstances*64+uli.count*16+16], vk.Float32ToBytes(e.Color[:]))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kEmissiveInstances, nil)
	}
}

func (e *EmissiveMaterial) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4) {
	scf := vscene.GetSimpleFrame(dc.Frame)
	if scf == nil {
		return // Not supported
	}
	rc := scf.GetCache()
	gp := dc.Pass.Get(rc.Ctx, kEmissiveSkinnedPipeline, func(ctx vk.APIContext) interface{} {
		return e.newPipeline(ctx, dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kEmissiveInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &emissiveInstances{ds: ds, sl: sl}
	}).(*emissiveInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	copy(uli.sl.Content[MaxInstances*64+uli.count*16:MaxInstances*64+uli.count*16+16], vk.Float32ToBytes(e.Color[:]))

	dsMesh, slMesh := uc.Alloc(rc.Ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kEmissiveInstances, nil)
	}
}

func (e *EmissiveMaterial) newPipeline(ctx vk.APIContext, dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
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
		gp.AddShader(ctx, vk.SHADERStageVertexBit, emissive_vert_skin_spv)
	} else {
		gp.AddShader(ctx, vk.SHADERStageVertexBit, emissive_vert_spv)
	}
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, emissive_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, dc.Pass)
	return gp
}

type emissiveInstances struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kEmissivePipeline = vk.NewKey()
var kEmissiveSkinnedPipeline = vk.NewKey()
var kEmissiveWorld = vk.NewKey()
var kEmissiveInstances = vk.NewKey()
