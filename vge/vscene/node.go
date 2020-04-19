package vscene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type ProcessInfo struct {
	Phase   Phase
	Time    float64
	Visible bool
	World   mgl32.Mat4
	parent  *ProcessInfo
	extras  map[vk.Key]interface{}
}

// Set extra value to processing info. Value is only valid while processing continue on this node or some of it's child nodes
// You can also override value from previous phases. Override remains while processing any child node of this node
func (pi *ProcessInfo) Set(key vk.Key, extra interface{}) {
	if pi.extras == nil {
		pi.extras = make(map[vk.Key]interface{})
	}
	pi.extras[key] = extra
}

// Get extra info from processing info. Value is nil if key is not available
func (pi *ProcessInfo) Get(key vk.Key) (extra interface{}) {
	if pi.extras != nil {
		v, ok := pi.extras[key]
		if ok {
			return v
		}
	}
	if pi.parent != nil {
		return pi.parent.Get(key)
	}
	return nil
}

type NodeControl interface {
	Process(pi *ProcessInfo)
}

type Node struct {
	Ctrl     NodeControl
	Children []*Node
}

func NewNode(ctrl NodeControl, children ...*Node) *Node {
	return &Node{Ctrl: ctrl, Children: children}
}

func NewNodeAt(at mgl32.Mat4, ctrl NodeControl, children ...*Node) *Node {
	if ctrl != nil {
		return &Node{Ctrl: NewMultiControl(&TransformControl{Transform: at}, ctrl), Children: children}
	}
	return &Node{Ctrl: &TransformControl{Transform: at}, Children: children}
}

type MultiControl struct {
	Controls []NodeControl
}

func NewMultiControl(ctrls ...NodeControl) *MultiControl {
	return &MultiControl{Controls: ctrls}
}

func (m *MultiControl) Process(pi *ProcessInfo) {
	for _, ctrl := range m.Controls {
		ctrl.Process(pi)
		if !pi.Visible {
			return
		}
	}
}

type TransformControl struct {
	Transform mgl32.Mat4
}

func (t *TransformControl) Process(pi *ProcessInfo) {
	pi.World = pi.World.Mul4(t.Transform)
}

type MeshNodeControl struct {
	Mat  vmodel.Shader
	Mesh vmodel.Mesh
}

func (m *MeshNodeControl) Process(pi *ProcessInfo) {
	phase := pi.Phase
	dr, ok := phase.(DrawPhase)
	if ok {
		dc := dr.GetDC(LAYER3D)
		if dc != nil {
			m.Mat.Draw(dc, m.Mesh, pi.World, pi)
		}
	}
	bb, ok := phase.(*BoudingBox)
	if ok {
		aabb := m.Mesh.AABB
		aabb = aabb.Translate(pi.World)
		bb.Add(aabb)
	}
	sd, ok := phase.(ShadowPhase)
	if ok {
		sd.DrawShadow(m.Mesh, pi.World, 0)
	}
}

func NodeFromModel(m *vmodel.Model, node vmodel.NodeIndex, recursive bool) *Node {
	n := &Node{}
	mn := m.GetNode(node)
	if mn.Material >= 0 && mn.Skin > 0 {
		an := &AnimatedNodeControl{Mat: m.GetMaterial(mn.Material).Shader, Mesh: m.GetMesh(mn.Mesh), Skin: m.GetSkin(mn.Skin)}
		n.Ctrl = an
		if len(an.Skin.Animations) > 0 {
			// Initialize first animation if available
			an.Animation = an.Skin.Animations[0]
		}
		if !mn.Transform.ApproxEqual(mgl32.Ident4()) {
			n.Ctrl = NewMultiControl(&TransformControl{Transform: mn.Transform}, n.Ctrl)
		}
	} else if mn.Material >= 0 {
		n.Ctrl = &MeshNodeControl{Mat: m.GetMaterial(mn.Material).Shader, Mesh: m.GetMesh(mn.Mesh)}
		if !mn.Transform.ApproxEqual(mgl32.Ident4()) {
			n.Ctrl = NewMultiControl(&TransformControl{Transform: mn.Transform}, n.Ctrl)
		}
	} else {
		if !mn.Transform.ApproxEqual(mgl32.Ident4()) {
			n.Ctrl = &TransformControl{Transform: mn.Transform}
		}
	}

	if recursive {
		for _, ch := range mn.Children {
			nc := NodeFromModel(m, ch, true)
			n.Children = append(n.Children, nc)
		}
	}
	return n
}
