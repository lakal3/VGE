package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestFilled_build(t *testing.T) {
	p := &Path{}
	p.AddRoundedRect(true, mgl32.Vec2{5, 5}, mgl32.Vec2{120, 50}, UniformCorners(8))
	p.AddRoundedRect(false, mgl32.Vec2{15, 15}, mgl32.Vec2{100, 30}, UniformCorners(4))
	fl := Filled{}
	f := fl.build(p)
	img := image.NewRGBA(image.Rect(0, 0, 130, 60))
	if f.err != nil {
		t.Error("Generate error ", f.err)
	}
	for idx := 0; idx < len(f.areas); idx++ {
		drawFillArea(t, img, f, idx)
	}

	fpath := filepath.Join(os.Getenv("VGE_TEST_DIR"), "vdraw_filled1.png")
	fPng, err := os.Create(fpath)
	if err != nil {
		t.Error(err)
		return
	}
	_ = png.Encode(fPng, img)
	_ = fPng.Close()
}

func TestPath_Fill(t *testing.T) {
	p := &Path{}
	p.AddRoundedRect(true, mgl32.Vec2{10, 10}, mgl32.Vec2{240, 240}, UniformCorners(10))
	p.AddCornerRect(false, mgl32.Vec2{25, 25}, mgl32.Vec2{220, 220}, UniformCorners(8))
	f := p.Fill()
	img := image.NewRGBA(image.Rect(0, 0, 260, 260))
	if f.err != nil {
		t.Error("Generate error ", f.err)
	}
	for idx := 0; idx < len(f.da); idx++ {
		drawArea(t, img, f, idx)
	}

	fpath := filepath.Join(os.Getenv("VGE_TEST_DIR"), "vdraw_filled2.png")
	fPng, err := os.Create(fpath)
	if err != nil {
		t.Error(err)
		return
	}
	_ = png.Encode(fPng, img)
	_ = fPng.Close()
}

func drawArea(t *testing.T, img *image.RGBA, f *Filled, idx int) {
	da := f.da[idx]
	var dsPrev DrawSegment
	for i := da.From; i < da.To; i++ {
		ds := f.ds[i]
		if i != da.From {
			drawLine(img, color.RGBA{B: 255, A: 255}, line{from: mgl32.Vec2{dsPrev.V1, dsPrev.V3}, to: mgl32.Vec2{ds.V1, ds.V3}})
			drawLine(img, color.RGBA{G: 255, A: 255}, line{from: mgl32.Vec2{dsPrev.V2, dsPrev.V3}, to: mgl32.Vec2{ds.V2, ds.V3}})
		}
		dsPrev = ds
	}
	// return
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: da.Min, to: mgl32.Vec2{da.Max[0], da.Min[1]}})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{da.Min[0], da.Max[1]}, to: da.Max})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: da.Min, to: mgl32.Vec2{da.Min[0], da.Max[1]}})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{da.Max[0], da.Min[1]}, to: da.Max})
}

func drawFillArea(t *testing.T, img *image.RGBA, f *filler, idx int) {
	var min float32
	var max float32
	for idx, lIdx := range f.areas[idx].lefts {
		var l = f.lines[lIdx]
		if idx == 0 {
			min, max = l.from[0], l.from[0]
		}
		min = min32(min, l.from[0])
		min = min32(min, l.to[0])
		max = max32(max, l.from[0])
		max = max32(max, l.to[0])
		drawLine(img, color.RGBA{G: 255, A: 255}, l)
	}
	for _, lIdx := range f.areas[idx].rights {
		var l = f.lines[lIdx]
		min = min32(min, l.from[0])
		min = min32(min, l.to[0])
		max = max32(max, l.from[0])
		max = max32(max, l.to[0])

		drawLine(img, color.RGBA{B: 255, A: 255}, l)
	}
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{min, f.areas[idx].fromY},
		to: mgl32.Vec2{max, f.areas[idx].fromY}})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{min, f.areas[idx].toY},
		to: mgl32.Vec2{max, f.areas[idx].toY}})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{min, f.areas[idx].fromY},
		to: mgl32.Vec2{min, f.areas[idx].toY}})
	drawLine(img, color.RGBA{R: 255, A: 255}, line{from: mgl32.Vec2{max, f.areas[idx].fromY},
		to: mgl32.Vec2{max, f.areas[idx].toY}})
}

func drawLine(img *image.RGBA, rgba color.RGBA, l line) {
	ls := l.to.Sub(l.from).Len()
	if ls < 0.1 {
		return
	}
	step := float32(0.5 / ls)
	for u := float32(0); u <= 1; u += step {
		p := l.at(u)
		img.Set(int(p[0]), int(p[1]), rgba)
	}
}
