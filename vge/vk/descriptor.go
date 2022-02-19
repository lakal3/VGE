package vk

import (
	"errors"
)

// VSlice interface describes slice of memory buffer than can be written to descriptor set
type VSlice interface {
	Bytes() []byte
	slice() (hBuffer uintptr, from uint64, size uint64)
}

type VImage interface {
	Describe() ImageDescription
	image() (hImage uintptr)
}

type VImageView interface {
	VImage() VImage
	Range() ImageRange
	imageView() (hView uintptr)
}

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
func NewDescriptorLayout(dev *Device, descriptorType DescriptorType, stages ShaderStageFlags, elements uint32) *DescriptorLayout {
	dl := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dev}
	call_Device_NewDescriptorLayout(dev, dev.hDev, descriptorType, stages, elements, 0, 0, &dl.hLayout)
	return dl
}

// NewDynamicDescriptorLayout creates a new dynamic descriptor layout. This will be binding slot 0 in descriptorset.
// You must add dynamics descriptor support using AddDynamicDescriptors
// or this api call will fail. Some older drivers may not support dynamics descriptors
func NewDynamicDescriptorLayout(dev *Device, descriptorType DescriptorType, stages ShaderStageFlags,
	elements uint32, flags DescriptorBindingFlagBitsEXT) *DescriptorLayout {
	if !dev.app.dynamicIndexing {
		dev.ReportError(errors.New("You must enable dynamic descriptors with AddDynamicDescriptors before initializing application"))
		return nil
	}

	dl := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dev, dynamic: true}
	call_Device_NewDescriptorLayout(dev, dev.hDev, descriptorType, stages, elements, flags|DESCRIPTORBindingUpdateAfterBindBitExt, 0, &dl.hLayout)
	return dl
}

// AddBinding creates a NEW descriptor binding that adds new binding to existing ones.
// Binding number will be automatically incremented
func (dl *DescriptorLayout) AddBinding(descriptorType DescriptorType, stages ShaderStageFlags, elements uint32) *DescriptorLayout {
	dlChild := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dl.dev, parent: dl}
	call_Device_NewDescriptorLayout(dl.dev, dl.dev.hDev, descriptorType, stages, elements, 0, dl.hLayout, &dlChild.hLayout)
	dl.owner.AddChild(dlChild)
	return dlChild
}

// AddDynamicBinding creates a NEW descriptor binding that adds new descriptor to existing ones.
// Binding number will be automatically incremented
// You must add dynamics descriptor support using AddDynamicDescriptors
// or this api call will fail. Some older drivers may not support dynamics descriptors
func (dl *DescriptorLayout) AddDynamicBinding(descriptorType DescriptorType, stages ShaderStageFlags,
	elements uint32, flags DescriptorBindingFlagBitsEXT) *DescriptorLayout {
	if !dl.dev.app.dynamicIndexing {
		dl.dev.ReportError(errors.New("You must enable dynamic descriptors with AddDynamicDescriptors before initializing application"))
		return nil
	}
	dlChild := &DescriptorLayout{descriptorType: descriptorType, elements: elements, dev: dl.dev, parent: dl}
	call_Device_NewDescriptorLayout(dl.dev, dl.dev.hDev, descriptorType, stages, elements, flags|DESCRIPTORBindingUpdateAfterBindBitExt, dl.hLayout, &dlChild.hLayout)
	dl.owner.AddChild(dlChild)
	return dlChild
}

// IsValid check that descriptor layout is valid (not disposed)
func (dl *DescriptorLayout) isValid() bool {
	if dl.hLayout == 0 {
		dl.dev.setError(ErrDisposed)
		return false
	}
	return true
}

