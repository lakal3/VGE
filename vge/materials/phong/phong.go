//

package phong

import (
	"github.com/lakal3/vge/vge/forward"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func PhongFactory(dev *vk.Device, props vmodel.MaterialProperties) (
	sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
	tx_diffuse := props.GetImage(vmodel.TxAlbedo)
	tx_specular := props.GetImage(vmodel.TxSpecular)
	tx_emissive := props.GetImage(vmodel.TxEmissive)
	tx_normal := props.GetImage(vmodel.TxBump)
	ub := phongMaterial{
		diffuseFactor:  props.GetColor(vmodel.CAlbedo, getFactor(tx_diffuse)),
		emissiveFactor: props.GetColor(vmodel.CEmissive, getFactor(tx_specular)),
		specularFactor: props.GetColor(vmodel.CSpecular, getFactor(tx_emissive)),
		specularPower:  props.GetFactor(vmodel.FSpeculaPower, 50),
	}
	if tx_normal > 0 {
		ub.normalMap = 1
	}

	b := *(*[unsafe.Sizeof(phongMaterial{})]byte)(unsafe.Pointer(&ub))
	return &PhongMaterial{}, getPhongLayout(dev), b[:], []vmodel.ImageIndex{tx_diffuse, tx_normal, tx_specular, tx_emissive}
}

func getFactor(imIndex vmodel.ImageIndex) mgl32.Vec4 {
	if imIndex > 0 {
		return mgl32.Vec4{1, 1, 1, 1}
	}
	return mgl32.Vec4{0, 0, 0, 1}
}

type PhongMaterial struct {
	dsMat *vk.DescriptorSet
}

func (u *PhongMaterial) DrawSkinned(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4,
	aniMatrix []mgl32.Mat4, extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(forward.ForwardFrame)
	if !ok {
		// Unsupported
		return
	}
	rc := ff.GetCache()
	gp := dc.Pass.Get(kPhongPipeline, func() interface{} {
		return u.NewPipeline(dc)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := ff.BindForwardFrame()
	uli := rc.GetPerFrame(kPhongInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &phongInstance{ds: ds, sl: sl}
	}).(*phongInstance)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dsMesh, slMesh := uc.Alloc()
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat, dsMesh).SetInstances(uli.count, 1)

	uli.count++
	if uli.count >= 256 {
		rc.SetPerFrame(kPhongInstances, nil)
	}
}

func (u *PhongMaterial) SetDescriptor(dsMat *vk.DescriptorSet) {
	u.dsMat = dsMat
}

func (u *PhongMaterial) Draw(dc *vmodel.DrawContext, mesh vmodel.Mesh, world mgl32.Mat4, extra vmodel.ShaderExtra) {
	ff, ok := dc.Frame.(forward.ForwardFrame)
	if !ok {
		// Unsupported
		return
	}
	rc := ff.GetCache()
	gp := dc.Pass.Get(kPhongPipeline, func() interface{} {
		return u.NewPipeline(dc)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := ff.BindForwardFrame()
	uli := rc.GetPerFrame(kPhongInstances, func() interface{} {
		ds, sl := uc.Alloc()
		return &phongInstance{ds: ds, sl: sl}
	}).(*phongInstance)
	copy(uli.sl.Content[uli.count*64:uli.count*64+64], vk.Float32ToBytes(world[:]))
	dc.DrawIndexed(gp, mesh.From, mesh.Count).AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).
		AddDescriptors(dsFrame, uli.ds, u.dsMat).SetInstances(uli.count, 1)

	uli.count++
	if uli.count >= 256 {
		rc.SetPerFrame(kPhongInstances, nil)
	}
}

func (u *PhongMaterial) NewPipeline(dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(rc.Device)
	vmodel.AddInput(gp, vmodel.MESHKindNormal)
	la := vscene.GetUniformLayout(rc.Device)
	laFrame := forward.GetFrameLayout(rc.Device)
	la2 := getPhongLayout(rc.Device)
	gp.AddLayout(laFrame)
	gp.AddLayout(la)
	gp.AddLayout(la2)
	gp.AddShader(vk.SHADERStageVertexBit, phong_vert_spv)
	gp.AddShader(vk.SHADERStageFragmentBit, phong_frag_spv)
	gp.AddDepth(true, true)
	gp.Create(dc.Pass)
	return gp
}

type phongMaterial struct {
	diffuseFactor  mgl32.Vec4
	specularFactor mgl32.Vec4
	emissiveFactor mgl32.Vec4
	normalMap      float32
	specularPower  float32
}

type phongInstance struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

var kPhongLayout = vk.NewKey()
var kPhongPipeline = vk.NewKey()
var kPhongInstances = vk.NewKey()

func getPhongLayout(dev *vk.Device) *vk.DescriptorLayout {
	la := vscene.GetUniformLayout(dev)
	return dev.Get(kPhongLayout, func() interface{} {
		return la.AddBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 4)
	}).(*vk.DescriptorLayout)
}
