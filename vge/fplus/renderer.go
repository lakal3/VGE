//

//go:generate glslangValidator.exe -V copy.vert.glsl -o copy.vert.spv
//go:generate glslangValidator.exe -V copy.frag.glsl -o copy.frag.spv
//go:generate packspv -p fplus .

package fplus

import (
	"errors"
	"github.com/lakal3/vge/vge/materials/predepth"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"runtime"
	"time"
)

const extraPhases = 2

type Renderer struct {
	// RenderDone is an optional function that is called each time after completing rendering of scene
	RenderDone func(started time.Time)

	imageFilter ImageFilter
	owner       vk.Owner
	dev         *vk.Device
	Ctx         vk.APIContext
	frp         *vk.FPlusRenderPass
	imagePool   *vk.MemoryPool
	imDepth     []*vk.Image
	imTemp      []*vk.Image
	dsPool      *vk.DescriptorPool
	// One set for each image and each extra phase
	dsSet        []*vk.DescriptorSet
	depthPrePass bool
}

// GetPhaseDescriptor retrieves descriptor for 2 image array.
// First image is output from previous phase, second image is depth buffer from first phase
func GetPhaseDescriptor(rc *vk.RenderCache) *vk.DescriptorSet {
	raw := rc.GetPerFrame(kPhaseDescriptor, func(ctx vk.APIContext) interface{} {
		return nil
	})
	if raw == nil {
		rc.Ctx.SetError(errors.New("Invalid phase to call GetPhaseDescriptor"))
		return nil
	}
	return raw.(*vk.DescriptorSet)
}

func NewRenderer(filter ImageFilter) *Renderer {
	return &Renderer{imageFilter: filter}
}

var kImageViews = vk.NewKeys(10)
var kFp = vk.NewKey()
var kCmd = vk.NewKey()
var kLayout = vk.NewKey()
var kSampler = vk.NewKey()
var kPhaseDescriptor = vk.NewKey()

func (f *Renderer) Dispose() {
	if f.imagePool != nil {
		f.imagePool.Dispose()
		f.imagePool = nil
	}
	if f.dsPool != nil {
		f.dsPool.Dispose()
		f.dsPool, f.dsSet = nil, nil
	}
	if f.frp != nil {
		f.frp.Dispose()
		f.frp = nil
	}

}

func (f *Renderer) GetRenderPass() vk.RenderPass {
	return f.frp
}

func (f *Renderer) AddDepthPrePass() *Renderer {
	f.depthPrePass = true
	return f
}

func (f *Renderer) Setup(ctx vk.APIContext, dev *vk.Device, mainImage vk.ImageDescription, images int) {
	fDepth := vk.FORMATD32Sfloat
	if f.frp != nil {
		f.imagePool.Dispose()
		f.imDepth, f.imTemp = nil, nil
		f.dsPool.Dispose()
		f.dsPool, f.dsSet = nil, nil
	} else {
		f.Ctx, f.dev = ctx, dev
		f.frp = vk.NewFPlusRenderPass(ctx, dev, extraPhases, mainImage.Format, vk.IMAGELayoutPresentSrcKhr, fDepth)
	}
	depthDesc := mainImage
	depthDesc.Format = vk.FORMATD32Sfloat
	la := GetBindImageLayout(ctx, dev)
	f.dsPool = vk.NewDescriptorPool(ctx, la, extraPhases*images)
	f.imagePool = vk.NewMemoryPool(dev)
	for idx := 0; idx < images; idx++ {
		f.imDepth = append(f.imDepth, f.imagePool.ReserveImage(ctx, depthDesc, vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageSampledBit))
	}
	for idx := 0; idx < images*extraPhases; idx++ {
		f.imTemp = append(f.imTemp, f.imagePool.ReserveImage(ctx, depthDesc,
			vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferSrcBit))
		f.dsSet = append(f.dsSet, f.dsPool.Alloc(ctx))
	}
	f.imagePool.Allocate(ctx)
}

func (f *Renderer) Render(sc *vscene.Scene, rc *vk.RenderCache, mainImage *vk.Image, imageIndex int, infos []vk.SubmitInfo) {
	mainView := rc.Get(kImageViews+vk.Key(imageIndex), func(ctx vk.APIContext) interface{} {
		return mainImage.NewView(ctx, 0, 0)
	}).(*vk.ImageView)
	views := make([]*vk.ImageView, 0, 2+extraPhases)
	views = append(views, mainView)
	for i := 0; i < extraPhases; i++ {
		views = append(views, f.imTemp[i+imageIndex*extraPhases].DefaultView(rc.Ctx))
	}
	views = append(views, f.imDepth[imageIndex].DefaultView(rc.Ctx))
	f.RenderView(sc, rc, views, imageIndex, infos)
}

func (f *Renderer) RenderView(sc *vscene.Scene, rc *vk.RenderCache, views []*vk.ImageView, imageIndex int, infos []vk.SubmitInfo) {
	fb := rc.Get(kFp, func(ctx vk.APIContext) interface{} {
		return vk.NewFramebuffer(ctx, f.frp, views)
	}).(*vk.Framebuffer)
	sa := GetSampler(rc.Ctx, rc.Device)
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
	dpl := vscene.NewPostProcessPhase(rc, f.frp, cmd, func(pp *vscene.PostProcessPhase) {
		cmd.NextSubpass(rc.Ctx)
		ds := f.dsSet[imageIndex*extraPhases]
		ds.WriteImage(rc.Ctx, 0, 0, f.imTemp[imageIndex*extraPhases].DefaultView(rc.Ctx), sa)
		ds.WriteImage(rc.Ctx, 0, 1, f.imDepth[imageIndex].DefaultView(rc.Ctx), sa)
		rc.SetPerFrame(kPhaseDescriptor, ds)
		copySrc(&pp.DrawContext)
	})

	ui := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERUI, cmd, nil, func() {
		cmd.EndRenderPass()
	})
	frame := vscene.GetFrame(rc)
	ppPhase := &vscene.PredrawPhase{Scene: sc, F: frame, Cache: rc, Cmd: cmd}
	dpg := &filterPass{DrawContext: vmodel.DrawContext{Cache: rc, Pass: f.frp}}
	if f.depthPrePass {
		pdp := &predepth.PreDepthPass{Cmd: cmd, DC: vmodel.DrawContext{Cache: rc, Pass: f.frp}}
		pdp.OnBegin = func() {
			cmd.BeginRenderPass(f.frp, fb)
		}
		sc.Process(sc.Time, &vscene.AnimatePhase{}, ppPhase, pdp, bg, dp, dpl, dpg, ui)
	} else {
		sc.Process(sc.Time, &vscene.AnimatePhase{}, ppPhase, bg, dp, dpl, dpg, ui)
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

func GetSampler(ctx vk.APIContext, dev *vk.Device) *vk.Sampler {
	return dev.Get(ctx, kSampler, func(ctx vk.APIContext) interface{} {
		return vk.NewSampler(ctx, dev, vk.SAMPLERAddressModeClampToEdge)
	}).(*vk.Sampler)
}

// GetBindImageLayout describe layout with array of 2 images, color attachment from previous pass and depth attachment from first pass
func GetBindImageLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(ctx, kLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 2)
	}).(*vk.DescriptorLayout)
}

type filterPass struct {
	vmodel.DrawContext
	r   *Renderer
	cmd *vk.Command
}

func (f *filterPass) Begin() (atEnd func()) {
	f.r.imageFilter.Filter(&f.DrawContext)
	return func() {
		f.cmd.Draw(f.List)
	}
}
