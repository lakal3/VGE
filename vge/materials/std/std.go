//

package std

import (
	"errors"
	"github.com/lakal3/vge/vge/deferred"
	"github.com/lakal3/vge/vge/forward"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const maxInstances = 1000 // Must match shader value

// MinEmission is minimum emission to interpret that material is emissive. Otherwise it is non emissive
const MinEmission = 0.01

func Factory(dev *vk.Device, props vmodel.MaterialProperties) (
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
		alphaCut:        props.GetFactor(vmodel.FAlphaCutoff, 0),
	}
	if tx_normal > 0 {
		ub.normalMap = 1
	}
	if ub.emissiveColor.Vec3().Len() < MinEmission {
		ub.emissiveColor = mgl32.Vec4{}
	}
	b := *(*[unsafe.Sizeof(stdMaterial{})]byte)(unsafe.Pointer(&ub))
	return &Material{alphaCut: ub.alphaCut, albedoTexture: tx_diffuse}, getStdLayout(dev), b[:], []vmodel.ImageIndex{tx_diffuse, tx_normal, tx_metalRoughness, tx_emissive}
}

func getColorFactor(imIndex vmodel.ImageIndex) mgl32.Vec4 {
	if imIndex > 0 {
		return mgl32.Vec4{1, 1, 1, 1}
	}
	return mgl32.Vec4{0, 0, 0, 1}
}

type Material struct {
	dsMat         *vk.DescriptorSet
	alphaCut      float32
	albedoTexture vmodel.ImageIndex
	model         *vmodel.Model
}

func (u *Material) GetAlphaTexture() (cutoff float32, view *vk.ImageView, sampler *vk.Sampler) {
	if u.alphaCut > 0 && u.albedoTexture > 0 && u.model != nil {
		view, sampler = u.model.GetImageView(u.albedoTexture)
		return u.alphaCut, view, sampler
	}
	return 0, nil, nil
}

func (u *Material) SetModel(model *vmodel.Model) {
	u.model = model
}

func (u *Material) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

var ErrNoDynamicFrame = errors.New("STD material required dynamic descriptor support and dynamics frames")

