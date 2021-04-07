package vscene

import (
	"image"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

var kView = vk.NewKey()
var kEyePos = vk.NewKey()

// var VulkanProj = mgl32.Mat4{1, 0, 0, 0, 0, -1, 0, 0, 0, 0, 0.5, 0.5, 0, 0, 0, 1}
var VulkanProj = mgl32.Mat4{1.5, 0, 0, 0, 0, -1.5, 0, 0, 0, 0, 0.5, 0.5, 0, 0, 0, 1}

type PerspectiveCamera struct {
	Near float32
	Far  float32
	FoV  float32

	// Position of camera
	Position mgl32.Vec3

	// Target vector
	Target mgl32.Vec3

	// Up vector
	Up mgl32.Vec3
}

func (pc *PerspectiveCamera) CameraProjection(size image.Point) (projection, view mgl32.Mat4) {
	aspect := float32(size.X) / float32(size.Y)
	// f.EyePos = pc.Position.Vec4(1)
	projection = mgl32.Perspective(pc.FoV, aspect, pc.Near, pc.Far)
	// f.Projection = VulkanProj.Mul4(proj)
	projection = projection.Mul4(mgl32.Scale3D(1, -1, 1))
	view = mgl32.LookAtV(pc.Position, pc.Target, pc.Up)
	return projection, view
}

func NewPerspectiveCamera(far float32) *PerspectiveCamera {
	pc := &PerspectiveCamera{Near: far / 10000, Far: far, FoV: 1,
		Position: mgl32.Vec3{0, 0, -1},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	return pc
}

func (pc *PerspectiveCamera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(pc.Position, pc.Target, pc.Up)
}
