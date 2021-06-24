package vasset

import (
	"errors"

	"github.com/lakal3/vge/vge/vk"
)

var ErrImageNotSupported = errors.New("Image kind is not supported")

func SaveImage(ctx vk.APIContext, kind string, desc vk.ImageDescription, buffer *vk.Buffer) []byte {
	if !buffer.IsValid(ctx) || !ctx.IsValid() {
		return nil
	}
	for _, il := range imageLoaders {
		_, w := il.SupportsImage(kind)
		if w {
			return il.SaveImage(ctx, kind, desc, buffer)
		}
	}
	ctx.SetError(ErrImageNotSupported)
	return nil
}

func SupportsImage(kind string) (read bool, write bool) {
	for _, il := range imageLoaders {
		r, w := il.SupportsImage(kind)
		read = r || read
		write = w || write
	}
	return
}

func DescribeImage(ctx vk.APIContext, kind string, desc *vk.ImageDescription, content []byte) {
	if !ctx.IsValid() {
		return
	}
	for _, il := range imageLoaders {
		r, _ := il.SupportsImage(kind)
		if r {
			il.DescribeImage(ctx, kind, desc, content)
			return
		}
	}
	ctx.SetError(ErrImageNotSupported)
}

func LoadImage(ctx vk.APIContext, kind string, content []byte, buffer *vk.Buffer) {
	if !buffer.IsValid(ctx) || !ctx.IsValid() {
		return
	}
	for _, il := range imageLoaders {
		r, _ := il.SupportsImage(kind)
		if r {
			il.LoadImage(ctx, kind, content, buffer)
			return
		}
	}
	ctx.SetError(ErrImageNotSupported)
}

var imageLoaders []vk.ImageLoader

// RegisterImageLoader register loader into image loaders chain. Loaders are processed in LIFO order.
// First suitable loader for given image type and action (save, load) is selected
func RegisterImageLoader(loader vk.ImageLoader) {
	imageLoaders = append([]vk.ImageLoader{loader}, imageLoaders...)
}

// RegisterNativeImageLoader register loader implemented in vgelib (using stb_image.h image loaders)
func RegisterNativeImageLoader(ctx vk.APIContext, app *vk.Application) {
	ld := vk.NewNativeImageLoader(ctx, app)
	if ld != nil {
		RegisterImageLoader(ld)
	}
}

type RawImageLoader struct {
}

func (r RawImageLoader) SaveImage(ctx vk.APIContext, kind string, desc vk.ImageDescription, buffer *vk.Buffer) []byte {
	tmp := make([]byte, desc.ImageSize())
	copy(tmp, buffer.Bytes(ctx))
	return tmp
}

func (r RawImageLoader) SupportsImage(kind string) (read bool, write bool) {
	if kind == "raw" {
		return true, true
	}
	return false, false
}

func (r RawImageLoader) DescribeImage(ctx vk.APIContext, kind string, desc *vk.ImageDescription, content []byte) {
	ctx.SetError(errors.New("Raw images don't support describe"))
}

func (r RawImageLoader) LoadImage(ctx vk.APIContext, kind string, content []byte, buffer *vk.Buffer) {
	copy(buffer.Bytes(ctx), content)
}

func init() {
	RegisterImageLoader(DdsImageLoader{})
	RegisterImageLoader(RawImageLoader{})
}
