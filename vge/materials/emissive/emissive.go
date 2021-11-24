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
	gp := dc.Pass.Get(kEmissivePipeline, func() interface{} {
		return e.newPipeline(dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kEmissiveInstances, func() interface{} {
		ds, sl := uc.Alloc()
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
	gp := dc.Pass.Get(kEmissiveSkinnedPipeline, func() interface{} {
		return e.newPipeline(dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kEmissiveInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &emissiveInstances{ds: ds, sl: sl}
	}).(*emissiveInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	copy(uli.sl.Content[MaxInstances*64+uli.count*16:MaxInstances*64+uli.count*16+16], vk.Float32ToBytes(e.Color[:]))

	dsMesh, slMesh := uc.Alloc()
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kEmissiveInstances, nil)
	}
}

func (e *EmissiveMaterial) newPipeline(dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
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
		gp.AddShader(vk.SHADERStageVertexBit, emissive_vert_skin_spv)
	} else {
		gp.AddShader(vk.SHADERStageVertexBit, emissive_vert_spv)
	}
	gp.AddShader(vk.SHADERStageFragmentBit, emissive_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(dc.Pass)
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
