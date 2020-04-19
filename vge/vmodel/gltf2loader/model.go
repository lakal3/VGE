package gltf2loader

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
)

const (
	BYTE                 = ComponentType(5120)
	UNSIGNED_BYTE        = ComponentType(5121)
	SHORT                = ComponentType(5122)
	UNSIGNED_SHORT       = ComponentType(5123)
	UNSIGNED_INT         = ComponentType(5125)
	FLOAT                = ComponentType(5126)
	ARRAY_BUFFER         = Target(34962)
	ELEMENT_ARRAY_BUFFER = Target(34963)
)

type GLTF2 struct {
	ExtensionsUsed     []string      `json:"extensionsUsed"`
	ExtensionsRequired []string      `json:"extensionsRequired"`
	Accessors          []*Accessor   `json:"accessors"`
	Animations         []*Animation  `json:"animations"`
	Buffers            []*Buffer     `json:"buffers"`
	BufferViews        []*BufferView `json:"bufferViews"`
	Meshes             []*Mesh       `json:"meshes"`
	Materials          []*Material   `json:"materials"`
	Nodes              []*Node       `json:"nodes"`
	Images             []*Image      `json:"images"`
	Skins              []*Skin       `json:"skins"`
	Scene              int           `json:"scene"`
	Scenes             []*Scene      `json:"scenes"`
}

type ComponentType uint32
type Target uint32

type GLTFBase struct {
	Name       string          `json:"name"`
	Extensions json.RawMessage `json:"extensions"`
	Extras     json.RawMessage `json:"extras"`
	gltf       *GLTF2
}

type Accessor struct {
	BufferView    int           `json:"bufferView"`
	ComponentType ComponentType `json:"componentType"`
	Normalized    bool          `json:"normalized"`
	ByteOffset    int           `json:"byteOffset"`
	Count         int           `json:"count"`
	Type          string        `json:"type"`
	Min           []float32     `json:"min"`
	Max           []float32     `json:"max"`
	GLTFBase
}

type Buffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri"`
	GLTFBase
	data []byte
}

type BufferView struct {
	Buffer     int    `json:"buffer"`
	ByteLength int    `json:"byteLength"`
	ByteOffset int    `json:"byteOffset"`
	ByteStride int    `json:"byteStride"`
	Target     Target `json:"target"`
	GLTFBase
}

type Image struct {
	URI        string `json:"uri,omitempty"`
	MimeType   string `json:"mimeType"`
	BufferView *int   `json:"bufferView,omitempty"`
	GLTFBase
	index   vmodel.ImageIndex
	kind    string
	content []byte
}

type Mesh struct {
	Primitives []Primitive `json:"primitives"`
	Weights    []float32   `json:"weights,omitempty"`
	GLTFBase
}

type Primitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    *int           `json:"indices,omitempty"`
	Material   *int           `json:"material,omitempty"`
	Mode       *int           `json:"mode,omitempty"`
	GLTFBase
}

type Node struct {
	Camera      *int      `json:"camera,omitempty"`
	Children    []int     `json:"children,omitempty"`
	Mesh        *int      `json:"mesh,omitempty"`
	Skin        *int      `json:"skin,omitempty"`
	Rotation    []float32 `json:"rotation,omitempty"`
	Scale       []float32 `json:"scale,omitempty"`
	Translation []float32 `json:"translation,omitempty"`
	Matrix      []float32 `json:"matrix,omitempty"`
	GLTFBase
}

type Skin struct {
	InverseBindMatrices int    `json:"inverseBindMatrices"`
	Joints              []int  `json:"joints"`
	Name                string `json:"name"`
	GLTFBase
}

type Texture struct {
	Index int `json:"index"`
}
type PbrMetallicRoughness struct {
	BaseColorFactor          []float32 `json:"baseColorFactor"`
	RoughnessFactor          *float32  `json:"roughnessFactor"`
	MetallicFactor           *float32  `json:"metallicFactor"`
	BaseColorTexture         *Texture  `json:"baseColorTexture"`
	MetallicRoughnessTexture *Texture  `json:"metallicRoughnessTexture"`
}

type Material struct {
	PbrMetallicRoughness *PbrMetallicRoughness `json:"pbrMetallicRoughness"`
	NormalTexture        *Texture              `json:"normalTexture"`
	OcclusionTexture     *Texture              `json:"occlusionTexture"`
	EmissiveFactor       []float32             `json:"emissiveFactor"`
	EmissiveTexture      *Texture              `json:"emissiveTexture"`
	Name                 string                `json:"name"`
}

