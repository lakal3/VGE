package vapp

import (
	"github.com/lakal3/vge/vge/materials/predepth"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"runtime"
	"time"
)

type ForwardRenderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

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

func (f *ForwardRenderer) RenderView(sc *vscene.Scene, rc *vk.RenderCache, mainView *vk.ImageView, depthView *vk.ImageView, infos []vk.SubmitInfo) {
	fb := rc.Get(kFp, func(ctx vk.APIContext) interface{} {
		if f.depth {
			return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView, depthView})
		}
		return vk.NewFramebuffer(ctx, f.frp, []*vk.ImageView{mainView})
	}).(*vk.Framebuffer)
	start := time.Now()
	cmd := rc.Get(kCmd, func(ctx vk.APIContext) interface{} {
		return vk.NewCommand(f.Ctx, f.dev, vk.QUEUEGraphicsBit, false)
	}).(*vk.Command)
	cmd.Begin()

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
	cmd.Submit(infos...)
	cmd.Wait()
	runtime.KeepAlive(infos)
	if f.RenderDone != nil {
		f.RenderDone(start)
	}

}
