package vk

type ImageLoader interface {
	SaveImage(kind string, desc ImageDescription, buffer *Buffer) ([]byte, error)

	SupportsImage(kind string) (read bool, write bool)

	DescribeImage(kind string, desc *ImageDescription, content []byte) error

	LoadImage(kind string, content []byte, buffer *Buffer) error
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

func NewNativeImageLoader(app *Application) ImageLoader {
	if !app.isValid() {
		return nil
	}
	ld := newNativeImageLoad(app)
	app.owner.AddChild(ld)
	return ld
}

func newNativeImageLoad(ctx apicontext) *nativeImageLoader {
	im := &nativeImageLoader{}
	call_NewImageLoader(ctx, &im.loader)
	return im
}

// Check if image type is supported. Kind must be in lowercase!
func (il *nativeImageLoader) SupportsImage(kind string) (read bool, write bool) {
	call_ImageLoader_Supported(nullContext{}, il.loader, []byte(kind), &read, &write)
	return
}

// Describe image
func (il *nativeImageLoader) DescribeImage(kind string, desc *ImageDescription, content []byte) error {
	var ec errContext
	call_ImageLoader_Describe(&ec, il.loader, []byte(kind), desc, content)
	return ec.err
}

func (il *nativeImageLoader) LoadImage(kind string, content []byte, buffer *Buffer) error {
	var ec errContext
	call_ImageLoader_Load(&ec, il.loader, []byte(kind), content, buffer.hBuf)
	return ec.err
}

func (il *nativeImageLoader) SaveImage(kind string, desc ImageDescription, buffer *Buffer) (content []byte, err error) {
	var ec errContext
	content = []byte{}
	var regSize uint64
	call_ImageLoader_Save(&ec, il.loader, []byte(kind), &desc, buffer.hBuf, content, &regSize)
	if ec.err != nil {
		return nil, ec.err
	}
	if regSize > 0 {
		content = make([]byte, regSize)
		call_ImageLoader_Save(&ec, il.loader, []byte(kind), &desc, buffer.hBuf, content, &regSize)
	}
	return content, ec.err
}
