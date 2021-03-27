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

type GeneralRenderPass struct {
	owner Owner
	dev   *Device
	hRp   hRenderPass
}

func (f *GeneralRenderPass) Get(ctx APIContext, key Key, cons Constructor) interface{} {
	return f.owner.Get(ctx, key, cons)
}

func (f *GeneralRenderPass) GetRenderPass() uintptr {
	return uintptr(f.hRp)
}

func (f *GeneralRenderPass) Dispose() {
	if f.hRp != 0 {
		f.owner.Dispose()
		call_Disposable_Dispose(hDisposable(f.hRp))
		f.hRp = 0
	}
}

func (f *GeneralRenderPass) IsValid(ctx APIContext) bool {
	if f.hRp == 0 {
		ctx.SetError(ErrDisposed)
		return false
	}
	return true
}

// Type alias to migrate old render pass definition
type ForwardRenderPass = GeneralRenderPass
type DepthRenderPass = GeneralRenderPass

// NewGeneralRenderPass creates a new render pass with one or many color attachments.
// If hasDepth is set, last attachment is depth or depth stencil attachment, otherwise there is no depth attachment for render pass
func NewGeneralRenderPass(ctx APIContext, dev *Device, hasDepth bool, attachments []AttachmentInfo) *GeneralRenderPass {
	gr := &GeneralRenderPass{dev: dev}
	call_NewRenderPass(ctx, dev.hDev, &gr.hRp, hasDepth, attachments)
	return gr
}

// NewForwardRenderPass created a new single phase pass that supports main image and optionally a depth image (Z-buffer).
// NewForwardRenderPass is badly named due to historical reason.
// ForwardRenderPass can be used to render anything with one color attachment and 0 - 1 depth attachments
func NewForwardRenderPass(ctx APIContext, dev *Device, mainImageFormat Format, finalLayout ImageLayout, depthImageFormat Format) *ForwardRenderPass {
	if !dev.IsValid(ctx) {
		return nil
	}
	fr := &ForwardRenderPass{dev: dev}
	var ai []AttachmentInfo
	ai = append(ai, AttachmentInfo{Format: mainImageFormat, FinalLayout: finalLayout, InitialLayout: IMAGELayoutUndefined,
		ClearColor: [4]float32{0.2, 0.2, 0.2, 1}})
	hasDepth := false
	if depthImageFormat != FORMATUndefined {
		hasDepth = true
		ai = append(ai, AttachmentInfo{Format: depthImageFormat, InitialLayout: IMAGELayoutUndefined, FinalLayout: IMAGELayoutUndefined,
			ClearColor: [4]float32{1, 0, 0, 0}})
	}
	call_NewRenderPass(ctx, dev.hDev, &fr.hRp, hasDepth, ai)
	return fr
}

// NewDepthRenderPass creates are single phase render pass that only supports depth image.
// This is mainly used for shadow map rendering
func NewDepthRenderPass(ctx APIContext, dev *Device, finalLayout ImageLayout, depthImageFormat Format) *DepthRenderPass {
	if !dev.IsValid(ctx) {
		return nil
	}
	fr := &DepthRenderPass{dev: dev}
	ai := []AttachmentInfo{{Format: depthImageFormat, FinalLayout: finalLayout, ClearColor: [4]float32{1, 0, 0, 0}}}
	call_NewRenderPass(ctx, dev.hDev, &fr.hRp, true, ai)
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