// NewDescriptorPool creates a new descriptor pool from where you can allocate descriptors. In VGE descriptor pools all
// descriptors must share same layout.
func NewDescriptorPool(dl *DescriptorLayout, maxDescriptors int) *DescriptorPool {
	if !dl.isValid() {
		return nil
	}
	pool := &DescriptorPool{dev: dl.dev, remaining: maxDescriptors}
	call_DescriptorLayout_NewPool(dl.dev, dl.hLayout, uint32(maxDescriptors), &pool.hPool)
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

func (dp *DescriptorPool) isValid() bool {
	if dp.hPool == 0 {
		dp.dev.setError(ErrDisposed)
		return false
	}
	return true
}

// Alloc allocates one descriptor from descriptor set. You can't unallocate single descriptor. Free descriptors disposing whole
// descriptor pool
func (dp *DescriptorPool) Alloc() *DescriptorSet {
	if !dp.isValid() {
		return nil
	}
	if dp.remaining <= 0 {
		dp.dev.ReportError(errors.New("No more descriptors left in pool"))
		return nil
	}
	ds := &DescriptorSet{pool: dp}
	call_DescriptorPool_Alloc(dp.dev, dp.hPool, &ds.hSet)
	dp.remaining--
	return ds
}

func (ds *DescriptorSet) isValid() bool {
	return ds.pool.isValid()
}

// WriteBuffer writes single buffer to descriptor. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support buffer in this binding and index
func (ds *DescriptorSet) WriteBuffer(biding uint32, index uint32, buffer *Buffer) {
	if !ds.isValid() || !buffer.isValid() {
		return
	}
	call_DescriptorSet_WriteBuffer(ds.pool.dev, ds.hSet, biding, index, buffer.hBuf, 0, 0)
}

// WriteSlice writes part of buffer (slice) to descriptor.
// Note that descriptor layout must support buffer in this binding and index
func (ds *DescriptorSet) WriteSlice(biding uint32, index uint32, sl VSlice) {
	if !ds.isValid() {
		return
	}
	hBuf, from, size := sl.slice()
	if hBuf == 0 {
		return
	}
	call_DescriptorSet_WriteDSSlice(ds.pool.dev, ds.hSet, biding, index, hBuf, from, size)
}

// WriteBufferView writes buffer view. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support buffer view in this binding and index
func (ds *DescriptorSet) WriteBufferView(biding uint32, index uint32, view *BufferView) {
	if !ds.isValid() || !view.isValid() {
		return
	}
	call_DescriptorSet_WriteBufferView(ds.pool.dev, ds.hSet, biding, index, view.hView)
}

// WriteImage writes buffer view. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support image or sampled image in this binding and index
// Samples may be nil if biding don't need sampler
func (ds *DescriptorSet) WriteImage(biding uint32, index uint32, view *ImageView, sampler *Sampler) {
	hs := hSampler(0)
	if sampler != nil {
		hs = sampler.hSampler
	}
	call_DescriptorSet_WriteImage(ds.pool.dev, ds.hSet, biding, index, view.view, hs)
}

// WriteView writes image view. Descriptors must be written before they can be bound to draw commands.
// Note that descriptor layout must support image or sampled image in this binding and index
// Samples may be nil if biding don't need sampler
func (ds *DescriptorSet) WriteView(biding uint32, index uint32, view VImageView, layout ImageLayout, sampler *Sampler) {
	hs := hSampler(0)
	if sampler != nil {
		hs = sampler.hSampler
	}
	hView := view.imageView()
	if hView == 0 {
		return
	}
	call_DescriptorSet_WriteDSImageView(ds.pool.dev, ds.hSet, biding, index, hView, layout, hs)
}

// NewSampler creates new image samples with given address mode
func NewSampler(dev *Device, mode SamplerAddressMode) *Sampler {
	s := &Sampler{dev: dev}
	call_Device_NewSampler(dev, dev.hDev, mode, &s.hSampler)
	return s
}

func (s *Sampler) Dispose() {
	if s.hSampler != 0 {
		call_Disposable_Dispose(hDisposable(s.hSampler))
		s.hSampler = 0
	}
}

func (s *Sampler) isValid() bool {
	if s.hSampler == 0 {
		s.dev.setError(ErrDisposed)
		return false
	}
	return true
}
