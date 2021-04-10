package forward

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"unsafe"
)

const MAX_LIGHTS = 64
const MAX_IMAGES = 48

type Frame struct {
	Projection mgl32.Mat4
	View       mgl32.Mat4
	EyePos     mgl32.Vec4
	SPH        [9]mgl32.Vec4
	NoLights   float32
	EnvMap     float32
	EnvLods    float32
	Far        float32
	Lights     [MAX_LIGHTS]vscene.Light
}

func (f *Frame) AddFrameImage(rc *vk.RenderCache, view *vk.ImageView, sampler *vk.Sampler) (imageIndex vmodel.ImageIndex) {
	return f.SetFrameImage(rc, view, sampler)
}

func (f *Frame) AddProbe(SPH [9]mgl32.Vec4, ubfImage vmodel.ImageIndex) (probeIndex int) {
	if f.EnvLods > 0 {
		return 0 // Currently only one probe
	}
	f.EnvMap, f.EnvLods = float32(ubfImage), 6
	f.SPH = SPH
	return 0
}

func (f *Frame) ViewProjection() (projection, view mgl32.Mat4) {
	return f.Projection, f.View
}

func (f *Frame) CopyTo(sl *vk.Slice) {
	b := *(*[unsafe.Sizeof(Frame{})]byte)(unsafe.Pointer(f))
	copy(sl.Content, b[:])
}

func (f *Frame) AddLight(l vscene.Light) {
	lPos := int(f.NoLights)
	if lPos >= MAX_LIGHTS-1 {
		return
	}
	f.NoLights++
	f.Lights[lPos] = l
}

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

// Convert frame to forward renderer frame. If frame is not forward frame return value is nil
func GetForwardFrame(f vscene.Frame) *Frame {
	fr, _ := f.(*Frame)
	return fr
}

func MustGetForwardFrame(ctx vk.APIContext, f vscene.Frame) *Frame {
	fr := GetForwardFrame(f)
	if fr == nil {
		ctx.SetError(errors.New("Current frame is not forward.*Frame. Most likely this material is not compatible with current renderer"))
	}
	return fr
}

// SetFrameImage sets image for whole frame (like environment) and returns its index. If imageIndex < 0 all image slots has been used
func (f *Frame) SetFrameImage(rc *vk.RenderCache, view *vk.ImageView, sampler *vk.Sampler) (ii vmodel.ImageIndex) {
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
	if fi.idx < MAX_IMAGES {
		fd.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
		ii = fi.idx
		fi.idx++
	} else {
		ii = -1
	}
	if fdDyn == nil {
		return ii
	}
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

func (f *Frame) writeFrame(rc *vk.RenderCache) {
	lt := GetFrameLayout(rc.Ctx, rc.Device)
	fd := rc.Get(kBoundFrame, func(ctx vk.APIContext) interface{} {
		return newFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	_ = rc.GetPerFrame(kBoundFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f.CopyTo(fd.sl)
		sf := vscene.SimpleFrame{Projection: f.Projection, View: f.View}
		sf.WriteFrame(rc)
		return fd.ds
	})
}

func (f *Frame) writeDynamicFrame(rc *vk.RenderCache) {
	lt := GetDynamicFrameLayout(rc.Ctx, rc.Device)
	if lt == nil {
		return
	}
	fd := rc.Get(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		return newDynamicFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	_ = rc.GetPerFrame(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f.CopyTo(fd.sl)
		return fd.ds
	})
}

func BindFrame(rc *vk.RenderCache) *vk.DescriptorSet {
	ds := rc.GetPerFrame(kBoundFrame, func(ctx vk.APIContext) interface{} {
		ctx.SetError(errors.New("Frame not bound. BindFrame called before draw phase!"))
		return nil
	}).(*vk.DescriptorSet)
	return ds
}

func BindDynamicFrame(rc *vk.RenderCache) *vk.DescriptorSet {
	ds := rc.GetPerFrame(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		return nil
	})
	if ds == nil {
		return nil
	}
	return ds.(*vk.DescriptorSet)
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
	fdTmp := &frameDescriptor{maxSize: vscene.FrameMaxDynamicSamplers}
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
	if vscene.FrameMaxDynamicSamplers == 0 {
		return nil
	}
	return dev.Get(ctx, kFrameDynamicLayout, func(ctx vk.APIContext) interface{} {
		dl := dev.NewDescriptorLayout(ctx, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAllGraphics, 1)
		return dl.AddDynamicBinding(ctx, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageAllGraphics, vscene.FrameMaxDynamicSamplers,
			vk.DESCRIPTORBindingPartiallyBoundBitExt|vk.DESCRIPTORBindingUpdateUnusedWhilePendingBitExt)
	}).(*vk.DescriptorLayout)
}
