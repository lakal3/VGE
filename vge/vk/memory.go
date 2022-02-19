package vk

import (
	"errors"
	"reflect"
	"unsafe"
)

type memoryObject interface {
	Disposable
	setAllocated()
	handle() hMemoryObject
}

type MemoryPool struct {
	blocks    []hMemoryBlock
	reserved  []memoryObject
	allocated []memoryObject
	dev       *Device
}

func (mp *MemoryPool) Dispose() {
	for _, bl := range mp.blocks {
		call_Disposable_Dispose(hDisposable(bl))
	}
	mp.blocks = nil
	for _, obj := range mp.allocated {
		obj.Dispose()
	}
	mp.allocated = nil
}

func NewMemoryPool(dev *Device) *MemoryPool {
	return &MemoryPool{dev: dev}
}

type Buffer struct {
	Host  bool
	Usage BufferUsageFlags
	Size  uint64

	allocated bool
	pool      *MemoryPool
	hBuf      hBuffer
	rawBuf    uintptr
	dev       *Device
	buf       []byte
}

func (b *Buffer) slice() (hBuffer uintptr, from uint64, size uint64) {
	if !b.isValid() {
		return 0, 0, 0
	}
	return b.rawBuf, 0, b.Size
}

type Slice struct {
	Content []byte
	buffer  *Buffer
	from    uint64
	size    uint64
}

func (s *Slice) Bytes() []byte {
	return s.Content
}

func (s *Slice) slice() (hBuffer uintptr, from uint64, size uint64) {
	if s.isValid() {
		return s.buffer.rawBuf, s.from, s.size
	}
	return 0, 0, 0
}

type Image struct {
	owner       Owner
	Usage       ImageUsageFlags
	Description ImageDescription
	allocated   bool
	pool        *MemoryPool
	hImage      hImage
	rawImage    uintptr
	dev         *Device
	defView     *ImageView
	swapbuffer  bool
}

func (i *Image) Describe() ImageDescription {
	return i.Description
}

func (i *Image) image() (hImage uintptr) {
	return i.rawImage
}

type ImageView struct {
	imRange ImageRange
	image   *Image
	view    hImageView
	rawView uintptr
}

func (im *ImageView) Range() ImageRange {
	return im.imRange
}

func (im *ImageView) VImage() VImage {
	return im.image
}

func (im *ImageView) imageView() (hView uintptr) {
	return im.rawView
}

func (im ImageView) Handle() uintptr {
	return uintptr(im.view)
}

type BufferView struct {
	b     *Buffer
	hView hBufferView
}

func (b *Buffer) handle() hMemoryObject {
	return hMemoryObject(b.hBuf)
}

func (b *Buffer) Dispose() {
	if b.hBuf != 0 {
		call_Disposable_Dispose(hDisposable(b.hBuf))
		b.hBuf = 0
		b.allocated = false
	}
}

func (b *Buffer) setAllocated() {
	b.allocated = true
}

func (b *Buffer) isValid() bool {
	if !b.allocated || b.hBuf == 0 {
		b.dev.setError(errors.New("Buffer not allocated or disposed"))
		return false
	}
	return true
}

func (b *Buffer) Bytes() []byte {
	if !b.isValid() {
		return nil
	}
	if !b.Host {
		b.dev.ReportError(errors.New("Bytes only available for host memory"))
		return nil
	}
	if b.buf == nil {
		sl := &reflect.SliceHeader{Len: int(b.Size), Cap: int(b.Size)}
		call_Buffer_GetPtr(b.dev, b.hBuf, &sl.Data)

		b.buf = *(*[]byte)(unsafe.Pointer(sl))
	}
	return b.buf
}

func (b *Buffer) Slice(from uint64, to uint64) *Slice {
	if to == 0 {
		to = b.Size
	}
	s := &Slice{buffer: b, from: from, size: to - from}
	if b.Host {
		tmp := b.Bytes()
		s.Content = tmp[from:to]
	}
	return s
}

// CopyFrom copies content from Go memory to buffer. Offset is starting point inside buffer where to copy memory
// CopyFrom should be used only when size of copied item is large (>64k). For small items call overhead outweighs performance gain
func (b *Buffer) CopyFrom(offset uint64, ptr unsafe.Pointer, size uint64) {
	if !b.isValid() {
		return
	}
	call_Buffer_CopyFrom(b.dev, b.hBuf, offset, uintptr(ptr), size)
}

func (s *Slice) isValid() bool {
	return s.buffer.isValid()
}

func (i *Image) Dispose() {
	if i.hImage != 0 {
		i.owner.Dispose()
		if !i.swapbuffer {
			call_Disposable_Dispose(hDisposable(i.hImage))
		}
		i.hImage = 0
		i.allocated = false
	}
}

func (i *Image) DefaultView() *ImageView {
	if i.defView == nil {
		i.defView = i.NewView(-1, -1)
	}
	return i.defView
}

func (i *Image) isValid() bool {
	if !i.allocated || i.hImage == 0 {
		i.dev.setError(errors.New("Image not allocated or disposed"))
		return false
	}
	return true
}

func (i *Image) setAllocated() {
	i.allocated = true
}

func (i *Image) handle() hMemoryObject {
	return hMemoryObject(i.hImage)
}

