//

package shadow

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const ShadowFormat = vk.FORMATD32Sfloat

type PointLight struct {
	// Maximum distance we can see lights shadows from. If light is longer that this distance away for camera,
	// we just turn shadows off
	vscene.PointLight
	key     vk.Key
	mapSize uint32
}

// Per renderer resources
type renderResources struct {
	pool         *vk.MemoryPool // Pool for shadow maps
	shadowImages []*vk.Image
	sampler      *vk.Sampler
	shadowViews  []*vk.ImageView
	lastImage    int
	updateCount  int
	rp           *vk.GeneralRenderPass
}

type plFrameResources struct {
	fbs []*vk.Framebuffer
}

func (f *plFrameResources) Dispose() {
	for _, fb := range f.fbs {
		fb.Dispose()
	}
	f.fbs = nil
}

func (r *renderResources) Dispose() {
	for _, v := range r.shadowViews {
		v.Dispose()
	}
	r.shadowViews = nil
	if r.pool != nil {
		r.pool.Dispose()
		r.pool = nil
	}
}

type plResources struct {
	cmd []*vk.Command
}

const maxInstances = 1000

func QuoternionFromYUp(direction mgl32.Vec3) mgl32.Vec4 {
	q := mgl32.QuatBetweenVectors(mgl32.Vec3{0, 1, 0}, direction)
	q = q.Normalize()
	return mgl32.Vec4{q.V[0], q.V[1], q.V[2], q.W}
}

type shaderFrame struct {
	plane     mgl32.Vec4
	lightPos  mgl32.Vec4
	minShadow float32
	maxShadow float32
	yFactor   float32
	dummy2    float32

	instances [maxInstances]mgl32.Mat4
}

type plShadowPass struct {
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
	yFactor     float32
}

func (s *plShadowPass) GetRenderer() vmodel.Renderer {
	return s.renderer
}

func (s *plShadowPass) BindFrame() *vk.DescriptorSet {
	uc := vscene.GetUniformCache(s.rc)
	if s.dsFrame == nil {
		s.dsFrame, s.slFrame = uc.Alloc(s.ctx)
		s.si = &shaderFrame{lightPos: s.pos.Vec4(1), maxShadow: s.maxDistance, minShadow: 0}
		if s.yFactor != 0 {
			s.si.yFactor = s.yFactor
		} else {
			s.si.plane = QuoternionFromYUp(s.dir)
		}

	}
	return s.dsFrame
}

func (s *plShadowPass) GetCache() *vk.RenderCache {
	return s.rc
}

// Objects under this node will not cast shadow!
type NoShadow struct {
}

func (n NoShadow) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(*plShadowPass)
	if ok {
		pi.Visible = false
	}
	_, ok = pi.Phase.(*dirShadowPass)
	if ok {
		pi.Visible = false
	}
	_, ok = pi.Phase.(*cubeShadowPass)
	if ok {
		pi.Visible = false
	}
}

func (s *plShadowPass) Begin() (atEnd func()) {
	return nil
}

func (s *plShadowPass) DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex) {
	s.BindFrame()
	s.si.instances[s.siCount] = world
	s.dl.DrawIndexed(s.pl, mesh.From, mesh.Count).AddDescriptors(s.dsFrame).
		AddInputs(mesh.Model.VertexBuffers(vmodel.MESHKindNormal)...).SetInstances(uint32(s.siCount), 1)
	s.siCount++
	if s.siCount >= maxInstances {
		s.flush()
	}
}

func (s *plShadowPass) DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex, aniMatrix []mgl32.Mat4) {
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

func (s *plShadowPass) flush() {
	if s.siCount > 0 {
		b := *(*[unsafe.Sizeof(shaderFrame{})]byte)(unsafe.Pointer(s.si))
		copy(s.slFrame.Content, b[:])
		s.cmd.Draw(s.dl)
		s.dl = &vk.DrawList{}
	}
	s.si, s.dsFrame, s.slFrame, s.siCount = nil, nil, nil, 0
}

func (s *plResources) Dispose() {

	for _, cmd := range s.cmd {
		cmd.Dispose()
	}
	s.cmd = nil
}

var kShadowLayout = vk.NewKey()
var kFrameBuffer = vk.NewKeys(2)

func NewPointLight(baseLight vscene.PointLight, mapSize uint32) *PointLight {
	return &PointLight{key: vk.NewKey(), PointLight: baseLight, mapSize: mapSize}
}

func (pl *PointLight) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if pl.MaxDistance == 0 {
			pl.MaxDistance = 10
		}

		_, ok := pi.Frame.(vscene.ImageFrame)
		if !ok {
			return
		}
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		sf := vscene.GetSimpleFrame(pi.Frame)
		if sf == nil {
			return
		}
		eyePos := sf.SSF.View.Inv().Col(3).Vec3()
		le := eyePos.Sub(pos.Vec3()).Len()
		if le > pl.MaxDistance*4 || le < 0.1 {
			// Skip shadow pass for light too long away or very close
			return
		}
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)

		pl.renderShadowMap(pd, pi, rsr, 0)
		pl.renderShadowMap(pd, pi, rsr, 1)
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := vscene.Light{Intensity: pl.Intensity.Vec4(1),
			Position: pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)}
		var imIndex vmodel.ImageIndex
		if ok && rsr.lastImage >= 0 {
			imIndex = imFrame.AddFrameImage(rsr.shadowImages[rsr.lastImage].DefaultView(pi.Frame.GetCache().Ctx), rsr.sampler)
		}
		if imIndex > 0 {
			l.ShadowMapMethod = 3
			l.ShadowMapIndex = float32(imIndex)
		}
		lp.AddLight(l, lp)
	}
}

