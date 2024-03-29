//

package shadow

import (
	"math"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type CubePointLight struct {
	// Maximum distance we can see lights shadows from. If light is longer that this distance away for camera,
	// we just turn shadows off
	MaxShadowDistance float32
	vscene.PointLight
	key     vk.Key
	mapSize uint32
}

type cubeShadowResources struct {
	pool        *vk.MemoryPool // Pool for shadow maps
	shadowImage *vk.Image
	fb          *vk.Framebuffer
	cubeView    *vk.ImageView
	cmd         *vk.Command
	sampler     *vk.Sampler
}

type cubeShadowInstance struct {
	projection mgl32.Mat4
	view       [6]mgl32.Mat4
	lightPos   mgl32.Vec4
	instances  [maxInstances]mgl32.Mat4
}

type cubeShadowPass struct {
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
	si          *cubeShadowInstance
	siCount     int
	renderer    vmodel.Renderer
}

func (s *cubeShadowPass) GetRenderer() vmodel.Renderer {
	return s.renderer
}

func (s *cubeShadowPass) ViewProjection() (projection, view mgl32.Mat4) {
	return mgl32.Perspective(math.Pi/2, 1, s.maxDistance/2000, s.maxDistance*2),
		mgl32.LookAtV(s.pos, s.pos.Add(mgl32.Vec3{1, 0, 0}), mgl32.Vec3{0, -1, 0})
}

func (s *cubeShadowPass) BindFrame() *vk.DescriptorSet {
	uc := vscene.GetUniformCache(s.rc)
	if s.ds == nil {
		s.ds, s.sl = uc.Alloc(s.ctx)
		s.si = &cubeShadowInstance{}
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
	return s.ds
}

func (s *cubeShadowPass) GetCache() *vk.RenderCache {
	return s.rc
}

func (s *cubeShadowPass) Begin() (atEnd func()) {
	return nil
}

func (s *cubeShadowPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader) {
	s.BindFrame()
	s.si.instances[s.siCount] = world
	s.dl.DrawIndexed(s.pl, mesh.From, mesh.Count).AddDescriptors(s.ds).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *cubeShadowPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader, aniMatrix []mgl32.Mat4) {
	s.BindFrame()
	uc := vscene.GetUniformCache(s.rc)
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

func (s *cubeShadowPass) flush() {
	if s.siCount > 0 {
		b := *(*[unsafe.Sizeof(cubeShadowInstance{})]byte)(unsafe.Pointer(s.si))
		copy(s.sl.Content, b[:])
		s.cmd.Draw(s.dl)
		s.dl = &vk.DrawList{}
	}
	s.si, s.ds, s.sl, s.siCount = nil, nil, nil, 0
}

func (s *cubeShadowResources) Dispose() {
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

// NewCubePointLight creates a shadow casting point light that uses cube map to draw shadows
// New point light uses a parabolic shadow map that reduces rendered sides from 6 to 2. It is recommeded to use s
func NewCubePointLight(baseLight vscene.PointLight, mapSize uint32) *CubePointLight {
	return &CubePointLight{key: vk.NewKey(), PointLight: baseLight, mapSize: mapSize}
}

func (pl *CubePointLight) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if pl.MaxDistance == 0 {
			pl.MaxDistance = 10
		}
		if pl.MaxShadowDistance == 0 {
			pl.MaxShadowDistance = pl.MaxDistance * 2
		}
		_, ok := pi.Frame.(vscene.ImageFrame)
		if !ok {
			return
		}
		eyePos, _ := vscene.GetEyePosition(pi.Frame)
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		if eyePos.Sub(pos.Vec3()).Len() > pl.MaxShadowDistance {
			// Skip shadow pass for light too long aways
			return
		}
		pl.renderShadowMap(pd, pi)
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		hasShadowmap := pi.Frame.GetCache().GetPerFrame(pl.key, func(ctx vk.APIContext) interface{} {
			return false
		}).(bool)
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := vscene.Light{Intensity: pl.Intensity.Vec4(1),
			Position: pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)}
		var imIndex vmodel.ImageIndex
		if ok && hasShadowmap {
			sr := pi.Frame.GetCache().Get(pl.key, func(ctx vk.APIContext) interface{} {
				return nil
			}).(*cubeShadowResources)
			imIndex = imFrame.AddFrameImage(sr.cubeView, sr.sampler)
		}
		if imIndex > 0 {
			l.ShadowMapMethod = 1
			l.ShadowMapIndex = float32(imIndex)
		}
		lp.AddLight(l, lp)
	}
}

func (pl *CubePointLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo) *cubeShadowResources {
	cache := pi.Frame.GetCache()
	rp := cache.Device.Get(cache.Ctx, kDepthPass, func(ctx vk.APIContext) interface{} {
		return vk.NewDepthRenderPass(ctx, cache.Device, vk.IMAGELayoutShaderReadOnlyOptimal, ShadowFormat)
	}).(*vk.DepthRenderPass)
	sr := cache.Get(pl.key, func(ctx vk.APIContext) interface{} {
		return pl.makeResources(ctx, cache.Device, rp)
	}).(*cubeShadowResources)

	gpl := rp.Get(cache.Ctx, kDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeShadowPipeline(ctx, cache.Device, rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rp.Get(cache.Ctx, kSkinnedDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeSkinnedShadowPipeline(ctx, cache.Device, rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd
	cmd.Begin()
	cmd.BeginRenderPass(rp, sr.fb)
	sp := &cubeShadowPass{ctx: cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: cache, renderer: pi.Frame.GetRenderer()}
	sp.pos = pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1}).Vec3()
	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		cmd.Wait()
	})
	sr.sampler = getShadowSampler(cache.Ctx, cache.Device)
	cache.SetPerFrame(pl.key, true)
	return sr
}

func (pl *CubePointLight) makeResources(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *cubeShadowResources {
	sr := &cubeShadowResources{}
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

func (pl *CubePointLight) makeShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
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

func (pl *CubePointLight) makeSkinnedShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
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
