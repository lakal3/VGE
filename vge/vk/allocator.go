package vk

import (
	"errors"
	"fmt"
	"unsafe"
)

type Allocator struct {
	hAllocator hAllocator
	dev        *Device
}

type AMemory struct {
	hMem    uintptr
	size    uint64
	memType uint32
	al      *Allocator
	bytes   uintptr
	dev     *Device
}

type ABuffer struct {
	Usage BufferUsageFlags

	hBuf      uintptr
	size      uint64
	offset    uint64
	alignment uint32
	memType   uint32
	mem       *AMemory
	al        *Allocator
	dev       *Device
}

type ASlice struct {
	from uint64
	size uint64
	buf  *ABuffer
	dev  *Device
}

type AImage struct {
	Usage       ImageUsageFlags
	Description ImageDescription

	hImage    uintptr
	size      uint64
	offset    uint64
	alignment uint32
	memType   uint32
	mem       *AMemory
	al        *Allocator
	dev       *Device
}

func (ai *AImage) Describe() ImageDescription {
	return ai.Description
}

func (ai *AImage) image() (hImage uintptr) {
	return ai.hImage
}

type AImageView struct {
	Range ImageRange

	image *AImage
	hView uintptr
	al    *Allocator
	dev   *Device
}

func (av *AImageView) VImage() VImage {
	return av.image
}

func (av *AImageView) imageView() (hView uintptr) {
	if !av.isValid() {
		return 0
	}
	return av.hView
}

var ErrNotBound = errors.New("Object not bound to memory")
var ErrBound = errors.New("Object bound to memory")

func (a *Allocator) Dispose() {
	if a.hAllocator != 0 {
		call_Disposable_Dispose(hDisposable(a.hAllocator))
		a.hAllocator = 0
	}
}

func NewAllocator(dev *Device) *Allocator {
	a := &Allocator{dev: dev}
	call_Device_NewAllocator(dev, dev.hDev, &a.hAllocator)
	return a
}

func (a *Allocator) AllocMemory(size uint64, memType uint32, hostMem bool) *AMemory {
	if !a.isValid() {
		return nil
	}
	am := &AMemory{al: a, size: size, dev: a.dev, memType: memType}
	call_Allocator_AllocMemory(a.dev, a.hAllocator, size, memType, hostMem, &am.hMem, &am.bytes)
	return am
}

func (a *Allocator) AllocImage(usage ImageUsageFlags, desc ImageDescription) *AImage {
	if !a.isValid() {
		return nil
	}
	ai := &AImage{Usage: usage, Description: desc, dev: a.dev, al: a}
	call_Allocator_AllocImage(a.dev, a.hAllocator, usage, &desc, &ai.hImage, &ai.size, &ai.memType, &ai.alignment)
	return ai
}

func (a *Allocator) AllocBuffer(usage BufferUsageFlags, size uint64, hostMem bool) *ABuffer {
	if !a.isValid() {
		return nil
	}
	ab := &ABuffer{al: a, size: size, dev: a.dev, Usage: usage}
	if hostMem {
		call_Allocator_AllocBuffer(a.dev, a.hAllocator, usage, size, &ab.hBuf, &ab.memType, &ab.alignment)
	} else {
		call_Allocator_AllocDeviceBuffer(a.dev, a.hAllocator, usage, size, &ab.hBuf, &ab.memType, &ab.alignment)
	}
	return ab
}

func (a *Allocator) isValid() bool {
	if a.hAllocator == 0 {
		a.dev.FatalError(ErrDisposed)
		return false
	}
	return true
}

func (a *AMemory) Dispose() {
	if a.hMem != 0 {
		if a.al.isValid() {
			call_Allocator_FreeMemory(a.dev, a.al.hAllocator, a.hMem, a.bytes != 0)
		}
		a.hMem, a.size, a.bytes = 0, 0, 0
	}
}

func (a *AMemory) HostMemory() bool {
	return a.bytes != 0
}

func (a *AMemory) Bytes() []byte {
	if !a.isValid() {
		return nil
	}
	if a.bytes == 0 {
		a.dev.FatalError(errors.New("Device memory"))
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(a.bytes)), a.size)
}

func (a *AMemory) isValid() bool {
	if a.hMem == 0 {
		a.dev.FatalError(ErrDisposed)
		return false
	}
	return a.al.isValid()
}

func (a *AMemory) Size() uint64 {
	return a.size
}

func (ab *ABuffer) Dispose() {
	if ab.hBuf != 0 {
		if ab.al.isValid() {
			call_Allocator_FreeBuffer(ab.dev, ab.al.hAllocator, ab.hBuf)
		}
		ab.hBuf = 0
	}
}

