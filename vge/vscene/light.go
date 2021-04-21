package vscene

import "github.com/go-gl/mathgl/mgl32"

type Light struct {
	Intensity   mgl32.Vec4
	Position    mgl32.Vec4 // w = 0 for directional light, 1 = point light, 2 = spot light
	Direction   mgl32.Vec4 // depending on light type
	Attenuation mgl32.Vec4 // 0, 1st and 2nd order, w is shadowmap index
	// InnerAngle for spotlight
	InnerAngle float32
	// OuterAngle for spotlight
	OuterAngle float32
	// Shadow mapping method. 0 - No map, 1 - Point line cube shadow map
	ShadowMapMethod float32
	// Index for shadow map
	ShadowMapIndex float32
}

type DirectionalLight struct {
	Intensity mgl32.Vec3
	Direction mgl32.Vec3
}

func (d *DirectionalLight) Process(pi *ProcessInfo) {
	bf, ok := pi.Phase.(LightPhase)
	if ok {
		bf.AddLight(Light{Intensity: d.Intensity.Vec4(1),
			Direction: d.Direction.Vec4(0), Attenuation: mgl32.Vec4{1, 0, 0, 0}}, d)
	}
}

type PointLight struct {
	// Light intensity (and color)
	Intensity mgl32.Vec3
	// Attenuation of light using formula a[0] + a[1]*d + a[2]*d^2 where d is distance from light.
	// Physically realisting point lights have only a[2]
	Attenuation mgl32.Vec3
	// Maximum distance where light should be visible
	MaxDistance float32
}

func (p *PointLight) Process(pi *ProcessInfo) {
	bf, ok := pi.Phase.(LightPhase)
	if ok {
		if p.MaxDistance == 0 {
			p.MaxDistance = 10
		}
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		bf.AddLight(Light{Intensity: p.Intensity.Vec4(1),
			Position: pos, Attenuation: p.Attenuation.Vec4(p.MaxDistance)}, p)
	}
}
