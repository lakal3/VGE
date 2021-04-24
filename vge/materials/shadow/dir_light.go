package shadow

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type DirectionalLight struct {
	// Maximum distance we can see lights shadows from. If light is longer that this distance away for camera,
	// we just turn shadows off
	vscene.DirectionalLight

	// MaxShadowDistance determines how large area shadow map will cover. Everything outside if will be fully lit
	MaxShadowDistance float32
	key               vk.Key
	mapSize           uint32
}

// SetMaxShadowDistance sets MaxShadowDistance
func (pl *DirectionalLight) SetMaxShadowDistance(maxDistance float32) *DirectionalLight {
	pl.MaxShadowDistance = maxDistance
	return pl
}

// NewDirectionalLight creates a new shadow casting direction light.
// Map size is size of shadow map used to sample lights distance to closest occluder
// Higher resolutions give better quality of shadows but greatly increases memory usage in GPU
func NewDirectionalLight(baseLight vscene.DirectionalLight, mapSize uint32) *DirectionalLight {
	return &DirectionalLight{key: vk.NewKey(), DirectionalLight: baseLight, mapSize: mapSize}
}

func (pl *DirectionalLight) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if pl.MaxShadowDistance == 0 {
			pl.MaxShadowDistance = 100
		}

		_, ok := pi.Frame.(vscene.ImageFrame)
		if !ok {
			return
		}
		eyePos, _ := vscene.GetEyePosition(pi.Frame)

		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)

		pl.renderShadowMap(pd, pi, rsr, eyePos)
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		eyePos, _ := vscene.GetEyePosition(pi.Frame)
		l := vscene.Light{Intensity: pl.Intensity.Vec4(1), Direction: pl.Direction.Vec4(0), Attenuation: mgl32.Vec4{1, 0, 0, pl.MaxShadowDistance}}
		l.Position = eyePos.Add(pl.Direction.Mul(-pl.MaxShadowDistance * 0.5)).Vec4(0)
		var imIndex vmodel.ImageIndex
		if ok && rsr.lastImage >= 0 {
			imIndex = imFrame.AddFrameImage(rsr.shadowImages[rsr.lastImage].DefaultView(pi.Frame.GetCache().Ctx), rsr.sampler)
		}
		if imIndex > 0 {
			l.ShadowMapMethod = 5
			l.ShadowPlane = QuoternionFromYUp(l.Direction.Vec3())
			l.ShadowMapIndex = float32(imIndex)
		}
		lp.AddLight(l, lp)
	}
}

type dirFrameResources struct {
	fb *vk.Framebuffer
}

func (f *dirFrameResources) Dispose() {
	if f.fb != nil {
		f.fb.Dispose()
		f.fb = nil
	}
}

type dirResources struct {
	cmd *vk.Command
}

type dirShadowPass struct {
	ctx         vk.APIContext
	cmd         *vk.Command
	maxDistance float32
	pos         mgl32.Vec3
	dl          *vk.DrawList
	pl          *vk.GraphicsPipeline
	plSkin      *vk.GraphicsPipeline
	rc          *vk.RenderCache
	dsFrame     *vk.DescriptorSet
	slFrame     *vk.Slice
	si          *shaderFrame
	siCount     int
	renderer    vmodel.Renderer
	dir         mgl32.Vec3
}

func (s *dirShadowPass) GetRenderer() vmodel.Renderer {
	return s.renderer
}

func (s *dirShadowPass) BindFrame() *vk.DescriptorSet {
	uc := vscene.GetUniformCache(s.rc)
	if s.dsFrame == nil {
		s.dsFrame, s.slFrame = uc.Alloc(s.ctx)
		s.si = &shaderFrame{plane: QuoternionFromYUp(s.dir), lightPos: s.pos.Vec4(1),
			maxShadow: s.maxDistance, minShadow: 0}

	}
	return s.dsFrame
}

func (s *dirShadowPass) GetCache() *vk.RenderCache {
	return s.rc
}

func (s *dirShadowPass) Begin() (atEnd func()) {
	return nil
}