func (i *Image) NewView(layer int32, mipLevel int32) *ImageView {
	r := &ImageRange{Layout: IMAGELayoutShaderReadOnlyOptimal}
	if layer < 0 {
		r.LayerCount = i.Description.Layers
	} else {
		r.LayerCount = 1
		r.FirstLayer = uint32(layer)
	}
	if mipLevel < 0 {
		r.LevelCount = i.Description.MipLevels
	} else {
		r.LevelCount = 1
		r.FirstMipLevel = uint32(mipLevel)
	}
	iv := NewImageView(i, r)
	i.owner.AddChild(iv)
	return iv
}

func (i *Image) NewCubeView(mipLevel int32) *ImageView {
	r := &ImageRange{Layout: IMAGELayoutShaderReadOnlyOptimal}
	r.LayerCount = 6
	r.FirstLayer = 0
	if mipLevel < 0 {
		r.LevelCount = i.Description.MipLevels
	} else {
		r.LevelCount = 1
		r.FirstMipLevel = uint32(mipLevel)
	}
	iv := NewCubeView(i, r)
	i.owner.AddChild(iv)
	return iv
}

func NewImageView(image *Image, imRange *ImageRange) *ImageView {

	if !image.isValid() {
		return nil
	}

	iv := &ImageView{image: image, imRange: *imRange}
	call_Image_NewView(image.dev, image.hImage, imRange, &iv.view, &iv.rawView)
	return iv
}

// Deprecated: Use NewImageView with ViewType settings in ImageRange
func NewCubeView(image *Image, imRange *ImageRange) *ImageView {

	if !image.isValid() {
		return nil
	}

	if imRange.LayerCount != 6 {
		image.dev.ReportError(errors.New("Cube hView must have 6 layers"))
		return nil
	}
	ir := *imRange
	ir.ViewType = CubeView
	iv := &ImageView{image: image, imRange: ir}
	call_Image_NewView(image.dev, image.hImage, &ir, &iv.view, &iv.rawView)
	return iv
}

func (i *Image) FullRange() ImageRange {
	return i.Description.FullRange()
}

func (iv *ImageView) Dispose() {
	if iv.view != 0 {
		call_Disposable_Dispose(hDisposable(iv.view))
		iv.view = 0
	}
}

func (iv *ImageView) isValid() bool {
	if iv.view == 0 {
		iv.image.dev.setError(ErrDisposed)
		return false
	}
	return iv.image.isValid()
}

func NewBufferView(b *Buffer, format Format) *BufferView {
	if !b.isValid() {
		return nil
	}
	bv := &BufferView{b: b}
	call_Buffer_NewView(b.dev, b.hBuf, format, 0, 0, &bv.hView)
	return bv
}

func (bv *BufferView) Dispose() {
	if bv.hView != 0 {
		call_Disposable_Dispose(hDisposable(bv.hView))
		bv.hView, bv.b = 0, nil
	}
}

func (bv *BufferView) isValid() bool {
	if bv.hView == 0 {
		bv.b.dev.setError(ErrDisposed)
		return false
	}
	return bv.b.isValid()
}

func (mp *MemoryPool) ReserveBuffer(size uint64, hostmem bool, usage BufferUsageFlags) *Buffer {
	b := &Buffer{dev: mp.dev, pool: mp, Host: hostmem, Size: size, Usage: usage}
	call_Device_NewBuffer(mp.dev, mp.dev.hDev, size, hostmem, usage, &b.hBuf, &b.rawBuf)
	mp.reserved = append(mp.reserved, b)
	return b
}

func (mp *MemoryPool) ReserveImage(desc ImageDescription, usage ImageUsageFlags) *Image {
	img := &Image{dev: mp.dev, pool: mp, Description: desc, Usage: usage}
	call_Device_NewImage(mp.dev, mp.dev.hDev, usage, &img.Description, &img.hImage, &img.rawImage)
	mp.reserved = append(mp.reserved, img)
	return img
}

func (mp *MemoryPool) Allocate() {
	if !mp.dev.isValid() {
		return
	}
	for len(mp.reserved) > 0 {
		mp.allocBlock()
	}
}
func (description *ImageDescription) ImageSize() uint64 {
	return description.ImageRangeSize(description.FullRange())
}

func (description *ImageDescription) ImageRangeSize(imRange ImageRange) uint64 {
	f, ok := Formats[description.Format]
	if !ok {
		return 0
	}
	w, h, d, sTotal := uint64(description.Width), uint64(description.Height), uint64(description.Depth), uint64(0)
	for mips := uint32(0); mips < description.MipLevels; mips++ {
		if mips >= imRange.FirstMipLevel && mips < imRange.FirstMipLevel+imRange.LevelCount {
			sTotal += uint64(f.PixelSize) * w * h * d * uint64(imRange.LayerCount) / 8
		}
		w, h, d = div2(w), div2(h), div2(d)
	}
	return sTotal
}

func (description *ImageDescription) FullRange() ImageRange {
	return ImageRange{LayerCount: description.Layers, LevelCount: description.MipLevels}
}

func (mp *MemoryPool) allocBlock() {
	var block hMemoryBlock
	var remaining []memoryObject

	for _, obj := range mp.reserved {
		if block == 0 {
			call_Device_NewMemoryBlock(mp.dev, mp.dev.hDev, &block)
		}
		var suitable bool
		call_MemoryBlock_Reserve(mp.dev, block, obj.handle(), &suitable)
		if suitable {
			mp.allocated = append(mp.allocated, obj)
			obj.setAllocated()
		} else {
			remaining = append(remaining, obj)
		}
	}
	call_MemoryBlock_Allocate(mp.dev, block)
	mp.reserved = remaining
	mp.blocks = append(mp.blocks, block)
}
