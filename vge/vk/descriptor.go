package vk

import (
	"errors"
)

// DescriptorLayout describes layout of single descriptor.
type DescriptorLayout struct {
	owner   Owner
	hLayout hDescriptorLayout

	parent         *DescriptorLayout
	descriptorType DescriptorType
	elements       uint32
	dev            *Device
	dynamic        bool
}

func (dl *DescriptorLayout) Dispose() {
	if dl.hLayout != 0 {
		dl.owner.Dispose()
		call_Disposable_Dispose(hDisposable(dl.hLayout))
		dl.hLayout = 0
	}
}

type DescriptorPool struct {
	remaining int
	dev       *Device
	hPool     hDescriptorPool
}

// Descriptor set that must be updated with WriteXXX methods and then bound to draw command
type DescriptorSet struct {
	pool *DescriptorPool
	hSet hDescriptorSet
}

// Sampler needed for combine image sampler bindings (WriteImage in DescriptorSet)
type Sampler struct {
	dev      *Device
	hSampler hSampler
}

// NewDescriptorLayout, created descriptor layout. This will be binding slot 0 in descriptorset.
func NewDescriptorLayout(ctx APIContext, dev *Device, descriptorType DescriptorType, stages ShaderStageFlags, elements uint32) *DescriptorLayout {
	dl := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dev}
	call_Device_NewDescriptorLayout(ctx, dev.hDev, descriptorType, stages, elements, 0, 0, &dl.hLayout)
	return dl
}

// NewDynamicDescriptorLayout creates a new dynamic descriptor layout. This will be binding slot 0 in descriptorset.
// You must add dynamics descriptor support using AddDynamicDescriptors
// or this api call will fail. Some older drivers may not support dynamics descriptors
func NewDynamicDescriptorLayout(ctx APIContext, dev *Device, descriptorType DescriptorType, stages ShaderStageFlags,
	elements uint32, flags DescriptorBindingFlagBitsEXT) *DescriptorLayout {
	if !dev.app.dynamicIndexing {
		ctx.SetError(errors.New("You must enable dynamic descriptors with AddDynamicDescriptors before initializing application"))
		return nil
	}

	dl := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dev, dynamic: true}
	call_Device_NewDescriptorLayout(ctx, dev.hDev, descriptorType, stages, elements, flags|DESCRIPTORBindingUpdateAfterBindBitExt, 0, &dl.hLayout)
	return dl
}

// AddBinding creates a NEW descriptor binding that adds new binding to existing ones.
// Binding number will be automatically incremented
func (dl *DescriptorLayout) AddBinding(ctx APIContext, descriptorType DescriptorType, stages ShaderStageFlags, elements uint32) *DescriptorLayout {
	dlChild := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dl.dev, parent: dl}
	call_Device_NewDescriptorLayout(ctx, dl.dev.hDev, descriptorType, stages, elements, 0, dl.hLayout, &dlChild.hLayout)
	dl.owner.AddChild(dlChild)
	return dlChild
}

// AddDynamicBinding creates a NEW descriptor binding that adds new descriptor to existing ones.
// Binding number will be automatically incremented
// You must add dynamics descriptor support using AddDynamicDescriptors
// or this api call will fail. Some older drivers may not support dynamics descriptors
func (dl *DescriptorLayout) AddDynamicBinding(ctx APIContext, descriptorType DescriptorType, stages ShaderStageFlags,
	elements uint32, flags DescriptorBindingFlagBitsEXT) *DescriptorLayout {
	if !dl.dev.app.dynamicIndexing {
		ctx.SetError(errors.New("You must enable dynamic descriptors with AddDynamicDescriptors before initializing application"))
		return nil
	}
	dlChild := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dl.dev, parent: dl}
	call_Device_NewDescriptorLayout(ctx, dl.dev.hDev, descriptorType, stages, elements, flags|DESCRIPTORBindingUpdateAfterBindBitExt, dl.hLayout, &dlChild.hLayout)
	dl.owner.AddChild(dlChild)
	return dlChild
}

