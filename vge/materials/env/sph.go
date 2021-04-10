package env

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

var kSPHLayout = vk.NewKeys(4)
var kSPHPipeline = vk.NewKey()

const SPHUnits = 32

func (p *probeRender) calcSPH(cmd *vk.Command) {
	rc := p.rcParent
	plSPH := p.p.getSPHPipeline(rc.Ctx, rc.Device)
	mip := p.p.desc.MipLevels - 1
	w := p.p.desc.Width >> mip
	h := p.p.desc.Height >> mip
	fls := []float32{float32(2.2), float32(1.0), float32(w * 4), float32(h * 2)}
	copy(p.sphBuf2.Bytes(rc.Ctx), vk.Float32ToBytes(fls))
	sampler := getEnvSampler(rc.Ctx, rc.Device)
	rInput := vk.ImageRange{LevelCount: 1, LayerCount: 6, FirstMipLevel: mip, FirstLayer: 0, Layout: vk.IMAGELayoutGeneral}
	// rInput.FirstMipLevel = 0
	im := p.p.imgs[p.p.currentImg]
	vInput := vk.NewCubeView(rc.Ctx, im, &rInput)
	p.views = append(p.views, vInput)
	p.dsSPH.WriteImage(rc.Ctx, 1, 0, vInput, sampler)
	cmd.Compute(plSPH, 1, 1, 1, p.dsSPH)
}

func (p *probeRender) accumulateSPH() {
	sphRaw := vk.BytesToFloat32(p.sphBuf.Bytes(p.rcParent.Ctx))
	for n := 0; n < 9; n++ {
		p.p.SPH[0] = mgl32.Vec4{}
	}
	var weight float32
	for idx := 0; idx < SPHUnits; idx++ {
		for n := 0; n < 9; n++ {
			pos := n*4 + idx*9*4
			p.p.SPH[n] = p.p.SPH[n].Add(mgl32.Vec4{sphRaw[pos], sphRaw[pos+1], sphRaw[pos+2], 0}.Mul(1 / math.Pi))
			weight += sphRaw[pos+3]
		}
	}

	/*
		Debug prints
		for idx := 0; idx < 9; idx++ {
			fmt.Println(idx, " ", p.SPH[idx])
		}
		fmt.Println("Weight ", weight)
	*/
}

func (p *Probe) getSPHLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	l1 := dev.Get(ctx, kSPHLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l2 := dev.Get(ctx, kSPHLayout+1, func(ctx vk.APIContext) interface{} {
		return l1.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	l3 := dev.Get(ctx, kSPHLayout+2, func(ctx vk.APIContext) interface{} {
		return l2.AddBinding(ctx, vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return l3
}

func (p *Probe) getSPHPipeline(ctx vk.APIContext, dev *vk.Device) *vk.ComputePipeline {
	lComp := p.getSPHLayout(ctx, dev)
	return dev.Get(ctx, kSPHPipeline, func(ctx vk.APIContext) interface{} {
		cp := vk.NewComputePipeline(ctx, dev)
		cp.AddLayout(ctx, lComp)
		cp.AddShader(ctx, sph_comp_spv)
		cp.Create(ctx)
		return cp
	}).(*vk.ComputePipeline)
}