func (s *dirShadowPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex) {
	s.BindFrame()
	s.si.instances[s.siCount] = world
	s.dl.DrawIndexed(s.pl, mesh.From, mesh.Count).AddDescriptors(s.dsFrame).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *dirShadowPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex, aniMatrix []mgl32.Mat4) {
	s.BindFrame()
	uc := vscene.GetUniformCache(s.rc)
	s.si.instances[s.siCount] = world
	dsMesh, slMesh := uc.Alloc(s.ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	s.dl.DrawIndexed(s.plSkin, mesh.From, mesh.Count).AddDescriptors(s.dsFrame, dsMesh).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *dirShadowPass) flush() {
	if s.siCount > 0 {
		b := *(*[unsafe.Sizeof(shaderFrame{})]byte)(unsafe.Pointer(s.si))
		copy(s.slFrame.Content, b[:])
		s.cmd.Draw(s.dl)
		s.dl = &vk.DrawList{}
	}
	s.si, s.dsFrame, s.slFrame, s.siCount = nil, nil, nil, 0
}

func (s *dirResources) Dispose() {
	if s.cmd != nil {
		s.cmd.Dispose()
		s.cmd = nil
	}
}

var kDirDepthPass = vk.NewKey()
var kDirDepthPipeline = vk.NewKey()
var kDirSkinnedDepthPipeline = vk.NewKey()

func (pl *DirectionalLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo, rsr *renderResources, eyePos mgl32.Vec3) {
	cache := pi.Frame.GetCache()
	sr := cache.Get(pl.key, func(ctx vk.APIContext) interface{} {
		return makeDirResources(ctx, cache.Device, rsr.rp)
	}).(*dirResources)
	gpl := rsr.rp.Get(cache.Ctx, kDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeShadowPipeline(ctx, cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rsr.rp.Get(cache.Ctx, kSkinnedDepthPipeline, func(ctx vk.APIContext) interface{} {
		return pl.makeSkinnedShadowPipeline(ctx, cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd
	cmd.Begin()
	imageIndex := rsr.lastImage + 1
	if imageIndex >= 2 {
		imageIndex = 0
	}
	fbs := cache.GetPerFrame(pl.key, func(ctx vk.APIContext) interface{} {
		return makeDirFrameResource(cache, rsr, imageIndex)
	}).(*dirFrameResources)
	fb := fbs.fb
	cmd.BeginRenderPass(rsr.rp, fb)
	sp := &dirShadowPass{ctx: cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxShadowDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: cache, renderer: pi.Frame.GetRenderer()}
	sp.dir = pl.Direction
	sp.pos = eyePos.Add(pl.Direction.Mul(-pl.MaxShadowDistance * 0.5))

	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		// cmd.Wait()
	})
	rsr.lastImage = imageIndex
	return
}

func (pl *DirectionalLight) makeRenderPass(ctx vk.APIContext, dev *vk.Device) *vk.GeneralRenderPass {
	rp := dev.Get(ctx, kDepthPass, func(ctx vk.APIContext) interface{} {
		return vk.NewGeneralRenderPass(ctx, dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{Format: ShadowFormat, InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				ClearColor: [4]float32{1, 0, 0, 0}},
		})
	}).(*vk.GeneralRenderPass)
	return rp
}

func makeDirResources(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *dirResources {
	sr := &dirResources{}
	sr.cmd = vk.NewCommand(ctx, dev, vk.QUEUEGraphicsBit, false)
	return sr
}

func (pl *DirectionalLight) makeShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, dir_shadow_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func (pl *DirectionalLight) makeSkinnedShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, dir_shadow_vert_skin_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func (pl *DirectionalLight) makeRenderResources(ctx vk.APIContext, dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = pl.makeRenderPass(ctx, dev)
	rsr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Depth: 1, Layers: 1, MipLevels: 1,
		Format: vk.FORMATD32Sfloat}
	for idx := 0; idx < 2; idx++ {
		rsr.shadowImages = append(rsr.shadowImages, rsr.pool.ReserveImage(ctx, desc, vk.IMAGEUsageDepthStencilAttachmentBit|
			vk.IMAGEUsageSampledBit))
	}
	rsr.pool.Allocate(ctx)
	rsr.sampler = vmodel.GetDefaultSampler(ctx, dev)
	for idx := 0; idx < 2; idx++ {
		rg := vk.ImageRange{FirstLayer: 0, LayerCount: 1, LevelCount: 1}
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(ctx, rsr.shadowImages[idx], &rg))
	}
	return &rsr
}

func makeDirFrameResource(cache *vk.RenderCache, rsr *renderResources, imageIndex int) *dirFrameResources {
	fr := dirFrameResources{}
	fr.fb = vk.NewFramebuffer(cache.Ctx, rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex]})
	return &fr
}
