//

package pbr

import (
	"github.com/lakal3/vge/vge/forward"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func PbrFactory(ctx vk.APIContext, dev *vk.Device, props vmodel.MaterialProperties) (
	sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
	tx_diffuse := props.GetImage(vmodel.TxAlbedo)
	tx_metalRoughness := props.GetImage(vmodel.TxMetallicRoughness)
	tx_emissive := props.GetImage(vmodel.TxEmissive)
	tx_normal := props.GetImage(vmodel.TxBump)
	mf := float32(0)
	if tx_metalRoughness != 0 {
		mf = 1
	}
	ub := pbrMaterial{
		albedoColor:     props.GetColor(vmodel.CAlbedo, getColorFactor(tx_diffuse)),
		emissiveColor:   props.GetColor(vmodel.CEmissive, getColorFactor(tx_emissive)),
		metalnessFactor: props.GetFactor(vmodel.FMetalness, mf),
		roughnessFactor: props.GetFactor(vmodel.FRoughness, 1),
	}
	if tx_normal > 0 {
		ub.normalMap = 1
	}

	b := *(*[unsafe.Sizeof(pbrMaterial{})]byte)(unsafe.Pointer(&ub))
	return &PbrMaterial{}, getPbrLayout(ctx, dev), b[:], []vmodel.ImageIndex{tx_diffuse, tx_normal, tx_metalRoughness, tx_emissive}
}

func getColorFactor(imIndex vmodel.ImageIndex) mgl32.Vec4 {
	if imIndex > 0 {
		return mgl32.Vec4{1, 1, 1, 1}
	}
	return mgl32.Vec4{0, 0, 0, 1}
}

type PbrMaterial struct {
	dsMat *vk.DescriptorSet
}

func (u *PbrMaterial) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

func (u *PbrMaterial) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	rc := dc.Cache
	gp := dc.Pass.Get(rc.Ctx, kPbrSkinnedPipeline, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := forward.BindFrame(rc)
	uli := rc.GetPerFrame(kPbrSkinnedInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &pbrInstance{ds: ds, sl: sl}
	}).(*pbrInstance)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dsMesh, slMesh := uc.Alloc(dc.Cache.Ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kPbrSkinnedInstances, nil)
	}
}

func (u *PbrMaterial) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	rc := dc.Cache
	gp := dc.Pass.Get(rc.Ctx, kPbrPipeline, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := forward.BindFrame(rc)
	uli := rc.GetPerFrame(kPbrInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &pbrInstance{ds: ds, sl: sl}
	}).(*pbrInstance)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kPbrInstances, nil)
	}
}

func (u *PbrMaterial) NewPipeline(ctx vk.APIContext, dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Cache
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	if skinned {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, pbr_vert_skin_spv)

	} else {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, pbr_vert_spv)
	}
	laFrame := forward.GetFrameLayout(ctx, rc.Device)
	la := vscene.GetUniformLayout(ctx, rc.Device)
	la2 := getPbrLayout(ctx, rc.Device)
	laUBF := vscene.GetUniformLayout(ctx, rc.Device)
	gp.AddLayout(ctx, laFrame)
	gp.AddLayout(ctx, la)
	gp.AddLayout(ctx, la2)
	if skinned {
		gp.AddLayout(ctx, laUBF) // Transform matrix
	}
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, pbr_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, dc.Pass)
	return gp
}

type pbrMaterial struct {
	albedoColor     mgl32.Vec4
	emissiveColor   mgl32.Vec4
	metalnessFactor float32
	roughnessFactor float32
	normalMap       float32
}

type pbrInstance struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kPbrLayout = vk.NewKey()
var kPbrPipeline = vk.NewKey()
var kPbrSkinnedPipeline = vk.NewKey()
var kPbrInstances = vk.NewKey()
var kPbrSkinnedInstances = vk.NewKey()

func getPbrLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(ctx, dev)
	return dev.Get(ctx, kPbrLayout, func(ctx vk.APIContext) interface{} {
		return la.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 4)
	}).(*vk.DescriptorLayout)
}
