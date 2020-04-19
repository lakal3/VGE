package vmodel

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"math"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type SkinIndex int
type MaterialIndex int
type ImageIndex int
type MeshIndex int
type NodeIndex int

// ModelBuilder allows us to load multiple meshes, materials, images and skins. When we have loaded all nessessary
// artifacts, model builder can be converted to a model. Model will upload all model information (except Skins) to
// GPU where Shaders can access them quickly
type ModelBuilder struct {
	MipLevels     uint32
	ShaderFactory ShaderFactory
	white         *ImageBuilder
	materials     []materialInfo
	images        []*ImageBuilder
	meshes        []*MeshBuilder
	root          *NodeBuilder
	nodeCount     uint32
	skins         []Skin
	wg            *sync.WaitGroup
	// joints       []MJoint
	// skins        []MSkin
	// channels     []MChannel
}

type ImageBuilder struct {
	index       ImageIndex
	kind        string
	content     []byte
	usage       vk.ImageUsageFlags
	desc        vk.ImageDescription
	orignalMips uint32
}

type NodeBuilder struct {
	Index    NodeIndex
	name     string
	children []*NodeBuilder
	location mgl32.Mat4
	mesh     MeshIndex
	material MaterialIndex
	skin     SkinIndex
}

// SetMesh assign a mesh with material to node
func (nb *NodeBuilder) SetMesh(m MeshIndex, mat MaterialIndex) *NodeBuilder {
	nb.mesh = m
	nb.material = mat
	return nb
}

// SetMesh assign a mesh with material and skin to node
func (nb *NodeBuilder) SetSkinnedMesh(m MeshIndex, mat MaterialIndex, skin SkinIndex) *NodeBuilder {
	nb.mesh = m
	nb.material = mat
	nb.skin = skin
	return nb
}

// Add childs adds child nodes to a node
func (nb *NodeBuilder) AddChild(child ...*NodeBuilder) *NodeBuilder {
	nb.children = append(nb.children, child...)
	return nb
}

// Add new material to model builders. Model builders ShaderFactory is finally used to convert material properties
// to shader. Most of shader also build a descriptor set that links all static assets like color and textures
// of a material to a single Vulkan descriptor set.
func (mb *ModelBuilder) AddMaterial(name string, props MaterialProperties) MaterialIndex {
	mi := MaterialIndex(len(mb.materials))
	mb.materials = append(mb.materials, materialInfo{props: props, name: name})
	return mi
}

// Attach image to model. This image can be them bound to material using material properties
func (mb *ModelBuilder) AddImage(kind string, content []byte, usage vk.ImageUsageFlags) ImageIndex {
	_ = mb.AddWhite()
	im := &ImageBuilder{kind: kind, content: content, index: ImageIndex(len(mb.images)), usage: usage}
	mb.images = append(mb.images, im)
	return im.index
}

// Add white adds simple white image to model. We can use while image instead of actual image when
// material don't have an image assigned to it. Vulkan requires that we bind something to each allocates image slot
// so we can use this white image when we don't have anything real to bind.
// (Vulkan 1.2 and Vulkan 1.1 with extensions support partially bound descriptor and this kind of binding is no longer necessary)

func (mb *ModelBuilder) AddWhite() ImageIndex {
	if mb.white == nil {
		mb.white = &ImageBuilder{kind: "dds", content: white_bin, index: ImageIndex(len(mb.images)),
			usage: vk.IMAGEUsageSampledBit | vk.IMAGEUsageTransferDstBit}
		mb.images = append(mb.images, mb.white)
	}
	return mb.white.index
}

// Add a mesh to model. Actual mesh content is built with MeshBuilder
func (mb *ModelBuilder) AddMesh(mesh *MeshBuilder) MeshIndex {
	mesh.buildUvs()
	mesh.buildNormals()
	mesh.buildTangets()
	// mesh.buildQTangent()
	hasWeights := mesh.buildWeights()
	var aabb AABB
	for idx, vb := range mesh.vextexies {
		aabb.Add(idx == 0, vb.position)
	}
	mesh.aabb = aabb
	if hasWeights {
		mesh.kind = MESHKindSkinned
	}
	mesh.index = MeshIndex(len(mb.meshes))
	mb.meshes = append(mb.meshes, mesh)
	return mesh.index
}

