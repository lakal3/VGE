package vk

import (
	"runtime"
	"unsafe"
)

type Command struct {
	dev       *Device
	hCmd      hCommand
	recording bool
}

type DrawList struct {
	list          []DrawItem
	pushConstants []byte
}

type SubmitInfo struct {
	info hSubmitInfo
}

func NewCommand(dev *Device, cmdQueue QueueFlags, once bool) *Command {
	c := &Command{dev: dev}
	call_Device_NewCommand(dev, dev.hDev, cmdQueue, once, &c.hCmd)
	return c
}

func (c *Command) isValid() bool {
	if c.hCmd == 0 {
		c.dev.setError(ErrDisposed)
		return false
	}
	return true
}

func (c *Command) Begin() {
	if c.isValid() {
		call_Command_Begin(c.dev, c.hCmd)
	}
}

func (c *Command) Submit(infos ...SubmitInfo) {
	if !c.isValid() {
		return
	}
	var infoList []hSubmitInfo
	for _, ifs := range infos {
		infoList = append(infoList, ifs.info)
	}
	c.dev.mxQueue.Lock()
	var dummy hSubmitInfo
	call_Device_Submit(c.dev, c.dev.hDev, c.hCmd, 0, infoList, 0, &dummy)
	c.dev.mxQueue.Unlock()
	runtime.KeepAlive(infoList)
}

func (c *Command) SubmitForWait(priority uint32, stage PipelineStageFlags, infos ...SubmitInfo) SubmitInfo {
	if !c.isValid() {
		return SubmitInfo{}
	}
	var infoList []hSubmitInfo
	for _, ifs := range infos {
		infoList = append(infoList, ifs.info)
	}
	var si hSubmitInfo
	c.dev.mxQueue.Lock()
	call_Device_Submit(c.dev, c.dev.hDev, c.hCmd, priority, infoList, stage, &si)
	c.dev.mxQueue.Unlock()
	return SubmitInfo{info: si}
}

func (c *Command) Wait() {
	if c.isValid() {
		call_Command_Wait(c.dev, c.hCmd)
	}
}

func (c *Command) Dispose() {
	if c.hCmd != 0 {
		call_Disposable_Dispose(hDisposable(c.hCmd))
		c.hCmd = 0
	}
}

func (c *Command) CopyBuffer(dst *Buffer, src *Buffer) {
	if c.isValid() && src.isValid() && dst.isValid() {
		call_Command_CopyBuffer(c.dev, c.hCmd, src.hBuf, dst.hBuf)
	}
}

func (c *Command) CopyImageToBuffer(dst *Buffer, src *Image, imRange *ImageRange) {
	if c.isValid() && src.isValid() && dst.isValid() {
		call_Command_CopyImageToBuffer(c.dev, c.hCmd, src.hImage, dst.hBuf, imRange, 0)
	}
}

func (c *Command) CopyImageToSlice(sl *Slice, src *Image, imRange *ImageRange) {
	if c.isValid() && src.isValid() && sl.isValid() {
		call_Command_CopyImageToBuffer(c.dev, c.hCmd, src.hImage, sl.buffer.hBuf, imRange, sl.from)
	}
}

func (c *Command) CopyBufferToImage(dst *Image, src *Buffer, imRange *ImageRange) {
	if c.isValid() && src.isValid() && dst.isValid() {
		call_Command_CopyBufferToImage(c.dev, c.hCmd, src.hBuf, dst.hImage, imRange, 0)
	}
}

func (c *Command) ClearImage(dst *Image, imRange *ImageRange, color float32, alpha float32) {
	if c.isValid() && dst.isValid() {
		call_Command_ClearImage(c.dev, c.hCmd, dst.hImage, imRange, imRange.Layout, color, alpha)
	}
}
func (c *Command) BeginRenderPass(renderPass *GeneralRenderPass, fb *Framebuffer) {
	if !renderPass.isValid() || !fb.isValid() || !c.isValid() {
		return
	}
	call_Command_BeginRenderPass(c.dev, c.hCmd, renderPass.hRp, fb.hFb)
}

func (c *Command) EndRenderPass() {
	call_Command_EndRenderPass(c.dev, c.hCmd)
}

func (c *Command) Draw(dl *DrawList) {
	if !c.isValid() {
		return
	}
	dl.optimize()
	call_Command_Draw(c.dev, c.hCmd, dl.list, dl.pushConstants)
}

func (c *Command) SetLayout(img *Image, imRange *ImageRange, newLayout ImageLayout) {
	call_Command_SetLayout(c.dev, c.hCmd, img.hImage, imRange, newLayout)
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
	call_Command_Compute(c.dev, c.hCmd, cp.hPl, x, y, z, hds)
}

// Write value of timer after all commands in stage has completed
func (c *Command) WriteTimer(tp *TimerPool, timerIndex uint32, stage PipelineStageFlags) {
	call_Command_WriteTimer(c.dev, c.hCmd, tp.pool, stage, timerIndex)
}

type TimerPool struct {
	dev  *Device
	pool hQueryPool
	size uint32
}

func (t *TimerPool) Dispose() {
	if t.pool != 0 {
		call_Disposable_Dispose(hDisposable(t.pool))
		t.pool = 0
	}
}

// Get retrieve recorded values converted to nanosecond. Values are valid only after command(s) have completed.
// Only difference of values make any sense and if you record time events from multiple queues, times may no be comparable
// Use WriteTimer in command to record actual timer values.
func (t *TimerPool) Get() []float64 {
	result := make([]uint64, t.size)
	var multiplier float32
	call_QueryPool_Get(t.dev, t.pool, result, &multiplier)
	fresult := make([]float64, t.size)
	for idx, r := range result {
		fresult[idx] = float64(r) * float64(multiplier)
	}
	return fresult
}

// NewTimerPool creates a QueryPool for timing.
// You must specify how many time events you want to write to in one pool
// Currently pool can be used only once. Dispose and create a new pool if you need multiple recording
func NewTimerPool(dev *Device, size uint32) *TimerPool {
	t := &TimerPool{dev: dev, size: size}
	call_Device_NewTimestampQuery(dev, dev.hDev, size, &t.pool)
	return t
}

func (dr *DrawList) AllocPushConstants(size uint32) (ptr unsafe.Pointer, offset uint64) {
	if len(dr.pushConstants)+int(size) > cap(dr.pushConstants) {
		ns := len(dr.pushConstants) * 2
		if ns < 65536 {
			ns = 65536
		}
		old := dr.pushConstants
		dr.pushConstants = make([]byte, len(dr.pushConstants), ns)
		copy(dr.pushConstants, old)
	}
	offset = uint64(len(dr.pushConstants))
	dr.pushConstants = dr.pushConstants[:offset+uint64(size)]
	return unsafe.Pointer(&dr.pushConstants[offset]), offset
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
			if di.pushlen != prev.pushlen || di.pushoffset != prev.pushoffset {
				anyDif = true
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

func (di *DrawItem) AddPushConstants(size uint32, offset uint64) {
	di.pushlen, di.pushoffset = size, offset
}
