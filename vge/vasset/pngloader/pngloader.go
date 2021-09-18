package pngloader

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"sync"

	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
)

var register = &sync.Once{}

// Register PNG loader that support saving and loading PNG images using Go's image/png package
func RegisterPngLoader() {
	register.Do(func() {
		vasset.RegisterImageLoader(PngLoader{})
	})
}

type PngLoader struct {
}

func (p PngLoader) SaveImage(ctx vk.APIContext, kind string, desc vk.ImageDescription, buffer *vk.Buffer) []byte {
	if kind != "png" {
		ctx.SetError(errors.New("Not supported " + kind))
		return nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, int(desc.Width), int(desc.Height)))
	copy(img.Pix, buffer.Bytes(ctx))
	wr := &bytes.Buffer{}
	err := png.Encode(wr, img)
	if err != nil {
		ctx.SetError(err)
	}
	return wr.Bytes()
}

func (p PngLoader) SupportsImage(kind string) (read bool, write bool) {
	if kind == "png" {
		return true, true
	}
	return false, false
}

func (p PngLoader) DescribeImage(ctx vk.APIContext, kind string, desc *vk.ImageDescription, content []byte) {
	if kind != "png" {
		ctx.SetError(errors.New("Not supported " + kind))
		return
	}
	config, err := png.DecodeConfig(bytes.NewBuffer(content))
	if err != nil {
		ctx.SetError(err)
		return
	}
	desc.Height = uint32(config.Height)
	desc.Width = uint32(config.Width)
	desc.Depth = 1
	desc.Format = vk.FORMATR8g8b8a8Unorm
	desc.Layers = 1
	desc.MipLevels = 1
}

func (p PngLoader) LoadImage(ctx vk.APIContext, kind string, content []byte, buffer *vk.Buffer) {
	if kind != "png" {
		ctx.SetError(errors.New("Not supported " + kind))
		return
	}
	img, err := png.Decode(bytes.NewBuffer(content))
	if err != nil {
		ctx.SetError(err)
		return
	}
	sl := buffer.Bytes(ctx)
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

}
