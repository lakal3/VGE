package vscene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"unsafe"
)

const MAX_LIGHTS = 64
const MAX_IMAGES = 48

// Maximum dynamics samplers per frame instance. You must add dynamics descriptors support to application before using
// dynamics samplers in frame. Also, check device.Props for MaxSamplersPerStage. This value combined with other samplers used in
// material descriptors may not exceed this limit.
var FrameMaxDynamicSamplers = uint32(0)

type Frame struct {
	Projection mgl32.Mat4
	View       mgl32.Mat4
	EyePos     mgl32.Vec4
	SPH        [9]mgl32.Vec4
	NoLights   float32
	EnvMap     float32
	EnvLods    float32
	Far        float32
	Lights     [MAX_LIGHTS]Light
}

type Light struct {
	Intensity   mgl32.Vec4
	Position    mgl32.Vec4 // w = 0 for directional light and this is shadowmap position
	Direction   mgl32.Vec4 // if w > 0, shadowmap index = w - 1
	Attenuation mgl32.Vec4 // 0, 1st and 2nd order, w is shadowmap index
}

func (f *Frame) CopyTo(sl *vk.Slice) {
	b := *(*[unsafe.Sizeof(Frame{})]byte)(unsafe.Pointer(f))
	copy(sl.Content, b[:])
}

func (f *Frame) AddLight(l Light) {
	lPos := int(f.NoLights)
	if lPos >= MAX_LIGHTS-1 {
		return
	}
	f.NoLights++
	f.Lights[lPos] = l
}

var kFrame = vk.NewKey()
var kBoundFrame = vk.NewKey()
var kBoundDynamicFrame = vk.NewKey()
var kFrameLayout = vk.NewKey()
var kFrameDynamicLayout = vk.NewKey()
var kFrameInfo = vk.NewKey()

type frameDescriptor struct {
	dsPool     *vk.DescriptorPool
	ds         *vk.DescriptorSet
	pool       *vk.MemoryPool
	buffer     *vk.Buffer
	sl         *vk.Slice
	whiteImage *vk.Image
	maxSize    uint32
}

type frameInfo struct {
	idx vmodel.ImageIndex
}

func (f *frameDescriptor) Dispose() {
	if f.dsPool != nil {
		f.dsPool.Dispose()
		f.dsPool = nil
		f.ds = nil
	}
	if f.pool != nil {
		f.pool.Dispose()
		f.pool, f.buffer = nil, nil
	}
}

func GetFrame(rc *vk.RenderCache) *Frame {
	return rc.GetPerFrame(kFrame, func(ctx vk.APIContext) interface{} {
		return &Frame{Projection: mgl32.Ident4(), View: mgl32.Ident4()}
	}).(*Frame)
}

// SetFrameImage sets image for whole frame (like environment) and returns its index. If imageIndex < 0 all image slots has been used
func SetFrameImage(rc *vk.RenderCache, view *vk.ImageView, sampler *vk.Sampler) (ii vmodel.ImageIndex) {
	lt := GetFrameLayout(rc.Ctx, rc.Device)
	ltDyn := GetDynamicFrameLayout(rc.Ctx, rc.Device)
	fd := rc.Get(kBoundFrame, func(ctx vk.APIContext) interface{} {
		return newFrameDescriptor(rc, lt)
	}).(*frameDescriptor)
	var fdDyn *frameDescriptor
	if ltDyn != nil {
		fdDyn = rc.Get(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
			return newDynamicFrameDescriptor(rc, ltDyn)
		}).(*frameDescriptor)
	}
	fi := rc.GetPerFrame(kFrameInfo, func(ctx vk.APIContext) interface{} {
		return &frameInfo{idx: 1}
	}).(*frameInfo)
	if fdDyn != nil {
		if fi.idx < vmodel.ImageIndex(fdDyn.maxSize) {
			fdDyn.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
			if fi.idx < MAX_IMAGES {
				fd.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
			}
			ii = fi.idx
			fi.idx++
		} else {
			ii = -1
		}
		return
	}
	if fi.idx < MAX_IMAGES {
		fd.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
		ii = fi.idx
		fi.idx++
	} else {
		ii = -1
	}
	return
}