// IsValid check that descriptor layout is valid (not disposed)
func (dl *DescriptorLayout) IsValid(ctx APIContext) bool {
	if dl.hLayout == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

// NewDescriptorPool creates a new descriptor pool from where you can allocate descriptors. In VGE descriptor pools all
// descriptors must share same layout.
func NewDescriptorPool(ctx APIContext, dl *DescriptorLayout, maxDescriptors int) *DescriptorPool {
	if !dl.IsValid(ctx) {
		return nil
	}
	pool := &DescriptorPool{dev: dl.dev, remaining: maxDescriptors}
	call_DescriptorLayout_NewPool(ctx, dl.hLayout, uint32(maxDescriptors), &pool.hPool)
	return pool
}

// Number of remaining descriptors
func (dp *DescriptorPool) Remaining() int {
	return dp.remaining
}

func (dp *DescriptorPool) Dispose() {
	if dp.hPool != 0 {
		call_Disposable_Dispose(hDisposable(dp.hPool))
		dp.hPool, dp.remaining = 0, 0
	}
}

func (dp *DescriptorPool) IsValid(ctx APIContext) bool {
	if dp.hPool == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

// Alloc allocates one descriptor from descriptor set. You can't unallocate single descriptor. Free descriptors disposing whole
// descriptor pool
func (dp *DescriptorPool) Alloc(ctx APIContext) *DescriptorSet {
	if !dp.IsValid(ctx) {
		return nil
	}
	if dp.remaining <= 0 {
		ctx.SetError(errors.New("No more descriptors left in pool"))
		return nil
	}
	ds := &DescriptorSet{pool: dp}
	call_DescriptorPool_Alloc(ctx, dp.hPool, &ds.hSet)
	dp.remaining--
	return ds
}

func (ds *DescriptorSet) IsValid(ctx APIContext) bool {
	return ds.pool.IsValid(ctx)
}

// WriteBuffer writes single buffer to descriptor. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support buffer in this binding and index
func (ds *DescriptorSet) WriteBuffer(ctx APIContext, biding uint32, index uint32, buffer *Buffer) {
	if !ds.IsValid(ctx) || !buffer.IsValid(ctx) {
		return
	}
	call_DescriptorSet_WriteBuffer(ctx, ds.hSet, biding, index, buffer.hBuf, 0, 0)
}

// WriteSlice writes part of buffer (slice) to descriptor.
// Note that descriptor layout must support buffer in this binding and index
func (ds *DescriptorSet) WriteSlice(ctx APIContext, biding uint32, index uint32, sl *Slice) {
	if !ds.IsValid(ctx) || !sl.IsValid(ctx) {
		return
	}
	call_DescriptorSet_WriteBuffer(ctx, ds.hSet, biding, index, sl.buffer.hBuf, sl.from, sl.size)
}

// WriteBufferView writes buffer view. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support buffer view in this binding and index
func (ds *DescriptorSet) WriteBufferView(ctx APIContext, biding uint32, index uint32, view *BufferView) {
	if !ds.IsValid(ctx) || !view.IsValid(ctx) {
		return
	}
	call_DescriptorSet_WriteBufferView(ctx, ds.hSet, biding, index, view.hView)
}

// WriteImage writes buffer view. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support image or sampled image in this binding and index
// Samples may be nil if biding don't need sampler
func (ds *DescriptorSet) WriteImage(ctx APIContext, biding uint32, index uint32, view *ImageView, sampler *Sampler) {
	hs := hSampler(0)
	if sampler != nil {
		hs = sampler.hSampler
	}
	call_DescriptorSet_WriteImage(ctx, ds.hSet, biding, index, view.view, hs)
}

// NewSampler creates new image samples with given address mode
func NewSampler(ctx APIContext, dev *Device, mode SamplerAddressMode) *Sampler {
	s := &Sampler{dev: dev}
	call_Device_NewSampler(ctx, dev.hDev, mode, &s.hSampler)
	return s
}

func (s *Sampler) Dispose() {
	if s.hSampler != 0 {
		call_Disposable_Dispose(hDisposable(s.hSampler))
		s.hSampler = 0
	}
}

func (s *Sampler) IsValid(ctx APIContext) bool {
	if s.hSampler == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}
