package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"os"
	"testing"
)

func TestBrush_Image(t *testing.T) {
	wood, err := os.ReadFile("../../assets/tests/wood.png")
	if err != nil {
		t.Fatal("Load wood ", err)
	}
	pngloader.RegisterPngLoader()
	err = vtestapp.Init("vdraw_brushimage", opt{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}

	mb := &vmodel.ModelBuilder{}
	woodIdx := mb.AddImage("png", wood, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	model, err := mb.ToModel(vtestapp.TestApp.Dev)
	if err != nil {
		t.Fatal("ToModel: ", err)
	}
	vtestapp.AddChild(model)
	c := NewCanvas(vtestapp.TestApp.Dev)
	brBlue := &Brush{Color: mgl32.Vec4{0, 0, 0.8, 1}}
	brGray := &Brush{Color: mgl32.Vec4{1, 1, 1, 0.5}}
	brRedGreen := &Brush{Color: mgl32.Vec4{1, 0, 0, 1}, Color2: mgl32.Vec4{-1, 1, 0, 1}, UVTransform: mgl32.Scale2D(1.0/240, 1)}
	brGreenRed := &Brush{Color: mgl32.Vec4{0, 1, 0, 1}, Color2: mgl32.Vec4{1, -1, 0, 1}, UVTransform: mgl32.Scale2D(1.0/240, 1)}
	brWood := &Brush{Color: mgl32.Vec4{1, 1, 1, 1}, UVTransform: mgl32.Scale2D(1.0/120, 1.0/120).Mul3(mgl32.Rotate3DZ(math.Pi / 4))}
	brWood.Image, brWood.Sampler = model.GetImageView(woodIdx)
	rc := vk.NewFrameCache(vtestapp.TestApp.Dev, 1)
	p := &Path{}
	p.AddRoundedRect(true, mgl32.Vec2{10, 10}, mgl32.Vec2{240, 120}, UniformCorners(10))
	dr := p.Fill()
	//	drawStroke(t, "gdraw3_rects", dr)
	mm := vtestapp.NewMainImage()
	renderTo(rc.Instances[0], mm, c, func(rp *vk.GeneralRenderPass) *vk.DrawList {
		dl := &vk.DrawList{}
		cp := c.BeginDraw(rc.Instances[0], rp, dl, c.Projection(mgl32.Vec2{0, 0}, mgl32.Vec2{1024, 768}))
		cp.Clip.To = mgl32.Vec2{1024, 768}

		cp.Draw(dr, mgl32.Vec2{10, 10}, mgl32.Vec2{1.2, 1.2}, brBlue)
		cp.Draw(dr, mgl32.Vec2{10, 180}, mgl32.Vec2{1.2, 1.2}, brRedGreen)
		cp.Draw(dr, mgl32.Vec2{10, 350}, mgl32.Vec2{1.2, 1.2}, brWood)
		cp.Draw(dr, mgl32.Vec2{10, 520}, mgl32.Vec2{1.2, 1.2}, brGray)
		cp.Clip.To = mgl32.Vec2{160, 768}
		cp.Draw(dr, mgl32.Vec2{10, 660}, mgl32.Vec2{1.2, 1.2}, brGreenRed)
		cp.Clip.From = mgl32.Vec2{160, 0}
		cp.Clip.To = mgl32.Vec2{1024, 768}
		cp.Draw(dr, mgl32.Vec2{10, 660}, mgl32.Vec2{1.2, 1.2}, brRedGreen)
		cp.End()
		return dl
	})
	rc.Dispose()

	mm.SaveKind("png", "vdraw_image", vk.IMAGELayoutTransferSrcOptimal)
	mm.Dispose()
	vtestapp.Terminate()
}
