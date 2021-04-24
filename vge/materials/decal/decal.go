package decal

import (
	"github.com/go-gl/mathgl/mgl32"
)

const MAX_DECALS = 256

type GPUDecals struct {
	NoDecals  float32
	Filler1   float32
	Filler2   float32
	Filler3   float32
	Instances [MAX_DECALS]GPUInstance
}

type GPUInstance struct {
	ToDecalSpace      mgl32.Mat4
	AlbedoFactor      mgl32.Vec4
	MetalnessFactor   float32
	RoughnessFactor   float32
	NormalAttenuation float32
	TxAlbedo          float32
	TxNormal          float32
	TxMetalRoughness  float32
	Filler1           float32
	Filler2           float32
}
