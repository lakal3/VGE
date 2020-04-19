//

//go:generate glslangValidator.exe -V shadow.vert.glsl -o shadow.vert.spv
//go:generate glslangValidator.exe -V -DSKINNED=1 shadow.vert.glsl -o shadow.vert_skin.spv
//go:generate glslangValidator.exe -V shadow.geom.glsl -o shadow.geom.spv
//go:generate glslangValidator.exe -V shadow.frag.glsl -o shadow.frag.spv
//go:generate packspv -p shadow .
package shadow

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"math"
	"unsafe"
)

const ShadowFormat = vk.FORMATD32Sfloat

type PointLight struct {
	// Maximum distance we can see lights shadows from. If light is longer that this distance away for camera,
	// we just turn shadows off
	MaxShadowDistance float32
	vscene.PointLight
	key     vk.Key
	mapSize uint32
}

type shadowResources struct {
	pool        *vk.MemoryPool // Pool for shadow maps
	shadowImage *vk.Image
	fb          *vk.Framebuffer
	cubeView    *vk.ImageView
	cmd         *vk.Command
}

const maxInstances = 100

type shadowInstance struct {
	projection mgl32.Mat4
	view       [6]mgl32.Mat4
	lightPos   mgl32.Vec4
	instances  [maxInstances]mgl32.Mat4
}

type shadowPass struct {
	ctx         vk.APIContext
	cmd         *vk.Command
	maxDistance float32
	pos         mgl32.Vec3
	dl          *vk.DrawList
	pl          *vk.GraphicsPipeline
	plSkin      *vk.GraphicsPipeline
	rc          *vk.RenderCache
	ds          *vk.DescriptorSet
	sl          *vk.Slice
	si          *shadowInstance
	siCount     int
}

// Objects under this node will not cast shadow!
type NoShadow struct {
}

func (n NoShadow) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(*shadowPass)
	if ok {
		pi.Visible = false
	}
}

func (s *shadowPass) Begin() (atEnd func()) {
	return nil
}

func (s *shadowPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex) {
	uc := vscene.GetUniformCache(s.rc)
	if s.ds == nil {
		s.ds, s.sl = uc.Alloc(s.ctx)
		s.si = &shadowInstance{}
		s.si.projection = mgl32.Perspective(math.Pi/2, 1, s.maxDistance/2000, s.maxDistance*2)
		pos := s.pos
		s.si.view[0] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{1, 0, 0}), mgl32.Vec3{0, -1, 0})
		s.si.view[1] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{-1, 0, 0}), mgl32.Vec3{0, -1, 0})
		s.si.view[2] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 1, 0}), mgl32.Vec3{0, 0, 1})
		s.si.view[3] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, -1, 0}), mgl32.Vec3{0, 0, -1})
		s.si.view[4] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, 1}), mgl32.Vec3{0, -1, 0})
		s.si.view[5] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, -1}), mgl32.Vec3{0, -1, 0})
		s.si.lightPos = pos.Vec4(s.maxDistance)

	}
	s.si.instances[s.siCount] = world
	s.dl.DrawIndexed(s.pl, mesh.From, mesh.Count).AddDescriptors(s.ds).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *shadowPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex, aniMatrix []mgl32.Mat4) {
	uc := vscene.GetUniformCache(s.rc)
	if s.ds == nil {
		s.ds, s.sl = uc.Alloc(s.ctx)
		s.si = &shadowInstance{}
		s.si.projection = mgl32.Perspective(math.Pi/2, 1, s.maxDistance/2000, s.maxDistance*2)
		pos := s.pos
		s.si.view[0] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{1, 0, 0}), mgl32.Vec3{0, -1, 0})
		s.si.view[1] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{-1, 0, 0}), mgl32.Vec3{0, -1, 0})
		s.si.view[2] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 1, 0}), mgl32.Vec3{0, 0, 1})
		s.si.view[3] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, -1, 0}), mgl32.Vec3{0, 0, -1})
		s.si.view[4] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, 1}), mgl32.Vec3{0, -1, 0})
		s.si.view[5] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, -1}), mgl32.Vec3{0, -1, 0})
		s.si.lightPos = pos.Vec4(s.maxDistance)

	}
	s.si.instances[s.siCount] = world
	dsMesh, slMesh := uc.Alloc(s.ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	s.dl.DrawIndexed(s.plSkin, mesh.From, mesh.Count).AddDescriptors(s.ds, dsMesh).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *shadowPass) flush() {
	if s.siCount > 0 {
		b := *(*[unsafe.Sizeof(shadowInstance{})]byte)(unsafe.Pointer(s.si))
		copy(s.sl.Content, b[:])
		s.cmd.Draw(s.dl)
		s.dl = &vk.DrawList{}
	}
	s.si, s.ds, s.sl, s.siCount = nil, nil, nil, 0
}

