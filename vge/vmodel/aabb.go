package vmodel

import "github.com/go-gl/mathgl/mgl32"

type AABB struct {
	Min mgl32.Vec3
	Max mgl32.Vec3
}

func (aabb AABB) Len() float32 {
	return aabb.Max.Sub(aabb.Min).Len()
}

func (aabb AABB) Center() mgl32.Vec3 {
	return aabb.Max.Add(aabb.Min).Mul(0.5)
}

func (aabb *AABB) Add(first bool, vertex mgl32.Vec3) {
	if first {
		aabb.Min, aabb.Max = vertex, vertex
	} else {
		for idx := 0; idx < 3; idx++ {
			if aabb.Min[idx] > vertex[idx] {
				aabb.Min[idx] = vertex[idx]
			}
			if aabb.Max[idx] < vertex[idx] {
				aabb.Max[idx] = vertex[idx]
			}
		}
	}
}

func (aabb *AABB) Translate(tr mgl32.Mat4) AABB {
	abNew := AABB{}
	for idx := 0; idx < 8; idx++ {
		v := aabb.Min
		if idx&1 == 1 {
			v[0] = aabb.Max[0]
		}
		if idx&2 == 2 {
			v[1] = aabb.Max[1]
		}
		if idx&4 == 4 {
			v[2] = aabb.Max[2]
		}
		abNew.Add(idx == 0, tr.Mul4x1(v.Vec4(1)).Vec3())
	}
	return abNew
}
