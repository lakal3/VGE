package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

const testChars = "VGEvgestio"

func TestCanvas_DrawChar(t *testing.T) {
	ft, err := LoadFontFile("../../assets/fonts/RobotoMono_Medium.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}
	for _, r := range testChars {
		p := &Path{}
		err = ft.DrawChar(p, 64, mgl32.Vec2{10, 70}, r)
		if err != nil {
			t.Fatal("Draw char: ", r, " ", err)
		}
		f := p.Fill()
		if f.err != nil {
			t.Error(f.err)
			return
		}
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		if f.err != nil {
			t.Error("Generate error ", f.err)
		}
		for idx := 0; idx < len(f.da); idx++ {
			drawArea(t, img, f, idx)
		}

		fpath := filepath.Join(os.Getenv("VGE_TEST_DIR"), "vdraw_"+string(r)+".png")
		fPng, err := os.Create(fpath)
		if err != nil {
			t.Error(err)
			return
		}
		_ = png.Encode(fPng, img)
		_ = fPng.Close()
	}
}
