package shadow

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

// AlphaTexture materials have a albedo texture with holes. Holes are marked with alpha < cutoff
// If cutoff is 0, material is opaque event if it support this interface. View and sampler can be null when cutoff is 0
type AlphaTexture interface {
	vmodel.Shader
	GetAlphaTexture() (cutoff float32, view *vk.ImageView, sampler *vk.Sampler)
}

type DirectionalLight struct {
	// Maximum distance we can see lights shadows from. If light is longer that this distance away for camera,
	// we just turn shadows off
	vscene.DirectionalLight

	// MaxShadowDistance determines how large area shadow map will cover. Everything outside if will be fully lit.
	// Default size is 100 but should match to size of your scene.
	// Currently VGE dont support multiple resolutions for depth map, but this may change in future.
	MaxShadowDistance float32

	// CenterPoint tells where to place light shadow map. Even though directional lights don't have any place we must
	// place shadowmap somewhere and create an orthographic projection from that point to lights direction. CenterPoint and MaxShadowDistance
	// will define a volume that will be contained in shadowmap. For directional lights all points outside shadowmap are considered visible.
	//
	// Centerpoint is percentage of MaxShadowDistance we move lights away from camera to view direction. Typically values
	// are around 0.5 - 0.9 (50% - 90%)
	CenterPoint float32

	// Number of frames to keep same shadow map
	UpdateDelay int

	key     vk.Key
	mapSize uint32
}

// SetMaxShadowDistance sets MaxShadowDistance
func (pl *DirectionalLight) SetMaxShadowDistance(maxDistance float32) *DirectionalLight {
	pl.MaxShadowDistance = maxDistance
	return pl
}

// SetCenterPoint sets CenterPoint
func (pl *DirectionalLight) SetCenterPoint(centerPoint float32) *DirectionalLight {
	pl.CenterPoint = centerPoint
	return pl
}

// SetUpdateDelay set delay between shadowmap updates. 0 - each frame, 1 - every second frame, 2 - every third frame...
func (pl *DirectionalLight) SetUpdateDelay(delayFrames int) *DirectionalLight {
	pl.UpdateDelay = delayFrames
	return pl
}

// NewDirectionalLight creates a new shadow casting direction light.
// Map size is size of shadow map used to sample lights distance to closest occluder
// Higher resolutions give better quality of shadows but greatly increases memory usage in GPU
func NewDirectionalLight(baseLight vscene.DirectionalLight, mapSize uint32) *DirectionalLight {
	return &DirectionalLight{key: vk.NewKey(), DirectionalLight: baseLight, mapSize: mapSize, MaxShadowDistance: 100, CenterPoint: 0.5}
}

func (pl *DirectionalLight) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if pl.MaxShadowDistance == 0 {
			return
		}

		_, ok := pi.Frame.(vscene.ImageFrame)
		if !ok {
			return
		}
		eyePos, _ := vscene.GetEyePosition(pi.Frame)

		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)
		if rsr.updateCount > 0 {
			rsr.updateCount--
		} else {
			pl.renderShadowMap(pd, pi, rsr, eyePos)
		}
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := vscene.Light{Intensity: pl.Intensity.Vec4(1), Direction: pl.Direction.Vec4(0), Attenuation: mgl32.Vec4{1, 0, 0, pl.MaxShadowDistance}}
		l.Position = pl.GetShadowmapPos(pi.Frame)
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

type shadowPass struct {
	ctx         vk.APIContext
	cmd         *vk.Command
	rc          *vk.RenderCache
	dl          *vk.DrawList
	updated     bool
	dsFrame     *vk.DescriptorSet
	slFrame     *vk.Slice
	siFrame     *shaderFrame
	siInstance  *shaderInstances
	siCount     int
	imCount     uint32
	renderer    vmodel.Renderer
	maxDistance float32
	pos         mgl32.Vec4
	dir         mgl32.Vec3
	yFactor     float32
	pl          *vk.GraphicsPipeline
	plSkin      *vk.GraphicsPipeline
	sampler     *vk.Sampler
	dsInst      *vk.DescriptorSet
	slInst      *vk.Slice
	imMap       map[uintptr]uint32
}

