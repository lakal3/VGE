//

package vmodel

import (
	"github.com/lakal3/vge/vge/vk"
)

type Compute struct {
	dev     *vk.Device
	cmd     *vk.Command
	ctx     vk.APIContext
	memPool *vk.MemoryPool
	bUbf    *vk.Buffer
}

func (cp *Compute) Dispose() {
	if cp.memPool != nil {
		cp.memPool.Dispose()
		cp.memPool, cp.bUbf = nil, nil
	}
	if cp.cmd != nil {
		cp.cmd.Dispose()
		cp.cmd = nil
	}
}

func NewCompute(ctx vk.APIContext, dev *vk.Device) *Compute {
	c := &Compute{dev: dev, ctx: ctx}
	c.cmd = vk.NewCommand(ctx, dev, vk.QUEUEComputeBit, false)
	c.memPool = vk.NewMemoryPool(dev)
	c.bUbf = c.memPool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit)
	c.memPool.Allocate(ctx)
	return c
}

func (cp *Compute) MipImage(img *vk.Image, layer uint32, mipTo uint32) {
	desc := img.Description
	w := desc.Width >> mipTo
	h := desc.Height >> mipTo
	la := cp.dev.Get(cp.ctx, kMipLayout+2, cp.newMipLayout).(*vk.DescriptorLayout)
	pl := cp.dev.Get(cp.ctx, kMipPipeline, func(ctx vk.APIContext) interface{} {
		return cp.newMipPipeline(la)
	}).(*vk.ComputePipeline)

	dp := vk.NewDescriptorPool(cp.ctx, la, 1)
	defer dp.Dispose()
	ds := dp.Alloc(cp.ctx)

	ubfs := []float32{float32(w), float32(h)}
	copy(cp.bUbf.Bytes(cp.ctx), vk.Float32ToBytes(ubfs))
	r := vk.ImageRange{LayerCount: 1, FirstLayer: layer, FirstMipLevel: mipTo - 1,
		LevelCount: 1, Layout: vk.IMAGELayoutGeneral}
	v1 := vk.NewImageView(cp.ctx, img, &r)
	defer v1.Dispose()
	r.FirstMipLevel = mipTo
	v2 := vk.NewImageView(cp.ctx, img, &r)
	defer v2.Dispose()
	ds.WriteBuffer(cp.ctx, 0, 0, cp.bUbf)
	ds.WriteImage(cp.ctx, 1, 0, v1, nil)
	ds.WriteImage(cp.ctx, 2, 0, v2, nil)
	cp.cmd.Begin()
	cp.cmd.Compute(pl, w/16+1, h/16+1, 1, ds)
	cp.cmd.Submit()
	cp.cmd.Wait()
}

func (cp *Compute) newMipLayout(ctx vk.APIContext) interface{} {
	ds0 := cp.dev.Get(cp.ctx, kMipLayout+0, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, cp.dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	ds1 := cp.dev.Get(cp.ctx, kMipLayout+1, func(ctx vk.APIContext) interface{} {
		return ds0.AddBinding(ctx, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return ds1.AddBinding(ctx, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
}

func (cp *Compute) newMipPipeline(la *vk.DescriptorLayout) *vk.ComputePipeline {
	cl := vk.NewComputePipeline(cp.ctx, cp.dev)
	cl.AddLayout(cp.ctx, la)
	cl.AddShader(cp.ctx, genmip_comp_spv)
	cl.Create(cp.ctx)
	return cl
}

var kMipPipeline = vk.NewKey()
var kMipLayout = vk.NewKeys(3)
