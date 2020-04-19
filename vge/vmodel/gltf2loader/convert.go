package gltf2loader

import (
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
)

type meshMaterial struct {
	mesh vmodel.MeshIndex
	mat  vmodel.MaterialIndex
}

type GLTF2Loader struct {
	Builder    *vmodel.ModelBuilder
	Parent     *vmodel.NodeBuilder
	Loader     vasset.Loader
	Model      *GLTF2
	ImageUsage vk.ImageUsageFlags

	materials []vmodel.MaterialIndex
	meshes    [][]meshMaterial
	skins     []vmodel.SkinIndex
	streams   map[int][]float32
}

const APosition = "POSITION"
const AUV0 = "TEXCOORD_0"
const AUV1 = "TEXCOORD_1"
const ANormal = "NORMAL"
const ATangent = "TANGENT"
const AWeights0 = "WEIGHTS_0"
const AWeights1 = "WEIGHTS_1"
const AJoints0 = "JOINTS_0"
const AJoints1 = "JOINTS_1"

func (cc *GLTF2Loader) Convert(sceneIndex int) error {
	if cc.Builder == nil || cc.Model == nil {
		return errors.New("Set Scene and Model")
	}
	if cc.ImageUsage == 0 {
		cc.ImageUsage = vk.IMAGEUsageSampledBit | vk.IMAGEUsageTransferDstBit
	}
	return cc.toScene(sceneIndex)
}

func (cc *GLTF2Loader) loadImage(tx *Texture) (vmodel.ImageIndex, error) {
	if tx == nil {
		return 0, nil
	}
	img := cc.Model.Images[tx.Index]
	if img.index != 0 {
		return img.index, nil
	}
	if img.content == nil {
		return 0, fmt.Errorf("No content for image %d", tx.Index)
	}
	img.index = cc.Builder.AddImage(img.kind, img.content, cc.ImageUsage)
	return img.index, nil
}

func (cc *GLTF2Loader) toScene(sceneIndex int) (err error) {
	if cc.materials == nil {
		err = cc.convertMaterials()
		if err != nil {
			return err
		}
	}
	cc.convertScene(sceneIndex)
	return
}

func (cc *GLTF2Loader) convertMaterials() error {
	var err error
	gltf := cc.Model
	cc.meshes = make([][]meshMaterial, len(gltf.Meshes))
	cc.materials = make([]vmodel.MaterialIndex, len(gltf.Materials))

	// Default material
	cc.Builder.AddMaterial("_default", vmodel.NewMaterialProperties())
	for idx, m := range gltf.Materials {

		mNew, err := cc.mapMaterial(m)
		if err != nil {
			return err
		}
		cc.materials[idx] = mNew

	}
	err = cc.buildMeshes(gltf)
	if err != nil {
		return err
	}
	// err = cc.buildSkins(gltf)
	return nil
}

func (cc *GLTF2Loader) buildMeshes(gltf *GLTF2) error {
	for idx, m := range gltf.Meshes {

		for _, p := range m.Primitives {
			mb := &vmodel.MeshBuilder{}
			aPos := p.Attributes[APosition]
			var vbs []*vmodel.VertexBuilder
			posAttrs, posElement, err := gltf.GetFloats(aPos)
			if err != nil {
				return err
			}
			if posElement != 3 {
				return fmt.Errorf("Invalid position attribute. It should have 3 elements")
			}
			for idx := 0; idx < len(posAttrs); idx += 3 {
				vbs = append(vbs, mb.AddVertex(mgl32.Vec3{posAttrs[idx], posAttrs[idx+1], posAttrs[idx+2]}))
			}
			err = cc.addNormal(vbs, p)
			if err != nil {
				return err
			}
			err = cc.addTangets(vbs, p)
			if err != nil {
				return err
			}
			err = cc.addUV(vbs, p)
			if err != nil {
				return err
			}
			err = cc.addWeightJoints(vbs, p)
			if err != nil {
				return err
			}

			var incides []uint32
			if p.Indices != nil {
				incides, err = gltf.GetIncides(*p.Indices)
				if err != nil {
					return err
				}
			} else {
				to := len(vbs)
				for idx := 0; idx < to; idx++ {
					incides = append(incides, uint32(idx))
				}
			}
			mb.AddIndex(incides...)
			n := cc.Builder.AddMesh(mb)
			mm := meshMaterial{mesh: n}
			if p.Material != nil {
				mm.mat = cc.materials[*p.Material]
			}
			cc.meshes[idx] = append(cc.meshes[idx], mm)
		}
	}
	return nil
}

