package env

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"image"
	"io/ioutil"
	"math"
)

type Probe struct {
	pool       *vk.MemoryPool
	desc       vk.ImageDescription
	imgs       []*vk.Image
	frp        *vk.ForwardRenderPass
	currentImg int

	SPH        [9]mgl32.Vec4
	needUpdate bool
	indexKey   vk.Key
}

// KProbe can be used to access current probe from Phase
var KProbe = vk.NewKey()

func (p *Probe) Update() {
	p.needUpdate = true
}

func (p *Probe) Dispose() {

	if p.pool != nil {
		p.pool.Dispose()
		p.pool, p.imgs = nil, nil
	}
	if p.frp != nil {
		p.frp.Dispose()
		p.frp = nil
	}
}

func NewProbe(ctx vk.APIContext, dev *vk.Device) *Probe {
	p := &Probe{}
	p.setup(ctx, dev)
	return p
}

func (p *Probe) Process(pi *vscene.ProcessInfo) {
	pre, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		if p.needUpdate {
			p.renderProbe(pre.Cache.Ctx, pre.Cache, pre.Scene, pi.World.Col(3).Vec3())
			// p.saveImg(pre.Cache, "d:/temp/prope.dds")
		}
		sampler := getEnvSampler(pre.Cache.Ctx, pre.Cache.Device)
		idx := pre.Frame.AddFrameImage(pre.Cache, p.imgs[p.currentImg].NewCubeView(pre.Cache.Ctx, -1), sampler)
		if idx >= 0 {
			probeIndex := pre.Frame.AddProbe(p.SPH, idx)
			pre.Cache.SetPerFrame(p.indexKey, probeIndex)
		} else {
			pre.Cache.SetPerFrame(p.indexKey, -1)
		}
	}
	dp, ok := pi.Phase.(*drawProbe)
	if ok && dp.p == p {
		pi.Visible = false
	} else {
		ci := pi.Phase.GetCache()
		if ci == nil {
			return
		}
		probeIndex := ci.GetPerFrame(p.indexKey, func(ctx vk.APIContext) interface{} {
			return -1
		}).(int)
		if probeIndex >= 0 {
			pi.Set(KProbe, probeIndex)
		}
	}
}

type probeSettings struct {
	gamma  float32
	white  float32
	width  float32
	height float32
}

type drawProbe struct {
	vmodel.DrawContext
	layer vscene.Layer
	cmd   *vk.Command
	fb    *vk.Framebuffer
	p     *Probe
}

func (d *drawProbe) GetCache() *vk.RenderCache {
	return d.DrawContext.Cache
}

func (d *drawProbe) Begin() (atEnd func()) {
	if d.layer == vscene.LAYERBackground {
		d.cmd.BeginRenderPass(d.Pass, d.fb)
	}
	return func() {
		if d.DrawContext.List != nil {
			d.cmd.Draw(d.DrawContext.List)
		}
		if d.layer == vscene.LAYER3DProbe {
			d.cmd.EndRenderPass()
		}
	}
}

func (d *drawProbe) GetDC(layer vscene.Layer) *vmodel.DrawContext {
	if d.layer == layer {
		return &d.DrawContext
	}
	return nil
}

var kBlurPipeline = vk.NewKey()

func (p *Probe) setup(ctx vk.APIContext, dev *vk.Device) {
	p.indexKey = vk.NewKey()
	p.currentImg, p.needUpdate = -1, true
	p.pool = vk.NewMemoryPool(dev)
	p.desc = vk.ImageDescription{
		Width:     1024,
		Height:    1024,
		Depth:     1,
		Format:    vk.FORMATR8g8b8a8Unorm,
		Layers:    6,
		MipLevels: 6,
	}
	for idx := 0; idx < 2; idx++ {
		img := p.pool.ReserveImage(ctx, p.desc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageSampledBit|
			vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageStorageBit)
		p.imgs = append(p.imgs, img)
	}
	p.pool.Allocate(ctx)
	// p.slUbf = ubf.Slice(ctx, 0, 256)
	p.frp = vk.NewForwardRenderPass(ctx, dev, p.desc.Format, vk.IMAGELayoutGeneral, vk.FORMATD32Sfloat)
	// lComp := p.getProbeLayout(ctx, dev)
	// p.dsPool = vk.NewDescriptorPool(ctx, lComp, 1)
	// p.ds = p.dsPool.Alloc(ctx)
}

type probeRender struct {
	p        *Probe
	rcParent *vk.RenderCache
	sc       *vscene.Scene
	probePos mgl32.Vec3
	image    *vk.Image
	// rc       *vk.RenderCache
	memPool   *vk.MemoryPool
	mipUbfs   []*vk.Slice
	mipPool   *vk.DescriptorPool
	mipDSs    []*vk.DescriptorSet
	sphBuf    *vk.Buffer
	dsSPHPool *vk.DescriptorPool
	dsSPH     *vk.DescriptorSet
	sphBuf2   *vk.Buffer

	views        []*vk.ImageView
	subRenderers []*layerRender
}

