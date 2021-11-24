package vk

import (
	"errors"
	"fmt"
)

var ErrInitialized = errors.New("Already initialized")

type Pipeline interface {
	handle() hPipeline
}

type GraphicsPipeline struct {
	hPl         hGraphicsPipeline
	initialized bool
	dev         *Device
}

type ComputePipeline struct {
	hPl         hComputePipeline
	initialized bool
	dev         *Device
}

func (gp *GraphicsPipeline) handle() hPipeline {
	return hPipeline(gp.hPl)
}

func NewGraphicsPipeline(dev *Device) *GraphicsPipeline {
	gp := &GraphicsPipeline{dev: dev}
	call_Device_NewGraphicsPipeline(dev, dev.hDev, &gp.hPl)
	return gp
}

func (gp *GraphicsPipeline) AddVextexInput(rate VertexInputRate, formats ...Format) {
	if gp.initialized {
		gp.dev.setError(ErrInitialized)
		return
	}
	stride := uint32(0)
	for _, ft := range formats {
		f, ok := Formats[ft]
		if !ok || f.PixelSize == 0 {
			gp.dev.setError(fmt.Errorf("Unknown format %d", ft))
		}
		stride += uint32(f.PixelSize / 8)
	}
	call_GraphicsPipeline_AddVertexBinding(gp.dev, gp.hPl, stride, rate)
	offset := uint32(0)
	for _, ft := range formats {
		call_GraphicsPipeline_AddVertexFormat(gp.dev, gp.hPl, ft, offset)
		offset += uint32(Formats[ft].PixelSize / 8)
	}
}

func (gp *GraphicsPipeline) AddLayout(dsLayout *DescriptorLayout) {
	addLayout(gp.dev, gp, dsLayout)
}

func (gp *GraphicsPipeline) AddShader(stage ShaderStageFlags, code []byte) {
	addShader(gp.dev, gp, stage, code)
}

func (gp *GraphicsPipeline) AddDepth(write bool, check bool) {
	call_GraphicsPipeline_AddDepth(gp.dev, gp.hPl, write, check)
}

func (gp *GraphicsPipeline) SetTopology(topology PrimitiveTopology) {
	call_GraphicsPipeline_SetTopology(gp.dev, gp.hPl, topology)
}

func (gp *GraphicsPipeline) Create(rp *GeneralRenderPass) {
	call_GraphicsPipeline_Create(gp.dev, gp.hPl, rp.hRp)
	gp.initialized = true
}

func (gp *GraphicsPipeline) Dispose() {
	if gp.hPl != 0 {
		call_Disposable_Dispose(hDisposable(gp.hPl))
		gp.initialized = false
	}
}

func (gp *GraphicsPipeline) AddAlphaBlend() {
	call_GraphicsPipeline_AddAlphaBlend(gp.dev, gp.hPl)
}

func NewComputePipeline(dev *Device) *ComputePipeline {
	cp := &ComputePipeline{dev: dev}
	call_Device_NewComputePipeline(dev, dev.hDev, &cp.hPl)
	return cp
}

func (c *ComputePipeline) Dispose() {
	if c.hPl != 0 {
		call_Disposable_Dispose(hDisposable(c.hPl))
		c.hPl, c.initialized = 0, false
	}
}

func (c *ComputePipeline) handle() hPipeline {
	return hPipeline(c.hPl)
}

func (cp *ComputePipeline) AddLayout(dsLayout *DescriptorLayout) {
	addLayout(cp.dev, cp, dsLayout)
}

func (cp *ComputePipeline) AddShader(code []byte) {
	addShader(cp.dev, cp, SHADERStageComputeBit, code)
}

func (cp *ComputePipeline) Create() {
	call_ComputePipeline_Create(cp.dev, cp.hPl)
	cp.initialized = true
}

func addShader(ctx apicontext, pl Pipeline, stage ShaderStageFlags, code []byte) {
	call_Pipeline_AddShader(ctx, pl.handle(), stage, code)
}

func addLayout(ctx apicontext, pl Pipeline, layout *DescriptorLayout) {
	call_Pipeline_AddDescriptorLayout(ctx, pl.handle(), layout.hLayout)
}
