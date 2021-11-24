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

type EnvFrame interface {
	// Add probe to frame
	AddEnvironment(SPH [9]mgl32.Vec4, ubfImage vmodel.ImageIndex, pi *vscene.ProcessInfo)
}

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

func NewProbe(dev *vk.Device) *Probe {
	p := &Probe{}
	p.setup(dev)
	return p
}

func (p *Probe) Process(pi *vscene.ProcessInfo) {
	pre, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		ev, ok := pi.Frame.(EnvFrame)
		if !ok {
			return
		}
		cache := pi.Frame.GetCache()
		if p.needUpdate {
			p.renderProbe(cache, pre.Scene, pi.World.Col(3).Vec3())
			// p.saveImg(pre.Cache, "d:/temp/prope.dds")
		}
		sampler := getEnvSampler(cache.Device)
		imFrame, ok := pi.Frame.(vscene.ImageFrame)
		if ok {
			idx := imFrame.AddFrameImage(p.imgs[p.currentImg].NewCubeView(-1), sampler)
			ev.AddEnvironment(p.SPH, idx, pi)
		} else {
			ev.AddEnvironment(p.SPH, 0, pi)
		}
	}
	dp, ok := pi.Phase.(*drawProbe)
	if ok && dp.p == p {
		pi.Visible = false
	}
	// TODO: Add multi probe support
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
	return d.DrawContext.Frame.GetCache()
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

func (p *Probe) setup(dev *vk.Device) {
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
		img := p.pool.ReserveImage(p.desc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageSampledBit|
			vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageStorageBit)
		p.imgs = append(p.imgs, img)
	}
	p.pool.Allocate()
	// p.slUbf = ubf.Slice( 0, 256)
	p.frp = vk.NewForwardRenderPass(dev, p.desc.Format, vk.IMAGELayoutGeneral, vk.FORMATD32Sfloat)
	// lComp := p.getProbeLayout( dev)
	// p.dsPool = vk.NewDescriptorPool( lComp, 1)
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

	p.memPool = vk.NewMemoryPool(p.rcParent.Device)
	depthDesc := vk.ImageDescription{Width: p.p.desc.Width, Height: p.p.desc.Height, Format: vk.FORMATD32Sfloat,
		MipLevels: 1, Depth: 1, Layers: 1}
	var depthImages []*vk.Image
	for layer := int32(0); layer < 6; layer++ {
		depthImage := p.memPool.ReserveImage(depthDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		depthImages = append(depthImages, depthImage)
	}
	subImages := 6 * int(p.p.desc.MipLevels)
	ubf := p.memPool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment*uint64(subImages),
		true, vk.BUFFERUsageUniformBufferBit)
	p.sphBuf = p.memPool.ReserveBuffer(SPHUnits*16*9, true, vk.BUFFERUsageStorageBufferBit)
	p.sphBuf2 = p.memPool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit)

	p.memPool.Allocate()
	lComp := p.p.getProbeLayout(p.rcParent.Device)
	p.mipPool = vk.NewDescriptorPool(lComp, subImages)

	for idx := 0; idx < subImages; idx++ {
		p.mipUbfs = append(p.mipUbfs, ubf.Slice(
			vk.MinUniformBufferOffsetAlignment*uint64(idx), vk.MinUniformBufferOffsetAlignment*uint64(idx+1)-1))
		p.mipDSs = append(p.mipDSs, p.mipPool.Alloc())
		p.mipDSs[idx].WriteSlice(0, 0, p.mipUbfs[idx])
	}
	lSHP := p.p.getSPHLayout(p.rcParent.Device)
	p.dsSPHPool = vk.NewDescriptorPool(lSHP, 1)
	p.dsSPH = p.dsSPHPool.Alloc()
	// p.ds.WriteSlice( 0, 0, p.slUbf)
	p.dsSPH.WriteBuffer(0, 0, p.sphBuf2)
	p.dsSPH.WriteBuffer(2, 0, p.sphBuf)

	cmd := vk.NewCommand(p.rcParent.Device, vk.QUEUEGraphicsBit, true)
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
	lr := &layerRender{rc: vk.NewRenderCache(p.rcParent.Device)}
	p.subRenderers = append(p.subRenderers, lr)
	lr.rc.NewFrame()
	lr.iv = p.image.NewView(layer, 0)
	lr.fb = vk.NewFramebuffer(p.p.frp, []*vk.ImageView{lr.iv, depthImage.DefaultView()})

	f := &vscene.SimpleFrame{Cache: lr.rc}
	dc1 := &drawProbe{DrawContext: vmodel.DrawContext{Frame: f, Pass: p.p.frp}, cmd: cmd,
		p: p.p, layer: vscene.LAYERBackground, fb: lr.fb}
	dc2 := &drawProbe{DrawContext: dc1.DrawContext, cmd: cmd, p: p.p, layer: vscene.LAYER3DProbe}
	size := image.Pt(int(p.p.desc.Width), int(p.p.desc.Height))
	f.SSF.Projection, f.SSF.View = pc.CameraProjection(size)
	p.sc.Process(0, f, dc1, dc2)

}

