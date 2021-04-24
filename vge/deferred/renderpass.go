package deferred

import (
	"errors"
	"image"
	"runtime"
	"time"

	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const LAYER3DSplit = vscene.LAYER3D + 1
const ShadowPoints = 1024 * 1024

type Renderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

	timedOutput func(started time.Time, gpuTimes []float64)

	owner   vk.Owner
	dev     *vk.Device
	Ctx     vk.APIContext
	rpFinal *vk.GeneralRenderPass
	rpSplit *vk.GeneralRenderPass
	rpBG    *vk.GeneralRenderPass

	// rpJoin     *vk.GeneralRenderPass
	mpImages   *vk.MemoryPool
	imDepth    []*vk.Image
	imColor    []*vk.Image
	imNormal   []*vk.Image
	imMaterial []*vk.Image

	imViews      []*vk.ImageView
	frameBuffers []*vk.Buffer
	whiteImage   *vk.Image
	laLights     *vk.DescriptorLayout
	dpLights     *vk.DescriptorPool
	dsLights     []*vk.DescriptorSet
	size         image.Point
	joinPipeline *vk.GraphicsPipeline

	debugMode  int
	debugIndex int
}

func (r *Renderer) GetPerRenderer(key vk.Key, ctor func(ctx vk.APIContext) interface{}) interface{} {
	return r.owner.Get(r.Ctx, key, ctor)
}