type Channels struct {
	Sampler int `json:"sampler"`
	Target  struct {
		Node int    `json:"node"`
		Path string `json:"path"`
	} `json:"target"`
}

type Sampler struct {
	Input         int    `json:"input"`
	Interpolation string `json:"interpolation"`
	Output        int    `json:"output"`
}

type Animation struct {
	Channels []*Channels `json:"channels"`
	Name     string      `json:"name"`
	Samplers []*Sampler  `json:"samplers"`
	GLTFBase
}

func (node *Node) GetMesh() *Mesh {
	if node.Mesh != nil {
		return node.gltf.Meshes[*node.Mesh]
	}
	return nil
}

type Scene struct {
	Nodes []int `json:"nodes"`
	GLTFBase
}

func (gltf *GLTF2) Bind() {
	for _, a := range gltf.Accessors {
		a.gltf = gltf
	}
	for _, b := range gltf.Buffers {
		b.gltf = gltf
	}
	for _, b := range gltf.BufferViews {
		b.gltf = gltf
	}
	for _, m := range gltf.Meshes {
		m.bind(gltf)
	}
	for _, i := range gltf.Images {
		i.gltf = gltf
	}
	for _, n := range gltf.Nodes {
		n.gltf = gltf
	}
	for _, s := range gltf.Scenes {
		s.gltf = gltf
	}
}

func (gltf *GLTF2) GetIncides(accessor int) ([]uint32, error) {
	ac := gltf.Accessors[accessor]
	bv := gltf.BufferViews[ac.BufferView]
	data := gltf.GetContent(bv)
	result := make([]uint32, ac.Count, ac.Count)
	for idx := 0; idx < ac.Count; idx++ {
		switch ac.ComponentType {
		case UNSIGNED_BYTE:
			result[idx] = uint32(data[ac.ByteOffset+idx])
		case UNSIGNED_SHORT:
			result[idx] = uint32(binary.LittleEndian.Uint16(data[ac.ByteOffset+idx*2 : ac.ByteOffset+idx*2+2]))
		case UNSIGNED_INT:
			result[idx] = uint32(binary.LittleEndian.Uint32(data[ac.ByteOffset+idx*4 : ac.ByteOffset+idx*4+4]))
		default:
			return nil, fmt.Errorf("Invalid incides type %d", ac.ComponentType)
		}
	}
	return result, nil
}

func (gltf *GLTF2) GetMatrix(accessor int) (vec []float32, elements int, err error) {
	ac := gltf.Accessors[accessor]
	if ac.ComponentType != FLOAT {
		return nil, 0, errors.New("Invalid type")
	}
	elements = 1
	switch ac.Type {
	case "MAT4":
		elements = 16
	default:
		return nil, 0, fmt.Errorf("Invalid accessor type: %s", ac.Type)
	}
	bv := gltf.BufferViews[ac.BufferView]
	data := gltf.GetContent(bv)
	result := make([]float32, ac.Count*elements, ac.Count*elements)
	stride := 0
	for idx := 0; idx < ac.Count*elements; idx++ {
		result[idx] = math.Float32frombits(binary.LittleEndian.Uint32(data[ac.ByteOffset+stride+idx*4 : ac.ByteOffset+stride+idx*4+4]))
		if (idx+1)%elements == 0 && bv.ByteStride > 0 {
			stride = stride + bv.ByteStride - 4*elements
		}
	}
	return result, elements, nil
}

func (gltf *GLTF2) GetFloats(accessor int) (vec []float32, elements int, err error) {
	ac := gltf.Accessors[accessor]
	if ac.ComponentType != FLOAT {
		return nil, 0, fmt.Errorf("Invalid component type %d", ac.ComponentType)
	}
	elements = 1
	switch ac.Type {
	case "SCALAR":
		elements = 1
	case "VEC2":
		elements = 2
	case "VEC3":
		elements = 3
	case "VEC4":
		elements = 4
	default:
		return nil, 0, fmt.Errorf("Invalid accessor type: %s", ac.Type)
	}
	bv := gltf.BufferViews[ac.BufferView]
	data := gltf.GetContent(bv)
	result := make([]float32, ac.Count*elements, ac.Count*elements)
	stride := 0
	for idx := 0; idx < ac.Count*elements; idx++ {
		result[idx] = math.Float32frombits(binary.LittleEndian.Uint32(data[ac.ByteOffset+stride+idx*4 : ac.ByteOffset+stride+idx*4+4]))
		if (idx+1)%elements == 0 && bv.ByteStride > 0 {
			stride = stride + bv.ByteStride - 4*elements
		}
	}
	return result, elements, nil
}

