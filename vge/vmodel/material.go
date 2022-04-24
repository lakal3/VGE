package vmodel

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"unsafe"
)

type Property uint32

const (
	Color               = Property(0x01000000)
	CAlbedo             = Color + 1
	CEmissive           = Color + 2
	CSpecular           = Color + 3
	CIntensity          = Color + 10
	CCustom1            = Color + 20
	CUser               = Color + 0x10000
	Texture             = Property(0x02000000)
	TxAlbedo            = Texture + 1
	TxEmissive          = Texture + 2
	TxSpecular          = Texture + 3
	TxBump              = Texture + 4
	TxMetallicRoughness = Texture + 5
	TxCustom1           = Texture + 10
	TxCustom2           = Texture + 11
	TxCustom3           = Texture + 12
	TxCustom4           = Texture + 14
	TxUser              = Texture + 0x10000
	Factor              = Property(0x03000000)
	FSpeculaPower       = Factor + 1
	FMetalness          = Factor + 2
	FRoughness          = Factor + 3
	// For decals to specify if normal difference from surface to decal will attenuate decal effect.
	// 0 - No attanuation, 1 - Full attenuation (factor = dot(normal, decalNormal))
	FNormalAttenuation = Factor + 4
	// Max level of alpha to discard pixel.
	FAlphaCutoff = Factor + 5
	// Light attenuation is calculate with formulate attn = FLightAttenuation0 + FLightAttenuation1 * distanceToLight + FLightAttenuation2 * distanceToLight^2
	// For directional lights, default is FLightAttenuation0 = 1, for point lights FLightAttenuation2 = 1
	FLightAttenuation0 = Factor + 10
	FLightAttenuation1 = Factor + 11
	FLightAttenuation2 = Factor + 12
	FInnerAngle        = Factor + 13
	FOuterAngle        = Factor + 14
	// FTransparent if != 0 requires transparent rendering of material
	FUser          = Factor + 0x10000
	Uint           = Property(0x04000000)
	UShadowMapSize = Uint + 1
	UMaterialID    = Uint + 2
	Special        = Property(0xFF000000)
	SMaxIndex      = Special + 1
	PropertyKind   = Property(0xFF000000)
)

type MaterialProperties map[Property]interface{}

func NewMaterialProperties() MaterialProperties {
	return make(MaterialProperties)
}

func (mp MaterialProperties) Clone() MaterialProperties {
	mpNew := make(MaterialProperties, len(mp))
	for k, v := range mp {
		mpNew[k] = v
	}
	return mpNew
}

func (mp MaterialProperties) SetColor(prop Property, val mgl32.Vec4) MaterialProperties {
	mp[prop] = val
	return mp
}

func (mp MaterialProperties) SetFactor(prop Property, val float32) MaterialProperties {
	mp[prop] = val
	return mp
}

func (mp MaterialProperties) SetImage(prop Property, im ImageIndex) MaterialProperties {
	mp[prop] = im
	return mp
}

func (mp MaterialProperties) SetUInt(prop Property, i uint32) MaterialProperties {
	mp[prop] = i
	return mp
}

func (mp MaterialProperties) GetColor(prop Property, defaultValue mgl32.Vec4) mgl32.Vec4 {
	v, ok := mp[prop]
	if ok {
		return v.(mgl32.Vec4)
	}
	return defaultValue
}

func (mp MaterialProperties) GetFactor(prop Property, defaultValue float32) float32 {
	v, ok := mp[prop]
	if ok {
		return v.(float32)
	}
	return defaultValue
}

func (mp MaterialProperties) GetUInt(prop Property, defaultValue uint32) uint32 {
	v, ok := mp[prop]
	if ok {
		return v.(uint32)
	}
	return defaultValue
}

func (mp MaterialProperties) GetImage(prop Property) ImageIndex {
	v, ok := mp[prop]
	if ok {
		return v.(ImageIndex)
	}
	return 0
}

type Renderer interface {
	// GetPerRenderer allows nodes to store information that is shared between all frames of renderer. Typically each frame has it's own render
	// cache where you can store frame relevant assets. Each image of swapchain will get it's own frame.
	GetPerRenderer(key vk.Key, ctor func() interface{}) interface{}
}

type Frame interface {
	// Retrieve renderer associated to this frame
	GetRenderer() Renderer

	// Retrieve render cache assosiate to this render instance
	GetCache() *vk.RenderCache
}

type DrawContext struct {
	Frame Frame
	Pass  *vk.GeneralRenderPass
	List  *vk.DrawList
}

func (dc *DrawContext) DrawIndexed(pl vk.Pipeline, from uint32, count uint32) *vk.DrawItem {
	if dc.List == nil {
		dc.List = new(vk.DrawList)
	}
	return dc.List.DrawIndexed(pl, from, count)
}

func (dc *DrawContext) Draw(pl vk.Pipeline, from uint32, count uint32) *vk.DrawItem {
	if dc.List == nil {
		dc.List = new(vk.DrawList)
	}
	return dc.List.Draw(pl, from, count)
}

func (dc *DrawContext) AllocPushConstants(size uint32) (ptr unsafe.Pointer, offset uint64) {
	if dc.List == nil {
		dc.List = new(vk.DrawList)
	}
	return dc.List.AllocPushConstants(size)
}

type ShaderExtra interface {
	Get(key vk.Key) interface{}
}

type Shader interface {
	SetDescriptor(dsMat *vk.DescriptorSet)
	Draw(ctx *DrawContext, mesh Mesh, world mgl32.Mat4, extra ShaderExtra)
	DrawSkinned(ctx *DrawContext, mesh Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4, extra ShaderExtra)
}

type BoundShader interface {
	Shader
	SetModel(model *Model)
}

type ShaderFactory func(dev *vk.Device, propSet MaterialProperties) (
	sh Shader, layout *vk.DescriptorLayout, ubf []byte, images []ImageIndex)
