package vmodel

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

type MeshKind int

const (
	MESHKindNormal  = MeshKind(0)
	MESHKindSkinned = MeshKind(1)
	MESHMax         = 2
)

func AddInput(ctx vk.APIContext, gr *vk.GraphicsPipeline, kind MeshKind) {
	switch kind {
	case MESHKindNormal:
		// Position, uv , normal, tangent, color
		gr.AddVextexInput(ctx, vk.VERTEXInputRateVertex, vk.FORMATR32g32b32Sfloat,
			vk.FORMATR32g32Sfloat, vk.FORMATR32g32b32Sfloat, vk.FORMATR32g32b32Sfloat, vk.FORMATR32g32b32a32Sfloat)
	case MESHKindSkinned:
		// Position, uv , normal, tangent, color, weights, joints
		gr.AddVextexInput(ctx, vk.VERTEXInputRateVertex, vk.FORMATR32g32b32Sfloat,
			vk.FORMATR32g32Sfloat, vk.FORMATR32g32b32Sfloat, vk.FORMATR32g32b32Sfloat, vk.FORMATR32g32b32a32Sfloat,
			vk.FORMATR32g32b32a32Sfloat, vk.FORMATR16g16b16a16Uint)
	}
}

type Model struct {
	owner     vk.Owner
	vertexies [MESHMax]vertexInfo
	meshes    []Mesh
	materials []Material
	nodes     []Node
	images    []*vk.Image
	sampler   *vk.Sampler
	views     []*vk.ImageView
	bUbf      *vk.Buffer
	memPool   *vk.MemoryPool
	skins     []Skin

	// joints         []MJoint
	// skins          []MSkin
	// channelStream  []float32
	// channels       []MChannel
}

type vertexInfo struct {
	bVertex *vk.Buffer
	bIndex  *vk.Buffer
}

func (m *Model) Dispose() {
	m.owner.Dispose()
	m.nodes, m.images, m.meshes, m.materials = nil, nil, nil, nil
}

func (m *Model) NodeCount() int {
	return len(m.nodes)
}

func (m *Model) FindNode(nodeName string) NodeIndex {
	return m.getNodeByName(nodeName)
}

func (m *Model) GetNode(idx NodeIndex) Node {
	return m.nodes[idx]
}

func (m *Model) GetMesh(idx MeshIndex) Mesh {
	return m.meshes[idx]
}

func (m *Model) GetMaterial(idx MaterialIndex) Material {
	return m.materials[idx]
}

func (m *Model) GetImage(idx ImageIndex) *vk.Image {
	return m.images[idx]
}

func (m *Model) GetImageView(idx ImageIndex) (view *vk.ImageView, sampler *vk.Sampler) {
	return m.views[idx], m.sampler
}

// FindMaterial finds material index for named material. Return is -1 if material was not found
func (m *Model) FindMaterial(name string) MaterialIndex {
	for idx, m := range m.materials {
		if m.Name == name {
			return MaterialIndex(idx)
		}
	}
	return -1
}

func (m *Model) Bounds(index NodeIndex, transform mgl32.Mat4, includeChild bool) (aabb AABB) {
	node := m.nodes[index]
	transform = transform.Mul4(node.Transform)
	first := true
	if node.Mesh >= 0 {
		mesh := m.meshes[node.Mesh]
		aabb = mesh.AABB.Translate(transform)
		first = false
	}
	if includeChild {
		for _, ch := range node.Children {
			chAabb := m.Bounds(ch, transform, true)
			aabb.Add(first, chAabb.Min)
			aabb.Add(false, chAabb.Max)
			first = false
		}
	}
	return aabb
}

func (m *Model) getNodeByName(s string) NodeIndex {
	for idx, n := range m.nodes {
		if n.Name == s {
			return NodeIndex(idx)
		}
	}
	return -1
}

func (m *Model) VertexBuffers(kind MeshKind) []*vk.Buffer {
	return []*vk.Buffer{m.vertexies[kind].bIndex, m.vertexies[kind].bVertex}
}

func (m *Model) GetSkin(index SkinIndex) *Skin {
	if index == 0 {
		return nil
	}
	return &(m.skins[index-1])
}

func (n Node) Enum(world mgl32.Mat4, action func(local mgl32.Mat4, n Node)) {
	world = world.Mul4(n.Transform)
	action(world, n)
	for _, ch := range n.Children {
		n.Model.GetNode(ch).Enum(world, action)
	}
}

type normalVertex struct {
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
	tangent  mgl32.Vec3
	color    mgl32.Vec4
}

type skinnedVertex struct {
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
	tangent  mgl32.Vec3
	color    mgl32.Vec4
	weights  mgl32.Vec4
	joints   [4]uint16
}

type Mesh struct {
	Kind  MeshKind
	AABB  AABB
	Model *Model
	From  uint32
	Count uint32
}

type Material struct {
	Props  MaterialProperties
	Shader Shader
	Name   string
}

type Node struct {
	Name      string
	Children  []NodeIndex
	Model     *Model
	Transform mgl32.Mat4
	Material  MaterialIndex
	Mesh      MeshIndex
	Skin      SkinIndex
	Parent    NodeIndex
}