// Add new node to model. If parent node is empty, predefined root node, named _root will be used.
func (mb *ModelBuilder) AddNode(name string, parent *NodeBuilder, transform mgl32.Mat4) *NodeBuilder {
	if parent == nil {
		if mb.root == nil {
			mb.root = &NodeBuilder{name: "_root", location: mgl32.Ident4(), Index: 0, mesh: -1, material: -1}
			mb.nodeCount++
		}
		parent = mb.root
	}
	n := &NodeBuilder{name: name, location: transform, Index: NodeIndex(mb.nodeCount), mesh: -1, material: -1}
	mb.nodeCount++
	parent.children = append(parent.children, n)
	return n
}

// Convert content of model builder to actual model and uploads model content (except skins) to GPU
func (mb *ModelBuilder) ToModel(ctx vk.APIContext, dev *vk.Device) *Model {
	mb.wg = &sync.WaitGroup{}
	if mb.ShaderFactory == nil {
		ctx.SetError(errors.New("Set ShaderFactory"))
		return nil
	}
	m := &Model{}
	mb.AddWhite()
	m.memPool = vk.NewMemoryPool(dev)
	m.owner.AddChild(m.memPool)
	m.images = make([]*vk.Image, 0, len(mb.images))
	var imMaxLen uint64
	for _, ib := range mb.images {
		vasset.DescribeImage(ctx, ib.kind, &ib.desc, ib.content)
		imLen := ib.desc.ImageSize()
		if imLen > imMaxLen {
			imMaxLen = imLen
		}
		desc := ib.desc
		ib.orignalMips = desc.MipLevels
		if desc.MipLevels < mb.MipLevels && mb.canDoMips(desc) {
			desc.MipLevels = mb.MipLevels
			ib.desc.MipLevels = mb.MipLevels
			ib.usage |= vk.IMAGEUsageStorageBit
		}
		img := m.memPool.ReserveImage(ctx, desc, ib.usage)
		m.images = append(m.images, img)
	}
	var iLen [MESHMax]uint64
	var vLen [MESHMax]uint64

	for _, ms := range mb.meshes {
		var vertexSize uint64
		switch ms.kind {
		case MESHKindNormal:
			vertexSize = uint64(unsafe.Sizeof(normalVertex{}))
		case MESHKindSkinned:
			vertexSize = uint64(unsafe.Sizeof(skinnedVertex{}))
		}
		iLen[ms.kind] += uint64(len(ms.incides)) * 4
		vLen[ms.kind] += vertexSize * uint64(len(ms.vextexies))
	}
	for idx := 0; idx < MESHMax; idx++ {
		if iLen[idx] > 0 {
			m.vertexies[idx].bIndex = m.memPool.ReserveBuffer(ctx, iLen[idx], false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageIndexBufferBit)
			m.vertexies[idx].bVertex = m.memPool.ReserveBuffer(ctx, vLen[idx], false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageVertexBufferBit)
		}
	}
	ubfLen := mb.buildMaterials(ctx, dev, m)
	m.bUbf = m.memPool.ReserveBuffer(ctx, ubfLen, false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageUniformBufferBit)
	m.memPool.Allocate(ctx)
	cp := NewCopier(ctx, dev)
	defer cp.Dispose()

	mb.copyNormalVertex(m, cp)
	mb.copySkinnedVertex(m, cp)
	for _, ib := range mb.images {
		mb.wg.Add(1)
		go mb.copyImage(m, ctx, dev, ib)
	}
	sampler := GetDefaultSampler(ctx, dev)
	mb.wg.Wait()
	mb.copyUbf(m, cp, ubfLen, sampler)
	mb.addNodes(mb.root, m)
	m.skins = mb.skins
	return m
}

