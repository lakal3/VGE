package vmodel

import (
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
)

type Copier struct {
	dev *vk.Device
	cmd *vk.Command
}

var kDefaultSampler = vk.NewKey()

func GetDefaultSampler(dev *vk.Device) *vk.Sampler {
	return dev.Get(kDefaultSampler, func() interface{} {
		return vk.NewSampler(dev, vk.SAMPLERAddressModeRepeat)
	}).(*vk.Sampler)
}

func (c *Copier) Dispose() {
	if c.cmd != nil {
		c.cmd.Dispose()
		c.cmd = nil
	}
}

func DescribeWhiteImage() (desc vk.ImageDescription) {
	_ = vasset.DescribeImage("dds", &desc, white_bin)
	return
}

func NewCopier(dev *vk.Device) *Copier {
	c := &Copier{dev: dev}
	c.cmd = vk.NewCommand(dev, vk.QUEUETransferBit, false)
	return c
}

func (c *Copier) CopyToBuffer(dst *vk.Buffer, content []byte) {
	mb := vk.NewMemoryPool(c.dev)
	defer mb.Dispose()
	bTmp := mb.ReserveBuffer(uint64(len(content)), true, vk.BUFFERUsageTransferSrcBit)
	mb.Allocate()
	copy(bTmp.Bytes(), content)
	c.cmd.Begin()
	c.cmd.CopyBuffer(dst, bTmp)
	c.cmd.Submit()
	c.cmd.Wait()
}

func (c *Copier) CopyToImage(dst *vk.Image, kind string, content []byte, imRange vk.ImageRange, finalLayout vk.ImageLayout) {
	mb := vk.NewMemoryPool(c.dev)
	defer mb.Dispose()
	bTmp := mb.ReserveBuffer(dst.Description.ImageRangeSize(imRange), true, vk.BUFFERUsageTransferSrcBit)
	mb.Allocate()
	vasset.LoadImage(kind, content, bTmp)
	c.cmd.Begin()
	c.cmd.SetLayout(dst, &imRange, vk.IMAGELayoutTransferDstOptimal)
	imRange.Layout = vk.IMAGELayoutTransferDstOptimal
	c.cmd.CopyBufferToImage(dst, bTmp, &imRange)
	c.cmd.SetLayout(dst, &imRange, finalLayout)
	c.cmd.Submit()
	c.cmd.Wait()
}

func (c *Copier) SetLayout(dst *vk.Image, imRange vk.ImageRange, finalLayout vk.ImageLayout) {
	c.cmd.Begin()
	c.cmd.SetLayout(dst, &imRange, finalLayout)
	c.cmd.Submit()
	c.cmd.Wait()
}

func (c *Copier) CopyFromImage(src *vk.Image, imRange vk.ImageRange, kind string, finalLayout vk.ImageLayout) (content []byte, err error) {
	mb := vk.NewMemoryPool(c.dev)
	defer mb.Dispose()
	dstSize := src.Description.ImageRangeSize(imRange)
	bTmp := mb.ReserveBuffer(dstSize, true, vk.BUFFERUsageTransferDstBit)
	mb.Allocate()
	c.cmd.Begin()
	c.cmd.SetLayout(src, &imRange, vk.IMAGELayoutTransferSrcOptimal)

	if imRange.LevelCount > 0 {
		subRange := imRange
		subRange.LevelCount = 1
		subRange.LayerCount = 1
		offset := uint64(0)
		for layer := imRange.FirstLayer; layer < imRange.FirstLayer+imRange.LayerCount; layer++ {
			for mip := imRange.FirstMipLevel; mip < imRange.FirstMipLevel+imRange.LevelCount; mip++ {
				subRange.FirstMipLevel = mip
				subRange.FirstLayer = layer
				mipSize := src.Description.ImageRangeSize(subRange)
				c.cmd.CopyImageToSlice(bTmp.Slice(offset, offset+mipSize), src, &subRange)
				offset += mipSize
			}
		}
	} else {
		c.cmd.CopyImageToBuffer(bTmp, src, &imRange)

	}
	if finalLayout != vk.IMAGELayoutUndefined {
		c.cmd.SetLayout(src, &imRange, finalLayout)
	}
	c.cmd.Submit()
	c.cmd.Wait()
	return vasset.SaveImage(kind, src.Description, bTmp)
}

func (c *Copier) CopyWhiteImage(dst *vk.Image) {
	ir := vk.ImageRange{LayerCount: 1, LevelCount: 1}
	c.CopyToImage(dst, "dds", white_bin, ir, vk.IMAGELayoutShaderReadOnlyOptimal)
}
