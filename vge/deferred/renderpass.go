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

// Renderer implement deferred shading that first render all objects into G-buffer. Light calculations are done from G-buffer
// data after all objects have been rendered to G-buffers.
// Deferred rendering uses much more resources from GPU than forward shader but it should be faster with scenes that have lots of light (that don't cast shadows).
// Deferred rendering may also later support post processing effects that are not easily done with forward shader but currently they both have nearly same features.
type Renderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

	timedOutput func(started time.Time, gpuTimes []float64)

	owner   vk.Owner
	dev     *vk.Device
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
	laDraw       *vk.DescriptorLayout
	dpDraw       *vk.DescriptorPool
	dsDraw       []*vk.DescriptorSet
	size         image.Point
	joinPipeline *vk.GraphicsPipeline

	debugMode  int
	debugIndex int
}

func (r *Renderer) GetPerRenderer(key vk.Key, ctor func() interface{}) interface{} {
	return r.owner.Get(key, ctor)
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
var kFpFinal = vk.NewKey()
var kCmd = vk.NewKey()
var kFrameLayout = vk.NewKey()

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

func (f *Renderer) GetRenderPass() *vk.GeneralRenderPass {
	return f.rpFinal
}

func (f *Renderer) Setup(dev *vk.Device, mainImage vk.ImageDescription, images int) {
	if vscene.FrameMaxDynamicSamplers == 0 {
		dev.ReportError(errors.New("you must enable DynamicDescriptor and set vscene.FrameMaxDynamicSamplers for DeferredRenderer"))
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
		f.dev = dev
		f.rpSplit = vk.NewGeneralRenderPass(dev, true, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutColorAttachmentOptimal, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: colorDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: normalDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutShaderReadOnlyOptimal, Format: materialDesc.Format},
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutGeneral, Format: depthDesc.Format,
				ClearColor: [4]float32{1, 0, 0, 0}},
		})
		f.rpFinal = vk.NewGeneralRenderPass(dev, false, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutPresentSrcKhr, Format: mainImage.Format},
		})
		f.rpBG = vk.NewGeneralRenderPass(dev, false, []vk.AttachmentInfo{
			{InitialLayout: vk.IMAGELayoutUndefined, FinalLayout: vk.IMAGELayoutColorAttachmentOptimal, Format: colorDesc.Format},
		})

		la := vscene.GetUniformLayout(dev)
		f.laLights = la.AddDynamicBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit,
			vscene.FrameMaxDynamicSamplers, vk.DESCRIPTORBindingPartiallyBoundBitExt)
		f.laDraw = GetFrameLayout(dev)
		f.dpLights = vk.NewDescriptorPool(f.laLights, images)
		f.dpDraw = vk.NewDescriptorPool(f.laDraw, images)
		for idx := 0; idx < images; idx++ {
			f.dsLights = append(f.dsLights, f.dpLights.Alloc())
			f.dsDraw = append(f.dsDraw, f.dpDraw.Alloc())
		}
		f.joinPipeline = f.newLightsPipeline(dev)
	}
	f.mpImages = vk.NewMemoryPool(dev)
	for idx := 0; idx < images; idx++ {
		f.imDepth = append(f.imDepth, f.mpImages.ReserveImage(depthDesc,
			vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imColor = append(f.imColor, f.mpImages.ReserveImage(colorDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imNormal = append(f.imNormal, f.mpImages.ReserveImage(normalDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images; idx++ {
		f.imMaterial = append(f.imMaterial, f.mpImages.ReserveImage(materialDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit))
	}

	for idx := 0; idx < images*2; idx++ {
		f.frameBuffers = append(f.frameBuffers, f.mpImages.ReserveBuffer(32768, true, vk.BUFFERUsageUniformBufferBit))
	}
	sampler := vmodel.GetDefaultSampler(f.dev)
	f.whiteImage = f.mpImages.ReserveImage(vmodel.DescribeWhiteImage(), vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
	f.mpImages.Allocate()
	cp := vmodel.NewCopier(dev)
	defer cp.Dispose()
	cp.CopyWhiteImage(f.whiteImage)
	for idx := 0; idx < images; idx++ {
		f.dsLights[idx].WriteImage(1, 0, f.whiteImage.DefaultView(), sampler)
		f.dsLights[idx].WriteImage(1, 1, f.imColor[idx].DefaultView(), sampler)
		f.dsLights[idx].WriteImage(1, 2, f.imNormal[idx].DefaultView(), sampler)
		f.dsLights[idx].WriteImage(1, 3, f.imMaterial[idx].DefaultView(), sampler)
		f.dsLights[idx].WriteImage(1, 4, f.imDepth[idx].DefaultView(), sampler)
	}

}

func GetFrameLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kFrameLayout, func() interface{} {
		la := vscene.GetUniformLayout(dev)
		return la.AddDynamicBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit,
			vscene.FrameMaxDynamicSamplers, vk.DESCRIPTORBindingPartiallyBoundBitExt)
	}).(*vk.DescriptorLayout)
}

func (f *Renderer) Render(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo) {
	mainView := rc.Get(kImageViews+vk.Key(imageIndex), func() interface{} {
		return mainImage.NewView(0, 0)
	}).(*vk.ImageView)
	f.RenderView(camera, sc, rc, mainView, imageIndex, infos)
}

var kTimer = vk.NewKey()
var kTimerCmd = vk.NewKey()

func (f *Renderer) RenderView(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainView *vk.ImageView, imageIndex int, infos []vk.SubmitInfo) {
	depthView := f.imDepth[imageIndex].DefaultView()
	colorView := f.imColor[imageIndex].DefaultView()
	normalView := f.imNormal[imageIndex].DefaultView()
	materialView := f.imMaterial[imageIndex].DefaultView()

	dsLight := f.dsLights[imageIndex]
	fbSplit := rc.Get(kFpSplit, func() interface{} {
		return vk.NewFramebuffer(f.rpSplit, []*vk.ImageView{colorView, normalView, materialView, depthView})
	}).(*vk.Framebuffer)
	fbFinal := rc.Get(kFpFinal, func() interface{} {
		return vk.NewFramebuffer(f.rpFinal, []*vk.ImageView{mainView})
	}).(*vk.Framebuffer)
	fbBG := rc.Get(kFpBG, func() interface{} {
		return vk.NewFramebuffer(f.rpBG, []*vk.ImageView{colorView})
	}).(*vk.Framebuffer)

	start := time.Now()
	var tp *vk.TimerPool
	if f.timedOutput != nil {
		tp = vk.NewTimerPool(rc.Device, 3)
		rc.SetPerFrame(kTimer, tp)
		timerCmd := vk.NewCommand(rc.Device, vk.QUEUEComputeBit, true)
		timerCmd.Begin()
		timerCmd.WriteTimer(tp, 0, vk.PIPELINEStageTopOfPipeBit)
		infos = append(infos, timerCmd.SubmitForWait(1, vk.PIPELINEStageTopOfPipeBit))
		rc.SetPerFrame(kTimerCmd, timerCmd)
	}
	cmd := rc.Get(kCmd, func() interface{} {
		return vk.NewCommand(f.dev, vk.QUEUEGraphicsBit, false)
	}).(*vk.Command)
	cmd.Begin()
	if f.timedOutput != nil {
		cmd.WriteTimer(tp, 1, vk.PIPELINEStageTopOfPipeBit)
	}
	frame := &DeferredFrame{dsLight: dsLight, dsDraw: f.dsDraw[imageIndex], imagesUsed: 4, cache: rc, renderer: f}
	frame.bfLightsFrame = f.frameBuffers[imageIndex*2]
	frame.bfDrawFrame = f.frameBuffers[imageIndex*2+1]
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
	transparent := vscene.NewDrawPhase(frame, f.rpFinal, vscene.LAYERTransparent, cmd, nil, nil)
	ui := vscene.NewDrawPhase(frame, f.rpFinal, vscene.LAYERUI, cmd, func() {
	}, func() {

		cmd.EndRenderPass()
	})
	ppPhase := &vscene.PredrawPhase{Scene: sc, Cmd: cmd}

	sc.Process(sc.Time, frame, &vscene.AnimatePhase{}, ppPhase, bgPhase, splitPhase, join, transparent, ui)

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
		f.timedOutput(start, tp.Get())
	}
	if f.RenderDone != nil {
		f.RenderDone(start)
	}

}