func (gltf *GLTF2) GetJointIndex(accessor int) (vec []uint16, elements int, err error) {
	ac := gltf.Accessors[accessor]
	size := 0
	if ac.ComponentType == UNSIGNED_SHORT {
		size = 2
	}
	if size == 0 {
		return nil, 0, fmt.Errorf("Invalid component type %d", ac.ComponentType)
	}
	elements = 1
	switch ac.Type {
	case "VEC4":
		elements = 4
	default:
		return nil, 0, fmt.Errorf("Invalid int accessor type: %s", ac.Type)
	}
	bv := gltf.BufferViews[ac.BufferView]
	data := gltf.GetContent(bv)
	result := make([]uint16, ac.Count*elements, ac.Count*elements)
	stride := 0
	for idx := 0; idx < ac.Count*elements; idx++ {
		var val uint16
		val = binary.LittleEndian.Uint16(data[ac.ByteOffset+stride+idx*2 : ac.ByteOffset+stride+idx*2+2])
		result[idx] = val
		if (idx+1)%elements == 0 && bv.ByteStride > 0 {
			stride = stride + bv.ByteStride - size*elements
		}
	}
	return result, elements, nil
}

func (mesh *Mesh) bind(gltf *GLTF2) {
	mesh.gltf = gltf
	for idx, _ := range mesh.Primitives {
		mesh.Primitives[idx].gltf = gltf
	}
}

func (b *Buffer) SetContent(data []byte) {
	b.data = data
}

func (gltf *GLTF2) GetContent(bv *BufferView) (data []byte) {
	return gltf.Buffers[bv.Buffer].data[bv.ByteOffset : bv.ByteLength+bv.ByteOffset]
}

func (p *Primitive) GetAttribute(attribute string) ([]float32, int, error) {
	ac, ok := p.Attributes[attribute]
	if !ok {
		return nil, 0, fmt.Errorf("No attribute: %s", attribute)
	}
	return p.gltf.GetFloats(ac)
}

// FindNodes from gltf scene. Scene number can be -1 for default scene
func (gltf *GLTF2) FindNodes(scene int, tr mgl32.Mat4, action func(n *Node, tr mgl32.Mat4) (cont bool)) {
	if scene < 0 || scene >= len(gltf.Scenes) {
		scene = gltf.Scene
	}
	sc := gltf.Scenes[scene]
	for _, n := range sc.Nodes {
		if !gltf.findNode(gltf.Nodes[n], tr, action) {
			return
		}
	}
}

func (gltf *GLTF2) findNode(n *Node, tr mgl32.Mat4, action func(n *Node, tr mgl32.Mat4) (cont bool)) bool {
	if !action(n, tr) {
		return false
	}
	n.ApplyTransform(&tr)
	for _, ch := range n.Children {
		cn := gltf.Nodes[ch]
		if !gltf.findNode(cn, tr, action) {
			return false
		}
	}
	return true
}

func (n *Node) ApplyTransform(tr *mgl32.Mat4) (hasTransform bool) {
	if len(n.Matrix) > 0 {
		tr2 := mgl32.Mat4{}
		copy(tr2[:], n.Matrix)
		*tr = tr.Mul4(tr2)
		hasTransform = true
	} else {
		if len(n.Translation) != 0 {
			*tr = tr.Mul4(mgl32.Translate3D(n.Translation[0], n.Translation[1], n.Translation[2]))
			hasTransform = true
		}
		if len(n.Rotation) != 0 {
			q := mgl32.Quat{V: mgl32.Vec3{n.Rotation[0], n.Rotation[1], n.Rotation[2]}, W: n.Rotation[3]}
			*tr = tr.Mul4(q.Mat4())
			hasTransform = true
		}
		if len(n.Scale) != 0 {
			*tr = tr.Mul4(mgl32.Scale3D(n.Scale[0], n.Scale[1], n.Scale[2]))
			hasTransform = true
		}
	}
	return
}