func (cc *GLTF2Loader) convertScene(sceneIndex int) {
	gltf := cc.Model
	nl := gltf.Scenes[sceneIndex].Nodes
	cc.convertLevel(nl, cc.Parent)
}

func (cc *GLTF2Loader) convertLevel(nodes []int, parent *vmodel.NodeBuilder) (err error) {

	for _, nIdx := range nodes {
		local := mgl32.Ident4()
		n := cc.Model.Nodes[nIdx]
		n.ApplyTransform(&local)
		sk := vmodel.SkinIndex(0)
		if n.Skin != nil {
			sk, err = cc.buildSkin(*n.Skin)
			if err != nil {
				return err
			}
		}

		newNode := cc.Builder.AddNode(n.Name, parent, local)
		if n.Mesh != nil {
			mList := cc.meshes[*n.Mesh]
			for chIdx, m := range mList {
				mName := fmt.Sprintf("%s_%d", n.Name, chIdx+1)
				if sk > 0 {

					cc.Builder.AddNode(mName, newNode, mgl32.Ident4()).SetSkinnedMesh(m.mesh, m.mat, sk)
				} else {
					cc.Builder.AddNode(mName, newNode, mgl32.Ident4()).SetMesh(m.mesh, m.mat)
				}
			}
		}
		if n.Children != nil {
			err = cc.convertLevel(n.Children, newNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cc *GLTF2Loader) mapMaterial(m *Material) (vmodel.MaterialIndex, error) {
	props := vmodel.NewMaterialProperties()
	var col mgl32.Vec4
	var txIndex vmodel.ImageIndex
	var err error
	if m.PbrMetallicRoughness != nil {
		pbr := m.PbrMetallicRoughness
		txIndex, err = cc.loadImage(pbr.BaseColorTexture)
		if err != nil {
			return 0, err
		}
		props.SetImage(vmodel.TxAlbedo, txIndex)
		if len(pbr.BaseColorFactor) > 0 {
			parseColor(&col, pbr.BaseColorFactor)
		} else {
			col = mgl32.Vec4{1, 1, 1, 1}
		}
		props.SetColor(vmodel.CAlbedo, col)
		txIndex, err = cc.loadImage(pbr.MetallicRoughnessTexture)
		if err != nil {
			return 0, err
		}
		props.SetImage(vmodel.TxMetallicRoughness, txIndex)
		if pbr.RoughnessFactor != nil {
			props.SetFactor(vmodel.FRoughness, *pbr.RoughnessFactor)
		} else {
			props.SetFactor(vmodel.FRoughness, 0.9)
		}
		if pbr.MetallicFactor != nil {
			props.SetFactor(vmodel.FMetalness, *pbr.MetallicFactor)
		} else {
			props.SetFactor(vmodel.FMetalness, 1)
		}
	} else {
		props.SetFactor(vmodel.FRoughness, 0.9)
		props.SetFactor(vmodel.FMetalness, 0.9)
	}
	txIndex, err = cc.loadImage(m.NormalTexture)
	if err != nil {
		return 0, err
	}
	props.SetImage(vmodel.TxBump, txIndex)
	txIndex, err = cc.loadImage(m.EmissiveTexture)
	if err != nil {
		return 0, err
	}
	props.SetImage(vmodel.TxEmissive, txIndex)
	if len(m.EmissiveFactor) > 0 {
		parseColor(&col, m.EmissiveFactor)
		props.SetColor(vmodel.CEmissive, col)
	} else if txIndex != 0 {
		props.SetColor(vmodel.CEmissive, mgl32.Vec4{1, 1, 1, 1})
	}
	return cc.Builder.AddMaterial(m.Name, props), nil
}

func (cc *GLTF2Loader) isRoot(sk *vmodel.Skin, joint int) bool {
	for _, j := range sk.Joints {
		for _, ch := range j.Children {
			if ch == joint {
				return false
			}
		}
	}
	return true
}

func (cc *GLTF2Loader) buildJoint(jNode int, joints map[int]int) vmodel.Joint {
	j := vmodel.Joint{}
	n := cc.Model.Nodes[jNode]
	for _, ch := range n.Children {
		chIdx, ok := joints[ch]
		if !ok {
			continue
		}
		j.Children = append(j.Children, chIdx)
	}
	j.Name = n.Name
	if len(n.Translation) > 2 {
		j.Translate = mgl32.Vec3{n.Translation[0], n.Translation[1], n.Translation[2]}
	}
	if len(n.Rotation) > 3 {
		j.Rotate = mgl32.Quat{V: mgl32.Vec3{n.Rotation[0], n.Rotation[1], n.Rotation[2]}, W: n.Rotation[3]}
	} else {
		j.Rotate = mgl32.Quat{W: 1}
	}
	if len(n.Scale) > 2 {
		j.Scale = mgl32.Vec3{n.Scale[0], n.Scale[1], n.Scale[2]}
	} else {
		j.Scale = mgl32.Vec3{1, 1, 1}
	}

	return j
}

func (cc *GLTF2Loader) addAnimations(sk *vmodel.Skin, jointMap map[int]int) error {
	for _, an := range cc.Model.Animations {
		err := cc.addAnimation(an, sk, jointMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cc *GLTF2Loader) addAnimation(animation *Animation, sk *vmodel.Skin, jointMap map[int]int) error {
	an := vmodel.Animation{Name: animation.Name}

	for _, ch := range animation.Channels {
		jt, ok := jointMap[ch.Target.Node]
		if !ok {
			continue
		}
		var target vmodel.ChannelTarget
		switch ch.Target.Path {
		case "translation":
			target = vmodel.TTranslation
		case "rotation":
			target = vmodel.TRotation
		case "scale":
			target = vmodel.TScale
		default:
			return fmt.Errorf("Unknown channel target %s", ch.Target.Path)
		}
		input, output, err := cc.getSampler(target, animation.Samplers[ch.Sampler])
		if err != nil {
			return err
		}
		if len(input) > 1 {
			an.Channels = append(an.Channels, vmodel.Channel{Target: target, Joint: jt, Input: input, Output: output})
		}
	}
	if len(an.Channels) > 0 {
		sk.Animations = append(sk.Animations, an)
	}
	return nil
}

func (cc *GLTF2Loader) getSampler(target vmodel.ChannelTarget, sampler *Sampler) (input []float32, output []float32, err error) {
	if cc.streams == nil {
		cc.streams = make(map[int][]float32)
	}
	input, ok := cc.streams[sampler.Input]
	if !ok {
		fl, el, err := cc.Model.GetFloats(sampler.Input)
		if err != nil {
			return nil, nil, err
		}
		if el != 1 {
			return nil, nil, errors.New("Input sample should have element size 1")
		}
		input = fl
		cc.streams[sampler.Input] = input
	}
	output, ok = cc.streams[sampler.Output]
	if !ok {
		fl, el, err := cc.Model.GetFloats(sampler.Output)
		if err != nil {
			return nil, nil, err
		}
		if (el != 3 && target != vmodel.TRotation) || (el != 4 && target == vmodel.TRotation) {
			return nil, nil, errors.New("Output sample should have element size 3 or 4")
		}
		output = fl
		cc.streams[sampler.Output] = output
	}
	return
}

func (cc *GLTF2Loader) buildSkin(skinNro int) (vmodel.SkinIndex, error) {
	gltf := cc.Model
	rawSk := gltf.Skins[skinNro]
	sk := vmodel.Skin{}
	jointMap := make(map[int]int)
	for idx, jIdx := range rawSk.Joints {
		jointMap[jIdx] = idx
		n := gltf.Nodes[jIdx]
		if len(n.Matrix) > 0 {
			return 0, errors.New("Joint nodes can't have full transform matrix")
		}
		if n.Skin != nil || n.Mesh != nil {
			return 0, errors.New("Joint nodes can't have skin or mesh")
		}
	}
	sk.Joints = make([]vmodel.Joint, len(rawSk.Joints))
	mxFloats, _, err := cc.Model.GetMatrix(rawSk.InverseBindMatrices)
	if err != nil {
		return 0, err
	}
	for idx, jIdx := range rawSk.Joints {
		sk.Joints[idx] = cc.buildJoint(jIdx, jointMap)
		copy(sk.Joints[idx].InverseMatrix[:], mxFloats[16*idx:])
	}
	err = cc.addAnimations(&sk, jointMap)
	if err != nil {
		return 0, err
	}
	for idx, _ := range rawSk.Joints {
		sk.Joints[idx].Root = cc.isRoot(&sk, idx)
	}
	return cc.Builder.AddSkin(sk), nil
}

func (cc *GLTF2Loader) addNormal(vbs []*vmodel.VertexBuilder, p Primitive) error {
	ac, ok := p.Attributes[ANormal]
	if !ok {
		return nil
	}
	attrs, elements, err := cc.Model.GetFloats(ac)
	if err != nil {
		return err
	}
	if elements != 3 {
		return fmt.Errorf("Normals should have 3 elements, not %d", elements)
	}
	if len(attrs) != 3*len(vbs) {
		return fmt.Errorf("%d normals and %d vertexies", len(attrs)/3, len(vbs))
	}
	for idx := 0; idx < len(attrs); idx += 3 {
		vbs[idx/3].AddNormal(mgl32.Vec3{attrs[idx], attrs[idx+1], attrs[idx+2]})
	}
	return nil
}

func (cc *GLTF2Loader) addTangets(vbs []*vmodel.VertexBuilder, p Primitive) error {
	ac, ok := p.Attributes[ATangent]
	if !ok {
		return nil
	}
	attrs, elements, err := cc.Model.GetFloats(ac)
	if err != nil {
		return err
	}

	if elements != 3 && elements != 4 {
		return fmt.Errorf("Tangets should have 3 or 4 elements, not %d", elements)
	}
	if len(attrs) != elements*len(vbs) {
		return fmt.Errorf("%d normals and %d vertexies", len(attrs)/3, len(vbs))
	}
	for idx := 0; idx < len(attrs); idx += elements {
		vbs[idx/elements].AddTangent(mgl32.Vec3{attrs[idx], attrs[idx+1], attrs[idx+2]})
	}
	return nil
}

func (cc *GLTF2Loader) addUV(vbs []*vmodel.VertexBuilder, p Primitive) error {
	ac, ok := p.Attributes[AUV0]
	if !ok {
		return nil
	}
	attrs, elements, err := cc.Model.GetFloats(ac)
	if err != nil {
		return err
	}

	if elements != 2 {
		return fmt.Errorf("UVs should have 2 elements, not %d", elements)
	}
	if len(attrs) != 2*len(vbs) {
		return fmt.Errorf("%d normals and %d vertexies", len(attrs)/3, len(vbs))
	}
	for idx := 0; idx < len(attrs); idx += 2 {
		vbs[idx/elements].AddUV(mgl32.Vec2{attrs[idx], attrs[idx+1]})
	}
	return nil
}

func (cc *GLTF2Loader) addWeightJoints(vbs []*vmodel.VertexBuilder, p Primitive) error {
	ac, ok := p.Attributes[AJoints0]
	if !ok {
		return nil
	}
	aw, ok := p.Attributes[AWeights0]
	if !ok {
		return nil
	}
	iAttrs, elements, err := cc.Model.GetJointIndex(ac)
	if err != nil {
		return err
	}

	if elements != 4 {
		return fmt.Errorf("Joints should have 4 elements, not %d", elements)
	}
	if len(iAttrs) != elements*len(vbs) {
		return fmt.Errorf("%d joins and %d vertexies", len(iAttrs)/4, len(vbs))
	}
	attrs, elements, err := cc.Model.GetFloats(aw)
	if elements != 4 {
		return fmt.Errorf("Weight should have 4 elements, not %d", elements)
	}
	if len(iAttrs) != elements*len(vbs) {
		return fmt.Errorf("%d weights and %d vertexies", len(attrs)/4, len(vbs))
	}
	for idx := 0; idx < len(attrs); idx += elements {
		vbs[idx/elements].AddWeights(mgl32.Vec4{attrs[idx], attrs[idx+1], attrs[idx+2], attrs[idx+3]},
			iAttrs[idx], iAttrs[idx+1], iAttrs[idx+2], iAttrs[idx+3])
	}
	return nil
}

func parseColor(colors *mgl32.Vec4, parts []float32) {
	for idx := 0; idx < len(parts) && idx < 4; idx++ {
		colors[idx] = parts[idx]
	}
}

func atan2(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}
