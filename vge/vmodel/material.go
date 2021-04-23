package vmodel

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

type Property uint32

const (
	Color               = Property(0x01000000)
	CAlbedo             = Color + 1
	CEmissive           = Color + 2
	CSpecular           = Color + 3
	Texture             = Property(0x02000000)
	TxAlbedo            = Texture + 1
	TxEmissive          = Texture + 2
	TxSpecular          = Texture + 3
	TxBump              = Texture + 4
	TxMetallicRoughness = Texture + 5
	Factor              = Property(0x03000000)
	FSpeculaPower       = Factor + 1
	FMetalness          = Factor + 2
	FRoughness          = Factor + 3
	// For decals to specify if normal difference from surface to decal will attenuate decal effect.
	// 0 - No attanuation, 1 - Full attenuation (factor = dot(normal, decalNormal))
	FNormalAttenuation = Factor + 4
	Special            = Property(0xFF000000)
	SMaxIndex          = Special + 1
	PropertyKind       = Property(0xFF000000)
)

type MaterialProperties map[Property]interface{}

func NewMaterialProperties() MaterialProperties {
	return make(MaterialProperties)
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

func (mp MaterialProperties) GetImage(prop Property) ImageIndex {
	v, ok := mp[prop]
	if ok {
		return v.(ImageIndex)
	}
	return 0
}

type Renderer interface {
	// GetPerRenderer allows phases share item for whole renderer
	GetPerRenderer(key vk.Key, ctor func(ctx vk.APIContext) interface{}) interface{}
}

type Frame interface {
	// Retrieve renderer associated to this frame
	GetRenderer() Renderer

	// Retrieve render cache assosiate to this render instance
	GetCache() *vk.RenderCache
}

type DrawContext struct {
	Frame Frame
	Pass  vk.RenderPass
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

type ShaderExtra interface {
	Get(key vk.Key) interface{}
}

type Shader interface {
	SetDescriptor(dsMat *vk.DescriptorSet)
	Draw(ctx *DrawContext, mesh Mesh, world mgl32.Mat4, extra ShaderExtra)
	DrawSkinned(ctx *DrawContext, mesh Mesh, world mgl32.Mat4, aniMatrix []mgl32.Mat4, extra ShaderExtra)
}

type ShaderFactory func(ctx vk.APIContext, dev *vk.Device, propSet MaterialProperties) (
	sh Shader, layout *vk.DescriptorLayout, ubf []byte, images []ImageIndex)
