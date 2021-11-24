//

package vmodel

import (
	"github.com/lakal3/vge/vge/vk"
)

type Compute struct {
	dev     *vk.Device
	cmd     *vk.Command
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

func NewCompute(dev *vk.Device) *Compute {
	c := &Compute{dev: dev}
	c.cmd = vk.NewCommand(dev, vk.QUEUEComputeBit, false)
	c.memPool = vk.NewMemoryPool(dev)
	c.bUbf = c.memPool.ReserveBuffer(vk.MinUniformBufferOffsetAlignment, true, vk.BUFFERUsageUniformBufferBit)
	c.memPool.Allocate()
	return c
}

func (cp *Compute) MipImage(img *vk.Image, layer uint32, mipTo uint32) {
	desc := img.Description
	w := desc.Width >> mipTo
	h := desc.Height >> mipTo
	la := cp.dev.Get(kMipLayout+2, cp.newMipLayout).(*vk.DescriptorLayout)
	pl := cp.dev.Get(kMipPipeline, func() interface{} {
		return cp.newMipPipeline(la)
	}).(*vk.ComputePipeline)

	dp := vk.NewDescriptorPool(la, 1)
	defer dp.Dispose()
	ds := dp.Alloc()

	ubfs := []float32{float32(w), float32(h)}
	copy(cp.bUbf.Bytes(), vk.Float32ToBytes(ubfs))
	r := vk.ImageRange{LayerCount: 1, FirstLayer: layer, FirstMipLevel: mipTo - 1,
		LevelCount: 1, Layout: vk.IMAGELayoutGeneral}
	v1 := vk.NewImageView(img, &r)
	defer v1.Dispose()
	r.FirstMipLevel = mipTo
	v2 := vk.NewImageView(img, &r)
	defer v2.Dispose()
	ds.WriteBuffer(0, 0, cp.bUbf)
	ds.WriteImage(1, 0, v1, nil)
	ds.WriteImage(2, 0, v2, nil)
	cp.cmd.Begin()
	cp.cmd.Compute(pl, w/16+1, h/16+1, 1, ds)
	cp.cmd.Submit()
	cp.cmd.Wait()
}

func (cp *Compute) newMipLayout() interface{} {
	ds0 := cp.dev.Get(kMipLayout+0, func() interface{} {
		return vk.NewDescriptorLayout(cp.dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	ds1 := cp.dev.Get(kMipLayout+1, func() interface{} {
		return ds0.AddBinding(vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return ds1.AddBinding(vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
}

func (cp *Compute) newMipPipeline(la *vk.DescriptorLayout) *vk.ComputePipeline {
	cl := vk.NewComputePipeline(cp.dev)
	cl.AddLayout(la)
	cl.AddShader(genmip_comp_spv)
	cl.Create()
	return cl
}

var kMipPipeline = vk.NewKey()
var kMipLayout = vk.NewKeys(3)