func (mb *ModelBuilder) copyNormalVertex(m *Model, cp *Copier) {
	var indices []uint32
	var vertexies []normalVertex
	for _, mesh := range mb.meshes {
		offset := uint32(len(vertexies))
		iOffset := uint32(len(indices))
		if mesh.kind == MESHKindNormal {
			for _, vb := range mesh.vextexies {
				vertexies = append(vertexies, normalVertex{position: vb.position, uv: vb.uv, normal: vb.normal, tangent: vb.tangent, color: vb.color})
			}
			for _, idx := range mesh.incides {
				indices = append(indices, idx+offset)
			}
			m.meshes = append(m.meshes, Mesh{Kind: MESHKindNormal, AABB: mesh.aabb,
				Model: m, From: iOffset, Count: uint32(len(mesh.incides))})
		}
	}
	if len(indices) > 0 {
		cp.CopyToBuffer(m.vertexies[MESHKindNormal].bIndex, vk.UInt32ToBytes(indices))
		cp.CopyToBuffer(m.vertexies[MESHKindNormal].bVertex, normalVertexToBytes(vertexies))
	}
}

func (mb *ModelBuilder) copySkinnedVertex(m *Model, cp *Copier) {
	var indices []uint32
	var vertexies []skinnedVertex
	for _, mesh := range mb.meshes {
		offset := uint32(len(vertexies))
		iOffset := uint32(len(indices))
		if mesh.kind == MESHKindSkinned {
			for _, vb := range mesh.vextexies {
				vertexies = append(vertexies, skinnedVertex{position: vb.position, uv: vb.uv, normal: vb.normal, tangent: vb.tangent, color: vb.color,
					weights: vb.weights, joints: vb.joints})
			}
			for _, idx := range mesh.incides {
				indices = append(indices, idx+offset)
			}
			m.meshes = append(m.meshes, Mesh{Kind: MESHKindSkinned, AABB: mesh.aabb,
				Model: m, From: iOffset, Count: uint32(len(mesh.incides))})
		}
	}
	if len(indices) > 0 {
		cp.CopyToBuffer(m.vertexies[MESHKindSkinned].bIndex, vk.UInt32ToBytes(indices))
		cp.CopyToBuffer(m.vertexies[MESHKindSkinned].bVertex, skinnedVertexToBytes(vertexies))
	}
}

