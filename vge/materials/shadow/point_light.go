//

package shadow

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const ShadowFormat = vk.FORMATD32Sfloat

type PointLight struct {
	vscene.PointLight

	// Number of frames to keep same shadow map
	UpdateDelay int

	key     vk.Key
	mapSize uint32
}

// SetUpdateDelay set delay between shadowmap updates. 0 - each frame, 1 - every second frame, 2 - every third frame...
func (pl *PointLight) SetUpdateDelay(delayFrames int) *PointLight {
	pl.UpdateDelay = delayFrames
	return pl
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
	dpFrame      *vk.DescriptorPool
	dsFrame      []*vk.DescriptorSet
	slFrame      []*vk.Slice
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
	if r.dpFrame != nil {
		r.dpFrame.Dispose()
		r.dsFrame, r.dpFrame, r.slFrame = nil, nil, nil
	}
}

type plResources struct {
	cmd []*vk.Command
}

const maxInstances = 800

func QuoternionFromYUp(direction mgl32.Vec3) mgl32.Vec4 {
	q := mgl32.QuatBetweenVectors(mgl32.Vec3{0, 1, 0}, direction)
	q = q.Normalize()
	return mgl32.Vec4{q.V[0], q.V[1], q.V[2], q.W}
}

type shaderInstance struct {
	world       mgl32.Mat4
	tx_albedo   float32
	alphaCutoff float32
	filler1     float32
	filler2     float32
}

type shaderFrame struct {
	plane     mgl32.Vec4
	lightPos  mgl32.Vec4
	minShadow float32
	maxShadow float32
	yFactor   float32
	dummy2    float32
}

type shaderInstances struct {
	instances [maxInstances]shaderInstance
}

// Objects under this node will not cast shadow!
type NoShadow struct {
}

func (n NoShadow) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(*shadowPass)
	if ok {
		pi.Visible = false
	}

	_, ok = pi.Phase.(*cubeShadowPass)
	if ok {
		pi.Visible = false
	}
}

func (s *plResources) Dispose() {

	for _, cmd := range s.cmd {
		cmd.Dispose()
	}
	s.cmd = nil
}

var kShadowLayout = vk.NewKey()
var kFrameBuffer = vk.NewKeys(2)

// NewPointLight will construct a point light that has a shadow. MapSize parameter sets size of parabloid shadow map.
// Higher resolution will produce more accurate shadow but will have higher memory and GPU rendering cost.
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
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func() interface{} {
			return pl.makeRenderResources(pi.Frame.GetCache().Device)
		}).(*renderResources)

		if rsr.updateCount > 0 {
			rsr.updateCount--
		} else {
			pl.renderShadowMap(pd, pi, rsr, 0)
			pl.renderShadowMap(pd, pi, rsr, 1)
		}
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func() interface{} {
			return pl.makeRenderResources(pi.Frame.GetCache().Device)
		}).(*renderResources)
		pos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := vscene.Light{Intensity: pl.Intensity.Vec4(1),
			Position: pos, Attenuation: pl.Attenuation.Vec4(pl.MaxDistance)}
		var imIndex vmodel.ImageIndex
		var imIndex2 vmodel.ImageIndex
		if ok && rsr.lastImage >= 0 {
			imIndex = imFrame.AddFrameImage(rsr.shadowViews[rsr.lastImage*2+4], rsr.sampler)
			imIndex2 = imFrame.AddFrameImage(rsr.shadowViews[rsr.lastImage*2+5], rsr.sampler)
		}
		if imIndex2 > 0 {
			l.ShadowMapMethod = 3
			l.ShadowMapIndex = float32(imIndex)
		}
		lp.AddLight(l, lp)
	}
}