func (ab *ABuffer) Bind(mem *AMemory, offset uint64) {
	if !mem.isValid() {
		return
	}
	if ab.hBuf == 0 {
		ab.dev.FatalError(ErrDisposed)
		return
	}
	if ab.mem != nil {
		ab.dev.FatalError(ErrBound)
	}
	if offset+ab.size > mem.size {
		ab.dev.FatalError(fmt.Errorf("Required size %d available %d", offset+ab.size, mem.size))
		return
	}
	if ab.memType != mem.memType {
		ab.dev.FatalError(fmt.Errorf("Memtype should be %d, not %d", ab.memType, mem.memType))
		return
	}
	call_Allocator_BindBuffer(ab.dev, ab.al.hAllocator, mem.hMem, ab.hBuf, offset)
	ab.offset = offset
	ab.mem = mem
}

func (ab *ABuffer) isValid() bool {
	if ab.hBuf == 0 {
		ab.dev.FatalError(ErrDisposed)
		return false
	}
	if ab.mem == nil {
		ab.dev.FatalError(ErrNotBound)
		return false
	}
	return ab.mem.isValid()
}

func (ab *ABuffer) Size() (size uint64, alignment uint32) {
	return ab.size, ab.alignment
}

func (ab *ABuffer) Slice(from uint64, size uint64) *ASlice {
	if !ab.isValid() {
		return nil
	}
	if from+size > ab.size {
		ab.dev.FatalError(fmt.Errorf("Requested size %d, available %d", from+size, ab.size))
		return nil
	}
	return &ASlice{dev: ab.dev, buf: ab, from: from, size: size}
}

func (ab *ABuffer) MemoryType() uint32 {
	return ab.memType
}

func (ab *ABuffer) bytes() []byte {
	return ab.mem.Bytes()[ab.offset : ab.offset+ab.size]
}

func (ai *AImage) Bind(mem *AMemory, offset uint64) {
	if !mem.isValid() {
		return
	}
	if ai.hImage == 0 {
		ai.dev.FatalError(ErrDisposed)
		return
	}
	if ai.mem != nil {
		ai.dev.FatalError(ErrBound)
	}
	if offset+ai.size > mem.size {
		ai.dev.FatalError(fmt.Errorf("Required size %d available %d", offset+ai.size, mem.size))
		return
	}
	if ai.memType != mem.memType {
		ai.dev.FatalError(fmt.Errorf("Memtype should be %d, not %d", ai.memType, mem.memType))
		return
	}
	call_Allocator_BindImage(ai.dev, ai.al.hAllocator, mem.hMem, ai.hImage, offset)
	ai.offset = offset
	ai.mem = mem
}

func (ai *AImage) Size() (size uint64, alignment uint32) {
	return ai.size, ai.alignment
}

func (ai *AImage) Dispose() {
	if ai.hImage != 0 {
		ai.al.isValid()
		call_Allocator_FreeImage(ai.dev, ai.al.hAllocator, ai.hImage)
		ai.hImage, ai.mem = 0, nil
	}
}

func (ai *AImage) MemoryType() uint32 {
	return ai.memType
}

func (ai *AImage) isValid() bool {
	if ai.hImage == 0 {
		ai.dev.FatalError(ErrDisposed)
		return false
	}
	if ai.mem == nil {
		ai.dev.FatalError(ErrNotBound)
		return false
	}
	return ai.mem.isValid()
}

func (ai *AImage) AllocView(ir ImageRange, cubeView bool) *AImageView {
	if !ai.isValid() {
		return nil
	}
	av := &AImageView{dev: ai.dev, al: ai.al, image: ai, Range: ir}
	call_Allocator_AllocView(ai.dev, ai.al.hAllocator, ai.hImage, &ir, &ai.Description, cubeView, &av.hView)
	return av
}

func (av *AImageView) Dispose() {
	if av.hView != 0 {
		if av.al.isValid() {
			call_Allocator_FreeView(av.dev, av.al.hAllocator, av.hView)
		}
		av.hView, av.image, av.al = 0, nil, nil
	}
}

func (av *AImageView) isValid() bool {
	if av.hView == 0 {
		av.dev.FatalError(ErrDisposed)
		return false
	}
	return av.image.isValid()
}

func (as *ASlice) slice() (hBuf uintptr, from, size uint64) {
	if as.buf == nil || !as.buf.isValid() {
		return 0, 0, 0
	}
	return as.buf.hBuf, as.from, as.size
}

func (as *ASlice) Bytes() []byte {
	if !as.isValid() {
		return nil
	}
	return as.buf.bytes()[as.from : as.from+as.size]
}

func (as *ASlice) isValid() bool {
	if as.buf == nil || !as.buf.isValid() {
		return false
	}
	return true
}
