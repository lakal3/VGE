package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

var MaxImages uint32 = 1024

var kFrameLayout = vk.NewKeys(3)
var kShadowFrameLayout = vk.NewKey()
var kRenderPasses = vk.NewKeys(4)
var kJointsLayout = vk.NewKey()

func GetFrameLayout(dev *vk.Device) *vk.DescriptorLayout {
	la1 := dev.Get(kFrameLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
	la2 := dev.Get(kFrameLayout+1, func() interface{} {
		return la1.AddBinding(vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
	la3 := dev.Get(kFrameLayout+2, func() interface{} {
		return la2.AddDynamicBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageAll, MaxImages,
			vk.DESCRIPTORBindingPartiallyBoundBitExt|vk.DESCRIPTORBindingUpdateAfterBindBitExt)
	}).(*vk.DescriptorLayout)
	return la3
}

func GetShadowFrameLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kShadowFrameLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
}

func GetJointsLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kJointsLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
}

func getDrawRenderPass1(dev *vk.Device) *vk.GeneralRenderPass {
	return dev.Get(kRenderPasses, func() interface{} {
		return vk.NewGeneralRenderPass(dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutColorAttachmentOptimal,
				Format: vk.FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.8, 0.8, 0.8, 1}},
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				Format: vk.FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.8, 0.8, 0.8, 1}},
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutShaderReadOnlyOptimal, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				Format: vk.FORMATD32Sfloat},
		})
	}).(*vk.GeneralRenderPass)
}

func getDrawRenderPass2(dev *vk.Device) *vk.GeneralRenderPass {
	return dev.Get(kRenderPasses+1, func() interface{} {
		return vk.NewGeneralRenderPass(dev, false, []vk.AttachmentInfo{
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutColorAttachmentOptimal, FinalLayout: vk.IMAGELayoutColorAttachmentOptimal,
				Format: vk.FORMATR8g8b8a8Unorm},
		})
	}).(*vk.GeneralRenderPass)
}

func getDepthRenderPass(dev *vk.Device) *vk.GeneralRenderPass {
	return dev.Get(kRenderPasses+2, func() interface{} {
		return vk.NewGeneralRenderPass(dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				Format: vk.FORMATD32Sfloat, ClearColor: [4]float32{1}},
		})
	}).(*vk.GeneralRenderPass)
}

func getProbeRenderPass(dev *vk.Device) *vk.GeneralRenderPass {
	return dev.Get(kRenderPasses+3, func() interface{} {
		return vk.NewGeneralRenderPass(dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutGeneral,
				Format: vk.FORMATB10g11r11UfloatPack32, ClearColor: [4]float32{0.2, 0.2, 0.2, 1}},
			vk.AttachmentInfo{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutUndefined,
				Format: vk.FORMATD32Sfloat, ClearColor: [4]float32{1}},
		})
	}).(*vk.GeneralRenderPass)
}

type frame struct {
	projection   mgl32.Mat4
	view         mgl32.Mat4
	cameraPos    mgl32.Vec4
	ambient      mgl32.Vec4
	ambienty     mgl32.Vec4
	viewPosition mgl32.Vec4
	lightPos     uint32
	lights       uint32
	decalPos     uint32
	decals       uint32
	debug        uint32
	temp2        float32
}

type shadowFrame struct {
	plane     mgl32.Vec4
	lightPos  mgl32.Vec4
	minShadow float32
	maxShadow float32
	yFactor   float32
	dummy1    float32
}

type probeFrame struct {
	projection mgl32.Mat4
	views      [6]mgl32.Mat4
	cameraPos  mgl32.Vec4
}

type material struct {
	world         mgl32.Mat4
	albedo        mgl32.Vec4
	emissive      mgl32.Vec4
	metalRoughess mgl32.Vec4
	cColor        mgl32.Vec4 // Custom color or any other fields
	textures1     mgl32.Vec4 // abbedo, emissive, metalRoughness, normal
	ctextures     mgl32.Vec4 // Custom textures
	alphaCutoff   float32
	probe         uint32
	frozenID      uint32
}