var _ vscene.Renderer = &Renderer{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (f *Renderer) SetTimedOutput(output func(started time.Time, gpuTimes []float64)) {
	f.timedOutput = output
}

var kImageViews = vk.NewKeys(10)
var kFpSplit = vk.NewKey()
var kFpBG = vk.NewKey()
var kFpShadow = vk.NewKey()
var kFpFinal = vk.NewKey()
var kCmd = vk.NewKey()
var vkShadowLayout = vk.NewKey()

func (f *Renderer) Dispose() {
	for _, v := range f.imViews {
		v.Dispose()
	}
	f.imViews = nil
	if f.mpImages != nil {
		f.mpImages.Dispose()
		f.mpImages = nil
	}
	if f.rpFinal != nil {
		f.rpFinal.Dispose()
		f.rpFinal = nil
	}
	if f.rpBG != nil {
		f.rpBG.Dispose()
		f.rpBG = nil
	}
	if f.rpSplit != nil {
		f.rpSplit.Dispose()
		f.rpSplit = nil
	}

}

func (f *Renderer) GetRenderPass() vk.RenderPass {
	return f.rpFinal
}

func (f *Renderer) SetDebugMode(mode int) {
	f.debugMode = mode
}

func (f *Renderer) Setup(ctx vk.APIContext, dev *vk.Device, mainImage vk.ImageDescription, images int) {
	if vscene.FrameMaxDynamicSamplers == 0 {
		ctx.SetError(errors.New("you must enable DynamicDescriptor and set vscene.FrameMaxDynamicSamplers for DeferredRenderer"))
	}
	depthDesc := mainImage
	f.size = image.Pt(int(mainImage.Width), int(mainImage.Height))
	depthDesc.Format = vk.FORMATD32Sfloat
	colorDesc := mainImage
	colorDesc.Format = vk.FORMATR8g8b8a8Unorm
	normalDesc := mainImage
	normalDesc.Format = vk.FORMATA2b10g10r10UnormPack32
	materialDesc := mainImage
	materialDesc.Format = vk.FORMATR8g8b8a8Uint
	if f.rpFinal != nil {
		f.mpImages.Dispose()
		f.imDepth, f.imColor, f.imMaterial, f.imNormal, f.frameBuffers = nil, nil, nil, nil, nil
	} else {
		f.Ctx, f.dev = ctx, dev
		f.rpSplit = vk.NewGeneralRenderPass(ctx, dev, true, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutColorAttachmentOptimal, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: colorDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: normalDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: materialDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutGeneral, Format: depthDesc.Format,
				ClearColor: [4]float32{1, 0, 0, 0}},
		})
		f.rpFinal = vk.NewGeneralRenderPass(ctx, dev, false, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutPresentSrcKhr, Format: mainImage.Format},
		})
		f.rpBG = vk.NewGeneralRenderPass(ctx, dev, false, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutColorAttachmentOptimal, Format: colorDesc.Format},
		})

		la := vscene.GetUniformLayout(ctx, dev)
		f.laLights = la.AddDynamicBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit,
			vscene.FrameMaxDynamicSamplers, vk.DESCRIPTORBindingPartiallyBoundBitExt)
		f.dpLights = vk.NewDescriptorPool(ctx, f.laLights, images)
		for idx := 0; idx < images; idx++ {
			f.dsLights = append(f.dsLights, f.dpLights.Alloc(ctx))
		}
		f.joinPipeline = f.newLightsPipeline(ctx, dev)
	}
	f.mpImages = vk.NewMemoryPool(dev)
	for idx := 0; idx < images; idx++ {
		f.imDepth = append(f.imDepth, f.mpImages.ReserveImage(ctx, depthDesc,
			vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imColor = append(f.imColor, f.mpImages.ReserveImage(ctx, colorDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imNormal = append(f.imNormal, f.mpImages.ReserveImage(ctx, normalDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imMaterial = append(f.imMaterial, f.mpImages.ReserveImage(ctx, materialDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}

	for idx := 0; idx < images; idx++ {
		f.frameBuffers = append(f.frameBuffers, f.mpImages.ReserveBuffer(ctx, 16384, true, vk.BUFFERUsageUniformBufferBit))
	}
	sampler := vmodel.GetDefaultSampler(ctx, f.dev)
	f.whiteImage = f.mpImages.ReserveImage(ctx, vmodel.DescribeWhiteImage(ctx), vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
	f.mpImages.Allocate(ctx)
	cp := vmodel.NewCopier(ctx, dev)
	defer cp.Dispose()
	cp.CopyWhiteImage(f.whiteImage)
	for idx := 0; idx < images; idx++ {
		f.dsLights[idx].WriteImage(ctx, 1, 0, f.whiteImage.DefaultView(ctx), sampler)
		f.dsLights[idx].WriteImage(ctx, 1, 1, f.imColor[idx].DefaultView(ctx), sampler)
		f.dsLights[idx].WriteImage(ctx, 1, 2, f.imNormal[idx].DefaultView(ctx), sampler)
		f.dsLights[idx].WriteImage(ctx, 1, 3, f.imMaterial[idx].DefaultView(ctx), sampler)
		f.dsLights[idx].WriteImage(ctx, 1, 4, f.imDepth[idx].DefaultView(ctx), sampler)
	}

}

func (f *Renderer) Render(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo) {
	mainView := rc.Get(kImageViews+vk.Key(imageIndex), func(ctx vk.APIContext) interface{} {
		return mainImage.NewView(ctx, 0, 0)
	}).(*vk.ImageView)
	f.RenderView(camera, sc, rc, mainView, imageIndex, infos)
}

var kTimer = vk.NewKey()
var kTimerCmd = vk.NewKey()

func (f *Renderer) RenderView(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainView *vk.ImageView, imageIndex int, infos []vk.SubmitInfo) {
	depthView := f.imDepth[imageIndex].DefaultView(rc.Ctx)
	colorView := f.imColor[imageIndex].DefaultView(rc.Ctx)
	normalView := f.imNormal[imageIndex].DefaultView(rc.Ctx)
	materialView := f.imMaterial[imageIndex].DefaultView(rc.Ctx)

	dsLight := f.dsLights[imageIndex]
	fbSplit := rc.Get(kFpSplit, func(ctx vk.APIContext) interface{} {
		return vk.NewFramebuffer(ctx, f.rpSplit, []*vk.ImageView{colorView, normalView, materialView, depthView})
	}).(*vk.Framebuffer)
	fbFinal := rc.Get(kFpFinal, func(ctx vk.APIContext) interface{} {
		return vk.NewFramebuffer(ctx, f.rpFinal, []*vk.ImageView{mainView})
	}).(*vk.Framebuffer)
	fbBG := rc.Get(kFpBG, func(ctx vk.APIContext) interface{} {
		return vk.NewFramebuffer(ctx, f.rpBG, []*vk.ImageView{colorView})
	}).(*vk.Framebuffer)

	start := time.Now()
	var tp *vk.TimerPool
	if f.timedOutput != nil {
		tp = vk.NewTimerPool(rc.Ctx, rc.Device, 3)
		rc.SetPerFrame(kTimer, tp)
		timerCmd := vk.NewCommand(rc.Ctx, rc.Device, vk.QUEUEComputeBit, true)
		timerCmd.Begin()
		timerCmd.WriteTimer(tp, 0, vk.PIPELINEStageTopOfPipeBit)
		infos = append(infos, timerCmd.SubmitForWait(1, vk.PIPELINEStageTopOfPipeBit))
		rc.SetPerFrame(kTimerCmd, timerCmd)
	}
	cmd := rc.Get(kCmd, func(ctx vk.APIContext) interface{} {
		return vk.NewCommand(f.Ctx, f.dev, vk.QUEUEGraphicsBit, false)
	}).(*vk.Command)
	cmd.Begin()
	if f.timedOutput != nil {
		cmd.WriteTimer(tp, 1, vk.PIPELINEStageTopOfPipeBit)
	}
	frame := &DeferredFrame{dsLight: dsLight, imagesUsed: 4, cache: rc, renderer: f}
	frame.bfLightsFrame = f.frameBuffers[imageIndex]
	frame.LightsFrame.Debug = float32(f.debugMode)
	frame.LightsFrame.Index = float32(f.debugIndex)
	f.debugIndex = (f.debugIndex + 1) % 256
	bgPhase := vscene.NewDrawPhase(frame, f.rpBG, vscene.LAYERBackground, cmd, func() {
		cmd.BeginRenderPass(f.rpBG, fbBG)
	}, func() {
		cmd.EndRenderPass()
	})
	splitPhase := vscene.NewDrawPhase(frame, f.rpSplit, vscene.LAYER3D, cmd, func() {
		cmd.BeginRenderPass(f.rpSplit, fbSplit)
	}, func() {
		cmd.EndRenderPass()
	})
	frame.DrawPhase.Projection, frame.DrawPhase.View = camera.CameraProjection(f.size)
	frame.DrawPhase.EyePos = frame.DrawPhase.View.Inv().Col(3)
	frame.LightsFrame.View = frame.DrawPhase.View

	join := &DrawLights{ds: dsLight, fb: fbFinal, rp: f.rpFinal, cache: rc, cmd: cmd, frame: frame, pipeline: f.joinPipeline}
	// splitPhase := vscene.NewDrawPhase(rc, f.rpFinal, vscene.LAYER3D, cmd, nil, nil)
	// dt := vscene.NewDrawPhase(rc, f.rpFinal, vscene.LAYERTransparent, cmd, nil, nil)
	ui := vscene.NewDrawPhase(frame, f.rpFinal, vscene.LAYERUI, cmd, func() {
	}, func() {

		cmd.EndRenderPass()
	})
	ppPhase := &vscene.PredrawPhase{Scene: sc, Cmd: cmd}

	sc.Process(sc.Time, frame, &vscene.AnimatePhase{}, ppPhase, bgPhase, splitPhase, join, ui)

	// Complete pendings from predraw phase
	for _, pd := range ppPhase.Pending {
		pd()
	}
	infos = append(infos, ppPhase.Needeed...)
	if tp != nil {
		cmd.WriteTimer(tp, 2, vk.PIPELINEStageAllCommandsBit)
	}
	cmd.Submit(infos...)
	cmd.Wait()
	runtime.KeepAlive(infos)
	if tp != nil {
		f.timedOutput(start, tp.Get(rc.Ctx))
	}
	if f.RenderDone != nil {
		f.RenderDone(start)
	}

}