func (s *shadowPass) GetRenderer() vmodel.Renderer {
	return s.renderer
}

func (s *shadowPass) BindFrame() *vk.DescriptorSet {
	if s.siFrame == nil {
		s.siFrame = &shaderFrame{lightPos: s.pos,
			maxShadow: s.maxDistance, minShadow: 0}
		s.imCount = 1
		if s.yFactor != 0 {
			s.siFrame.yFactor = s.yFactor
		} else {
			s.siFrame.plane = QuoternionFromYUp(s.dir)
		}
		b := *(*[unsafe.Sizeof(shaderFrame{})]byte)(unsafe.Pointer(s.siFrame))
		copy(s.slFrame.Content, b[:])
	}
	if s.siInstance == nil {
		uc := vscene.GetUniformCache(s.rc)
		s.siInstance = &shaderInstances{}
		s.dsInst, s.slInst = uc.Alloc(s.rc.Ctx)
	}
	return s.dsFrame
}

func (s *shadowPass) GetCache() *vk.RenderCache {
	return s.rc
}

func (s *shadowPass) Begin() (atEnd func()) {
	return nil
}

func (s *shadowPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader) {
	_ = s.BindFrame()
	s.siInstance.instances[s.siCount] = s.makeInstance(world, mesh, material)
	s.dl.DrawIndexed(s.pl, mesh.From, mesh.Count).AddDescriptors(s.dsFrame, s.dsInst).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *shadowPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, material vmodel.Shader, aniMatrix []mgl32.Mat4) {
	_ = s.BindFrame()
	uc := vscene.GetUniformCache(s.rc)
	s.siInstance.instances[s.siCount] = s.makeInstance(world, mesh, material)
	dsMesh, slMesh := uc.Alloc(s.ctx)
	copy(slMesh.Content, vscene.Mat4ToBytes(aniMatrix))
	s.dl.DrawIndexed(s.plSkin, mesh.From, mesh.Count).AddDescriptors(s.dsFrame, s.dsInst, dsMesh).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindSkinned)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *shadowPass) makeInstance(world mgl32.Mat4, mesh vmodel.Mesh, material vmodel.Shader) shaderInstance {
	dsFrame := s.BindFrame()
	inst := shaderInstance{world: world}
	if vscene.FrameMaxDynamicSamplers == 0 {
		return inst
	}
	at, ok := material.(AlphaTexture)
	if !ok {
		return inst
	}
	cutoff, view, sampler := at.GetAlphaTexture()
	if cutoff <= 0 {
		return inst
	}
	if s.imCount >= vscene.FrameMaxDynamicSamplers {
		return inst
	}
	if s.imMap == nil {
		s.imMap = make(map[uintptr]uint32)
	}
	imIndex, ok := s.imMap[view.Handle()]
	inst.alphaCutoff = cutoff
	if ok {
		inst.tx_albedo = float32(imIndex)
		return inst
	}
	dsFrame.WriteImage(s.ctx, 1, s.imCount, view, sampler)
	inst.tx_albedo = float32(s.imCount)
	s.imMap[view.Handle()] = s.imCount
	s.imCount++
	return inst
}

