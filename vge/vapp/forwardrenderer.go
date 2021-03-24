package vapp

import (
	"runtime"
	"time"

	"github.com/lakal3/vge/vge/materials/predepth"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type ForwardRenderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

	// GPSTiming if set, records GPU timings for each frame
	// First value is start time, second at start of main cmd, third at end of main cmd
	// NOTE! First timing if from different submit and maybe different submit queue. See Vulkan documentation for vkCmdWriteTimestamp.
	GPUTiming func([]float64)

	owner        vk.Owner
	dev          *vk.Device
	Ctx          vk.APIContext
	depth        bool
	frp          *vk.ForwardRenderPass
	mpDepth      *vk.MemoryPool
	imDepth      []*vk.Image
	depthPrePass bool
}

func NewForwardRenderer(depthBuffer bool) *ForwardRenderer {
	return &ForwardRenderer{depth: depthBuffer}
}

var kImageViews = vk.NewKeys(10)
var kFp = vk.NewKey()
var kCmd = vk.NewKey()

func (f *ForwardRenderer) Dispose() {
	if f.mpDepth != nil {
		f.mpDepth.Dispose()
		f.mpDepth = nil
	}
	if f.frp != nil {
		f.frp.Dispose()
		f.frp = nil
	}

}

func (f *ForwardRenderer) GetRenderPass() vk.RenderPass {
	return f.frp
}

func (f *ForwardRenderer) AddDepthPrePass() *ForwardRenderer {
	f.depthPrePass = true
	return f
}

func (f *ForwardRenderer) Setup(ctx vk.APIContext, dev *vk.Device, mainImage vk.ImageDescription, images int) {
	fDepth := vk.FORMATUndefined
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

func (f *ForwardRenderer) Render(sc *vscene.Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo) {
	mainView := rc.Get(kImageViews+vk.Key(imageIndex), func(ctx vk.APIContext) interface{} {
		return mainImage.NewView(ctx, 0, 0)
	}).(*vk.ImageView)
	var depthView *vk.ImageView
	if f.depth {
		depthView = f.imDepth[imageIndex].DefaultView(rc.Ctx)
	}
	f.RenderView(sc, rc, mainView, depthView, infos)
}

var kTimer = vk.NewKey()
var kTimerCmd = vk.NewKey()

func (f *ForwardRenderer) RenderView(sc *vscene.Scene, rc *vk.RenderCache, mainView *vk.ImageView, depthView *vk.ImageView, infos []vk.SubmitInfo) {
	fb := rc.Get(kFp, func(ctx vk.APIContext) interface{} {
		if f.depth {
			return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView, depthView})
		}
		return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView})
	}).(*vk.Framebuffer)
	start := time.Now()
	var tp *vk.TimerPool
	if f.GPUTiming != nil {
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
	if f.GPUTiming != nil {
		cmd.WriteTimer(tp, 1, vk.PIPELINEStageTopOfPipeBit)
	}
	bg := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERBackground, cmd, func() {
		if !f.depthPrePass {
			cmd.BeginRenderPass(f.frp, fb)
		}
	}, nil)
	dp := vscene.NewDrawPhase(rc, f.frp, vscene.LAYER3D, cmd, nil, nil)
	dt := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERTransparent, cmd, nil, nil)
	ui := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERUI, cmd, nil, func() {
		cmd.EndRenderPass()
	})
	frame := vscene.GetFrame(rc)
	ppPhase := &vscene.PredrawPhase{Scene: sc, F: frame, Cache: rc, Cmd: cmd}
	if f.depthPrePass {
		pdp := &predepth.PreDepthPass{Cmd: cmd, DC: vmodel.DrawContext{Cache: rc, Pass: f.frp}}
		pdp.OnBegin = func() {
			cmd.BeginRenderPass(f.frp, fb)
		}
		sc.Process(sc.Time, &vscene.AnimatePhase{}, ppPhase, pdp, bg, dp, dt, ui)
	} else {
		sc.Process(sc.Time, &vscene.AnimatePhase{}, ppPhase, bg, dp, dt, ui)
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
		f.GPUTiming(tp.Get(rc.Ctx))
	}
	if f.RenderDone != nil {
		f.RenderDone(start)
	}

}
