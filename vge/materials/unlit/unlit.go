//

package unlit

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func UnlitFactory(dev *vk.Device, props vmodel.MaterialProperties) (
	sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
	ub := unlitMaterial{
		color: props.GetColor(vmodel.CAlbedo, mgl32.Vec4{0, 0, 0, 1}),
	}
	tx_albedo := props.GetImage(vmodel.TxAlbedo)
	if tx_albedo > 0 {
		ub.textured = 1
	}
	b := *(*[unsafe.Sizeof(unlitMaterial{})]byte)(unsafe.Pointer(&ub))
	return &UnlitMaterial{}, getUnlitLayout(dev), b[:], []vmodel.ImageIndex{tx_albedo}
}

type UnlitMaterial struct {
	dsMat *vk.DescriptorSet
}

func (u *UnlitMaterial) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

func (u *UnlitMaterial) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	scf := vscene.GetSimpleFrame(dc.Frame)
	if scf == nil {
		return // Simple frame not supported
	}
	rc := scf.GetCache()
	gp := dc.Pass.Get(kUnlitPipeline, func() interface{} {
		return u.NewPipeline(dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kUnlitInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &unlitInstances{ds: ds, sl: sl}
	}).(*unlitInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, u.dsMat).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kUnlitInstances, nil)
	}
}

func (u *UnlitMaterial) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	scf := vscene.GetSimpleFrame(dc.Frame)
	if scf == nil {
		return // Simple frame not supported
	}
	rc := scf.GetCache()
	gp := dc.Pass.Get(kUnlitSkinnedPipeline, func() interface{} {
		return u.NewPipeline(dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsWorld := scf.BindFrame()
	uli := rc.GetPerFrame(kUnlitInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &unlitInstances{ds: ds, sl: sl}
	}).(*unlitInstances)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dsMesh, slMesh := uc.Alloc()
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsWorld, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= 200 {
		rc.SetPerFrame(kUnlitInstances, nil)
	}
}

func (u *UnlitMaterial) NewPipeline(dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(rc.Device)
	if skinned {
		vmodel.AddInput(gp, vmodel.MESHKindSkinned)
	} else {
		vmodel.AddInput(gp, vmodel.MESHKindNormal)
	}
	la := vscene.GetUniformLayout(rc.Device)
	la2 := getUnlitLayout(rc.Device)
	gp.AddLayout(la)
	gp.AddLayout(la)
	gp.AddLayout(la2)
	if skinned {
		gp.AddLayout(la)
		gp.AddShader(vk.SHADERStageVertexBit, unlit_vert_skin_spv)
	} else {
		gp.AddShader(vk.SHADERStageVertexBit, unlit_vert_spv)
	}
	gp.AddShader(vk.SHADERStageFragmentBit, unlit_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(dc.Pass)
	return gp
}

type unlitMaterial struct {
	color    mgl32.Vec4
	textured float32
}

type unlitInstances struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kUnlitLayout = vk.NewKey()
var kUnlitPipeline = vk.NewKey()
var kUnlitSkinnedPipeline = vk.NewKey()
var kUnlitWorld = vk.NewKey()
var kUnlitInstances = vk.NewKey()

func getUnlitLayout(dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(dev)
	return dev.Get(kUnlitLayout, func() interface{} {
		return la.AddBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 1)
	}).(*vk.DescriptorLayout)
}
