package vk

type Framebuffer struct {
	hFb hFramebuffer
}

func (f *Framebuffer) Dispose() {
	if f.hFb != 0 {
		call_Disposable_Dispose(hDisposable(f.hFb))
		f.hFb = 0
	}
}

func (f *Framebuffer) IsValid(ctx APIContext) bool {
	if f.hFb == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

type RenderPass interface {
	Disposable
	IsValid(ctx APIContext) bool
	GetRenderPass() uintptr
	Get(ctx APIContext, key Key, cons Constructor) interface{}
}

type ForwardRenderPass struct {
	owner Owner
	dev   *Device
	hRp   hRenderPass
}

func (f *ForwardRenderPass) Get(ctx APIContext, key Key, cons Constructor) interface{} {
	return f.owner.Get(ctx, key, cons)
}

func (f *ForwardRenderPass) GetRenderPass() uintptr {
	return uintptr(f.hRp)
}

func (f *ForwardRenderPass) Dispose() {
	if f.hRp != 0 {
		f.owner.Dispose()
		call_Disposable_Dispose(hDisposable(f.hRp))
		f.hRp = 0
	}
}

// NewForwardRenderPass created a new single phase pass that supports main image and optionally a depth image (Z-buffer).
func NewForwardRenderPass(ctx APIContext, dev *Device, mainImageFormat Format, finalLayout ImageLayout, depthImageFormat Format) *ForwardRenderPass {
	if !dev.IsValid(ctx) {
		return nil
	}
	fr := &ForwardRenderPass{dev: dev}
	call_NewForwardRenderPass(ctx, dev.hDev, finalLayout, mainImageFormat, depthImageFormat, &fr.hRp)
	call_RenderPass_Init(ctx, fr.hRp)
	return fr
}

func (f *ForwardRenderPass) IsValid(ctx APIContext) bool {
	if f.hRp == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

type DepthRenderPass struct {
	owner Owner
	dev   *Device
	hRp   hRenderPass
}

func (dp *DepthRenderPass) Get(ctx APIContext, key Key, cons Constructor) interface{} {
	return dp.owner.Get(ctx, key, cons)
}

func (dp *DepthRenderPass) GetRenderPass() uintptr {
	return uintptr(dp.hRp)
}

func (dp *DepthRenderPass) Dispose() {
	if dp.hRp != 0 {
		dp.owner.Dispose()
		call_Disposable_Dispose(hDisposable(dp.hRp))
		dp.hRp = 0
	}
}
func (dp *DepthRenderPass) IsValid(ctx APIContext) bool {
	if dp.hRp == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

// NewDepthRenderPass creates are single phase render pass that only supports depth image.
// This is mainly used for shadow map rendering
func NewDepthRenderPass(ctx APIContext, dev *Device, finalLayout ImageLayout, depthImageFormat Format) *DepthRenderPass {
	if !dev.IsValid(ctx) {
		return nil
	}
	fr := &DepthRenderPass{dev: dev}
	call_NewDepthRenderPass(ctx, dev.hDev, finalLayout, depthImageFormat, &fr.hRp)
	call_RenderPass_Init(ctx, fr.hRp)
	return fr
}

// NewFramebuffer initialized new framebuffer with given attachments.
// Attachment count and types must match renderpass attachment count
func NewFramebuffer(ctx APIContext, rp RenderPass, attachments []*ImageView) *Framebuffer {
	if !rp.IsValid(ctx) {
		return nil
	}
	att := make([]hImageView, len(attachments))
	for idx, at := range attachments {
		if !at.IsValid(ctx) {
			return nil
		}
		att[idx] = at.view
	}
	fb := &Framebuffer{}
	call_RenderPass_NewFrameBuffer(ctx, hRenderPass(rp.GetRenderPass()), att, &fb.hFb)
	return fb
}

type FPlusRenderPass struct {
	owner       Owner
	dev         *Device
	hRp         hRenderPass
	extraPhases uint32
}

func (f *FPlusRenderPass) Get(ctx APIContext, key Key, cons Constructor) interface{} {
	return f.owner.Get(ctx, key, cons)
}

func (f *FPlusRenderPass) GetRenderPass() uintptr {
	return uintptr(f.hRp)
}

func (f *FPlusRenderPass) Dispose() {
	if f.hRp != 0 {
		f.owner.Dispose()
		call_Disposable_Dispose(hDisposable(f.hRp))
		f.hRp = 0
	}
}

// NewFPlusRenderPass created a new multi phase pass that supports main image and a depth image (Z-buffer) in first pass.
// In framebuffer, final image will be attachment extraPhases, depth image will be extraPhase + 1.
// You must supply extraPhase number of images that matches final image format and size. These images are used to store temporary image from each subpass
func NewFPlusRenderPass(ctx APIContext, dev *Device, extraPhases uint32, mainImageFormat Format,
	finalLayout ImageLayout, depthImageFormat Format) *FPlusRenderPass {
	if !dev.IsValid(ctx) {
		return nil
	}

	fr := &FPlusRenderPass{dev: dev, extraPhases: extraPhases}
	call_NewFPlusRenderPass(ctx, dev.hDev, extraPhases, finalLayout, mainImageFormat, depthImageFormat, &fr.hRp)
	call_RenderPass_Init(ctx, fr.hRp)
	return fr
}

func (f *FPlusRenderPass) IsValid(ctx APIContext) bool {
	if f.hRp == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}
