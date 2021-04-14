package vscene

import (
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

// Maximum dynamics samplers (=sampled images) per frame instance. You must add dynamics descriptors support to application before using
// dynamics samplers in frame. Dynamic descriptor sets and pool are allocated with BINDING_UPDATE_AFTER_BIND that should allow
// relatively large values on all Windows / Linux cards that support dynamic descriptors
var FrameMaxDynamicSamplers = uint32(0)

// ImageFrame is frame that supports binding images to Frames descriptor set
type ImageFrame interface {

	// Add image bound with frame descriptor set. If imageIndex < 0, there where no more slots left
	AddFrameImage(view *vk.ImageView, sampler *vk.Sampler) (imageIndex vmodel.ImageIndex)
}

type Camera interface {
	CameraProjection(size image.Point) (projection, view mgl32.Mat4)
}

type NullFrame struct {
}

func (n NullFrame) GetCache() *vk.RenderCache {
	return nil
}

func (n NullFrame) ViewProjection() (projection, view mgl32.Mat4) {
	return mgl32.Ident4(), mgl32.Ident4()
}

func (n NullFrame) BindFrame() *vk.DescriptorSet {
	return nil
}

type SimpleShaderFrame struct {
	Projection mgl32.Mat4
	View       mgl32.Mat4
}

type AsSimpleFrame interface {
	GetSimpleFrame() *SimpleFrame
}

func GetSimpleFrame(f vmodel.Frame) *SimpleFrame {
	asf := f.(AsSimpleFrame)
	if asf != nil {
		return asf.GetSimpleFrame()
	}
	return nil
}

type SimpleFrame struct {
	Cache *vk.RenderCache
	ds    *vk.DescriptorSet
	SSF   SimpleShaderFrame
}

func (s *SimpleFrame) GetSimpleFrame() *SimpleFrame {
	return s
}

func (s *SimpleFrame) BindFrame() *vk.DescriptorSet {
	if s.ds == nil {
		s.WriteFrame()
	}
	return s.ds
}

func (s *SimpleFrame) GetCache() *vk.RenderCache {
	return s.Cache
}

func (s *SimpleFrame) ViewProjection() (projection, view mgl32.Mat4) {
	return s.SSF.Projection, s.SSF.View
}

func (s *SimpleFrame) WriteFrame() *vk.DescriptorSet {
	uc := GetUniformCache(s.Cache)
	var sl *vk.Slice
	s.ds, sl = uc.Alloc(s.Cache.Ctx)
	s.CopyTo(sl)
	return s.ds
}

func (s *SimpleFrame) CopyTo(sl *vk.Slice) {
	b := *(*[unsafe.Sizeof(SimpleShaderFrame{})]byte)(unsafe.Pointer(&s.SSF))
	copy(sl.Content, b[:])
}
