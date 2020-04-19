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
	dsPool     *vk.DescriptorPool
	ds         *vk.DescriptorSet
	slUbf      *vk.Slice
	sphBuf     *vk.Buffer
	dsSPHPool  *vk.DescriptorPool
	dsSPH      *vk.DescriptorSet
	SPH        [9]mgl32.Vec4
	needUpdate bool
}

func (p *Probe) Update() {
	p.needUpdate = true
}

func (p *Probe) Dispose() {
	if p.dsPool != nil {
		p.dsPool.Dispose()
		p.dsPool, p.ds = nil, nil
	}
	if p.pool != nil {
		p.pool.Dispose()
		p.pool, p.imgs, p.slUbf = nil, nil, nil
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
			p.renderProbe(pre.Cache.Ctx, pre.Cache.Device, pre.Scene, mgl32.Vec3{})
			// p.saveImg(pre.Cache, "d:/temp/prope.dds")
		}
		sampler := getEnvSampler(pre.Cache.Ctx, pre.Cache.Device)
		idx := vscene.SetFrameImage(pre.Cache, p.imgs[p.currentImg].NewCubeView(pre.Cache.Ctx, -1), sampler)
		f := vscene.GetFrame(pre.Cache)
		if idx >= 0 {
			f.EnvMap, f.EnvLods = float32(idx), float32(p.desc.MipLevels)
		}
		f.SPH = p.SPH
	}
	dp, ok := pi.Phase.(*drawProbe)
	if ok && dp.p == p {
		pi.Visible = false
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

func (d *drawProbe) Begin() (atEnd func()) {
	if d.layer == vscene.LAYERBackground {
		d.cmd.BeginRenderPass(d.Pass, d.fb)
	}
	return func() {
		if d.DrawContext.List != nil {
			d.cmd.Draw(d.DrawContext.List)
		}
		if d.layer == vscene.LAYER3D {
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
	ubf := p.pool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit)
	p.sphBuf = p.pool.ReserveBuffer(ctx, SPHUnits*16*9, true, vk.BUFFERUsageStorageBufferBit)
	p.pool.Allocate(ctx)
	p.slUbf = ubf.Slice(ctx, 0, 256)
	p.frp = vk.NewForwardRenderPass(ctx, dev, p.desc.Format, vk.IMAGELayoutGeneral, vk.FORMATD32Sfloat)
	lComp := p.getProbeLayout(ctx, dev)
	lSHP := p.getSPHLayout(ctx, dev)
	p.dsPool = vk.NewDescriptorPool(ctx, lComp, 1)
	p.ds = p.dsPool.Alloc(ctx)
	p.dsSPHPool = vk.NewDescriptorPool(ctx, lSHP, 1)
	p.dsSPH = p.dsSPHPool.Alloc(ctx)
	p.ds.WriteSlice(ctx, 0, 0, p.slUbf)
	p.dsSPH.WriteSlice(ctx, 0, 0, p.slUbf)
	p.dsSPH.WriteBuffer(ctx, 2, 0, p.sphBuf)
}

func (p *Probe) renderProbe(ctx vk.APIContext, dev *vk.Device, sc *vscene.Scene, probePos mgl32.Vec3) {
	rc := vk.NewRenderCache(ctx, dev)
	defer rc.Dispose()
	depthPool := vk.NewMemoryPool(dev)
	defer depthPool.Dispose()
	depthDesc := vk.ImageDescription{Width: p.desc.Width, Height: p.desc.Height, Format: vk.FORMATD32Sfloat,
		MipLevels: 1, Depth: 1, Layers: 1}
	depthImage := depthPool.ReserveImage(ctx, depthDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
	depthPool.Allocate(ctx)
	p.currentImg++
	if p.currentImg >= len(p.imgs) {
		p.currentImg = 0
	}
	for layer := int32(0); layer < 6; layer++ {
		p.renderMainLayer(rc, layer, sc, probePos, depthImage)
	}
	cmd := vk.NewCommand(ctx, dev, vk.QUEUEComputeBit, false)
	defer cmd.Dispose()
	for layer := int32(0); layer < 6; layer++ {
		for mip := uint32(1); mip < p.desc.MipLevels; mip++ {
			p.renderSubimage(rc, cmd, layer, mip)
		}
	}
	p.calcSPH(rc, cmd)
	r := p.imgs[p.currentImg].FullRange()
	r.Layout = vk.IMAGELayoutGeneral
	cmd.Begin()
	cmd.SetLayout(p.imgs[p.currentImg], &r, vk.IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
	p.needUpdate = false
}

func (p *Probe) getCamera(layer int32, pos mgl32.Vec3) *vscene.PerspectiveCamera {
	pc := vscene.NewPerspectiveCamera(1000)
	pc.FoV = math.Pi / 2
	switch layer {
	case 0:
		pc.Target = pc.Position.Add(mgl32.Vec3{1, 0, 0})
	case 1:
		pc.Target = pc.Position.Add(mgl32.Vec3{-1, 0, 0})
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

func (p *Probe) renderMainLayer(rc *vk.RenderCache, layer int32, sc *vscene.Scene, probePos mgl32.Vec3, depthImage *vk.Image) {
	pc := p.getCamera(layer, probePos)
	rc.NewFrame()
	img := p.imgs[p.currentImg]
	iv := img.NewView(rc.Ctx, layer, 0)
	defer iv.Dispose()
	fb := vk.NewFramebuffer(rc.Ctx, p.frp, []*vk.ImageView{iv, depthImage.DefaultView(rc.Ctx)})
	defer fb.Dispose()
	cmd := vk.NewCommand(rc.Ctx, rc.Device, vk.QUEUEGraphicsBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	dc1 := &drawProbe{DrawContext: vmodel.DrawContext{Cache: rc, Pass: p.frp}, cmd: cmd,
		p: p, layer: vscene.LAYERBackground, fb: fb}
	dc2 := &drawProbe{DrawContext: dc1.DrawContext, cmd: cmd, p: p, layer: vscene.LAYER3D}
	f := vscene.GetFrame(rc)
	pc.SetupFrame(f, image.Pt(int(p.desc.Width), int(p.desc.Height)))
	// Mirror image over x axis
	f.View = mgl32.Scale3D(-1, 1, 1).Mul4(f.View)
	sc.Process(0, dc1, dc2)
	cmd.Submit()
	cmd.Wait()
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

func (p *Probe) renderSubimage(rc *vk.RenderCache, cmd *vk.Command, layer int32, mip uint32) {
	plProbe := p.getProbePipeline(rc.Ctx, rc.Device)
	w := p.desc.Width >> mip
	h := p.desc.Height >> mip
	fls := []float32{float32(2.2), float32(1.0), float32(w * 2), float32(h * 2)}
	copy(p.slUbf.Content, vk.Float32ToBytes(fls))
	im := p.imgs[p.currentImg]
	rOutput := vk.ImageRange{LevelCount: 1, LayerCount: 1, FirstMipLevel: mip, FirstLayer: uint32(layer), Layout: vk.IMAGELayoutGeneral}
	rInput := rOutput
	rInput.FirstMipLevel--
	vInput := vk.NewImageView(rc.Ctx, im, &rInput)
	vOutput := vk.NewImageView(rc.Ctx, im, &rOutput)
	defer vInput.Dispose()
	defer vOutput.Dispose()
	p.ds.WriteImage(rc.Ctx, 1, 0, vInput, nil)
	p.ds.WriteImage(rc.Ctx, 2, 0, vOutput, nil)
	cmd.Begin()
	rOutput.Layout = vk.IMAGELayoutUndefined
	cmd.SetLayout(im, &rOutput, vk.IMAGELayoutGeneral)
	cmd.Compute(plProbe, w/32+1, h/32+1, 1, p.ds)
	cmd.Submit()
	cmd.Wait()
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