func (pl *PointLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo, rsr *renderResources, side int) *plResources {
	cache := pi.Frame.GetCache()
	sr := cache.Get(pl.key, func() interface{} {
		return pl.makeResources(cache.Device, rsr.rp)
	}).(*plResources)
	gpl := rsr.rp.Get(kDepthPipeline, func() interface{} {
		return makePlShadowPipeline(cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rsr.rp.Get(kSkinnedDepthPipeline, func() interface{} {
		return makePlSkinnedShadowPipeline(cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd[side]
	cmd.Begin()
	imageIndex := rsr.lastImage + 1
	if imageIndex >= 2 {
		imageIndex = 0
	}
	fbs := cache.GetPerFrame(pl.key, func() interface{} {
		return pl.makeFrameResource(cache, rsr, imageIndex)
	}).(*plFrameResources)
	fb := fbs.fbs[side]
	cmd.BeginRenderPass(rsr.rp, fb)
	sp := &shadowPass{cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		rc: cache, renderer: pi.Frame.GetRenderer(), pl: gpl, plSkin: gSkinnedPl, sampler: rsr.sampler,
		dsFrame: rsr.dsFrame[imageIndex*2+side], slFrame: rsr.slFrame[2*imageIndex+side]}
	lightPos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
	sp.pos = lightPos
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
		rsr.updateCount = pl.UpdateDelay
	}
	return sr
}

func makeRenderPass(dev *vk.Device) *vk.GeneralRenderPass {
	rp := dev.Get(kDepthPass, func() interface{} {
		return vk.NewGeneralRenderPass(dev, true, []vk.AttachmentInfo{
			vk.AttachmentInfo{Format: ShadowFormat, InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal,
				ClearColor: [4]float32{1, 0, 0, 0}},
		})
	}).(*vk.GeneralRenderPass)
	return rp
}

func (pl *PointLight) makeResources(dev *vk.Device, rp *vk.DepthRenderPass) *plResources {
	sr := &plResources{}
	sr.cmd = append(sr.cmd, vk.NewCommand(dev, vk.QUEUEGraphicsBit, false), vk.NewCommand(dev, vk.QUEUEGraphicsBit, false))
	return sr
}

func makePlShadowPipeline(dev *vk.Device, rp *vk.GeneralRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(dev)
	vmodel.AddInput(gp, vmodel.MESHKindNormal)
	gp.AddLayout(getShadowFrameLayout(dev))
	gp.AddLayout(vscene.GetUniformLayout(dev))
	gp.AddShader(vk.SHADERStageVertexBit, point_shadow_vert_spv)
	if vscene.FrameMaxDynamicSamplers > 0 {
		gp.AddShader(vk.SHADERStageFragmentBit, point_shadow_dyn_frag_spv)

	} else {
		gp.AddShader(vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	}
	gp.AddDepth(true, true)
	gp.Create(rp)
	return gp
}

func makePlSkinnedShadowPipeline(dev *vk.Device, rp *vk.DepthRenderPass) *vk.GraphicsPipeline {
	gp := vk.NewGraphicsPipeline(dev)
	vmodel.AddInput(gp, vmodel.MESHKindSkinned)
	gp.AddLayout(getShadowFrameLayout(dev))
	gp.AddLayout(vscene.GetUniformLayout(dev))
	gp.AddLayout(vscene.GetUniformLayout(dev))
	gp.AddShader(vk.SHADERStageVertexBit, point_shadow_vert_skin_spv)
	if vscene.FrameMaxDynamicSamplers > 0 {
		gp.AddShader(vk.SHADERStageFragmentBit, point_shadow_dyn_frag_spv)

	} else {
		gp.AddShader(vk.SHADERStageFragmentBit, point_shadow_frag_spv)
	}
	gp.AddDepth(true, true)
	gp.Create(rp)
	return gp
}

func (pl *PointLight) makeRenderResources(dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = makeRenderPass(dev)
	rsr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Depth: 1, Layers: 2, MipLevels: 1,
		Format: vk.FORMATD32Sfloat}
	var buffers []*vk.Buffer
	for idx := 0; idx < 2; idx++ {
		rsr.shadowImages = append(rsr.shadowImages, rsr.pool.ReserveImage(desc, vk.IMAGEUsageDepthStencilAttachmentBit|
			vk.IMAGEUsageSampledBit))
		buffers = append(buffers,
			rsr.pool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit))
		buffers = append(buffers,
			rsr.pool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit))
	}
	rsr.pool.Allocate()
	rsr.dpFrame = vk.NewDescriptorPool(getShadowFrameLayout(dev), 4)
	rsr.sampler = vmodel.GetDefaultSampler(dev)
	for idx := 0; idx < 2; idx++ {
		rsr.dsFrame = append(rsr.dsFrame, rsr.dpFrame.Alloc(), rsr.dpFrame.Alloc())
		sl := buffers[idx*2].Slice(0, vk.MinUniformBufferOffsetAlignment)
		rsr.dsFrame[idx*2].WriteSlice(0, 0, sl)
		rsr.slFrame = append(rsr.slFrame, sl)
		sl = buffers[idx*2+1].Slice(0, vk.MinUniformBufferOffsetAlignment)
		rsr.dsFrame[idx*2+1].WriteSlice(0, 0, sl)
		rsr.slFrame = append(rsr.slFrame, sl)
		rg := vk.ImageRange{FirstLayer: 0, LayerCount: 1, LevelCount: 1}
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(rsr.shadowImages[idx], &rg))
		rg2 := rg
		rg2.FirstLayer = 1
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(rsr.shadowImages[idx], &rg2))
	}

	for idx := 0; idx < 2; idx++ {
		rg := vk.ImageRange{FirstLayer: 0, LayerCount: 1, LevelCount: 1, Layout: vk.IMAGELayoutShaderReadOnlyOptimal}
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(rsr.shadowImages[idx], &rg))
		rg2 := rg
		rg2.FirstLayer = 1
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(rsr.shadowImages[idx], &rg2))
	}
	return &rsr
}

func (pl *PointLight) makeFrameResource(cache *vk.RenderCache, rsr *renderResources, imageIndex int) *plFrameResources {
	fr := plFrameResources{}
	fr.fbs = append(fr.fbs,
		vk.NewFramebuffer(rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex*2]}),
		vk.NewFramebuffer(rsr.rp, []*vk.ImageView{rsr.shadowViews[imageIndex*2+1]}),
	)
	return &fr
}