func BindFrame(rc *vk.RenderCache) *vk.DescriptorSet {
	lt := GetFrameLayout(rc.Ctx, rc.Device)
	fd := rc.Get(kBoundFrame, func(ctx vk.APIContext) interface{} {
		return newFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	_ = rc.GetPerFrame(kBoundFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f := GetFrame(rc)
		f.CopyTo(fd.sl)
		return f
	})
	return fd.ds
}

func BindDynamicFrame(rc *vk.RenderCache) *vk.DescriptorSet {
	lt := GetDynamicFrameLayout(rc.Ctx, rc.Device)
	if lt == nil {
		return nil
	}
	fd := rc.Get(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		return newDynamicFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	_ = rc.GetPerFrame(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f := GetFrame(rc)
		f.CopyTo(fd.sl)
		return f
	})
	return fd.ds
}

func newFrameDescriptor(rc *vk.RenderCache, lt *vk.DescriptorLayout) *frameDescriptor {
	ctx := rc.Ctx
	fdTmp := &frameDescriptor{}
	fdTmp.dsPool = vk.NewDescriptorPool(ctx, lt, 1)
	fdTmp.ds = fdTmp.dsPool.Alloc(ctx)
	fdTmp.pool = vk.NewMemoryPool(rc.Device)
	lf := uint64(unsafe.Sizeof(Frame{}))
	fdTmp.buffer = fdTmp.pool.ReserveBuffer(rc.Ctx, lf, true, vk.BUFFERUsageUniformBufferBit)
	fdTmp.whiteImage = fdTmp.pool.ReserveImage(ctx, vmodel.DescribeWhiteImage(ctx), vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
	fdTmp.pool.Allocate(rc.Ctx)
	fdTmp.sl = fdTmp.buffer.Slice(rc.Ctx, 0, lf)
	cp := vmodel.NewCopier(ctx, rc.Device)
	defer cp.Dispose()
	cp.CopyWhiteImage(fdTmp.whiteImage)
	sampler := vmodel.GetDefaultSampler(ctx, rc.Device)
	for idx := uint32(0); idx < MAX_IMAGES; idx++ {
		fdTmp.ds.WriteImage(ctx, 1, idx, fdTmp.whiteImage.DefaultView(ctx), sampler)
	}
	fdTmp.ds.WriteSlice(ctx, 0, 0, fdTmp.sl)
	return fdTmp
}

func newDynamicFrameDescriptor(rc *vk.RenderCache, lt *vk.DescriptorLayout) *frameDescriptor {
	ctx := rc.Ctx
	fdTmp := &frameDescriptor{maxSize: FrameMaxDynamicSamplers}
	fdTmp.dsPool = vk.NewDescriptorPool(ctx, lt, 1)
	fdTmp.ds = fdTmp.dsPool.Alloc(ctx)
	fdTmp.pool = vk.NewMemoryPool(rc.Device)
	lf := uint64(unsafe.Sizeof(Frame{}))
	fdTmp.buffer = fdTmp.pool.ReserveBuffer(rc.Ctx, lf, true, vk.BUFFERUsageUniformBufferBit)
	fdTmp.whiteImage = fdTmp.pool.ReserveImage(ctx, vmodel.DescribeWhiteImage(ctx), vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
	fdTmp.pool.Allocate(rc.Ctx)
	fdTmp.sl = fdTmp.buffer.Slice(rc.Ctx, 0, lf)
	cp := vmodel.NewCopier(ctx, rc.Device)
	defer cp.Dispose()
	cp.CopyWhiteImage(fdTmp.whiteImage)
	sampler := vmodel.GetDefaultSampler(ctx, rc.Device)
	fdTmp.ds.WriteImage(ctx, 1, 0, fdTmp.whiteImage.DefaultView(ctx), sampler)
	fdTmp.ds.WriteSlice(ctx, 0, 0, fdTmp.sl)
	return fdTmp
}

func GetFrameLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(ctx, kFrameLayout, func(ctx vk.APIContext) interface{} {
		dl := dev.NewDescriptorLayout(ctx, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAllGraphics, 1)
		return dl.AddBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, MAX_IMAGES)
	}).(*vk.DescriptorLayout)
}

func GetDynamicFrameLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	if FrameMaxDynamicSamplers == 0 {
		return nil
	}
	return dev.Get(ctx, kFrameDynamicLayout, func(ctx vk.APIContext) interface{} {
		dl := dev.NewDescriptorLayout(ctx, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAllGraphics, 1)
		return dl.AddDynamicBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageAllGraphics, FrameMaxDynamicSamplers,
			vk.DESCRIPTORBindingPartiallyBoundBitExt|vk.DESCRIPTORBindingUpdateUnusedWhilePendingBitExt)
	}).(*vk.DescriptorLayout)
}

type Camera interface {
	SetupFrame(f *Frame, size image.Point)
}
