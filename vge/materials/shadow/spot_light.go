package shadow

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

// NewSpotLight will construct a spot light that has a shadow. MapSize parameter sets size of parabloid shadow map.
// Higher resolution will produce more accurate shadow but will have higher memory and GPU rendering cost.
func NewSpotLight(base vscene.SpotLight, mapSize uint32) *SpotLight {
	return &SpotLight{SpotLight: base, key: vk.NewKey(), mapSize: mapSize}
}

type SpotLight struct {
	vscene.SpotLight

	// Number of frames to keep same shadow map
	UpdateDelay int

	key     vk.Key
	mapSize uint32
}

// SetUpdateDelay set delay between shadowmap updates. 0 - each frame, 1 - every second frame, 2 - every third frame...
func (pl *SpotLight) SetUpdateDelay(delayFrames int) *SpotLight {
	pl.UpdateDelay = delayFrames
	return pl
}

func (pl *SpotLight) Process(pi *vscene.ProcessInfo) {
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
			pl.renderShadowMap(pd, pi, rsr)
		}
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func() interface{} {
			return pl.makeRenderResources(pi.Frame.GetCache().Device)
		}).(*renderResources)
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := pl.AsStdLight(pi.World)
		var imIndex vmodel.ImageIndex
		if ok && rsr.lastImage >= 0 {
			imIndex = imFrame.AddFrameImage(rsr.shadowImages[rsr.lastImage].DefaultView(), rsr.sampler)
		}
		if imIndex >= 0 {
			l.ShadowMapMethod = 4
			l.ShadowMapIndex = float32(imIndex)
			l.ShadowPlane = QuoternionFromYUp(l.Direction.Vec3())
		}
		lp.AddLight(l, lp)
	}
}

func (pl *SpotLight) renderShadowMap(pd *vscene.PredrawPhase, pi *vscene.ProcessInfo, rsr *renderResources) *dirResources {
	cache := pi.Frame.GetCache()
	sr := cache.Get(pl.key, func() interface{} {
		return makeDirResources(cache.Device, rsr.rp)
	}).(*dirResources)
	gpl := rsr.rp.Get(kDepthPipeline, func() interface{} {
		return makePlShadowPipeline(cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rsr.rp.Get(kSkinnedDepthPipeline, func() interface{} {
		return makePlShadowPipeline(cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	cmd := sr.cmd
	cmd.Begin()
	imageIndex := rsr.lastImage + 1
	if imageIndex >= 2 {
		imageIndex = 0
	}
	fbs := cache.GetPerFrame(pl.key, func() interface{} {
		return makeDirFrameResource(cache, rsr, imageIndex)
	}).(*dirFrameResources)
	fb := fbs.fb
	cmd.BeginRenderPass(rsr.rp, fb)
	sp := &shadowPass{cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: cache, renderer: pi.Frame.GetRenderer(), sampler: rsr.sampler,
		dsFrame: rsr.dsFrame[imageIndex], slFrame: rsr.slFrame[imageIndex]}
	l := pl.AsStdLight(pi.World)
	sp.pos = l.Position
	sp.dir = l.Direction.Vec3()

	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		// cmd.Wait()
	})
	rsr.lastImage, rsr.updateCount = imageIndex, pl.UpdateDelay
	return sr
}

func (pl *SpotLight) makeRenderResources(dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = makeRenderPass(dev)
	rsr.pool = vk.NewMemoryPool(dev)
	desc := vk.ImageDescription{Width: pl.mapSize, Height: pl.mapSize, Depth: 1, Layers: 1, MipLevels: 1,
		Format: vk.FORMATD32Sfloat}
	var buffers []*vk.Buffer
	for idx := 0; idx < 2; idx++ {
		rsr.shadowImages = append(rsr.shadowImages, rsr.pool.ReserveImage(desc, vk.IMAGEUsageDepthStencilAttachmentBit|
			vk.IMAGEUsageSampledBit))
		buffers = append(buffers, rsr.pool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit))
	}
	rsr.dpFrame = vk.NewDescriptorPool(getShadowFrameLayout(dev), 2)

	rsr.pool.Allocate()
	rsr.sampler = vmodel.GetDefaultSampler(dev)
	for idx := 0; idx < 2; idx++ {
		rsr.dsFrame = append(rsr.dsFrame, rsr.dpFrame.Alloc())
		sl := buffers[idx].Slice(0, vk.MinUniformBufferOffsetAlignment)
		rsr.dsFrame[idx].WriteSlice(0, 0, sl)
		rsr.slFrame = append(rsr.slFrame, sl)
		rg := vk.ImageRange{FirstLayer: 0, LayerCount: 1, LevelCount: 1}
		rsr.shadowViews = append(rsr.shadowViews, vk.NewImageView(rsr.shadowImages[idx], &rg))
	}
	return &rsr
}
