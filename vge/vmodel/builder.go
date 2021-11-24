package vmodel

import (
	"errors"
	"math"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
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
	Materials     []MaterialInfo
	Images        []*ImageBuilder
	Meshes        []*MeshBuilder
	Root          *NodeBuilder
	Skins         []Skin
	wg            *sync.WaitGroup
	// joints       []MJoint
	// skins        []MSkin
	// channels     []MChannel
}

type ImageBuilder struct {
	index       ImageIndex
	Kind        string
	Content     []byte
	Usage       vk.ImageUsageFlags
	Desc        vk.ImageDescription
	orignalMips uint32
}

type NodeBuilder struct {
	Name     string
	Children []*NodeBuilder
	Location mgl32.Mat4
	Mesh     MeshIndex
	Material MaterialIndex
	Skin     SkinIndex
}

// SetMesh assign a mesh with material to node
func (nb *NodeBuilder) SetMesh(m MeshIndex, mat MaterialIndex) *NodeBuilder {
	nb.Mesh = m
	nb.Material = mat
	return nb
}

// SetMesh assign a mesh with material and skin to node
func (nb *NodeBuilder) SetSkinnedMesh(m MeshIndex, mat MaterialIndex, skin SkinIndex) *NodeBuilder {
	nb.Mesh = m
	nb.Material = mat
	nb.Skin = skin
	return nb
}

// Add childs adds child nodes to a node
func (nb *NodeBuilder) AddChild(child ...*NodeBuilder) *NodeBuilder {
	nb.Children = append(nb.Children, child...)
	return nb
}

// Add new material to model builders. Model builders ShaderFactory is finally used to convert material properties
// to shader. Most of shader also build a descriptor set that links all static assets like color and textures
// of a material to a single Vulkan descriptor set.
func (mb *ModelBuilder) AddMaterial(name string, props MaterialProperties) MaterialIndex {
	mi := MaterialIndex(len(mb.Materials))
	mb.Materials = append(mb.Materials, MaterialInfo{Props: props, Name: name})
	return mi
}

// FindMaterial retrieves converted named material. Return is -1, nil if no material was found
func (mb *ModelBuilder) FindMaterial(name string) (index MaterialIndex) {
	for idx, mi := range mb.Materials {
		if mi.Name == name {
			return MaterialIndex(idx)
		}
	}
	return -1
}

// ForNodes recursively enumerates though all nodes child nodes
func (nb *NodeBuilder) ForNodes(action func(n *NodeBuilder, parent *NodeBuilder, index int)) {
	for idx, ch := range nb.Children {
		action(ch, nb, idx)
		ch.ForNodes(action)
	}
}

// Add new Decal material to model builders. For Decal material model builder will not create a shader nor allocate any
// descriptor sets.
func (mb *ModelBuilder) AddDecalMaterial(name string, props MaterialProperties) MaterialIndex {
	mi := MaterialIndex(len(mb.Materials))
	mb.Materials = append(mb.Materials, MaterialInfo{Props: props, Name: name, Decal: true})
	return mi
}

// Attach image to model. This image can be them bound to material using material properties
func (mb *ModelBuilder) AddImage(kind string, content []byte, usage vk.ImageUsageFlags) ImageIndex {
	_ = mb.AddWhite()
	im := &ImageBuilder{Kind: kind, Content: content, index: ImageIndex(len(mb.Images)), Usage: usage}
	mb.Images = append(mb.Images, im)
	return im.index
}

// Add white adds simple white image to model. We can use while image instead of actual image when
// material don't have an image assigned to it. Vulkan requires that we bind something to each allocates image slot
// so we can use this white image when we don't have anything real to bind.
// (Vulkan 1.2 and Vulkan 1.1 with extensions support partially bound descriptor and this kind of binding is no longer necessary)

func (mb *ModelBuilder) AddWhite() ImageIndex {
	if mb.white == nil {
		mb.white = &ImageBuilder{Kind: "dds", Content: white_bin, index: ImageIndex(len(mb.Images)),
			Usage: vk.IMAGEUsageSampledBit | vk.IMAGEUsageTransferDstBit}
		mb.Images = append(mb.Images, mb.white)
	}
	return mb.white.index
}

// Add a mesh to model. Actual mesh content is built with MeshBuilder
func (mb *ModelBuilder) AddMesh(mesh *MeshBuilder) MeshIndex {

	index := MeshIndex(len(mb.Meshes))
	mb.Meshes = append(mb.Meshes, mesh)
	return index
}

