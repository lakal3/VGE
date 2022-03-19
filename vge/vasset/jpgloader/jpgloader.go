package jpgloader

import (
	"bytes"
	"errors"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"image/jpeg"
	"sync"
)

var register = &sync.Once{}

// Register JPEG loader that support saving and loading JPEG images using Go's image/jpeg package
func RegisterJPEGLoader() {
	register.Do(func() {
		vasset.RegisterImageLoader(JPEGLoader{})
	})
}

type JPEGLoader struct {
}

func (p JPEGLoader) SaveImage(kind string, desc vk.ImageDescription, buffer *vk.Buffer) ([]byte, error) {
	if kind != "jpeg" && kind != "jpg" {
		return nil, errors.New("Not supported " + kind)
	}
	img := image.NewNRGBA(image.Rect(0, 0, int(desc.Width), int(desc.Height)))
	copy(img.Pix, buffer.Bytes())
	wr := &bytes.Buffer{}
	err := jpeg.Encode(wr, img, &jpeg.Options{Quality: 95})
	if err != nil {
		return nil, err
	}
	return wr.Bytes(), nil
}

func (p JPEGLoader) SupportsImage(kind string) (read bool, write bool) {
	if kind == "jpeg" {
		return true, true
	}
	if kind == "jpg" {
		return true, true
	}
	return false, false
}

func (p JPEGLoader) DescribeImage(kind string, desc *vk.ImageDescription, content []byte) error {
	if kind != "jpeg" && kind != "jpg" {
		return errors.New("Not supported " + kind)
	}
	config, err := jpeg.DecodeConfig(bytes.NewBuffer(content))
	if err != nil {
		return err
	}
	desc.Height = uint32(config.Height)
	desc.Width = uint32(config.Width)
	desc.Depth = 1
	desc.Format = vk.FORMATR8g8b8a8Unorm
	desc.Layers = 1
	desc.MipLevels = 1
	return nil
}

func (p JPEGLoader) LoadImage(kind string, content []byte, buffer *vk.Buffer) error {
	if kind != "jpeg" && kind != "jpg" {
		return errors.New("Not supported " + kind)
	}
	img, err := jpeg.Decode(bytes.NewBuffer(content))
	if err != nil {

		return err
	}
	sl := buffer.Bytes()
	switch it := img.(type) {
	case *image.RGBA:
		copy(sl, it.Pix)
	case *image.NRGBA:
		copy(sl, it.Pix)
	default:
		b := img.Bounds().Max
		for y := 0; y < b.Y; y++ {
			for x := 0; x < b.X; x++ {
				off := (y*b.X + x) * 4
				c := img.At(x, y)
				r, g, b, a := c.RGBA()
				sl[off] = byte(r >> 8)
				sl[off+1] = byte(g >> 8)
				sl[off+2] = byte(b >> 8)
				sl[off+3] = byte(a >> 8)
			}
		}
	}
	return nil
}
