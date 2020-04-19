package vmodel

import (
	"github.com/go-gl/mathgl/mgl32"
)

type ChannelTarget uint32

const (
	TTranslation = ChannelTarget(1)
	TScale       = ChannelTarget(2)
	TRotation    = ChannelTarget(3)
)

type Joint struct {
	Translate     mgl32.Vec3
	Scale         mgl32.Vec3
	Rotate        mgl32.Quat
	InverseMatrix mgl32.Mat4
	Root          bool
	Children      []int
	Name          string
}

type Channel struct {
	Joint  int
	Input  []float32
	Output []float32
	Target ChannelTarget
}

type Animation struct {
	Name     string
	Channels []Channel
}

type Skin struct {
	Joints     []Joint
	Matrix     []mgl32.Mat4
	Animations []Animation
}