func (p *probeRender) Dispose() {
	for _, v := range p.subRenderers {
		v.Dispose()
	}
	p.subRenderers = nil
	if p.memPool != nil {
		p.memPool.Dispose()
		p.memPool = nil
	}
	if p.mipPool != nil {
		p.mipPool.Dispose()
		p.mipPool, p.mipDSs = nil, nil
	}
	for _, v := range p.views {
		v.Dispose()
	}
	p.views = nil

}

var kRpProbe = vk.NewKey()

func (p *probeRender) renderProbe() {
	ctx := p.rcParent.Ctx
	p.memPool = vk.NewMemoryPool(p.rcParent.Device)
	depthDesc := vk.ImageDescription{Width: p.p.desc.Width, Height: p.p.desc.Height, Format: vk.FORMATD32Sfloat,
		MipLevels: 1, Depth: 1, Layers: 1}
	var depthImages []*vk.Image
	for layer := int32(0); layer < 6; layer++ {
		depthImage := p.memPool.ReserveImage(ctx, depthDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		depthImages = append(depthImages, depthImage)
	}
	subImages := 6 * int(p.p.desc.MipLevels)
	ubf := p.memPool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment*uint64(subImages),
		true, vk.BUFFERUsageUniformBufferBit)
	p.sphBuf = p.memPool.ReserveBuffer(ctx, SPHUnits*16*9, true, vk.BUFFERUsageStorageBufferBit)
	p.sphBuf2 = p.memPool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit)

	p.memPool.Allocate(ctx)
	lComp := p.p.getProbeLayout(ctx, p.rcParent.Device)
	p.mipPool = vk.NewDescriptorPool(ctx, lComp, subImages)

	for idx := 0; idx < subImages; idx++ {
		p.mipUbfs = append(p.mipUbfs, ubf.Slice(ctx,
			vk.MinUniformBufferOffsetAlignment*uint64(idx), vk.MinUniformBufferOffsetAlignment*uint64(idx+1)-1))
		p.mipDSs = append(p.mipDSs, p.mipPool.Alloc(ctx))
		p.mipDSs[idx].WriteSlice(ctx, 0, 0, p.mipUbfs[idx])
	}
	lSHP := p.p.getSPHLayout(ctx, p.rcParent.Device)
	p.dsSPHPool = vk.NewDescriptorPool(ctx, lSHP, 1)
	p.dsSPH = p.dsSPHPool.Alloc(ctx)
	// p.ds.WriteSlice(ctx, 0, 0, p.slUbf)
	p.dsSPH.WriteBuffer(ctx, 0, 0, p.sphBuf2)
	p.dsSPH.WriteBuffer(ctx, 2, 0, p.sphBuf)

	cmd := vk.NewCommand(ctx, p.rcParent.Device, vk.QUEUEGraphicsBit, true)
	cmd.Begin()
	for layer := int32(0); layer < 6; layer++ {
		p.renderMainLayer(cmd, layer, depthImages[layer])
	}
	var siIndex = 0
	for layer := int32(0); layer < 6; layer++ {
		for mip := uint32(1); mip < p.p.desc.MipLevels; mip++ {
			p.renderSubimage(siIndex, cmd, layer, mip)
			siIndex++
		}
	}
	p.calcSPH(cmd)
	r := p.image.FullRange()
	r.Layout = vk.IMAGELayoutGeneral
	cmd.SetLayout(p.image, &r, vk.IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
	p.accumulateSPH()
}

func (p *probeRender) getCamera(layer int32) *vscene.PerspectiveCamera {
	pc := vscene.NewPerspectiveCamera(1000)
	pc.Position = p.probePos
	pc.FoV = math.Pi / 2

	switch layer {
	case 0:
		pc.Target = pc.Position.Add(mgl32.Vec3{-1, 0, 0})
	case 1:
		pc.Target = pc.Position.Add(mgl32.Vec3{1, 0, 0})
	case 2:
		pc.Target = pc.Position.Add(mgl32.Vec3{0, 1, 0})
		pc.Up = mgl32.Vec3{0, 0, -1}
	case 3:
		pc.Target = pc.Position.Add(mgl32.Vec3{0, -1, 0})
		pc.Up = mgl32.Vec3{0, 0, 1}
	case 4:
		pc.Target = pc.Position.Add(mgl32.Vec3{0, 0, 1})
	case 5:
		pc.Target = pc.Position.Add(mgl32.Vec3{0, 0, -1})
	}
	return pc
}

type layerRender struct {
	rc *vk.RenderCache
	iv *vk.ImageView
	fb *vk.Framebuffer
}

func (l *layerRender) Dispose() {
	if l.fb != nil {
		l.fb.Dispose()
		l.fb = nil
	}
	if l.iv != nil {
		l.iv.Dispose()
		l.iv = nil
	}
	if l.rc != nil {
		l.rc.Dispose()
		l.rc = nil
	}
}

