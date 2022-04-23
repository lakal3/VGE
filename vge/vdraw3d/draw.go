package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"sort"
	"unsafe"
)

type FrozenMesh struct {
	mesh         vmodel.Mesh
	mat          material
	transparent  bool
	colorShader  string
	depthShader  string
	shadowShader string
	probeShader  string
	pickShader   string

	views   [8]vk.VImageView
	sampler [8]*vk.Sampler
}

type AnimatedMesh struct {
	FrozenMesh
	mxAnims   []mgl32.Mat4
	joints    []vmodel.Joint
	dsJoints  *vk.DescriptorSet
	ubfJoints *vk.ASlice
}

func (f *FrozenMesh) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	return storageOffset
}

func (am *AnimatedMesh) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, 64*uint64(len(am.mxAnims)))
	fi.ReserveDescriptor(GetJointsLayout(fi.Device()))
	return storageOffset
}

func (f *FrozenMesh) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(RenderColor)
	if ok {
		return len(f.colorShader) > 0
	}
	_, ok = phase.(UpdateFrame)
	if ok {
		return true
	}
	_, ok = phase.(RenderDepth)
	if ok {
		return len(f.depthShader) > 0
	}
	_, ok = phase.(RenderProbe)
	if ok {
		return len(f.probeShader) > 0
	}
	_, ok = phase.(RenderShadow)
	if ok {
		return len(f.shadowShader) > 0
	}
	_, ok = phase.(RenderPick)
	if ok {
		return len(f.pickShader) > 0
	}

	return ok
}

func (f *FrozenMesh) Clone() Frozen {
	return f
}

