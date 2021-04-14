//

package debugmat

import (
	"github.com/lakal3/vge/vge/forward"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const maxInstaces = 200 // Must match to shader definition!

func DebugMatFactory(ctx vk.APIContext, dev *vk.Device, props vmodel.MaterialProperties) (
	sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {

	tx_diffuse := props.GetImage(vmodel.TxAlbedo)
	tx_specular := props.GetImage(vmodel.TxSpecular)
	tx_emissive := props.GetImage(vmodel.TxEmissive)
	tx_metalRoughness := props.GetImage(vmodel.TxMetallicRoughness)
	tx_normal := props.GetImage(vmodel.TxBump)
	ub := debugMaterial{
		diffuseFactor: props.GetColor(vmodel.CAlbedo, getFactor(tx_diffuse)),
	}
	if tx_normal > 0 {
		ub.normalMap = 1
	}

	b := *(*[unsafe.Sizeof(debugMaterial{})]byte)(unsafe.Pointer(&ub))
	return &DebugMaterial{}, getDebugLayout(ctx, dev), b[:], []vmodel.ImageIndex{tx_diffuse, tx_normal, 1, tx_specular, tx_emissive,
		tx_metalRoughness, 0, 0}
}

func getFactor(imIndex vmodel.ImageIndex) mgl32.Vec4 {
	if imIndex > 0 {
		return mgl32.Vec4{1, 1, 1, 1}
	}
	return mgl32.Vec4{0, 0, 0, 1}
}

var DebugModes mgl32.Vec4

type DebugMaterial struct {
	dsMat *vk.DescriptorSet
}

func (u *DebugMaterial) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(*forward.Frame)
	if !ok {
		return // Not supported
	}
	rc := ff.GetCache()
	dsFrame := ff.BindDynamicFrame()
	if dsFrame == nil {
		return
	}
	gp := dc.Pass.Get(rc.Ctx, kDebugPipelineSkinned, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)

	uli := rc.GetPerFrame(kDebugInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &debugInstance{ds: ds, sl: sl}
	}).(*debugInstance)
	dm := debugMatInstance{world: world, modes: DebugModes}
	lInst := uint32(unsafe.Sizeof(debugMatInstance{}))
	b := *(*[unsafe.Sizeof(debugMatInstance{})]byte)(unsafe.Pointer(&dm))
	copy(uli.sl.Content[uli.count*lInst:(uli.count+1)*lInst], b[:])
	dsMesh, slMesh := uc.Alloc(rc.Ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstaces {
		rc.SetPerFrame(kDebugInstances, nil)
	}
}

func (u *DebugMaterial) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

func (u *DebugMaterial) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(*forward.Frame)
	if !ok {
		return // Not supported
	}
	rc := ff.GetCache()
	dsFrame := ff.BindDynamicFrame()
	if dsFrame == nil {
		return
	}
	gp := dc.Pass.Get(rc.Ctx, kDebugPipeline, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)

	uli := rc.GetPerFrame(kDebugInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &debugInstance{ds: ds, sl: sl}
	}).(*debugInstance)
	dsMesh, slMesh := uc.Alloc(rc.Ctx)
	decals := decal.GetDecals(extra)
	dm := debugMatInstance{world: world, modes: DebugModes}
	if len(decals) > 0 {
		dm.decalIndex = mgl32.Vec2{0, float32(len(decals) / 2)}
		copy(slMesh.Content, vscene.Mat4ToBytes(decals))
	}

	lInst := uint32(unsafe.Sizeof(debugMatInstance{}))
	b := *(*[unsafe.Sizeof(debugMatInstance{})]byte)(unsafe.Pointer(&dm))
	copy(uli.sl.Content[uli.count*lInst:(uli.count+1)*lInst], b[:])

	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstaces {
		rc.SetPerFrame(kDebugInstances, nil)
	}

}

type debugMatInstance struct {
	world      mgl32.Mat4
	modes      mgl32.Vec4
	decalIndex mgl32.Vec2
	dummy      mgl32.Vec2
}

func (u *DebugMaterial) NewPipeline(ctx vk.APIContext, dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	if skinned {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, debugmat_vert_skinned_spv)

	} else {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, debugmat_vert_spv)
	}
	laFrame := forward.GetDynamicFrameLayout(ctx, rc.Device)
	laUBF := vscene.GetUniformLayout(ctx, rc.Device)

	la2 := getDebugLayout(ctx, rc.Device)
	gp.AddLayout(ctx, laFrame)
	gp.AddLayout(ctx, laUBF)
	gp.AddLayout(ctx, la2)
	gp.AddLayout(ctx, laUBF) // Transform matrix
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, debugmat_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, dc.Pass)
	return gp
}

type debugMaterial struct {
	diffuseFactor mgl32.Vec4
	normalMap     float32
}

type debugInstance struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kDebugLayout = vk.NewKey()
var kDebugPipeline = vk.NewKey()
var kDebugPipelineSkinned = vk.NewKey()
var kDebugInstances = vk.NewKey()

func getDebugLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(ctx, dev)
	return dev.Get(ctx, kDebugLayout, func(ctx vk.APIContext) interface{} {
		return la.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 8)
	}).(*vk.DescriptorLayout)
}
