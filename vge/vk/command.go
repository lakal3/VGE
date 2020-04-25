package vk

import "runtime"

type Command struct {
	dev       *Device
	hCmd      hCommand
	Ctx       APIContext
	recording bool
}

type DrawList struct {
	list []DrawItem
}

type SubmitInfo struct {
	info hSubmitInfo
}

func NewCommand(ctx APIContext, dev *Device, cmdQueue QueueFlags, once bool) *Command {
	dev.IsValid(ctx)
	if !ctx.IsValid() {
		return nil
	}
	c := &Command{dev: dev, Ctx: ctx}
	call_Device_NewCommand(ctx, dev.hDev, cmdQueue, once, &c.hCmd)
	return c
}

func (c *Command) IsValid(ctx APIContext) bool {
	if c.hCmd == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

func (c *Command) Begin() {
	if c.IsValid(c.Ctx) {
		call_Command_Begin(c.Ctx, c.hCmd)
	}
}

func (c *Command) Submit(infos ...SubmitInfo) {
	if !c.IsValid(c.Ctx) {
		return
	}
	var infoList []hSubmitInfo
	for _, ifs := range infos {
		infoList = append(infoList, ifs.info)
	}
	c.dev.mxQueue.Lock()
	var dummy hSubmitInfo
	call_Device_Submit(c.Ctx, c.dev.hDev, c.hCmd, 0, infoList, 0, &dummy)
	c.dev.mxQueue.Unlock()
	runtime.KeepAlive(infoList)
}

func (c *Command) SubmitForWait(priority uint32, stage PipelineStageFlags, infos ...SubmitInfo) SubmitInfo {
	if !c.IsValid(c.Ctx) {
		return SubmitInfo{}
	}
	var infoList []hSubmitInfo
	for _, ifs := range infos {
		infoList = append(infoList, ifs.info)
	}
	var si hSubmitInfo
	c.dev.mxQueue.Lock()
	call_Device_Submit(c.Ctx, c.dev.hDev, c.hCmd, priority, infoList, stage, &si)
	c.dev.mxQueue.Unlock()
	return SubmitInfo{info: si}
}

func (c *Command) Wait() {
	if c.IsValid(c.Ctx) {
		call_Command_Wait(c.Ctx, c.hCmd)
	}
}

func (c *Command) Dispose() {
	if c.hCmd != 0 {
		call_Disposable_Dispose(hDisposable(c.hCmd))
		c.hCmd = 0
	}
}

func (c *Command) CopyBuffer(dst *Buffer, src *Buffer) {
	if c.IsValid(c.Ctx) && src.IsValid(c.Ctx) && dst.IsValid(c.Ctx) {
		call_Command_CopyBuffer(c.Ctx, c.hCmd, src.hBuf, dst.hBuf)
	}
}

func (c *Command) CopyImageToBuffer(dst *Buffer, src *Image, imRange *ImageRange) {
	if c.IsValid(c.Ctx) && src.IsValid(c.Ctx) && dst.IsValid(c.Ctx) {
		call_Command_CopyImageToBuffer(c.Ctx, c.hCmd, src.hImage, dst.hBuf, imRange, 0)
	}
}

func (c *Command) CopyImageToSlice(sl *Slice, src *Image, imRange *ImageRange) {
	if c.IsValid(c.Ctx) && src.IsValid(c.Ctx) && sl.IsValid(c.Ctx) {
		call_Command_CopyImageToBuffer(c.Ctx, c.hCmd, src.hImage, sl.buffer.hBuf, imRange, sl.from)
	}
}

func (c *Command) CopyBufferToImage(dst *Image, src *Buffer, imRange *ImageRange) {
	if c.IsValid(c.Ctx) && src.IsValid(c.Ctx) && dst.IsValid(c.Ctx) {
		call_Command_CopyBufferToImage(c.Ctx, c.hCmd, src.hBuf, dst.hImage, imRange, 0)
	}
}

func (c *Command) BeginRenderPass(renderPass RenderPass, fb *Framebuffer) {
	if !renderPass.IsValid(c.Ctx) || !fb.IsValid(c.Ctx) || !c.IsValid(c.Ctx) {
		return
	}
	call_Command_BeginRenderPass(c.Ctx, c.hCmd, hRenderPass(renderPass.GetRenderPass()), fb.hFb)
}

func (c *Command) EndRenderPass() {
	call_Command_EndRenderPass(c.Ctx, c.hCmd)
}

func (c *Command) Draw(dl *DrawList) {
	if !c.IsValid(c.Ctx) {
		return
	}
	if len(dl.list) == 0 {
		return
	}
	dl.optimize()
	call_Command_Draw(c.Ctx, c.hCmd, dl.list)
}

func (c *Command) SetLayout(img *Image, imRange *ImageRange, newLayout ImageLayout) {
	call_Command_SetLayout(c.Ctx, c.hCmd, img.hImage, imRange, newLayout)
	imRange.Layout = newLayout
}

func (c *Command) Compute(cp *ComputePipeline, x uint32, y uint32, z uint32, descs ...*DescriptorSet) {
	// call_command_co
	hds := make([]hDescriptorSet, len(descs))
	for idx, ds := range descs {
		if ds != nil {
			hds[idx] = ds.hSet
		}
	}
	call_Command_Compute(c.Ctx, c.hCmd, cp.hPl, x, y, z, hds)
}

// NextSubpass begins new subpass of multi pass render phase
func (c *Command) NextSubpass(ctx APIContext) {
	call_Command_NextSubpass(ctx, c.hCmd)
}

func (dr *DrawList) Draw(pl Pipeline, from, count uint32) *DrawItem {
	di := DrawItem{pipeline: pl.handle(), from: from, count: count, instances: 1}
	dr.list = append(dr.list, di)
	return &(dr.list[len(dr.list)-1])
}

func (dr *DrawList) DrawIndexed(pl Pipeline, from, count uint32) *DrawItem {
	di := DrawItem{pipeline: pl.handle(), from: from, count: count, instances: 1, indexed: 1}
	dr.list = append(dr.list, di)
	return &(dr.list[len(dr.list)-1])
}

func (dr *DrawList) optimize() {
	if len(dr.list) < 2 {
		return
	}
	var prev DrawItem
	var prevIndex int
	for idxItem, di := range dr.list {
		if idxItem > 0 {
			anyDif := false
			anyChange := false
			for idx := 0; idx < 8; idx++ {
				if di.descriptors[idx].hSet != 0 && di.descriptors[idx].hSet == prev.descriptors[idx].hSet {
					di.descriptors[idx].hSet = 0
					anyChange = true
				}
				if di.descriptors[idx].hSet != 0 {
					anyDif = true
				}
				if di.inputs[idx] != 0 && di.inputs[idx] == prev.inputs[idx] {
					di.inputs[idx] = 0
					anyChange = true
				}
				if di.inputs[idx] != 0 {
					anyDif = true
				}

			}
			if di.pipeline == prev.pipeline {
				di.pipeline = 0
				anyChange = true
			} else {
				anyDif = true
			}
			if !anyDif && prev.from == di.from && prev.count == di.count {
				// Merge instances
				dr.list[prevIndex].instances++
				dr.list[idxItem].instances = 0
				continue
			}
			if anyChange {
				dr.list[idxItem] = di
			}
		}
		prev, prevIndex = di, idxItem
	}
}

func (di *DrawItem) AddInputs(inputs ...*Buffer) *DrawItem {
	for idx, input := range inputs {
		di.AddInput(idx, input)
	}
	return di
}

func (di *DrawItem) AddDescriptors(descriptors ...*DescriptorSet) *DrawItem {
	for idx, ds := range descriptors {
		di.AddDescriptor(idx, ds)
	}
	return di
}

func (di *DrawItem) AddInput(idx int, input *Buffer) *DrawItem {
	if input != nil {
		di.inputs[idx] = input.hBuf
	}
	return di
}

func (di *DrawItem) AddDescriptor(idx int, set *DescriptorSet) *DrawItem {
	di.descriptors[idx] = descriptorInfo{hSet: set.hSet}
	return di
}

func (di *DrawItem) AddDynamicDescriptor(idx int, set *DescriptorSet, offset uint32) *DrawItem {
	di.descriptors[idx] = descriptorInfo{hSet: set.hSet, hasOffset: 1, offset: offset}
	return di
}
