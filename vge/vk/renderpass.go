package vk

type Framebuffer struct {
	hFb hFramebuffer
	rp  *GeneralRenderPass
}

func (f *Framebuffer) Dispose() {
	if f.hFb != 0 {
		call_Disposable_Dispose(hDisposable(f.hFb))
		f.hFb = 0
	}
}

func (f *Framebuffer) isValid() bool {
	if f.hFb == 0 {
		f.rp.dev.setError(ErrDisposed)
		return false
	}
	return f.rp.isValid()
}

type GeneralRenderPass struct {
	owner Owner
	dev   *Device
	hRp   hRenderPass
}

func (f *GeneralRenderPass) Get(key Key, cons Constructor) interface{} {
	return f.owner.Get(key, cons)
}

func (f *GeneralRenderPass) Dispose() {
	if f.hRp != 0 {
		f.owner.Dispose()
		call_Disposable_Dispose(hDisposable(f.hRp))
		f.hRp = 0
	}
}

func (f *GeneralRenderPass) isValid() bool {
	if f.hRp == 0 {
		f.dev.setError(ErrDisposed)
		return false
	}
	return true
}

// Type alias to migrate old render pass definition
type ForwardRenderPass = GeneralRenderPass
type DepthRenderPass = GeneralRenderPass

// NewGeneralRenderPass creates a new render pass with one or many color attachments.
// If hasDepth is set, last attachment is depth or depth stencil attachment, otherwise there is no depth attachment for render pass
func NewGeneralRenderPass(dev *Device, hasDepth bool, attachments []AttachmentInfo) *GeneralRenderPass {
	gr := &GeneralRenderPass{dev: dev}
	call_NewRenderPass(dev, dev.hDev, &gr.hRp, hasDepth, attachments)
	return gr
}

// NewForwardRenderPass created a new single phase pass that supports main image and optionally a depth image (Z-buffer).
// NewForwardRenderPass is badly named due to historical reason.
// ForwardRenderPass can be used to render anything with one color attachment and 0 - 1 depth attachments
func NewForwardRenderPass(dev *Device, mainImageFormat Format, finalLayout ImageLayout, depthImageFormat Format) *ForwardRenderPass {
	if !dev.isValid() {
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
	call_NewRenderPass(dev, dev.hDev, &fr.hRp, hasDepth, ai)
	return fr
}

// NewDepthRenderPass creates are single phase render pass that only supports depth image.
// This is mainly used for shadow map rendering
func NewDepthRenderPass(dev *Device, finalLayout ImageLayout, depthImageFormat Format) *DepthRenderPass {
	if !dev.IsValid() {
		return nil
	}
	fr := &DepthRenderPass{dev: dev}
	ai := []AttachmentInfo{{Format: depthImageFormat, FinalLayout: finalLayout, ClearColor: [4]float32{1, 0, 0, 0}}}
	call_NewRenderPass(dev, dev.hDev, &fr.hRp, true, ai)
	return fr
}

// NewFramebuffer initialized new framebuffer with given attachments.
// Attachment count and types must match renderpass attachment count
func NewFramebuffer(rp *GeneralRenderPass, attachments []*ImageView) *Framebuffer {
	if !rp.isValid() {
		return nil
	}
	att := make([]hImageView, len(attachments))
	for idx, at := range attachments {
		if !at.isValid() {
			return nil
		}
		att[idx] = at.view
	}
	fb := &Framebuffer{rp: rp}
	call_RenderPass_NewFrameBuffer(rp.dev, rp.hRp, att, &fb.hFb)
	return fb
}

// NewNullFramebuffer initialized new framebuffer without any attachments.
// Instead you must give size for framebuffer that is normally fetch from first image
func NewNullFramebuffer(rp *GeneralRenderPass, width, height uint32) *Framebuffer {
	if !rp.isValid() {
		return nil
	}
	fb := &Framebuffer{rp: rp}
	call_RenderPass_NewNullFrameBuffer(rp.dev, rp.hRp, width, height, &fb.hFb)
	return fb
}
