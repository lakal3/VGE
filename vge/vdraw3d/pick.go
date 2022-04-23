package vdraw3d

import "github.com/go-gl/mathgl/mgl32"

type PickInfo struct {
	MeshID uint32
	Vertex uint32
	Depth  float32
	ColorR float32
}

type pickFrame struct {
	count    uint32
	max      uint32
	filler1  uint32
	filler2  uint32
	pickArea mgl32.Vec4
}

// Pick will instruct view to pick any freezables drawn on next render cycle.
// Max value indicates how many hits we record. Depending on geometry complexity and
func (v *View) Pick(max uint32, pickArea mgl32.Vec4, picked func(picks []PickInfo)) {
	v.lock.Lock()
	v.picked, v.pickFrame = picked, &pickFrame{count: 0, max: max, pickArea: pickArea}
	v.lock.Unlock()
}