func (s *shadowPass) flush() {
	if s.siCount > 0 {
		b := *(*[unsafe.Sizeof(shaderInstances{})]byte)(unsafe.Pointer(s.siInstance))
		copy(s.slInst.Content, b[:])
		s.cmd.Draw(s.dl)
		s.dl = &vk.DrawList{}
	}
	s.siInstance, s.dsInst, s.slInst, s.siCount = nil, nil, nil, 0
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
	sp := &shadowPass{ctx: cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, rc: cache, renderer: pi.Frame.GetRenderer(),
		pl: gpl, plSkin: gSkinnedPl, maxDistance: pl.MaxShadowDistance, sampler: rsr.sampler,
		dsFrame: rsr.dsFrame[imageIndex], slFrame: rsr.slFrame[imageIndex]}
	sp.dir = pl.Direction
	sp.pos = pl.GetShadowmapPos(pi.Frame)

	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		// cmd.Wait()
	})
	rsr.lastImage, rsr.updateCount = imageIndex, pl.UpdateDelay
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
	gp.AddLayout(ctx, getShadowFrameLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, dir_shadow_vert_spv)
	if vscene.FrameMaxDynamicSamplers > 0 {
		gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_dyn_frag_spv)
	} else {
		gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	}
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func (pl *DirectionalLight) makeSkinnedShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
	gp.AddLayout(ctx, getShadowFrameLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, dir_shadow_vert_skin_spv)
	if vscene.FrameMaxDynamicSamplers > 0 {
		gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_dyn_frag_spv)

	} else {
		gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	}
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func getShadowFrameLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(ctx, kShadowLayout, func(ctx vk.APIContext) interface{} {
		la := vscene.GetUniformLayout(ctx, dev)
		if vscene.FrameMaxDynamicSamplers > 0 {
			return la.AddDynamicBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, vscene.FrameMaxDynamicSamplers,
				vk.DESCRIPTORBindingUpdateAfterBindBitExt)
		}
		return la
	}).(*vk.DescriptorLayout)
}

func (pl *DirectionalLight) makeRenderResources(ctx vk.APIContext, dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = makeRenderPass(ctx, dev)
	rsr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Depth: 1, Layers: 1, MipLevels: 1,
		Format: vk.FORMATD32Sfloat}
	var buffers []*vk.Buffer
	for idx := 0; idx < 2; idx++ {
		rsr.shadowImages = append(rsr.shadowImages, rsr.pool.ReserveImage(ctx, desc, vk.IMAGEUsageDepthStencilAttachmentBit|
			vk.IMAGEUsageSampledBit))
		buffers = append(buffers, rsr.pool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit))
	}
	rsr.dpFrame = vk.NewDescriptorPool(ctx, getShadowFrameLayout(ctx, dev), 2)

	rsr.pool.Allocate(ctx)
	rsr.sampler = vmodel.GetDefaultSampler(ctx, dev)
	for idx := 0; idx < 2; idx++ {
		rsr.dsFrame = append(rsr.dsFrame, rsr.dpFrame.Alloc(ctx))
		sl := buffers[idx].Slice(ctx, 0, vk.MinUniformBufferOffsetAlignment)
		rsr.dsFrame[idx].WriteSlice(ctx, 0, 0, sl)
		rsr.slFrame = append(rsr.slFrame, sl)
		rg := vk.ImageRange{FirstLayer: 0, LayerCount: 1, LevelCount: 1}
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(ctx, rsr.shadowImages[idx], &rg))
	}
	return &rsr
}

func (pl *DirectionalLight) GetShadowmapPos(f vmodel.Frame) mgl32.Vec4 {
	sf := vscene.GetSimpleFrame(f)
	if sf == nil {
		return mgl32.Vec4{}
	}
	v := sf.SSF.View.Inv().Mul4x1(mgl32.Vec4{0, 0, -pl.CenterPoint * pl.MaxShadowDistance, 1})
	cp := v.Vec3()

	pos := cp.Add(pl.Direction.Normalize().Mul(-pl.MaxShadowDistance * 0.5)).Vec4(0)
	return pos
}

func makeDirFrameResource(cache *vk.RenderCache, rsr *renderResources, imageIndex int) *dirFrameResources {
	fr := dirFrameResources{}
	fr.fb = vk.NewFramebuffer(cache.Ctx, rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex]})
	return &fr
}