func (p *Probe) saveImg(rc *vk.RenderCache, path string) error {
	cp := vmodel.NewCopier(rc.Device)
	content, err := cp.CopyFromImage(p.imgs[p.currentImg], p.imgs[p.currentImg].FullRange(), "dds", vk.IMAGELayoutShaderReadOnlyOptimal)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, content, 0660)
	if err != nil {
		return err
	}
	fmt.Println("Saved probe to ", path)
	return nil
}

var kProbeLayouts = vk.NewKeys(4)

func (p *probeRender) renderSubimage(siIndex int, cmd *vk.Command, layer int32, mip uint32) {
	plProbe := p.p.getProbePipeline(p.rcParent.Device)
	w := p.p.desc.Width >> mip
	h := p.p.desc.Height >> mip
	fls := []float32{float32(2.2), float32(1.0), float32(w * 2), float32(h * 2)}
	copy(p.mipUbfs[siIndex].Content, vk.Float32ToBytes(fls))
	rOutput := vk.ImageRange{LevelCount: 1, LayerCount: 1, FirstMipLevel: mip, FirstLayer: uint32(layer), Layout: vk.IMAGELayoutGeneral}
	rInput := rOutput
	rInput.FirstMipLevel--
	vInput := vk.NewImageView(p.image, &rInput)
	p.views = append(p.views, vInput)
	vOutput := vk.NewImageView(p.image, &rOutput)
	p.views = append(p.views, vOutput)
	p.mipDSs[siIndex].WriteImage(1, 0, vInput, nil)
	p.mipDSs[siIndex].WriteImage(2, 0, vOutput, nil)
	rOutput.Layout = vk.IMAGELayoutUndefined
	cmd.SetLayout(p.image, &rOutput, vk.IMAGELayoutGeneral)
	cmd.Compute(plProbe, w/32+1, h/32+1, 1, p.mipDSs[siIndex])
}

func (p *Probe) getProbeLayout(dev *vk.Device) *vk.DescriptorLayout {
	l1 := dev.Get(kProbeLayouts, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l2 := dev.Get(kProbeLayouts+1, func() interface{} {
		return l1.AddBinding(vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l3 := dev.Get(kProbeLayouts+2, func() interface{} {
		return l2.AddBinding(vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return l3
}

func (p *Probe) getProbePipeline(dev *vk.Device) *vk.ComputePipeline {
	lComp := p.getProbeLayout(dev)
	return dev.Get(kBlurPipeline, func() interface{} {
		cp := vk.NewComputePipeline(dev)
		cp.AddLayout(lComp)
		cp.AddShader(probe_comp_spv)
		cp.Create()
		return cp
	}).(*vk.ComputePipeline)
}

func (p *Probe) renderProbe(rcParent *vk.RenderCache, scene *vscene.Scene, pos mgl32.Vec3) {
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
