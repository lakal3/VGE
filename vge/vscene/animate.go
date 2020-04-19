package vscene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"sort"
)

type RotateAnimate struct {
	Axis  mgl32.Vec3
	Speed float32
}

func (r *RotateAnimate) Process(pi *ProcessInfo) {
	rot := mgl32.HomogRotate3D(r.Speed*float32(pi.Time), r.Axis)
	pi.World = pi.World.Mul4(rot)
}

type AnimatedNodeControl struct {
	Mat       vmodel.Shader
	Mesh      vmodel.Mesh
	StartTime float64
	Skin      *vmodel.Skin
	Animation vmodel.Animation
	Joints    []vmodel.Joint

	calcTime float64
	mxAnims  []mgl32.Mat4
}

func (a *AnimatedNodeControl) Process(pi *ProcessInfo) {
	phase := pi.Phase
	_, ok := phase.(*AnimatePhase)
	if ok {
		time := pi.Time
		if a.StartTime == 0 {
			a.StartTime = pi.Time
		}
		if time-a.calcTime > 0.01 || len(a.mxAnims) == 0 {
			a.calcTime = time
			a.recalc(time - a.StartTime)
		}
	}
	dr, ok := phase.(DrawPhase)
	if ok {
		dc := dr.GetDC(LAYER3D)
		if dc != nil {
			a.Mat.DrawSkinned(dc, a.Mesh, pi.World, a.mxAnims, pi)
		}
	}
	bb, ok := phase.(*BoudingBox)
	if ok {
		aabb := a.Mesh.AABB
		aabb = aabb.Translate(pi.World)
		bb.Add(aabb)
	}

	sd, ok := phase.(ShadowPhase)
	if ok {
		sd.DrawSkinnedShadow(a.Mesh, pi.World, 0, a.mxAnims)
	}
}

// SetAnimationIndex pick new animation from Skin
func (a *AnimatedNodeControl) SetAnimationIndex(fromTime float64, animIndex int) {
	if len(a.Skin.Animations) > animIndex {
		a.Animation = a.Skin.Animations[animIndex]
		a.mxAnims = nil
		a.StartTime = fromTime
	}
}

func (a *AnimatedNodeControl) recalc(animTime float64) {
	if len(a.mxAnims) != len(a.Skin.Joints) {
		a.mxAnims = make([]mgl32.Mat4, len(a.Skin.Joints))
		a.Joints = make([]vmodel.Joint, len(a.Skin.Joints))
		for idx, j := range a.Skin.Joints {
			a.Joints[idx] = j
		}
	}
	for _, ch := range a.Animation.Channels {
		a.applyChannel(ch, animTime)
	}
	for jIdx, j := range a.Joints {
		if j.Root {
			a.recalcJoint(jIdx, j, mgl32.Ident4())
		}
	}
}

func (a *AnimatedNodeControl) recalcJoint(jNro int, joint vmodel.Joint, local mgl32.Mat4) {
	local = local.Mul4(mgl32.Translate3D(joint.Translate[0], joint.Translate[1], joint.Translate[2]))
	local = local.Mul4(joint.Rotate.Mat4())
	local = local.Mul4(mgl32.Scale3D(joint.Scale[0], joint.Scale[1], joint.Scale[2]))
	a.mxAnims[jNro] = local.Mul4(joint.InverseMatrix)
	for _, chJ := range joint.Children {
		a.recalcJoint(chJ, a.Joints[chJ], local)
	}
}

func (a *AnimatedNodeControl) applyChannel(ch vmodel.Channel, animTime float64) {
	chLen := ch.Input[len(ch.Input)-1]
	f := float32(math.Mod(animTime, float64(chLen)))
	idx := sort.Search(len(ch.Input), func(i int) bool {
		return ch.Input[i] >= f
	})
	idx--
	var lin float32
	if idx < 0 {
		idx = 0
		lin = 0
	} else if idx-1 >= len(ch.Input) {
		idx--
		lin = 1
	} else {
		lin = (f - ch.Input[idx]) / (ch.Input[idx+1] - ch.Input[idx])
	}
	switch ch.Target {
	case vmodel.TTranslation:
		a.translate(ch, lin, idx)
	case vmodel.TRotation:
		a.rotate(ch, lin, idx)
	}
}

func (a *AnimatedNodeControl) translate(ch vmodel.Channel, f float32, idx int) {
	min := mgl32.Vec3{ch.Output[idx*3], ch.Output[idx*3+1], ch.Output[idx*3+2]}
	max := mgl32.Vec3{ch.Output[idx*3+3], ch.Output[idx*3+4], ch.Output[idx*3+5]}
	a.Joints[ch.Joint].Translate = min.Mul(1 - f).Add(max.Mul(f))
}

func (a *AnimatedNodeControl) rotate(ch vmodel.Channel, f float32, idx int) {
	qMin := mgl32.Quat{V: mgl32.Vec3{ch.Output[idx*4], ch.Output[idx*4+1], ch.Output[idx*4+2]},
		W: ch.Output[idx*4+3]}
	qMax := mgl32.Quat{V: mgl32.Vec3{ch.Output[idx*4+4], ch.Output[idx*4+5], ch.Output[idx*4+6]},
		W: ch.Output[idx*4+7]}
	a.Joints[ch.Joint].Rotate = mgl32.QuatLerp(qMin, qMax, f)
}