func (pl *PointLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo, rsr *renderResources, side int) *plResources {
	cache := pi.Frame.GetCache()
	sr := cache.Get(pl.key, func(ctx vk.APIContext) interface{} {
		return pl.makeResources(ctx, cache.Device, rsr.rp)
	}).(*plResources)
	gpl := rsr.rp.Get(cache.Ctx, kDepthPipeline, func(ctx vk.APIContext) interface{} {
		return makePlShadowPipeline(ctx, cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rsr.rp.Get(cache.Ctx, kSkinnedDepthPipeline, func(ctx vk.APIContext) interface{} {
		return makePlSkinnedShadowPipeline(ctx, cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd[side]
	cmd.Begin()
	imageIndex := rsr.lastImage + 1
	if imageIndex >= 2 {
		imageIndex = 0
	}
	fbs := cache.GetPerFrame(pl.key, func(ctx vk.APIContext) interface{} {
		return pl.makeFrameResource(cache, rsr, imageIndex)
	}).(*plFrameResources)
	fb := fbs.fbs[side]
	cmd.BeginRenderPass(rsr.rp, fb)
	sp := &plShadowPass{ctx: cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: cache, renderer: pi.Frame.GetRenderer()}
	lightPos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
	sp.pos = lightPos.Vec3()
	sp.yFactor = -1
	if side == 1 {
		sp.yFactor = 1
	}

	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		// cmd.Wait()
	})
	if side == 1 {
		rsr.lastImage = imageIndex
	}
	return sr
}

func makeRenderPass(ctx vk.APIContext, dev *vk.Device) *vk.GeneralRenderPass {
	rp := dev.Get(ctx, kDepthPass, func(ctx vk.APIContext) interface{} {
		return vk.NewGeneralRenderPass(ctx, dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{Format: ShadowFormat, InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				ClearColor: [4]float32{1, 0, 0, 0}},
		})
	}).(*vk.GeneralRenderPass)
	return rp
}

func (pl *PointLight) makeResources(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *plResources {
	sr := &plResources{}
	sr.cmd = append(sr.cmd, vk.NewCommand(ctx, dev, vk.QUEUEGraphicsBit, false), vk.NewCommand(ctx, dev, vk.QUEUEGraphicsBit, false))
	return sr
}

func makePlShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.GeneralRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindNormal)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, point_shadow_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func makePlSkinnedShadowPipeline(ctx vk.APIContext, dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(ctx, dev)
	vmodel.AddInput(ctx, gp, vmodel.MESHKindSkinned)
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddLayout(ctx, vscene.GetUniformLayout(ctx, dev))
	gp.AddShader(ctx, vk.SHADERStageVertexBit, point_shadow_vert_skin_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	gp.AddDepth(ctx, true, true)
	gp.Create(ctx, rp)
	return gp
}

func (pl *PointLight) makeRenderResources(ctx vk.APIContext, dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = makeRenderPass(ctx, dev)
	rsr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Depth: 1, Layers: 2, MipLevels: 1,
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
		rg2 := rg
		rg2.FirstLayer = 1
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(ctx, rsr.shadowImages[idx], &rg2))
	}
	return &rsr
}

func (pl *PointLight) makeFrameResource(cache *vk.RenderCache, rsr *renderResources, imageIndex int) *plFrameResources {
	fr := plFrameResources{}
	fr.fbs = append(fr.fbs,
		vk.NewFramebuffer(cache.Ctx, rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex*2]}),
		vk.NewFramebuffer(cache.Ctx, rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex*2+1]}),
	)
	return &fr
}
