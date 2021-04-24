package shadow

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func NewSpotLight(base vscene.SpotLight, mapSize uint32) *SpotLight {
	return &SpotLight{SpotLight: base, key: vk.NewKey(), mapSize: mapSize}
}

type SpotLight struct {
	vscene.SpotLight
	key     vk.Key
	mapSize uint32
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
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)

		pl.renderShadowMap(pd, pi, rsr)
	}

	lp, ok := pi.Phase.(vscene.LightPhase)
	if ok {
		rsr := pi.Frame.GetRenderer().GetPerRenderer(pl.key, func(ctx vk.APIContext) interface{} {
			return pl.makeRenderResources(ctx, pi.Frame.GetCache().Device)
		}).(*renderResources)
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		l := pl.AsStdLight(pi.World)
		var imIndex vmodel.ImageIndex
		if ok && rsr.lastImage >= 0 {
			imIndex = imFrame.AddFrameImage(rsr.shadowImages[rsr.lastImage].DefaultView(pi.Frame.GetCache().Ctx), rsr.sampler)
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
	sr := cache.Get(pl.key, func(ctx vk.APIContext) interface{} {
		return makeDirResources(ctx, cache.Device, rsr.rp)
	}).(*dirResources)
	gpl := rsr.rp.Get(cache.Ctx, kDepthPipeline, func(ctx vk.APIContext) interface{} {
		return makePlShadowPipeline(ctx, cache.Device, rsr.rp)
	}).(*vk.GraphicsPipeline)
	gSkinnedPl := rsr.rp.Get(cache.Ctx, kSkinnedDepthPipeline, func(ctx vk.APIContext) interface{} {
		return makePlShadowPipeline(ctx, cache.Device, rsr.rp)
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
	sp := &plShadowPass{ctx: cache.Ctx, cmd: cmd, dl: &vk.DrawList{}, maxDistance: pl.MaxDistance,
		pl: gpl, plSkin: gSkinnedPl, rc: cache, renderer: pi.Frame.GetRenderer()}
	lightPos := pi.World.Mul4x1(mgl32.Vec4{0, 0, 0, 1})
	sp.pos = lightPos.Vec3()
	sp.dir = pi.World.Mul4x1(pl.Direction.Vec4(0)).Vec3()

	pd.Scene.Process(pi.Time, sp, sp)
	sp.flush()
	cmd.EndRenderPass()
	waitFor := cmd.SubmitForWait(1, vk.PIPELINEStageFragmentShaderBit)
	pd.Pending = append(pd.Pending, func() {
		pd.Needeed = append(pd.Needeed, waitFor)
		// cmd.Wait()
	})
	rsr.lastImage = imageIndex
	return sr
}

func (pl *SpotLight) makeRenderResources(ctx vk.APIContext, dev *vk.Device) *renderResources {
	rsr := renderResources{lastImage: -1}
	rsr.rp = makeRenderPass(ctx, dev)
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