func (p *probeRender) renderMainLayer(cmd *vk.Command, layer int32, depthImage *vk.Image) {
	pc := p.getCamera(layer)
	ctx := p.rcParent.Ctx
	lr := &layerRender{rc: vk.NewRenderCache(ctx, p.rcParent.Device)}
	p.subRenderers = append(p.subRenderers, lr)
	lr.rc.NewFrame()
	lr.iv = p.image.NewView(ctx, layer, 0)
	lr.fb = vk.NewFramebuffer(ctx, p.p.frp, []*vk.ImageView{lr.iv, depthImage.DefaultView(ctx)})

	dc1 := &drawProbe{DrawContext: vmodel.DrawContext{Cache: lr.rc, Pass: p.p.frp}, cmd: cmd,
		p: p.p, layer: vscene.LAYERBackground, fb: lr.fb}
	dc2 := &drawProbe{DrawContext: dc1.DrawContext, cmd: cmd, p: p.p, layer: vscene.LAYER3DProbe}
	f := &vscene.SimpleFrame{}
	size := image.Pt(int(p.p.desc.Width), int(p.p.desc.Height))
	f.Projection, f.View = pc.CameraProjection(size)
	// Mirror image over x axis
	f.WriteFrame(lr.rc)
	p.sc.Process(0, dc1, dc2)

}

func (p *Probe) saveImg(rc *vk.RenderCache, path string) {
	cp := vmodel.NewCopier(rc.Ctx, rc.Device)
	content := cp.CopyFromImage(p.imgs[p.currentImg], p.imgs[p.currentImg].FullRange(), "dds", vk.IMAGELayoutShaderReadOnlyOptimal)
	err := ioutil.WriteFile(path, content, 0660)
	if err != nil {
		fmt.Println("Failed to save ", path, ": ", err)
	} else {
		fmt.Println("Saved probe to ", path)
	}
}

var kProbeLayouts = vk.NewKeys(4)

func (p *probeRender) renderSubimage(siIndex int, cmd *vk.Command, layer int32, mip uint32) {
	ctx := p.rcParent.Ctx
	plProbe := p.p.getProbePipeline(ctx, p.rcParent.Device)
	w := p.p.desc.Width >> mip
	h := p.p.desc.Height >> mip
	fls := []float32{float32(2.2), float32(1.0), float32(w * 2), float32(h * 2)}
	copy(p.mipUbfs[siIndex].Content, vk.Float32ToBytes(fls))
	rOutput := vk.ImageRange{LevelCount: 1, LayerCount: 1, FirstMipLevel: mip, FirstLayer: uint32(layer), Layout: vk.IMAGELayoutGeneral}
	rInput := rOutput
	rInput.FirstMipLevel--
	vInput := vk.NewImageView(ctx, p.image, &rInput)
	p.views = append(p.views, vInput)
	vOutput := vk.NewImageView(ctx, p.image, &rOutput)
	p.views = append(p.views, vOutput)
	p.mipDSs[siIndex].WriteImage(ctx, 1, 0, vInput, nil)
	p.mipDSs[siIndex].WriteImage(ctx, 2, 0, vOutput, nil)
	rOutput.Layout = vk.IMAGELayoutUndefined
	cmd.SetLayout(p.image, &rOutput, vk.IMAGELayoutGeneral)
	cmd.Compute(plProbe, w/32+1, h/32+1, 1, p.mipDSs[siIndex])
}

func (p *Probe) getProbeLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	l1 := dev.Get(ctx, kProbeLayouts, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l2 := dev.Get(ctx, kProbeLayouts+1, func(ctx vk.APIContext) interface{} {
		return l1.AddBinding(ctx, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l3 := dev.Get(ctx, kProbeLayouts+2, func(ctx vk.APIContext) interface{} {
		return l2.AddBinding(ctx, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return l3
}

func (p *Probe) getProbePipeline(ctx vk.APIContext, dev *vk.Device) *vk.ComputePipeline {
	lComp := p.getProbeLayout(ctx, dev)
	return dev.Get(ctx, kBlurPipeline, func(ctx vk.APIContext) interface{} {
		cp := vk.NewComputePipeline(ctx, dev)
		cp.AddLayout(ctx, lComp)
		cp.AddShader(ctx, probe_comp_spv)
		cp.Create(ctx)
		return cp
	}).(*vk.ComputePipeline)
}

func (p *Probe) renderProbe(ctx vk.APIContext, rcParent *vk.RenderCache, scene *vscene.Scene, pos mgl32.Vec3) {
	pr := probeRender{p: p, rcParent: rcParent, sc: scene, probePos: pos}
	defer pr.Dispose()
	p.currentImg++
	if p.currentImg >= len(p.imgs) {
		p.currentImg = 0
	}
	pr.image = p.imgs[p.currentImg]
	pr.renderProbe()
	p.needUpdate = false
}
