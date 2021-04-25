package forward

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"unsafe"
)

const MAX_LIGHTS = 64
const MAX_IMAGES = 48

type ShaderFrame struct {
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

type Frame struct {
	SF        ShaderFrame
	ds        *vk.DescriptorSet
	dsDynamic *vk.DescriptorSet
	cache     *vk.RenderCache
	sf        *vscene.SimpleFrame
	renderer  vmodel.Renderer
}

func (f *Frame) GetRenderer() vmodel.Renderer {
	return f.renderer
}

func NewFrame(cache *vk.RenderCache, renderer vmodel.Renderer) *Frame {
	return &Frame{cache: cache, renderer: renderer}
}

func (f *Frame) GetSimpleFrame() *vscene.SimpleFrame {
	if f.sf == nil {
		f.sf = &vscene.SimpleFrame{SSF: vscene.SimpleShaderFrame{Projection: f.SF.Projection, View: f.SF.View}, Cache: f.cache}
	}
	return f.sf
}

func (f *Frame) GetCache() *vk.RenderCache {
	return f.cache
}

func (f *Frame) AddEnvironment(SPH [9]mgl32.Vec4, ubfImage vmodel.ImageIndex, pi *vscene.ProcessInfo) {
	if f.SF.EnvLods > 0 || f.ds != nil {
		return // Currently only one probe that must be added at prepare phase
	}
	f.SF.EnvMap, f.SF.EnvLods = float32(ubfImage), 6
	f.SF.SPH = SPH
}

func (f *Frame) AddFrameImage(view *vk.ImageView, sampler *vk.Sampler) (imageIndex vmodel.ImageIndex) {
	return f.SetFrameImage(f.cache, view, sampler)
}

func (f *Frame) ViewProjection() (projection, view mgl32.Mat4) {
	return f.SF.Projection, f.SF.View
}

func (f *Frame) copyTo(sl *vk.Slice) {
	b := *(*[unsafe.Sizeof(ShaderFrame{})]byte)(unsafe.Pointer(&f.SF))
	copy(sl.Content, b[:])
}

func (f *Frame) AddLight(l vscene.Light) {
	lPos := int(f.SF.NoLights)
	if lPos >= MAX_LIGHTS-1 {
		return
	}
	f.SF.NoLights++
	f.SF.Lights[lPos] = l
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

var kFrameImages = vk.NewKey()

// SetFrameImage sets image for whole frame (like environment) and returns its index. If imageIndex < 0 all image slots has been used
func (f *Frame) SetFrameImage(rc *vk.RenderCache, view *vk.ImageView, sampler *vk.Sampler) (ii vmodel.ImageIndex) {
	hm := rc.GetPerFrame(kFrameImages, func(ctx vk.APIContext) interface{} {
		return make(map[uintptr]vmodel.ImageIndex)
	}).(map[uintptr]vmodel.ImageIndex)
	imageIndex, ok := hm[view.Handle()]
	if ok {
		return imageIndex
	}
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
	} else {
		ii = -1
	}
	if fdDyn == nil {
		hm[view.Handle()] = ii
		fi.idx++
		return ii
	}
	if fi.idx < vmodel.ImageIndex(fdDyn.maxSize) {
		fdDyn.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
		if fi.idx < MAX_IMAGES {
			fd.ds.WriteImage(rc.Ctx, 1, uint32(fi.idx), view, sampler)
		}
		ii = fi.idx
	} else {
		ii = -1
	}
	hm[view.Handle()] = ii
	fi.idx++
	return
}

func (f *Frame) writeFrame() {
	rc := f.cache
	lt := GetFrameLayout(rc.Ctx, rc.Device)
	fd := rc.Get(kBoundFrame, func(ctx vk.APIContext) interface{} {
		return newFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	f.ds = rc.GetPerFrame(kBoundFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f.copyTo(fd.sl)
		return fd.ds
	}).(*vk.DescriptorSet)
}

func (f *Frame) writeDynamicFrame() {
	rc := f.cache
	lt := GetDynamicFrameLayout(rc.Ctx, rc.Device)
	if lt == nil {
		return
	}
	fd := rc.Get(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		return newDynamicFrameDescriptor(rc, lt)

	}).(*frameDescriptor)
	f.dsDynamic = rc.GetPerFrame(kBoundDynamicFrame, func(ctx vk.APIContext) interface{} {
		fd.buffer.Bytes(rc.Ctx)
		f.copyTo(fd.sl)
		return fd.ds
	}).(*vk.DescriptorSet)
}

func (f *Frame) BindFrame() *vk.DescriptorSet {
	if f.ds == nil {
		f.writeFrame()
	}
	return f.ds
}

func (f *Frame) BindDynamicFrame() *vk.DescriptorSet {
	if f.dsDynamic != nil {
		return f.dsDynamic
	}
	if vscene.FrameMaxDynamicSamplers == 0 {
		return nil
	}
	f.writeDynamicFrame()
	return f.dsDynamic
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
