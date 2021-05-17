// Forward package implements standard Forward renderer

package forward

import (
	"image"
	"runtime"
	"time"

	"github.com/lakal3/vge/vge/materials/predepth"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

// Renderer is forward renderer that handles individual phases of rendering and also gathers frame relevant settings like camera position
// Renderer has quite limitted options. If you want different phases etc, you should copy renderer logic and match it to your needs.
// See example webview on how to implement custom renderer.
type Renderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

	// API context attached to renderer
	Ctx vk.APIContext

	timedOutput func(started time.Time, gpuTimes []float64)

	size         image.Point
	owner        vk.Owner
	dev          *vk.Device
	depth        bool
	frp          *vk.ForwardRenderPass
	mpDepth      *vk.MemoryPool
	imDepth      []*vk.Image
	depthPrePass bool
}

func (f *Renderer) GetPerRenderer(key vk.Key, ctor func(ctx vk.APIContext) interface{}) interface{} {
	return f.owner.Get(f.Ctx, key, ctor)
}

// SetTimedOutput allow you to set function that is called after each frame. Started will be Go time when frame rendering was started
// GPU times will contain gpu timer as start, before main render pass and after rendering all phases.
func (f *Renderer) SetTimedOutput(output func(started time.Time, gpuTimes []float64)) {
	f.timedOutput = output
}

// NewRenderer create new forward renderer.
// DepthBuffer settings will if renderer uses depth buffer to limit visibility of objects behind other objects. This should be false only if you try
// to render something more or less 2D
func NewRenderer(depthBuffer bool) *Renderer {
	return &Renderer{depth: depthBuffer}
}

var kImageViews = vk.NewKeys(10)
var kFp = vk.NewKey()
var kCmd = vk.NewKey()

func (f *Renderer) Dispose() {
	if f.mpDepth != nil {
		f.mpDepth.Dispose()
		f.mpDepth = nil
	}
	if f.frp != nil {
		f.frp.Dispose()
		f.frp = nil
	}

}

func (f *Renderer) GetRenderPass() vk.RenderPass {
	return f.frp
}

// AddDepthPrePass will render z-buffer with very slight offset back before rendering scene. This should speed up
// rendering scene with lots of lights as we don't calculate expensive light calculation for most of pixel that will be culled of by depth check
// (they are behind an other object).
// Alternatively you can try new deferred.Renderer for scenes with lots of lights
func (f *Renderer) AddDepthPrePass() *Renderer {
	f.depthPrePass = true
	return f
}

func (f *Renderer) Setup(ctx vk.APIContext, dev *vk.Device, mainImage vk.ImageDescription, images int) {
	fDepth := vk.FORMATUndefined
	f.size.X, f.size.Y = int(mainImage.Width), int(mainImage.Height)
	if f.depth {
		fDepth = vk.FORMATD32Sfloat
	}
	if f.frp != nil {
		if f.depth {
			f.mpDepth.Dispose()
			f.imDepth = nil
		}
	} else {
		f.Ctx, f.dev = ctx, dev
		f.frp = vk.NewForwardRenderPass(ctx, dev, mainImage.Format, vk.IMAGELayoutPresentSrcKhr, fDepth)
	}
	if f.depth {
		depthDesc := mainImage
		depthDesc.Format = vk.FORMATD32Sfloat
		f.mpDepth = vk.NewMemoryPool(dev)
		for idx := 0; idx < images; idx++ {
			f.imDepth = append(f.imDepth, f.mpDepth.ReserveImage(ctx, depthDesc, vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageTransferSrcBit))
		}
		f.mpDepth.Allocate(ctx)
	}
}

func (f *Renderer) Render(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo) {
	mainView := rc.Get(kImageViews+vk.Key(imageIndex), func(ctx vk.APIContext) interface{} {
		return mainImage.NewView(ctx, 0, 0)
	}).(*vk.ImageView)
	var depthView *vk.ImageView
	if f.depth {
		depthView = f.imDepth[imageIndex].DefaultView(rc.Ctx)
	}
	f.RenderView(camera, sc, rc, mainView, depthView, infos)
}

var kTimer = vk.NewKey()
var kTimerCmd = vk.NewKey()

func (f *Renderer) RenderView(camera vscene.Camera, sc *vscene.Scene, rc *vk.RenderCache, mainView *vk.ImageView, depthView *vk.ImageView, infos []vk.SubmitInfo) {
	fb := rc.Get(kFp, func(ctx vk.APIContext) interface{} {
		if f.depth {
			return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView, depthView})
		}
		return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView})
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
	frame := &Frame{cache: rc, renderer: f}
	frame.SF.Projection, frame.SF.View = camera.CameraProjection(f.size)
	frame.SF.EyePos = frame.SF.View.Inv().Col(3)
	bg := vscene.NewDrawPhase(frame, f.frp, vscene.LAYERBackground, cmd, func() {
		if !f.depthPrePass {
			cmd.BeginRenderPass(f.frp, fb)
		}
	}, nil)
	dp := vscene.NewDrawPhase(frame, f.frp, vscene.LAYER3D, cmd, nil, nil)
	dt := vscene.NewDrawPhase(frame, f.frp, vscene.LAYERTransparent, cmd, nil, nil)
	ui := vscene.NewDrawPhase(frame, f.frp, vscene.LAYERUI, cmd, nil, func() {
		cmd.EndRenderPass()
	})
	ppPhase := &vscene.PredrawPhase{Scene: sc, Cmd: cmd}
	lightPhase := FrameLightPhase{F: frame, Cache: rc}
	if f.depthPrePass {
		pdp := &predepth.PreDepthPass{Cmd: cmd, DC: vmodel.DrawContext{Frame: frame, Pass: f.frp}}
		pdp.BindFrame = func() *vk.DescriptorSet {
			return frame.BindForwardFrame()
		}
		pdp.OnBegin = func() {
			cmd.BeginRenderPass(f.frp, fb)
		}
		sc.Process(sc.Time, frame, &vscene.AnimatePhase{}, ppPhase, lightPhase, pdp, bg, dp, dt, ui)
	} else {
		sc.Process(sc.Time, frame, &vscene.AnimatePhase{}, ppPhase, lightPhase, bg, dp, dt, ui)
	}
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