func (s *shadowResources) Dispose() {
	if s.pool != nil {
		s.fb.Dispose()
		s.shadowImage.Dispose()
		s.pool.Dispose()
		s.pool, s.fb, s.shadowImage = nil, nil, nil
	}
	if s.cmd != nil {
		s.cmd.Dispose()
		s.cmd = nil
	}
}

var kDepthPass = vk.NewKey()
var kDepthPipeline = vk.NewKey()
var kSkinnedDepthPipeline = vk.NewKey()
var kDepthInstance = vk.NewKey()
var kShadowSampler = vk.NewKey()

func NewPointLight(baseLight vscene.PointLight, mapSize uint32) *PointLight {
	return &PointLight{key: vk.NewKey(), PointLight: baseLight, mapSize: mapSize}
}

func (pl *PointLight) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if pl.MaxDistance == 0 {
			pl.MaxDistance = 10
		}
		if pl.MaxShadowDistance == 0 {
			pl.MaxShadowDistance = pl.MaxDistance * 2
		}
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		if pd.F.EyePos.Vec3().Sub(pos.Vec3()).Len() > pl.MaxShadowDistance {
			// Skip shadow pass for light too long aways
			pd.F.AddLight(vscene.Light{Intensity: pl.Intensity.Vec4(1),
				Position: pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)})
			return
		}
		sr := pl.renderShadowMap(pd, pi)
		sampler := getShadowSampler(pd.Cache.Ctx, pd.Cache.Device)
		idx := vscene.SetFrameImage(pd.Cache, sr.cubeView, sampler)
		if idx < 0 {
			pd.F.AddLight(vscene.Light{Intensity: pl.Intensity.Vec4(1),
				Position: pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)})
			return
		}
		pd.F.AddLight(vscene.Light{Intensity: pl.Intensity.Vec4(1),
			Direction: mgl32.Vec4{0, 0, 0, float32(idx)},
			Position:  pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)})
	}
}

func (pl *PointLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo) *shadowResources {
	rp := pd.Cache.Device.Get(pd.Cache.Ctx, kDepthPass, func(ctx vk.APIContext) interface{} {
		return vk.NewDepthRenderPass(ctx, pd.Cache.Device, vk.IMAGELayoutShaderReadOnlyOptimal, ShadowFormat)
	}).(*vk.DepthRenderPass)
	sr := pd.Cache.Get(pl.key, func(ctx vk.APIContext) interface{} {
		return pl.makeResources(ctx, pd.Cache.Device, rp)
	}).(*shadowResources)

	gpl := rp.Get(pd.Cache.Ctx, kDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeShadowPipeline(ctx, pd.Cache.Device, rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rp.Get(pd.Cache.Ctx, kSkinnedDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeSkinnedShadowPipeline(ctx, pd.Cache.Device, rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd
	cmd.Begin()
	cmd.BeginRenderPass(rp, sr.fb)
	sp := &shadowPass{ctx: pd.Cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: pd.Cache}
	sp.pos = pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1}).Vec3()
	pd.Scene.Process(pi.Time, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		cmd.Wait()
	})
	return sr
}

func (pl *PointLight) makeResources(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *shadowResources {
	sr := &shadowResources{}
	sr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Layers: 6, MipLevels: 1,
		Format: ShadowFormat, Depth: 1}
	sr.shadowImage = sr.pool.ReserveImage(ctx, desc, vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageDepthStencilAttachmentBit|
		vk.IMAGEUsageSampledBit)
	sr.pool.Allocate(ctx)
	sr.cubeView = sr.shadowImage.NewCubeView(ctx, 0)
	sr.fb = vk.NewFramebuffer(ctx, rp, []*vk.ImageView{sr.cubeView})
	sr.cmd = vk.NewCommand(ctx, dev, vk.QUEUEGraphicsBit, false)
	return sr
}

func (pl *PointLight) makeShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
	gp.AddDepth(ctx, true, true)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, shadow_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageGeometryBit, shadow_geom_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, shadow_frag_spv)
	gp.Create(ctx, rp)
	return gp
}

func (pl *PointLight) makeSkinnedShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
	gp.AddDepth(ctx, true, true)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, shadow_vert_skin_spv)
	gp.AddShader(ctx, vk.SHADERStageGeometryBit, shadow_geom_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, shadow_frag_spv)
	gp.Create(ctx, rp)
	return gp
}

func getShadowSampler(ctx vk.APIContext, dev *vk.Device) *vk.Sampler {
	sampler := dev.Get(ctx, kShadowSampler, func(ctx vk.APIContext) interface{} {
		return vk.NewSampler(ctx, dev, vk.SAMPLERAddressModeClampToEdge)
	}).(*vk.Sampler)
	return sampler
}