func (f *FrozenMesh) Render(fi *vk.FrameInstance, phase Phase) {
	ri, ok := phase.(UpdateFrame)
	if ok {
		for idx := 0; idx < 8; idx++ {
			if f.views[idx] != nil {
				if idx >= 4 {
					f.mat.ctextures[idx-4] = ri.AddView(f.views[idx], f.sampler[idx])
				} else {
					f.mat.textures1[idx] = ri.AddView(f.views[idx], f.sampler[idx])
				}
			}
		}
	}
	rc, ok := phase.(RenderColor)
	if ok {
		if f.transparent {
			center := rc.ViewTransform.Mul4(f.mat.world).Mul4x1(mgl32.Vec4{0, 0, 0, 1})
			rc.RenderTransparent(center[2], func(dl *vk.DrawList, pass *vk.GeneralRenderPass) {
				f.renderTransparent(dl, pass, fi, rc.DSFrame, *rc.Probe, rc.Shaders)
			})
			return
		}

		if len(f.colorShader) == 0 {
			return
		}
		pl := rc.Pass.Get(vk.NewHashKey(f.colorShader), func() interface{} {
			return f.buildPass(fi, rc.Pass, f.colorShader, false, false, rc.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rc.DL
		f.mat.probe = *rc.Probe
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = f.mat
		dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
			AddDescriptors(rc.DSFrame).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rd, ok := phase.(RenderDepth)
	if ok {
		if len(f.depthShader) == 0 {
			return
		}
		pl := rd.Pass.Get(vk.NewHashKey(f.depthShader), func() interface{} {
			return f.buildDepthPass(fi, rd.Pass, f.depthShader, false, rd.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rd.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = f.mat
		dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
			AddDescriptors(rd.DSFrame).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rs, ok := phase.(RenderShadow)
	if ok {
		if len(f.shadowShader) == 0 {
			return
		}
		pl := rs.Pass.Get(vk.NewHashKey(f.shadowShader), func() interface{} {
			return f.buildShadowPass(fi, rs.Pass, f.shadowShader, false, rs.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rs.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = f.mat
		dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
			AddDescriptors(rs.DSFrame, rs.DSShadowFrame).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rPick, ok := phase.(RenderPick)
	if ok {
		if len(f.pickShader) == 0 {
			return
		}
		pl := rPick.Pass.Get(vk.NewHashKey(f.pickShader), func() interface{} {
			return f.buildPickPass(fi, rPick.Pass, f.pickShader, false, rPick.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rPick.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = f.mat
		dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
			AddDescriptors(rPick.DSFrame, rPick.DSPick).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rp, ok := phase.(RenderProbe)
	if ok {
		if len(f.probeShader) > 0 {
			return
		}
		pl := rp.Pass.Get(vk.NewHashKey(f.probeShader), func() interface{} {
			return f.buildProbePass(fi, rp.Pass, f.probeShader, false, rp.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rp.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = f.mat
		dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
			AddDescriptors(rp.DSFrame, rp.DSProbeFrame).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
}

func (am *AnimatedMesh) Render(fi *vk.FrameInstance, phase Phase) {
	ri, ok := phase.(UpdateFrame)
	if ok {
		for idx := 0; idx < 4; idx++ {
			if am.views[idx] != nil {
				if idx >= 4 {
					am.mat.ctextures[idx-4] = ri.AddView(am.views[idx], am.sampler[idx])
				} else {
					am.mat.textures1[idx] = ri.AddView(am.views[idx], am.sampler[idx])
				}

			}
		}
		am.dsJoints = fi.AllocDescriptor(GetJointsLayout(fi.Device()))
		am.ubfJoints = fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, 64*uint64(len(am.mxAnims)))
		pMat := unsafe.Slice((*mgl32.Mat4)(unsafe.Pointer(&am.ubfJoints.Bytes()[0])), len(am.mxAnims))
		copy(pMat, am.mxAnims)
		am.dsJoints.WriteSlice(0, 0, am.ubfJoints)
	}
	rc, ok := phase.(RenderColor)
	if ok {
		if len(am.colorShader) == 0 {
			return
		}
		pl := rc.Pass.Get(vk.NewHashKey(am.colorShader), func() interface{} {
			return am.buildPass(fi, rc.Pass, am.colorShader, true, false, rc.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rc.DL
		am.mat.probe = *rc.Probe
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = am.mat
		dl.DrawIndexed(pl, am.mesh.From, am.mesh.Count).AddInputs(am.mesh.Model.VertexBuffers(am.mesh.Kind)...).
			AddDescriptors(rc.DSFrame, am.dsJoints).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rd, ok := phase.(RenderDepth)
	if ok {
		if len(am.depthShader) == 0 {
			return
		}
		pl := rd.Pass.Get(vk.NewHashKey(am.depthShader), func() interface{} {
			return am.buildDepthPass(fi, rd.Pass, am.depthShader, true, rd.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rd.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = am.mat
		dl.DrawIndexed(pl, am.mesh.From, am.mesh.Count).AddInputs(am.mesh.Model.VertexBuffers(am.mesh.Kind)...).
			AddDescriptors(rd.DSFrame, am.dsJoints).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rPick, ok := phase.(RenderPick)
	if ok {
		if len(am.pickShader) == 0 {
			return
		}
		pl := rPick.Pass.Get(vk.NewHashKey(am.pickShader), func() interface{} {
			return am.buildPickPass(fi, rPick.Pass, am.pickShader, true, rPick.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rPick.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = am.mat
		dl.DrawIndexed(pl, am.mesh.From, am.mesh.Count).AddInputs(am.mesh.Model.VertexBuffers(am.mesh.Kind)...).
			AddDescriptors(rPick.DSFrame, rPick.DSPick, am.dsJoints).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rs, ok := phase.(RenderShadow)
	if ok {
		if len(am.shadowShader) == 0 {
			return
		}
		pl := rs.Pass.Get(vk.NewHashKey(am.shadowShader), func() interface{} {
			return am.buildShadowPass(fi, rs.Pass, am.shadowShader, true, rs.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rs.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = am.mat
		dl.DrawIndexed(pl, am.mesh.From, am.mesh.Count).AddInputs(am.mesh.Model.VertexBuffers(am.mesh.Kind)...).
			AddDescriptors(rs.DSFrame, rs.DSShadowFrame, am.dsJoints).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
	rp, ok := phase.(RenderProbe)
	if ok {
		if len(am.probeShader) == 0 {
			return
		}
		pl := rp.Pass.Get(vk.NewHashKey(am.probeShader), func() interface{} {
			return am.buildProbePass(fi, rp.Pass, am.probeShader, true, rp.Shaders)
		}).(*vk.GraphicsPipeline)
		dl := rp.DL
		ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
		*(*material)(ptr) = am.mat
		dl.DrawIndexed(pl, am.mesh.From, am.mesh.Count).AddInputs(am.mesh.Model.VertexBuffers(am.mesh.Kind)...).
			AddDescriptors(rp.DSFrame, rp.DSProbeFrame, am.dsJoints).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
	}
}

func (f *FrozenMesh) buildPass(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, name string, animated bool, transparent bool, sp *shaders.Pack) *vk.GraphicsPipeline {
	dev := fi.Device()
	pl := vk.NewGraphicsPipeline(dev)
	pl.AddPushConstants(vk.SHADERStageAll, uint32(unsafe.Sizeof(material{})))
	pl.AddLayout(GetFrameLayout(dev))
	if animated {
		pl.AddLayout(GetJointsLayout(dev))
	}
	if transparent {
		pl.AddAlphaBlend()
	} else {
		pl.AddDepth(false, true)
	}
	code := sp.MustGet(dev, name)
	vmodel.AddInput(pl, f.mesh.Kind)

	pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
	pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
	pl.Create(pass)
	return pl
}

func (f *FrozenMesh) buildDepthPass(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, name string, animated bool, sp *shaders.Pack) *vk.GraphicsPipeline {
	dev := fi.Device()
	pl := vk.NewGraphicsPipeline(dev)
	pl.AddPushConstants(vk.SHADERStageAll, uint32(unsafe.Sizeof(material{})))
	pl.AddLayout(GetFrameLayout(dev))
	if animated {
		pl.AddLayout(GetJointsLayout(dev))
	}
	pl.AddDepth(true, true)
	vmodel.AddInput(pl, f.mesh.Kind)
	code := sp.MustGet(dev, name)
	pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
	pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
	pl.Create(pass)
	return pl
}

func (f *FrozenMesh) buildShadowPass(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, name string, animated bool, sp *shaders.Pack) *vk.GraphicsPipeline {
	dev := fi.Device()
	pl := vk.NewGraphicsPipeline(dev)
	pl.AddPushConstants(vk.SHADERStageAll, uint32(unsafe.Sizeof(material{})))
	pl.AddLayout(GetFrameLayout(dev))
	pl.AddLayout(GetShadowFrameLayout(dev))
	if animated {
		pl.AddLayout(GetJointsLayout(dev))
	}
	pl.AddDepth(true, true)
	vmodel.AddInput(pl, f.mesh.Kind)
	code := sp.MustGet(dev, name)
	pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
	pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
	pl.Create(pass)
	return pl
}

func (f *FrozenMesh) buildPickPass(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, name string, animated bool, sp *shaders.Pack) interface{} {
	dev := fi.Device()
	pl := vk.NewGraphicsPipeline(dev)
	pl.AddPushConstants(vk.SHADERStageAll, uint32(unsafe.Sizeof(material{})))
	pl.AddLayout(GetFrameLayout(dev))
	pl.AddLayout(GetPickFrameLayout(dev))
	if animated {
		pl.AddLayout(GetJointsLayout(dev))
	}
	pl.AddDepth(false, false)
	vmodel.AddInput(pl, f.mesh.Kind)
	code := sp.MustGet(dev, name)
	pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
	pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
	pl.Create(pass)
	return pl

}

func (f *FrozenMesh) buildProbePass(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, name string, animated bool, sp *shaders.Pack) *vk.GraphicsPipeline {
	dev := fi.Device()
	pl := vk.NewGraphicsPipeline(dev)
	pl.AddPushConstants(vk.SHADERStageAll, uint32(unsafe.Sizeof(material{})))
	pl.AddLayout(GetFrameLayout(dev))
	pl.AddLayout(GetShadowFrameLayout(dev))
	if animated {
		pl.AddLayout(GetJointsLayout(dev))
	}
	pl.AddDepth(true, true)
	vmodel.AddInput(pl, f.mesh.Kind)
	code := sp.MustGet(dev, name)
	pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
	pl.AddShader(vk.SHADERStageGeometryBit, code.Geometry)
	pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
	pl.Create(pass)
	return pl
}

func (fm *FrozenMesh) fillProps(mesh vmodel.Mesh, props vmodel.MaterialProperties) {
	fm.mat.albedo = props.GetColor(vmodel.CAlbedo, mgl32.Vec4{0, 0, 0, 1})
	fm.mat.emissive = props.GetColor(vmodel.CEmissive, mgl32.Vec4{0, 0, 0, 0})
	fm.mat.metalRoughess = mgl32.Vec4{props.GetFactor(vmodel.FMetalness, 0), props.GetFactor(vmodel.FRoughness, 1)}
	fm.mat.meshID = props.GetUInt(vmodel.UMeshID, 0)
	txIdx := props.GetImage(vmodel.TxAlbedo)
	if txIdx != 0 {
		fm.views[0], fm.sampler[0] = mesh.Model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxEmissive)
	if txIdx != 0 {
		fm.views[1], fm.sampler[1] = mesh.Model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxMetallicRoughness)
	if txIdx != 0 {
		fm.views[2], fm.sampler[2] = mesh.Model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxBump)
	if txIdx != 0 {
		fm.views[3], fm.sampler[3] = mesh.Model.GetImageView(txIdx)
	}
	for idx := vmodel.Property(0); idx < 4; idx++ {
		txIdx = props.GetImage(vmodel.TxCustom1 + idx)
		if txIdx != 0 {
			fm.views[4+idx], fm.sampler[4+idx] = mesh.Model.GetImageView(txIdx)
		}
	}
}

func (fm *FrozenMesh) fillShaders(mesh vmodel.Mesh, props vmodel.MaterialProperties) {

	fm.colorShader = "mesh_color"
	if mesh.Kind == vmodel.MESHKindSkinned {
		fm.colorShader = "mesh_color_skinned"
	}
	if fm.mat.textures1[3] != 0 {
		fm.colorShader += "_normap"
	}

	fm.shadowShader = "shadow_mesh"
	if mesh.Kind == vmodel.MESHKindSkinned {
		fm.shadowShader = "shadow_mesh_skinned"
	}

	fm.depthShader = "mesh_depth"
	if mesh.Kind == vmodel.MESHKindSkinned {
		fm.depthShader = "mesh_depth_skinned"
	}
	fm.probeShader = "probe_mesh"
	if mesh.Kind == vmodel.MESHKindSkinned {
		fm.probeShader = "probe_mesh_skinned"
	}
	fm.pickShader = "pick_mesh"
	if mesh.Kind == vmodel.MESHKindSkinned {
		fm.pickShader = "pick_mesh_skinned"
	}
}

func (am *AnimatedMesh) fillShaders(mesh vmodel.Mesh, props vmodel.MaterialProperties) {

	am.colorShader = "mesh_color_animated"
	if am.mat.textures1[3] != 0 {
		am.colorShader += "_normap"
	}

	am.shadowShader = "shadow_mesh_animated"
	am.depthShader = "mesh_depth_animated"
	am.probeShader = "probe_mesh_animated"
	am.pickShader = "pick_mesh_animated"
}

func (f *FrozenMesh) renderTransparent(dl *vk.DrawList, pass *vk.GeneralRenderPass, fi *vk.FrameInstance,
	dsFrame *vk.DescriptorSet, probe uint32, shaders *shaders.Pack) {
	if len(f.colorShader) == 0 {
		return
	}
	pl := pass.Get(vk.NewHashKey(f.colorShader), func() interface{} {
		return f.buildPass(fi, pass, f.colorShader, false, true, shaders)
	}).(*vk.GraphicsPipeline)
	f.mat.probe = probe
	ptr, offset := dl.AllocPushConstants(uint32(unsafe.Sizeof(material{})))
	*(*material)(ptr) = f.mat
	dl.DrawIndexed(pl, f.mesh.From, f.mesh.Count).AddInputs(f.mesh.Model.VertexBuffers(f.mesh.Kind)...).
		AddDescriptors(dsFrame).AddPushConstants(uint32(unsafe.Sizeof(material{})), offset)
}

type DrawOption interface {
	apply(fm *FrozenMesh)
	applyAnimated(am *AnimatedMesh)
}

type Transparent struct {
}

func (t Transparent) apply(fm *FrozenMesh) {
	fm.transparent = true
	fm.colorShader = "tr_" + fm.colorShader
	fm.depthShader, fm.shadowShader, fm.probeShader = "", "", ""
}

func (t Transparent) applyAnimated(am *AnimatedMesh) {
	am.transparent = true
}

type ColorShader struct {
	Shader string
}

func (c ColorShader) apply(fm *FrozenMesh) {
	fm.colorShader = c.Shader
}

func (c ColorShader) applyAnimated(am *AnimatedMesh) {
	am.colorShader = c.Shader
}

func DrawMesh(fl *FreezeList, mesh vmodel.Mesh, world mgl32.Mat4, props vmodel.MaterialProperties, options ...DrawOption) FrozenID {
	fm := &FrozenMesh{mesh: mesh}
	fm.mat.world = world
	fm.fillProps(mesh, props)
	fm.fillShaders(mesh, props)
	for _, o := range options {
		o.apply(fm)
	}
	id := fl.Add(fm)
	return id
}

func DrawAnimated(fl *FreezeList, mesh vmodel.Mesh, sk *vmodel.Skin, animation vmodel.Animation,
	animTime float64, world mgl32.Mat4, props vmodel.MaterialProperties, options ...DrawOption) FrozenID {
	am := &AnimatedMesh{FrozenMesh: FrozenMesh{mesh: mesh}}
	am.mat.world = world
	am.fillProps(mesh, props)
	am.fillShaders(mesh, props)
	am.mxAnims = make([]mgl32.Mat4, len(sk.Joints))
	am.joints = make([]vmodel.Joint, len(sk.Joints))
	for idx, j := range sk.Joints {
		am.joints[idx] = j
	}
	for _, ch := range animation.Channels {
		am.applyChannel(ch, animTime)
	}
	for jIdx, j := range am.joints {
		if j.Root {
			am.recalcJoint(jIdx, j, mgl32.Ident4())
		}
	}
	for _, o := range options {
		o.applyAnimated(am)
	}
	id := fl.Add(am)
	return id
}

func (am *AnimatedMesh) recalcJoint(jNro int, joint vmodel.Joint, local mgl32.Mat4) {
	local = local.Mul4(mgl32.Translate3D(joint.Translate[0], joint.Translate[1], joint.Translate[2]))
	local = local.Mul4(joint.Rotate.Mat4())
	local = local.Mul4(mgl32.Scale3D(joint.Scale[0], joint.Scale[1], joint.Scale[2]))
	am.mxAnims[jNro] = local.Mul4(joint.InverseMatrix)
	for _, chJ := range joint.Children {
		am.recalcJoint(chJ, am.joints[chJ], local)
	}
}

func (am *AnimatedMesh) applyChannel(ch vmodel.Channel, animTime float64) {
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
		am.translate(ch, lin, idx)
	case vmodel.TRotation:
		am.rotate(ch, lin, idx)
	}
}

func (am *AnimatedMesh) translate(ch vmodel.Channel, f float32, idx int) {
	min := mgl32.Vec3{ch.Output[idx*3], ch.Output[idx*3+1], ch.Output[idx*3+2]}
	max := mgl32.Vec3{ch.Output[idx*3+3], ch.Output[idx*3+4], ch.Output[idx*3+5]}
	am.joints[ch.Joint].Translate = min.Mul(1 - f).Add(max.Mul(f))
}

func (am *AnimatedMesh) rotate(ch vmodel.Channel, f float32, idx int) {
	qMin := mgl32.Quat{V: mgl32.Vec3{ch.Output[idx*4], ch.Output[idx*4+1], ch.Output[idx*4+2]},
		W: ch.Output[idx*4+3]}
	qMax := mgl32.Quat{V: mgl32.Vec3{ch.Output[idx*4+4], ch.Output[idx*4+5], ch.Output[idx*4+6]},
		W: ch.Output[idx*4+7]}
	am.joints[ch.Joint].Rotate = mgl32.QuatLerp(qMin, qMax, f)
}

func DrawNodes(fl *FreezeList, model *vmodel.Model, node vmodel.NodeIndex, world mgl32.Mat4) map[string]FrozenID {
	n := model.GetNode(node)
	result := make(map[string]FrozenID)
	n.Enum(world, func(local mgl32.Mat4, n vmodel.Node) {
		if n.Mesh >= 0 {
			result[n.Name] = DrawMesh(fl, model.GetMesh(n.Mesh), local, model.GetMaterial(n.Material).Props)
		}
	})
	return result
}
