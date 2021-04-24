package vscene

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

type SpotLight struct {
	Intensity mgl32.Vec3

	Direction mgl32.Vec3

	Attenuation mgl32.Vec3

	// Outer angle in degrees
	OuterAngle float32

	// Optional inner angle in degreen. Same as outerangle if not set
	InnerAngle float32

	MaxDistance float32
}

func (p *SpotLight) Process(pi *ProcessInfo) {
	bf, ok := pi.Phase.(LightPhase)
	if ok {
		if p.MaxDistance == 0 {
			p.MaxDistance = 10
		}
		bf.AddLight(p.AsStdLight(pi.World), p)
	}
}

func (p *SpotLight) AsStdLight(world mgl32.Mat4) Light {
	pos := world.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
	dir := world.Mul4x1(p.Direction.Vec4(0)).Vec3().Normalize()
	return Light{Intensity: p.Intensity.Vec4(1),
		Direction: dir.Vec4(0), OuterAngle: p.getOuterAngle(), InnerAngle: p.getInnerAngle(),
		Position: pos.Vec3().Vec4(2), Attenuation: p.Attenuation.Vec4(p.MaxDistance)}
}

func (p *SpotLight) getInnerAngle() float32 {
	if p.InnerAngle != 0 {
		return float32(math.Cos(float64(p.InnerAngle * math.Pi / 180)))
	}
	return p.getOuterAngle()
}

func (p *SpotLight) getOuterAngle() float32 {
	return float32(math.Cos(float64(p.OuterAngle * math.Pi / 180)))
}