// Add new node to model. If parent node is empty, predefined root node, named _root will be used.
func (mb *ModelBuilder) AddNode(name string, parent *NodeBuilder, transform mgl32.Mat4) *NodeBuilder {
	if parent == nil {
		if mb.Root == nil {
			mb.Root = &NodeBuilder{Name: "_root", Location: mgl32.Ident4(), Mesh: -1, Material: -1}
		}
		parent = mb.Root
	}
	n := &NodeBuilder{Name: name, Location: transform, Mesh: -1, Material: -1}
	parent.Children = append(parent.Children, n)
	return n
}

// Convert content of model builder to actual model and uploads model content (except skins) to GPU
func (mb *ModelBuilder) ToModel(dev *vk.Device) (*Model, error) {
	mb.wg = &sync.WaitGroup{}
	var err error
	m := &Model{}
	mb.AddWhite()
	m.memPool = vk.NewMemoryPool(dev)
	m.owner.AddChild(m.memPool)
	m.images = make([]*vk.Image, 0, len(mb.Images))
	var imMaxLen uint64
	for _, ib := range mb.Images {
		if ib.Kind != "raw" {
			err = vasset.DescribeImage(ib.Kind, &ib.Desc, ib.Content)
			if err != nil {
				return nil, err
			}
		} else {
			if ib.Desc.Layers == 0 {
				return nil, errors.New("Describe raw images before ToModel")
			}
		}
		imLen := ib.Desc.ImageSize()
		if imLen > imMaxLen {
			imMaxLen = imLen
		}
		desc := ib.Desc
		ib.orignalMips = desc.MipLevels
		if desc.MipLevels < mb.MipLevels && mb.canDoMips(desc) {
			desc.MipLevels = mb.MipLevels
			ib.Desc.MipLevels = mb.MipLevels
			ib.Usage |= vk.IMAGEUsageStorageBit
		}
		img := m.memPool.ReserveImage(desc, ib.Usage)
		m.images = append(m.images, img)
	}
	var iLen [MESHMax]uint64
	var vLen [MESHMax]uint64

	for _, ms := range mb.Meshes {
		ms.buildUvs()
		err = ms.buildNormals()
		if err != nil {
			return nil, err
		}
		ms.buildTangets()
		// mesh.buildQTangent()
		hasWeights := ms.buildWeights()
		var aabb AABB
		for idx, vb := range ms.Vextexies {
			aabb.Add(idx == 0, vb.Position)
		}
		ms.aabb = aabb
		if hasWeights {
			ms.kind = MESHKindSkinned
		}
		var vertexSize uint64
		switch ms.kind {
		case MESHKindNormal:
			vertexSize = uint64(unsafe.Sizeof(normalVertex{}))
		case MESHKindSkinned:
			vertexSize = uint64(unsafe.Sizeof(skinnedVertex{}))
		}
		iLen[ms.kind] += uint64(len(ms.Incides)) * 4
		vLen[ms.kind] += vertexSize * uint64(len(ms.Vextexies))
	}
	for idx := 0; idx < MESHMax; idx++ {
		if iLen[idx] > 0 {
			m.vertexies[idx].bIndex = m.memPool.ReserveBuffer(iLen[idx], false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageIndexBufferBit)
			m.vertexies[idx].bVertex = m.memPool.ReserveBuffer(vLen[idx], false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageVertexBufferBit)
		}
	}
	ubfLen, err := mb.buildMaterials(dev, m)
	if err != nil {
		return nil, err
	}
	if ubfLen > 0 {
		m.bUbf = m.memPool.ReserveBuffer(ubfLen, false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageUniformBufferBit)
	}
	m.memPool.Allocate()
	cp := NewCopier(dev)
	defer cp.Dispose()

	mb.copyNormalVertex(m, cp)
	mb.copySkinnedVertex(m, cp)
	for _, ib := range mb.Images {
		mb.wg.Add(1)
		go mb.copyImage(m, dev, ib)
	}
	m.sampler = GetDefaultSampler(dev)
	mb.wg.Wait()
	if len(m.images) > 0 {
		m.views = make([]*vk.ImageView, len(m.images))
		for idx, img := range m.images {
			m.views[idx] = img.DefaultView()
		}
	}
	mb.copyUbf(m, cp, ubfLen)
	mb.addNodes(mb.Root, m)
	m.skins = mb.Skins
	return m, nil
}