func (mb *ModelBuilder) copyImage(m *Model, ctx vk.APIContext, dev *vk.Device, ib *ImageBuilder) {
	defer func() {
		mb.wg.Done()
	}()
	cp := NewCopier(ctx, dev)
	defer cp.Dispose()
	if ib.desc.MipLevels > ib.orignalMips {
		r := ib.desc.FullRange()
		cp.SetLayout(m.images[ib.index], r, vk.IMAGELayoutGeneral)
		r = vk.ImageRange{LayerCount: ib.desc.Layers, LevelCount: 1}
		cp.CopyToImage(m.images[ib.index], ib.kind, ib.content, r, vk.IMAGELayoutGeneral)

		comp := NewCompute(ctx, dev)
		defer comp.Dispose()
		for mip := ib.orignalMips; mip < mb.MipLevels; mip++ {
			for l := uint32(0); l < ib.desc.Layers; l++ {
				comp.MipImage(m.images[ib.index], l, mip)
			}
		}
		r = ib.desc.FullRange()
		r.Layout = vk.IMAGELayoutGeneral
		cp.SetLayout(m.images[ib.index], r, vk.IMAGELayoutShaderReadOnlyOptimal)
	} else {
		cp.CopyToImage(m.images[ib.index], ib.kind, ib.content, ib.desc.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	}
}

func (mb *ModelBuilder) buildMaterials(ctx vk.APIContext, dev *vk.Device, m *Model) uint64 {
	mCounts := make(map[*vk.DescriptorLayout]int)
	mPools := make(map[*vk.DescriptorLayout]*vk.DescriptorPool)
	for idx, mt := range mb.materials {
		mt.mat, mt.layout, mt.ubf, mt.images = mb.ShaderFactory(ctx, dev, mt.props)
		mCounts[mt.layout] = mCounts[mt.layout] + 1
		mb.materials[idx] = mt

	}

	for idx, mCount := range mCounts {
		pool := vk.NewDescriptorPool(ctx, idx, mCount)
		m.owner.AddChild(pool)
		mPools[idx] = pool
	}

	offset := uint64(0)
	for idx, mi := range mb.materials {
		mi.ds = mPools[mi.layout].Alloc(ctx)
		mi.offset = offset
		mb.materials[idx] = mi
		rem := uint64(len(mi.ubf) % vk.MinUniformBufferOffsetAlignment)
		if rem > 0 {
			offset += uint64(len(mi.ubf)) + vk.MinUniformBufferOffsetAlignment - rem
		} else {
			offset += uint64(len(mi.ubf))
		}
		m.materials = append(m.materials, Material{Shader: mi.mat, Props: mi.props, Name: mi.name})
	}
	return offset
}

func (mb *ModelBuilder) copyUbf(m *Model, cp *Copier, ubfLen uint64, sampler *vk.Sampler) {
	ubfs := make([]byte, ubfLen)
	for mIndex, mi := range mb.materials {
		copy(ubfs[mi.offset:], mi.ubf)
		mi.ds.WriteSlice(cp.ctx, 0, 0, m.bUbf.Slice(cp.ctx, mi.offset, mi.offset+uint64(len(mi.ubf))))
		for idx, ib := range mi.images {
			mi.ds.WriteImage(cp.ctx, 1, uint32(idx), m.images[ib].DefaultView(cp.ctx), sampler)
			m.materials[mIndex].Shader.SetDescriptor(mi.ds)
		}
	}
	cp.CopyToBuffer(m.bUbf, ubfs)
}

func (mb *ModelBuilder) addNodes(n *NodeBuilder, m *Model) NodeIndex {
	node := Node{Model: m, Name: n.name, Transform: n.location}
	result := NodeIndex(len(m.nodes))
	node.Mesh = n.mesh
	node.Material = n.material
	node.Skin = n.skin
	m.nodes = append(m.nodes, node)
	if len(n.children) > 0 {
		for _, ch := range n.children {
			node.Children = append(node.Children, mb.addNodes(ch, m))
		}
		m.nodes[result] = node
	}
	return result
}

// Add skin to model
func (mb *ModelBuilder) AddSkin(skin Skin) SkinIndex {
	mb.skins = append(mb.skins, skin)
	idx := SkinIndex(len(mb.skins))
	return idx
}

func (mb *ModelBuilder) canDoMips(desc vk.ImageDescription) bool {
	w, h := desc.Width, desc.Height
	return (w>>mb.MipLevels) > 1 && (h>>mb.MipLevels) > 1
}

type materialInfo struct {
	mat    Shader
	offset uint64
	ubf    []byte
	images []ImageIndex
	ds     *vk.DescriptorSet
	layout *vk.DescriptorLayout
	props  MaterialProperties
	name   string
}

// MeshBuilder is used to construct one mesh. Mesh is then added to model builder
type MeshBuilder struct {
	index     MeshIndex
	vextexies []*VertexBuilder
	incides   []uint32
	aabb      AABB
	kind      MeshKind
}

type VertexFlags int

const (
	VFUV      = VertexFlags(1)
	VFNormal  = VertexFlags(2)
	VFTangent = VertexFlags(4)
	VFWeights = VertexFlags(16)
)

// Vertex builder builds one vertex of a mesh
type VertexBuilder struct {
	Index    uint32
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
	tangent  mgl32.Vec3
	color    mgl32.Vec4
	weights  mgl32.Vec4
	// qtangent mgl32.Quat
	joints [4]uint16
	flags  VertexFlags
}

// Add new vertex to mesh. You must provide at least position of mesh
func (mb *MeshBuilder) AddVertex(position mgl32.Vec3) *VertexBuilder {
	vb := &VertexBuilder{position: position, Index: uint32(len(mb.vextexies))}
	mb.vextexies = append(mb.vextexies, vb)
	return vb
}

// Add indexes to vertexes. Mesh can be used without index information, in witch case each set of three vertexes will make
// one triagle. VGE don't support any other modes to combine vertexes. Use indexes if you wan't to reuse same vertex multiple
// times
func (mb *MeshBuilder) AddIndex(points ...uint32) (index int) {
	idx := len(mb.incides) / len(points)
	mb.incides = append(mb.incides, points...)
	return idx
}

// AddCube will create simple unit cube (1x1x1) transformed with given transformation
// This method is used to make simple test models.
func (mb *MeshBuilder) AddCube(tr mgl32.Mat4) {
	mb.AddPlane(tr)
	mb.AddPlane(tr.Mul4(mgl32.HomogRotate3DX(math.Pi / 2)))
	mb.AddPlane(tr.Mul4(mgl32.HomogRotate3DX(math.Pi / -2)))
	mb.AddPlane(tr.Mul4(mgl32.HomogRotate3DY(math.Pi / 2)))
	mb.AddPlane(tr.Mul4(mgl32.HomogRotate3DY(math.Pi)))
	mb.AddPlane(tr.Mul4(mgl32.HomogRotate3DY(-math.Pi / 2)))
	return
}

// AddPlane adds one unit plane to to mesh
func (mb *MeshBuilder) AddPlane(tr mgl32.Mat4) {
	vb1 := mb.addPos(tr, mgl32.Vec3{-1, -1, 1}, mgl32.Vec3{0, 0, 1}, mgl32.Vec2{0, 0})
	vb2 := mb.addPos(tr, mgl32.Vec3{1, -1, 1}, mgl32.Vec3{0, 0, 1}, mgl32.Vec2{1, 0})
	vb3 := mb.addPos(tr, mgl32.Vec3{1, 1, 1}, mgl32.Vec3{0, 0, 1}, mgl32.Vec2{1, 1})
	vb4 := mb.addPos(tr, mgl32.Vec3{-1, 1, 1}, mgl32.Vec3{0, 0, 1}, mgl32.Vec2{0, 1})
	mb.AddIndex(vb1.Index, vb2.Index, vb3.Index, vb1.Index, vb3.Index, vb4.Index)
}

// AddIcosahedron adds simple unit icosahedron to mesh
func (mb *MeshBuilder) AddIcosahedron(tr mgl32.Mat4) {
	// Top
	top := mb.addPos(tr, mgl32.Vec3{0, 1, 0}, mgl32.Vec3{0, 1, 0}, mgl32.Vec2{0.5, 1})
	bottom := mb.addPos(tr, mgl32.Vec3{0, -1, 0}, mgl32.Vec3{0, -1, 0}, mgl32.Vec2{0.5, -1})
	// Layers
	y := float32(math.Atan(0.5))
	var layer1 []uint32
	var layer2 []uint32
	for idx := 0; idx <= 5; idx++ {
		x := float32(math.Sin(math.Pi * 2 / 5 * float64(idx)))
		z := float32(math.Cos(math.Pi * 2 / 5 * float64(idx)))
		ts := mb.addPos(tr, mgl32.Vec3{x, y, z}, mgl32.Vec3{x, y, z}, mgl32.Vec2{float32(idx) / 5, y})
		layer1 = append(layer1, ts.Index)
		ts = mb.addPos(tr, mgl32.Vec3{x, -y, z}, mgl32.Vec3{x, -y, z}, mgl32.Vec2{float32(idx) / 5, -y})
		layer2 = append(layer2, ts.Index)
		if idx > 0 {
			mb.AddIndex(top.Index, layer1[idx-1], layer1[idx])
			mb.AddIndex(bottom.Index, layer2[idx], layer2[idx-1])
			mb.AddIndex(layer1[idx-1], layer2[idx-1], layer2[idx])
			mb.AddIndex(layer1[idx-1], layer2[idx], layer1[idx])
		}
	}
}

func (mb *MeshBuilder) addPos(tr mgl32.Mat4, v mgl32.Vec3, n mgl32.Vec3, uv mgl32.Vec2) *VertexBuilder {
	v2 := tr.Mul4x1(v.Vec4(1)).Vec3()
	vn2 := tr.Mul4x1(n.Vec4(0)).Vec3()
	vb := mb.AddVertex(v2).AddNormal(vn2).AddUV(uv)
	return vb
}

func (mb *MeshBuilder) buildUvs() {
	for _, vb := range mb.vextexies {
		if vb.flags&VFUV != 0 {
			continue
		}
		pos := vb.position
		if pos.Len() < 0.001 {
			vb.AddUV(mgl32.Vec2{0, 0})
		} else {
			angle := math.Atan2(float64(pos.Z()), float64(pos.X()))
			pitch := math.Asin(float64(pos.Y() / pos.Len()))
			vb.AddUV(mgl32.Vec2{float32(angle/2/math.Pi*math.Cos(pitch) + 0.5), float32(pitch/math.Pi + 0.5)})
		}
	}
}

func (mb *MeshBuilder) buildNormals() error {
	for _, vb := range mb.vextexies {
		if vb.flags&VFNormal != 0 {
			continue
		}
		pos := vb.position

		if pos.Len() < 0.001 {
			vb.AddNormal(mgl32.Vec3{0, 1, 0})
		} else {
			np := pos.Normalize()
			vb.AddNormal(np)
		}
	}
	return nil
}

func (mb *MeshBuilder) buildTangets() {
	if len(mb.vextexies) == 0 || mb.vextexies[0].flags&VFTangent != 0 {
		return
	}
	incides := mb.incides
	bitangents := make([]mgl32.Vec3, len(mb.vextexies))
	for x := 0; x < len(incides); x += 3 {
		vb1 := mb.vextexies[incides[x]]
		vb2 := mb.vextexies[incides[x+1]]
		vb3 := mb.vextexies[incides[x+2]]

		deltuvs1 := vb2.uv.Sub(vb1.uv)
		deltuvs2 := vb3.uv.Sub(vb1.uv)
		deltpositions1 := vb2.position.Sub(vb1.position)
		deltpositions2 := vb3.position.Sub(vb1.position)
		r := 1.0 / (deltuvs1.X()*deltuvs2.Y() - deltuvs1.Y()*deltuvs2.X())
		var tangent mgl32.Vec3
		var bitangent mgl32.Vec3
		if math.IsInf(float64(r), 0) {
			pos1 := vb1.position
			tangent = mgl32.Vec3{-pos1.Y(), pos1.Z(), pos1.X()}
			bitangent = mgl32.Vec3{pos1.Z(), pos1.X(), -pos1.Y()}
		} else {
			tangent = deltpositions1.Mul(deltuvs2.Y()).Sub(deltpositions2.Mul(deltuvs1.Y())).Mul(r)
			// bitangent = (deltpositions2 * deltuvs1.x   - deltpositions1 * deltuvs2.x)*r;
			bitangent = deltpositions2.Mul(deltuvs1.X()).Sub(deltpositions1.Mul(deltuvs2.X())).Mul(r)
		}
		vb1.tangent = vb1.tangent.Add(tangent)
		vb2.tangent = vb2.tangent.Add(tangent)
		vb3.tangent = vb3.tangent.Add(tangent)
		bitangents[incides[x]] = bitangents[incides[x]].Add(bitangent)
		bitangents[incides[x+1]] = bitangents[incides[x+1]].Add(bitangent)
		bitangents[incides[x+2]] = bitangents[incides[x+2]].Add(bitangent)
	}

	for idx, vb := range mb.vextexies {
		n := vb.normal
		t := vb.tangent
		// t = (t - n * dot(n, t)).normalize();
		t2 := t.Sub(n.Mul(n.Dot(t))).Normalize()
		// if (glm::dot(glm::cross(n, t), b) < 0.0f){
		//     t = t * -1.0f;
		// }
		if n.Cross(t).Dot(bitangents[idx]) > 0 {
			t2 = t2.Mul(-1)
		}
		vb.tangent = t2
		vb.flags |= VFTangent
	}
}

/* Maybe later
func (mb *MeshBuilder) buildQTangent() {
	for _, vb := range mb.vextexies {
		vn := vb.normal.Normalize()
		vt := vb.tangent.Normalize()
		var rot mgl32.Mat3
		// Rotate around Z and X axis so that normal equals Y after rotation
		if abs32(vn.X()) > abs32(vn.Z()) {
			r := math.Atan2(float64(vn.X()), float64(vn.Y()))
			rot = mgl32.Rotate3DZ(float32(r))
			t1 := rot.Mul3x1(vn)
			r = math.Atan2(float64(t1.Z()), float64(t1.Y()))
			rot2 := mgl32.Rotate3DX(float32(-r))
			rot = rot2.Mul3(rot)
		} else {
			r := math.Atan2(float64(vn.Z()), float64(vn.Y()))
			rot = mgl32.Rotate3DX(float32(-r))
			t1 := rot.Mul3x1(vn)
			r = math.Atan2(float64(t1.X()), float64(t1.Y()))
			rot2 := mgl32.Rotate3DZ(float32(r))
			rot = rot2.Mul3(rot)
		}
		// Convert tangent to new rotated space and align tangent (that should be planar to X/Z axis) to X
		tn1 := rot.Mul3x1(vt)
		ry := math.Atan2(float64(tn1.Z()), float64(tn1.X()))
		rot3 := mgl32.Rotate3DY(float32(ry))
		rot = rot3.Mul3(rot)

		// Take inverse of rotation matrix. This should map Y=1 vector to normal , X=1 to tangent and B = 1 to bitangent
		iRot := rot.Inv()
		// Quoternion
		q := mgl32.Mat4ToQuat(iRot.Mat4())
		vb.qtangent = q
	}
}

*/

func (mb *MeshBuilder) buildWeights() (hasWeights bool) {
	for _, vb := range mb.vextexies {
		if vb.flags&VFWeights != 0 {
			hasWeights = true
			break
		}
	}
	if hasWeights {
		for idx, vb := range mb.vextexies {
			w1 := 1 - vb.weights[1] + vb.weights[2] + vb.weights[3]
			mb.vextexies[idx].weights[0] = w1
		}
	}
	return
}

// GetPosition retrieves all mesh positions as float32 array
func (mb *MeshBuilder) GetPositions() (positions []float32) {
	for _, vb := range mb.vextexies {
		positions = append(positions, vb.position[:]...)
	}
	return
}

func (vb *VertexBuilder) AddUV(uv mgl32.Vec2) *VertexBuilder {
	vb.uv = uv
	vb.flags |= VFUV
	return vb
}

func (vb *VertexBuilder) AddNormal(normal mgl32.Vec3) *VertexBuilder {
	vb.normal = normal
	vb.flags |= VFNormal
	return vb
}

func (vb *VertexBuilder) AddTangent(tangent mgl32.Vec3) *VertexBuilder {
	vb.tangent = tangent
	vb.flags |= VFTangent
	return vb
}

// AddColor adds vertex color. This is not used in current shaders but can be like an extra vec4 value is vertex input
func (vb *VertexBuilder) AddColor(color mgl32.Vec4) *VertexBuilder {
	vb.color = color
	return vb
}

// Add weight and joints to vertex. If you add weights and joints to vertex it will convert whole
// mesh to a skinned mesh. You must also provide Skin to model nodes that uses skinned mesh.
// Skinned meshes won't be shown without skin and of cause normal meshes can't handle skins.
func (vb *VertexBuilder) AddWeights(weights mgl32.Vec4, joints ...uint16) *VertexBuilder {
	vb.weights = weights
	copy(vb.joints[:], joints)
	vb.flags |= VFWeights
	return vb
}

func abs32(f float32) float32 {
	if f >= 0 {
		return f
	}
	return -f
}

func normalVertexToBytes(src []normalVertex) []byte {
	dPtr := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d2Ptr := &reflect.SliceHeader{Len: dPtr.Len * int(unsafe.Sizeof(normalVertex{})), Cap: dPtr.Cap * int(unsafe.Sizeof(normalVertex{})), Data: dPtr.Data}
	d2 := (*[]byte)(unsafe.Pointer(d2Ptr))
	runtime.KeepAlive(src)
	return *d2
}

func skinnedVertexToBytes(src []skinnedVertex) []byte {
	dPtr := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d2Ptr := &reflect.SliceHeader{Len: dPtr.Len * int(unsafe.Sizeof(skinnedVertex{})), Cap: dPtr.Cap * int(unsafe.Sizeof(skinnedVertex{})), Data: dPtr.Data}
	d2 := (*[]byte)(unsafe.Pointer(d2Ptr))
	runtime.KeepAlive(src)
	return *d2
}
