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
}

type ComputePipeline struct {
	hPl         hComputePipeline
	initialized bool
}

func (gp *GraphicsPipeline) handle() hPipeline {
	return hPipeline(gp.hPl)
}

func NewGraphicsPipeline(ctx APIContext, dev *Device) *GraphicsPipeline {
	gp := &GraphicsPipeline{}
	call_Device_NewGraphicsPipeline(ctx, dev.hDev, &gp.hPl)
	return gp
}

func (gp *GraphicsPipeline) AddVextexInput(ctx APIContext, rate VertexInputRate, formats ...Format) {
	if gp.initialized {
		ctx.SetError(ErrInitialized)
		return
	}
	stride := uint32(0)
	for _, ft := range formats {
		f, ok := Formats[ft]
		if !ok || f.PixelSize == 0 {
			ctx.SetError(fmt.Errorf("Unknown format %d", ft))
		}
		stride += uint32(f.PixelSize / 8)
	}
	call_GraphicsPipeline_AddVertexBinding(ctx, gp.hPl, stride, rate)
	offset := uint32(0)
	for _, ft := range formats {
		call_GraphicsPipeline_AddVertexFormat(ctx, gp.hPl, ft, offset)
		offset += uint32(Formats[ft].PixelSize / 8)
	}
}

func (gp *GraphicsPipeline) AddLayout(ctx APIContext, dsLayout *DescriptorLayout) {
	addLayout(ctx, gp, dsLayout)
}

func (gp *GraphicsPipeline) AddShader(ctx APIContext, stage ShaderStageFlags, code []byte) {
	addShader(ctx, gp, stage, code)
}

func (gp *GraphicsPipeline) AddDepth(ctx APIContext, write bool, check bool) {
	call_GraphicsPipeline_AddDepth(ctx, gp.hPl, write, check)
}

func (gp *GraphicsPipeline) SetTopology(ctx APIContext, topology PrimitiveTopology) {
	call_GraphicsPipeline_SetTopology(ctx, gp.hPl, topology)
}

func (gp *GraphicsPipeline) Create(ctx APIContext, rp RenderPass) {
	call_GraphicsPipeline_Create(ctx, gp.hPl, hRenderPass(rp.GetRenderPass()))
	gp.initialized = true
}

func (gp *GraphicsPipeline) Dispose() {
	if gp.hPl != 0 {
		call_Disposable_Dispose(hDisposable(gp.hPl))
		gp.initialized = false
	}
}

func (gp *GraphicsPipeline) AddAlphaBlend(ctx APIContext) {
	call_GraphicsPipeline_AddAlphaBlend(ctx, gp.hPl)
}

func NewComputePipeline(ctx APIContext, dev *Device) *ComputePipeline {
	cp := &ComputePipeline{}
	call_Device_NewComputePipeline(ctx, dev.hDev, &cp.hPl)
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

func (cp *ComputePipeline) AddLayout(ctx APIContext, dsLayout *DescriptorLayout) {
	addLayout(ctx, cp, dsLayout)
}

func (cp *ComputePipeline) AddShader(ctx APIContext, code []byte) {
	addShader(ctx, cp, SHADERStageComputeBit, code)
}

func (cp *ComputePipeline) Create(ctx APIContext) {
	call_ComputePipeline_Create(ctx, cp.hPl)
	cp.initialized = true
}

func addShader(ctx APIContext, pl Pipeline, stage ShaderStageFlags, code []byte) {
	call_Pipeline_AddShader(ctx, pl.handle(), stage, code)
}

func addLayout(ctx APIContext, pl Pipeline, layout *DescriptorLayout) {
	call_Pipeline_AddDescriptorLayout(ctx, pl.handle(), layout.hLayout)
}
