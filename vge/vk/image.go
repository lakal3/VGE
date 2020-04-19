package vk

type ImageLoader interface {
	SaveImage(ctx APIContext, kind string, desc ImageDescription, buffer *Buffer) []byte

	SupportsImage(kind string) (read bool, write bool)

	DescribeImage(ctx APIContext, kind string, desc *ImageDescription, content []byte)

	LoadImage(ctx APIContext, kind string, content []byte, buffer *Buffer)
}

type nativeImageLoader struct {
	loader hImageLoader
}

func (i *nativeImageLoader) Dispose() {
	if i.loader != 0 {
		call_Disposable_Dispose(hDisposable(i.loader))
		i.loader = 0
	}
}

func NewNativeImageLoader(ctx APIContext, app *Application) ImageLoader {
	if !app.IsValid(ctx) {
		return nil
	}
	ld := newNativeImageLoad(ctx)
	app.owner.AddChild(ld)
	return ld
}

func newNativeImageLoad(ctx APIContext) *nativeImageLoader {
	im := &nativeImageLoader{}
	call_NewImageLoader(ctx, &im.loader)
	return im
}

// Check if image type is supported. Kind must be in lowercase!
func (il *nativeImageLoader) SupportsImage(kind string) (read bool, write bool) {
	call_ImageLoader_Supported(PanicContext{}, il.loader, []byte(kind), &read, &write)
	return
}

// Describe image
func (il *nativeImageLoader) DescribeImage(ctx APIContext, kind string, desc *ImageDescription, content []byte) {
	call_ImageLoader_Describe(ctx, il.loader, []byte(kind), desc, content)
}

func (il *nativeImageLoader) LoadImage(ctx APIContext, kind string, content []byte, buffer *Buffer) {
	call_ImageLoader_Load(ctx, il.loader, []byte(kind), content, buffer.hBuf)
}

func (il *nativeImageLoader) SaveImage(ctx APIContext, kind string, desc ImageDescription, buffer *Buffer) (content []byte) {
	content = []byte{}
	var regSize uint64
	call_ImageLoader_Save(ctx, il.loader, []byte(kind), &desc, buffer.hBuf, content, &regSize)
	if regSize > 0 {
		content = make([]byte, regSize)
		call_ImageLoader_Save(ctx, il.loader, []byte(kind), &desc, buffer.hBuf, content, &regSize)
	}
	return
}
