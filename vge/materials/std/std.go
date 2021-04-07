//

package std

import (
	"errors"
	"github.com/lakal3/vge/vge/forward"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const maxInstances = 200 // Must match shader value

func Factory(ctx vk.APIContext, dev *vk.Device, props vmodel.MaterialProperties) (
	sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
	tx_diffuse := props.GetImage(vmodel.TxAlbedo)
	tx_metalRoughness := props.GetImage(vmodel.TxMetallicRoughness)
	tx_emissive := props.GetImage(vmodel.TxEmissive)
	tx_normal := props.GetImage(vmodel.TxBump)
	mf := float32(0)
	if tx_metalRoughness != 0 {
		mf = 1
	}
	ub := stdMaterial{
		albedoColor:     props.GetColor(vmodel.CAlbedo, getColorFactor(tx_diffuse)),
		emissiveColor:   props.GetColor(vmodel.CEmissive, getColorFactor(tx_emissive)),
		metalnessFactor: props.GetFactor(vmodel.FMetalness, mf),
		roughnessFactor: props.GetFactor(vmodel.FRoughness, 1),
	}
	if tx_normal > 0 {
		ub.normalMap = 1
	}

	b := *(*[unsafe.Sizeof(stdMaterial{})]byte)(unsafe.Pointer(&ub))
	return &Material{}, getStdLayout(ctx, dev), b[:], []vmodel.ImageIndex{tx_diffuse, tx_normal, tx_metalRoughness, tx_emissive}
}

func getColorFactor(imIndex vmodel.ImageIndex) mgl32.Vec4 {
	if imIndex > 0 {
		return mgl32.Vec4{1, 1, 1, 1}
	}
	return mgl32.Vec4{0, 0, 0, 1}
}

type Material struct {
	dsMat *vk.DescriptorSet
}

func (u *Material) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

var ErrNoDynamicFrame = errors.New("STD material required dynamic descriptor support and dynamics frames")

func (u *Material) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	rc := dc.Cache
	gp := dc.Pass.Get(rc.Ctx, kStdSkinnedPipeline, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := forward.BindDynamicFrame(rc)
	if dsFrame == nil {
		dc.Cache.Ctx.SetError(ErrNoDynamicFrame)
		return
	}
	uli := rc.GetPerFrame(kStdSkinnedInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	dm := stdMatInstance{world: world}
	dsMesh, slMesh := uc.Alloc(dc.Cache.Ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	lInst := uint32(unsafe.Sizeof(stdMatInstance{}))
	b := *(*[unsafe.Sizeof(stdMatInstance{})]byte)(unsafe.Pointer(&dm))
	copy(uli.sl.Content[uli.count*lInst:(uli.count+1)*lInst], b[:])
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstances {
		rc.SetPerFrame(kStdSkinnedInstances, nil)
	}
}

func (u *Material) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	rc := dc.Cache
	gp := dc.Pass.Get(rc.Ctx, kStdPipeline, func(ctx vk.APIContext) interface{} {
		return u.NewPipeline(ctx, dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := forward.BindDynamicFrame(rc)
	if dsFrame == nil {
		dc.Cache.Ctx.SetError(ErrNoDynamicFrame)
		return
	}
	uli := rc.GetPerFrame(kStdInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	dsMesh := uli.ds
	decals := decal.GetDecals(extra)
	dm := stdMatInstance{world: world}
	if len(decals) > 0 {
		var slMesh *vk.Slice
		dsMesh, slMesh = uc.Alloc(dc.Cache.Ctx)
		dm.decalIndex = mgl32.Vec2{0, float32(len(decals) / 2)}
		copy(slMesh.Content, vscene.Mat4ToBytes(decals))
	}

	lInst := uint32(unsafe.Sizeof(stdMatInstance{}))
	b := *(*[unsafe.Sizeof(stdMatInstance{})]byte)(unsafe.Pointer(&dm))
	copy(uli.sl.Content[uli.count*lInst:(uli.count+1)*lInst], b[:])
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstances {
		rc.SetPerFrame(kStdInstances, nil)
	}
}

type stdMatInstance struct {
	world      mgl32.Mat4
	decalIndex mgl32.Vec2
	dummy      mgl32.Vec2
}

func (u *Material) NewPipeline(ctx vk.APIContext, dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Cache
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	if skinned {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, std_vert_skin_spv)

	} else {
		vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
		gp.AddShader(ctx, vk.SHADERStageVertexBit, std_vert_spv)
	}
	laFrame := forward.GetDynamicFrameLayout(ctx, rc.Device)
	if laFrame == nil {
		ctx.SetError(ErrNoDynamicFrame)
		return nil
	}
	la := vscene.GetUniformLayout(ctx, rc.Device)
	la2 := getStdLayout(ctx, rc.Device)
	laUBF := vscene.GetUniformLayout(ctx, rc.Device)
	gp.AddLayout(ctx, laFrame)
	gp.AddLayout(ctx, la)
	gp.AddLayout(ctx, la2)
	gp.AddLayout(ctx, laUBF) // Transform & decal matrix
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, std_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, dc.Pass)
	return gp
}

type stdMaterial struct {
	albedoColor     mgl32.Vec4
	emissiveColor   mgl32.Vec4
	metalnessFactor float32
	roughnessFactor float32
	normalMap       float32
}

type stdInstance struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kStdLayout = vk.NewKey()
var kStdPipeline = vk.NewKey()
var kStdSkinnedPipeline = vk.NewKey()
var kStdInstances = vk.NewKey()
var kStdSkinnedInstances = vk.NewKey()

func getStdLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(ctx, dev)
	return dev.Get(ctx, kStdLayout, func(ctx vk.APIContext) interface{} {
		return la.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 4)
	}).(*vk.DescriptorLayout)
}
