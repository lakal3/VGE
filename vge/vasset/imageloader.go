package vasset

import (
	"errors"

	"github.com/lakal3/vge/vge/vk"
)

var ErrImageNotSupported = errors.New("Image kind is not supported")

func SaveImage(kind string, desc vk.ImageDescription, buffer *vk.Buffer) ([]byte, error) {
	for _, il := range imageLoaders {
		_, w := il.SupportsImage(kind)
		if w {
			return il.SaveImage(kind, desc, buffer)
		}
	}
	return nil, ErrImageNotSupported
}

func SupportsImage(kind string) (read bool, write bool) {
	for _, il := range imageLoaders {
		r, w := il.SupportsImage(kind)
		read = r || read
		write = w || write
	}
	return
}

func DescribeImage(kind string, desc *vk.ImageDescription, content []byte) error {
	for _, il := range imageLoaders {
		r, _ := il.SupportsImage(kind)
		if r {
			return il.DescribeImage(kind, desc, content)
		}
	}
	return ErrImageNotSupported
}

func LoadImage(kind string, content []byte, buffer *vk.Buffer) error {
	for _, il := range imageLoaders {
		r, _ := il.SupportsImage(kind)
		if r {
			return il.LoadImage(kind, content, buffer)
		}
	}
	return ErrImageNotSupported
}

var imageLoaders []vk.ImageLoader

// RegisterImageLoader register loader into image loaders chain. Loaders are processed in LIFO order.
// First suitable loader for given image type and action (save, load) is selected
func RegisterImageLoader(loader vk.ImageLoader) {
	imageLoaders = append([]vk.ImageLoader{loader}, imageLoaders...)
}

// RegisterNativeImageLoader register loader implemented in vgelib (using stb_image.h image loaders)
func RegisterNativeImageLoader(app *vk.Application) {
	ld := vk.NewNativeImageLoader(app)
	if ld != nil {
		RegisterImageLoader(ld)
	}
}

type RawImageLoader struct {
}

func (r RawImageLoader) SaveImage(kind string, desc vk.ImageDescription, buffer *vk.Buffer) ([]byte, error) {
	tmp := make([]byte, desc.ImageSize())
	copy(tmp, buffer.Bytes())
	return tmp, nil
}

func (r RawImageLoader) SupportsImage(kind string) (read bool, write bool) {
	if kind == "raw" {
		return true, true
	}
	return false, false
}

func (r RawImageLoader) DescribeImage(kind string, desc *vk.ImageDescription, content []byte) error {
	return errors.New("Raw images don't support describe")
}

func (r RawImageLoader) LoadImage(kind string, content []byte, buffer *vk.Buffer) error {
	copy(buffer.Bytes(), content)
	return nil
}

func init() {
	RegisterImageLoader(DdsImageLoader{})
	RegisterImageLoader(RawImageLoader{})
}