func (u *Material) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(forward.ForwardFrame)
	if !ok {
		u.drawSkinnedDeferred(dc, mesh, world, aniMatrix, extra)
		return
	}
	rc := ff.GetCache()
	gp := dc.Pass.Get(kStdSkinnedPipeline, func() interface{} {
		return u.NewPipeline(dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := ff.BindDynamicFrame()
	if dsFrame == nil {
		// dc.Cache.Ctx.SetError(ErrNoDynamicFrame)
		return // Only dynamic frame supported
	}
	uli := rc.GetPerFrame(kStdSkinnedInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	dsDecal := decal.BindPainter(rc, extra)
	dm := stdMatInstance{world: world}
	dsMesh, slMesh := uc.Alloc()
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	uli.writeInstance(dm)
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsDecal, dsMesh).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstances {
		rc.SetPerFrame(kStdSkinnedInstances, nil)
	}
}

func (u *Material) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(forward.ForwardFrame)
	if !ok {
		u.drawDeferred(dc, mesh, world, extra)
		return
	}
	rc := ff.GetCache()
	gp := dc.Pass.Get(kStdPipeline, func() interface{} {
		return u.NewPipeline(dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := ff.BindDynamicFrame()
	if dsFrame == nil {
		// dc.Cache.Ctx.SetError(ErrNoDynamicFrame)
		return // Not supported
	}
	uli := rc.GetPerFrame(kStdInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	dsDecal := decal.BindPainter(rc, extra)
	dm := stdMatInstance{world: world}
	uli.writeInstance(dm)
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsDecal).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstances {
		rc.SetPerFrame(kStdInstances, nil)
	}
}

func (u *Material) drawDeferred(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	frame, ok := dc.Frame.(deferred.DeferredLayout)
	if !ok {
		return
	}
	rc := dc.Frame.GetCache()
	gp := dc.Pass.Get(kDefPipeline, func() interface{} {
		return u.NewDeferredPipeline(dc, false)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := frame.BindDeferredFrame()
	uli := rc.GetPerFrame(kStdInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	uli.writeInstance(stdMatInstance{world: world})
	dsDecal := decal.BindPainter(rc, extra)
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsDecal).SetInstances(uli.count, 1)
	uli.count++
	if uli.count >= maxInstances {
		rc.SetPerFrame(kStdInstances, nil)
	}
}

func (u *Material) drawSkinnedDeferred(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4,
	extra vmodel.ShaderExtra) {
	frame, ok := dc.Frame.(deferred.DeferredLayout)
	if !ok {
		return
	}
	rc := dc.Frame.GetCache()
	gp := dc.Pass.Get(kDefSkinnedPipeline, func() interface{} {
		return u.NewDeferredPipeline(dc, true)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := frame.BindDeferredFrame()
	uli := rc.GetPerFrame(kStdInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &stdInstance{ds: ds, sl: sl}
	}).(*stdInstance)
	uli.writeInstance(stdMatInstance{world: world})
	dsMesh, slMesh := uc.Alloc()
	dsDecal := decal.BindPainter(rc, extra)

	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsDecal, dsMesh).SetInstances(uli.count, 1)
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

func (u *Material) NewPipeline(dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(rc.Device)
	if skinned {
		vmodel.AddInput(gp, vmodel.MESHKindSkinned)
		gp.AddShader(vk.SHADERStageVertexBit, std_vert_skin_spv)

	} else {
		vmodel.AddInput(gp, vmodel.MESHKindNormal)
		gp.AddShader(vk.SHADERStageVertexBit, std_vert_spv)
	}
	laFrame := forward.GetDynamicFrameLayout(rc.Device)
	if laFrame == nil {
		rc.Device.ReportError(ErrNoDynamicFrame)
		return nil
	}
	la := vscene.GetUniformLayout(rc.Device)
	la2 := getStdLayout(rc.Device)
	laUBF := vscene.GetUniformLayout(rc.Device)
	gp.AddLayout(laFrame)
	gp.AddLayout(la)
	gp.AddLayout(la2)
	gp.AddLayout(la) // Decals
	if skinned {
		gp.AddLayout(laUBF) // Transform & decal matrix
	}
	gp.AddShader(vk.SHADERStageFragmentBit, std_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(dc.Pass)
	return gp
}

func (u *Material) NewDeferredPipeline(dc *vmodel.DrawContext, skinned bool) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(rc.Device)
	if skinned {
		vmodel.AddInput(gp, vmodel.MESHKindSkinned)
		gp.AddShader(vk.SHADERStageVertexBit, defmat_vert_skin_spv)

	} else {
		vmodel.AddInput(gp, vmodel.MESHKindNormal)
		gp.AddShader(vk.SHADERStageVertexBit, defmat_vert_spv)
	}
	laFrame := deferred.GetFrameLayout(rc.Device)
	la := vscene.GetUniformLayout(rc.Device)
	la2 := getStdLayout(rc.Device)
	laUBF := vscene.GetUniformLayout(rc.Device)
	gp.AddLayout(laFrame)
	gp.AddLayout(la)
	gp.AddLayout(la2)
	if skinned {
		gp.AddLayout(laUBF) // Transform matrix
	}
	gp.AddLayout(la) // Decals
	gp.AddShader(vk.SHADERStageFragmentBit, defmat_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(dc.Pass)
	return gp
}

type stdMaterial struct {
	albedoColor     mgl32.Vec4
	emissiveColor   mgl32.Vec4
	metalnessFactor float32
	roughnessFactor float32
	normalMap       float32
	alphaCut        float32
}

type stdInstance struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

func (i stdInstance) writeInstance(dm stdMatInstance) {
	lInst := uint32(unsafe.Sizeof(stdMatInstance{}))
	b := *(*[unsafe.Sizeof(stdMatInstance{})]byte)(unsafe.Pointer(&dm))
	copy(i.sl.Content[i.count*lInst:(i.count+1)*lInst], b[:])
}

var kStdLayout = vk.NewKey()
var kStdPipeline = vk.NewKey()
var kStdSkinnedPipeline = vk.NewKey()
var kStdInstances = vk.NewKey()
var kStdSkinnedInstances = vk.NewKey()
var kDefPipeline = vk.NewKey()
var kDefSkinnedPipeline = vk.NewKey()

func getStdLayout(dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(dev)
	return dev.Get(kStdLayout, func() interface{} {
		return la.AddBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 4)
	}).(*vk.DescriptorLayout)
}