func (mb *ModelBuilder) copyNormalVertex(m *Model, cp *Copier) {
	var indices []uint32
	var vertexies []normalVertex
	for _, mesh := range mb.Meshes {
		offset := uint32(len(vertexies))
		iOffset := uint32(len(indices))
		if mesh.kind == MESHKindNormal {
			for _, vb := range mesh.Vextexies {
				vertexies = append(vertexies, normalVertex{position: vb.Position, uv: vb.Uv, normal: vb.Normal, tangent: vb.Tangent, color: vb.Color})
			}
			for _, idx := range mesh.Incides {
				indices = append(indices, idx+offset)
			}
			m.meshes = append(m.meshes, Mesh{Kind: MESHKindNormal, AABB: mesh.aabb,
				Model: m, From: iOffset, Count: uint32(len(mesh.Incides))})
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
	for _, mesh := range mb.Meshes {
		offset := uint32(len(vertexies))
		iOffset := uint32(len(indices))
		if mesh.kind == MESHKindSkinned {
			for _, vb := range mesh.Vextexies {
				vertexies = append(vertexies, skinnedVertex{position: vb.Position, uv: vb.Uv, normal: vb.Normal, tangent: vb.Tangent, color: vb.Color,
					weights: vb.Weights, joints: vb.Joints})
			}
			for _, idx := range mesh.Incides {
				indices = append(indices, idx+offset)
			}
			m.meshes = append(m.meshes, Mesh{Kind: MESHKindSkinned, AABB: mesh.aabb,
				Model: m, From: iOffset, Count: uint32(len(mesh.Incides))})
		}
	}
	if len(indices) > 0 {
		cp.CopyToBuffer(m.vertexies[MESHKindSkinned].bIndex, vk.UInt32ToBytes(indices))
		cp.CopyToBuffer(m.vertexies[MESHKindSkinned].bVertex, skinnedVertexToBytes(vertexies))
	}
}

func (mb *ModelBuilder) copyImage(m *Model, dev *vk.Device, ib *ImageBuilder) {
	defer func() {
		mb.wg.Done()
	}()
	cp := NewCopier(dev)
	defer cp.Dispose()
	if ib.Desc.MipLevels > ib.orignalMips {
		r := ib.Desc.FullRange()
		cp.SetLayout(m.images[ib.index], r, vk.IMAGELayoutGeneral)
		r = vk.ImageRange{LayerCount: ib.Desc.Layers, LevelCount: 1}
		cp.CopyToImage(m.images[ib.index], ib.Kind, ib.Content, r, vk.IMAGELayoutGeneral)

		comp := NewCompute(dev)
		defer comp.Dispose()
		for mip := ib.orignalMips; mip < mb.MipLevels; mip++ {
			for l := uint32(0); l < ib.Desc.Layers; l++ {
				comp.MipImage(m.images[ib.index], l, mip)
			}
		}
		r = ib.Desc.FullRange()
		r.Layout = vk.IMAGELayoutGeneral
		cp.SetLayout(m.images[ib.index], r, vk.IMAGELayoutShaderReadOnlyOptimal)
	} else {
		cp.CopyToImage(m.images[ib.index], ib.Kind, ib.Content, ib.Desc.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	}
}

func (mb *ModelBuilder) buildMaterials(dev *vk.Device, m *Model) (uint64, error) {
	mCounts := make(map[*vk.DescriptorLayout]int)
	mPools := make(map[*vk.DescriptorLayout]*vk.DescriptorPool)
	for idx, mt := range mb.Materials {
		if !mt.Decal {
			if mb.ShaderFactory == nil {

				return 0, errors.New("Set ShaderFactory")
			}
			mt.mat, mt.layout, mt.ubf, mt.images = mb.ShaderFactory(dev, mt.Props)
			mCounts[mt.layout] = mCounts[mt.layout] + 1
		}
		mb.Materials[idx] = mt

	}

	for idx, mCount := range mCounts {
		pool := vk.NewDescriptorPool(idx, mCount)
		m.owner.AddChild(pool)
		mPools[idx] = pool
	}

	offset := uint64(0)
	for idx, mi := range mb.Materials {
		if mi.Decal {
			m.materials = append(m.materials, Material{Props: mi.Props, Name: mi.Name})
			continue
		}
		mi.ds = mPools[mi.layout].Alloc()
		mi.offset = offset
		mb.Materials[idx] = mi
		rem := uint64(len(mi.ubf) % vk.MinUniformBufferOffsetAlignment)
		if rem > 0 {
			offset += uint64(len(mi.ubf)) + vk.MinUniformBufferOffsetAlignment - rem
		} else {
			offset += uint64(len(mi.ubf))
		}
		m.materials = append(m.materials, Material{Shader: mi.mat, Props: mi.Props, Name: mi.Name})
	}
	return offset, nil
}

func (mb *ModelBuilder) copyUbf(m *Model, cp *Copier, ubfLen uint64) {
	if ubfLen == 0 {
		return
	}
	ubfs := make([]byte, ubfLen)
	for mIndex, mi := range mb.Materials {
		if mi.Decal {
			continue
		}
		copy(ubfs[mi.offset:], mi.ubf)
		mi.ds.WriteSlice(0, 0, m.bUbf.Slice(mi.offset, mi.offset+uint64(len(mi.ubf))))
		for idx, ib := range mi.images {
			mi.ds.WriteImage(1, uint32(idx), m.views[ib], m.sampler)
			m.materials[mIndex].Shader.SetDescriptor(mi.ds)
			sm, ok := m.materials[mIndex].Shader.(BoundShader)
			if ok {
				sm.SetModel(m)
			}
		}
	}
	cp.CopyToBuffer(m.bUbf, ubfs)
}

func (mb *ModelBuilder) addNodes(n *NodeBuilder, m *Model) NodeIndex {
	if n == nil {
		return -1
	}
	node := Node{Model: m, Name: n.Name, Transform: n.Location}
	result := NodeIndex(len(m.nodes))
	node.Mesh = n.Mesh
	node.Material = n.Material
	node.Skin = n.Skin
	m.nodes = append(m.nodes, node)
	if len(n.Children) > 0 {
		for _, ch := range n.Children {
			node.Children = append(node.Children, mb.addNodes(ch, m))
		}
		m.nodes[result] = node
	}
	return result
}

// Add skin to model
func (mb *ModelBuilder) AddSkin(skin Skin) SkinIndex {
	mb.Skins = append(mb.Skins, skin)
	idx := SkinIndex(len(mb.Skins))
	return idx
}

func (mb *ModelBuilder) canDoMips(desc vk.ImageDescription) bool {
	w, h := desc.Width, desc.Height
	return (w>>mb.MipLevels) > 1 && (h>>mb.MipLevels) > 1
}

type MaterialInfo struct {
	Props  MaterialProperties
	Name   string
	Decal  bool
	mat    Shader
	offset uint64
	ubf    []byte
	images []ImageIndex
	ds     *vk.DescriptorSet
	layout *vk.DescriptorLayout
}

// MeshBuilder is used to construct one mesh. Mesh is then added to model builder
type MeshBuilder struct {
	Vextexies []*VertexBuilder
	Incides   []uint32
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
	Position mgl32.Vec3
	Uv       mgl32.Vec2
	Normal   mgl32.Vec3
	Tangent  mgl32.Vec3
	Color    mgl32.Vec4
	Weights  mgl32.Vec4
	// qtangent mgl32.Quat
	Joints [4]uint16
	flags  VertexFlags
}

// Add new vertex to mesh. You must provide at least position of mesh
func (mb *MeshBuilder) AddVertex(position mgl32.Vec3) *VertexBuilder {
	vb := &VertexBuilder{Position: position, Index: uint32(len(mb.Vextexies))}
	mb.Vextexies = append(mb.Vextexies, vb)
	return vb
}

// Add indexes to vertexes. Mesh can be used without index information, in witch case each set of three vertexes will make
// one triagle. VGE don't support any other modes to combine vertexes. Use indexes if you wan't to reuse same vertex multiple
// times
func (mb *MeshBuilder) AddIndex(points ...uint32) (index int) {
	idx := len(mb.Incides) / len(points)
	mb.Incides = append(mb.Incides, points...)
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
	for _, vb := range mb.Vextexies {
		if vb.flags&VFUV != 0 {
			continue
		}
		pos := vb.Position
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
	for _, vb := range mb.Vextexies {
		if vb.flags&VFNormal != 0 {
			continue
		}
		pos := vb.Position

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
	if len(mb.Vextexies) == 0 || mb.Vextexies[0].flags&VFTangent != 0 {
		return
	}
	incides := mb.Incides
	bitangents := make([]mgl32.Vec3, len(mb.Vextexies))
	for x := 0; x < len(incides); x += 3 {
		vb1 := mb.Vextexies[incides[x]]
		vb2 := mb.Vextexies[incides[x+1]]
		vb3 := mb.Vextexies[incides[x+2]]

		deltuvs1 := vb2.Uv.Sub(vb1.Uv)
		deltuvs2 := vb3.Uv.Sub(vb1.Uv)
		deltpositions1 := vb2.Position.Sub(vb1.Position)
		deltpositions2 := vb3.Position.Sub(vb1.Position)
		r := 1.0 / (deltuvs1.X()*deltuvs2.Y() - deltuvs1.Y()*deltuvs2.X())
		var tangent mgl32.Vec3
		var bitangent mgl32.Vec3
		if math.IsInf(float64(r), 0) {
			pos1 := vb1.Position
			tangent = mgl32.Vec3{-pos1.Y(), pos1.Z(), pos1.X()}
			bitangent = mgl32.Vec3{pos1.Z(), pos1.X(), -pos1.Y()}
		} else {
			tangent = deltpositions1.Mul(deltuvs2.Y()).Sub(deltpositions2.Mul(deltuvs1.Y())).Mul(r)
			// bitangent = (deltpositions2 * deltuvs1.x   - deltpositions1 * deltuvs2.x)*r;
			bitangent = deltpositions2.Mul(deltuvs1.X()).Sub(deltpositions1.Mul(deltuvs2.X())).Mul(r)
		}
		vb1.Tangent = vb1.Tangent.Add(tangent)
		vb2.Tangent = vb2.Tangent.Add(tangent)
		vb3.Tangent = vb3.Tangent.Add(tangent)
		bitangents[incides[x]] = bitangents[incides[x]].Add(bitangent)
		bitangents[incides[x+1]] = bitangents[incides[x+1]].Add(bitangent)
		bitangents[incides[x+2]] = bitangents[incides[x+2]].Add(bitangent)
	}

	for idx, vb := range mb.Vextexies {
		n := vb.Normal
		t := vb.Tangent
		// t = (t - n * dot(n, t)).normalize();
		t2 := t.Sub(n.Mul(n.Dot(t))).Normalize()
		// if (glm::dot(glm::cross(n, t), b) < 0.0f){
		//     t = t * -1.0f;
		// }
		if n.Cross(t).Dot(bitangents[idx]) > 0 {
			t2 = t2.Mul(-1)
		}
		vb.Tangent = t2
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
	for _, vb := range mb.Vextexies {
		if vb.flags&VFWeights != 0 {
			hasWeights = true
			break
		}
	}
	if hasWeights {
		for idx, vb := range mb.Vextexies {
			w1 := 1 - vb.Weights[1] + vb.Weights[2] + vb.Weights[3]
			mb.Vextexies[idx].Weights[0] = w1
		}
	}
	return
}

// GetPosition retrieves all mesh positions as float32 array
func (mb *MeshBuilder) GetPositions() (positions []float32) {
	for _, vb := range mb.Vextexies {
		positions = append(positions, vb.Position[:]...)
	}
	return
}

func (vb *VertexBuilder) AddUV(uv mgl32.Vec2) *VertexBuilder {
	vb.Uv = uv
	vb.flags |= VFUV
	return vb
}

func (vb *VertexBuilder) AddNormal(normal mgl32.Vec3) *VertexBuilder {
	vb.Normal = normal
	vb.flags |= VFNormal
	return vb
}

func (vb *VertexBuilder) AddTangent(tangent mgl32.Vec3) *VertexBuilder {
	vb.Tangent = tangent
	vb.flags |= VFTangent
	return vb
}

// AddColor adds vertex color. This is not used in current shaders but can be like an extra vec4 value is vertex input
func (vb *VertexBuilder) AddColor(color mgl32.Vec4) *VertexBuilder {
	vb.Color = color
	return vb
}

// Add weight and joints to vertex. If you add weights and joints to vertex it will convert whole
// mesh to a skinned mesh. You must also provide Skin to model nodes that uses skinned mesh.
// Skinned meshes won't be shown without skin and of cause normal meshes can't handle skins.
func (vb *VertexBuilder) AddWeights(weights mgl32.Vec4, joints ...uint16) *VertexBuilder {
	vb.Weights = weights
	copy(vb.Joints[:], joints)
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
