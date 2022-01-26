package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
	"image"
	"strings"
	"testing"
)

const testLipsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Duis sit amet augue risus. 
Cras ultricies, nunc ut interdum volutpat, lorem sem porttitor arcu, a pellentesque purus mi ut libero. 
Vestibulum ornare varius mauris, a sodales augue sodales eget. Praesent consectetur nisl sapien, vitae venenatis odio feugiat eget. 
Nullam molestie maximus eros a ullamcorper. In sit amet pharetra tortor. 
Integer venenatis velit non mauris suscipit, sit amet luctus purus volutpat. Nunc pharetra semper hendrerit. 
Aenean lacinia sem sed eros iaculis aliquam. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae;
Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Cras tincidunt purus eu odio tempus,
non imperdiet urna lacinia. Vivamus id mi magna. Curabitur turpis nibh, faucibus eget velit non, ullamcorper hendrerit neque.
Morbi porttitor justo commodo leo fringilla tristique. Proin sollicitudin congue diam, nec pulvinar nibh sollicitudin vel.

Mauris a commodo sem. Nullam lobortis nunc vel varius gravida. In tincidunt aliquet pellentesque. Praesent bibendum tempus molestie.
Curabitur ornare malesuada rhoncus. Duis in fringilla lacus, sed auctor enim. Duis ut sem quis risus ullamcorper blandit.

Suspendisse volutpat leo est, at venenatis quam vulputate a. Donec pulvinar congue enim, vitae posuere velit sagittis at.
Maecenas eget aliquet nibh. Morbi eget congue ligula, quis ornare orci. Donec convallis ipsum vel condimentum ullamcorper.
Nullam accumsan ex vel egestas ultrices. Sed rutrum magna at turpis molestie, at malesuada mauris bibendum. Mauris ac gravida sapien.
Sed dapibus ornare dapibus. Aliquam ut pellentesque eros. Nam ut risus feugiat, dignissim diam vitae, rutrum nulla.

Sed non quam dui. Aliquam erat volutpat. Quisque vel euismod lorem, quis ullamcorper leo. Lorem ipsum dolor sit amet, 
consectetur adipiscing elit. Suspendisse at ultrices libero. Pellentesque cursus accumsan mattis. Nulla malesuada enim nulla,
vitae accumsan libero mattis vel. Duis semper ex varius leo dictum ornare. Donec malesuada consectetur sagittis. 
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi lobortis, elit nec porttitor pellentesque, eros odio faucibus augue,
a malesuada ipsum enim at odio. Sed lacinia dignissim purus sit amet vestibulum. Aliquam facilisis ipsum et nisl fringilla tempus.
Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Donec commodo dolor ante,
vitae vestibulum velit hendrerit ac. Nam et commodo orci.

Mauris a commodo sem. Nullam lobortis nunc vel varius gravida. In tincidunt aliquet pellentesque. Praesent bibendum tempus molestie.
Curabitur ornare malesuada rhoncus. Duis in fringilla lacus, sed auctor enim. Duis ut sem quis risus ullamcorper blandit.

Suspendisse volutpat leo est, at venenatis quam vulputate a. Donec pulvinar congue enim, vitae posuere velit sagittis at.
Maecenas eget aliquet nibh. Morbi eget congue ligula, quis ornare orci. Donec convallis ipsum vel condimentum ullamcorper.
Nullam accumsan ex vel egestas ultrices. Sed rutrum magna at turpis molestie, at malesuada mauris bibendum. Mauris ac gravida sapien.
Sed dapibus ornare dapibus. Aliquam ut pellentesque eros. Nam ut risus feugiat, dignissim diam vitae, rutrum nulla.

Sed non quam dui. Aliquam erat volutpat. Quisque vel euismod lorem, quis ullamcorper leo. Lorem ipsum dolor sit amet, 
consectetur adipiscing elit. Suspendisse at ultrices libero. Pellentesque cursus accumsan mattis. Nulla malesuada enim nulla,
vitae accumsan libero mattis vel. Duis semper ex varius leo dictum ornare. Donec malesuada consectetur sagittis. 
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi lobortis, elit nec porttitor pellentesque, eros odio faucibus augue,
a malesuada ipsum enim at odio. Sed lacinia dignissim purus sit amet vestibulum. Aliquam facilisis ipsum et nisl fringilla tempus.
Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Donec commodo dolor ante,
vitae vestibulum velit hendrerit ac. Nam et commodo orci.
`

type opt struct {
}

func (o opt) InitOption() {
	vtestapp.TestApp.App.AddDynamicDescriptors()
	vscene.FrameMaxDynamicSamplers = 1024
}

func TestCanvas_DrawRect(t *testing.T) {
	pngloader.RegisterPngLoader()
	err := vtestapp.Init("drawrect", opt{}, vtestapp.UnitTest{T: t})
	if err != nil {
		t.Fatal("Init app ", err)
	}
	err = CompileShapes(vtestapp.TestApp.Dev)
	if err != nil {
		t.Fatal("Compile shapes ", err)
	}

	c := NewCanvas(vtestapp.TestApp.Dev)
	brBlue := Brush{Color: mgl32.Vec4{0, 0, 0.8, 1}}
	brGreen := Brush{Color: mgl32.Vec4{0, 0.8, 0, 1}}
	brRed := Brush{Color: mgl32.Vec4{0.8, 0, 0, 1}}
	fc := vk.NewFrameCache(vtestapp.TestApp.Dev, 1)
	p := &Path{}
	p.AddRoundedRect(true, mgl32.Vec2{10, 10}, mgl32.Vec2{240, 240}, UniformCorners(10))
	p.AddCornerRect(false, mgl32.Vec2{25, 25}, mgl32.Vec2{220, 220}, UniformCorners(8))
	r := Rect{Area: Area{From: mgl32.Vec2{150, 150}, To: mgl32.Vec2{200, 200}}}
	er := Border{Area: Area{From: mgl32.Vec2{0}, To: mgl32.Vec2{50, 50}}, Edges: Edges{Left: 1, Right: 2, Top: 3, Bottom: 4}}
	rr := RoudedRect{
		Area:    Area{From: mgl32.Vec2{150, 300}, To: mgl32.Vec2{200, 400}},
		Corners: Corners{TopLeft: 5, TopRight: 8, BottomLeft: 0, BottomRight: 15}}
	rb := RoundedBorder{
		Area:    Area{From: mgl32.Vec2{450, 300}, To: mgl32.Vec2{600, 400}},
		Corners: Corners{TopLeft: 5, TopRight: 8, BottomLeft: 15, BottomRight: 15},
		Edges:   Edges{Left: 4, Right: 3, Top: 2, Bottom: 1.5},
	}
	l := Line{From: mgl32.Vec2{500, 520}, To: mgl32.Vec2{700, 540}, Thickness: 3}
	dr := p.Fill()
	//	drawStroke(t, "gdraw3_rects", dr)
	mm := vtestapp.NewMainImage()
	renderTo(fc.Instances[0], mm, c, func(rp *vk.GeneralRenderPass) *vk.DrawList {
		dl := &vk.DrawList{}
		cp := c.BeginDraw(fc.Instances[0], rp, dl, c.Projection(mgl32.Vec2{0, 0}, mgl32.Vec2{1024, 768}))
		cp.Clip.To = mgl32.Vec2{1024, 768}
		cp.Draw(dr, mgl32.Vec2{0, 0}, mgl32.Vec2{3, 3}, &brBlue)
		cp.Draw(r, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &brGreen)
		cp.Draw(r, mgl32.Vec2{100, 0}, mgl32.Vec2{1, 1}, &brGreen)
		cp.Draw(r, mgl32.Vec2{200, 200}, mgl32.Vec2{0.5, 0.5}, &brGreen)
		cp.Draw(er, mgl32.Vec2{350, 150}, mgl32.Vec2{1, 1}, &brBlue)
		cp.Draw(rr, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &brRed)
		cp.Draw(rb, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &brRed)
		cp.Draw(l, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, &brGreen)
		cp.End()
		return dl
	})
	fc.Dispose()

	mm.SaveKind("png", "vdraw_rect", vk.IMAGELayoutTransferSrcOptimal)
	mm.Dispose()
	vtestapp.Terminate()
}

func TestCanvas_DrawVGE(t *testing.T) {
	f, err := LoadFontFile("../../assets/fonts/Lato-Regular.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}
	pngloader.RegisterPngLoader()
	err = vtestapp.Init("drawtext", opt{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}
	c := NewCanvas(vtestapp.TestApp.Dev)
	fc := vk.NewFrameCache(vtestapp.TestApp.Dev, 1)
	p := &Path{}
	p.AddCornerRect(true, mgl32.Vec2{5, 5}, mgl32.Vec2{1014, 758}, UniformCorners(8))
	dr := p.Fill()
	mm := vtestapp.NewMainImage()
	brWrite := &Brush{Color: mgl32.Vec4{1, 1, 1, 1}}
	brBlack := &Brush{Color: mgl32.Vec4{0, 0, 0, 1}}
	renderTo(fc.Instances[0], mm, c, func(rp *vk.GeneralRenderPass) *vk.DrawList {
		dl := &vk.DrawList{}
		cp := c.BeginDraw(fc.Instances[0], rp, dl, c.Projection(mgl32.Vec2{0, 0}, mgl32.Vec2{1024, 768}))
		cp.Clip.To = mgl32.Vec2{1024, 768}
		cp.Draw(dr, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, brWrite)
		at := mgl32.Vec2{50, 20}
		for size := float32(8); size <= 64; size += 4 {
			cp.DrawText(f, size, at, brBlack, "VGE vgem")
			at = at.Add(mgl32.Vec2{0, size * 1.3})
		}
		cp.End()
		return dl
	})
	fc.Dispose()

	mm.SaveKind("png", "vdraw_vge", vk.IMAGELayoutTransferSrcOptimal)
	mm.Dispose()
	vtestapp.Terminate()
}

func TestCanvas_DrawVGEGlyph(t *testing.T) {
	f, err := LoadFontFile("../../assets/fonts/Lato-Regular.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}
	pngloader.RegisterPngLoader()
	err = vtestapp.Init("drawtext", opt{}, vtestapp.UnitTest{T: t})
	if err != nil {
		t.Fatal("Init app: ", err)
	}

	c := NewCanvas(vtestapp.TestApp.Dev)
	rc := vk.NewFrameCache(vtestapp.TestApp.Dev, 1)
	p := &Path{}
	p.AddCornerRect(true, mgl32.Vec2{5, 5}, mgl32.Vec2{1014, 758}, UniformCorners(8))
	dr := p.Fill()
	gs := &GlyphSet{}
	for _, ch := range "WGEAM vgem" {
		err = f.AddToSet(ch, gs)
		if err != nil {
			t.Fatal("Add to set ", err)
		}
	}
	gs.Build(vtestapp.TestApp.Dev, image.Pt(64, 64))
	vtestapp.AddChild(gs)
	mm := vtestapp.NewMainImage()
	brWrite := &Brush{Color: mgl32.Vec4{1, 1, 1, 1}}
	brBlack := &Brush{Color: mgl32.Vec4{0, 0, 0, 1}}
	renderTo(rc.Instances[0], mm, c, func(rp *vk.GeneralRenderPass) *vk.DrawList {
		dl := &vk.DrawList{}
		cp := c.BeginDraw(rc.Instances[0], rp, dl, c.Projection(mgl32.Vec2{0, 0}, mgl32.Vec2{1024, 768}))
		cp.Clip.To = mgl32.Vec2{1024, 768}

		cp.Draw(dr, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, brWrite)
		at := mgl32.Vec2{50, 20}
		for size := float32(8); size <= 64; size += 4 {
			cp.DrawText(f, size, at, brBlack, "WGEAM vgem")
			if err != nil {
				t.Error(err)
			}
			at = at.Add(mgl32.Vec2{0, size * 1.3})
		}
		cp.End()
		return dl
	})
	rc.Dispose()

	mm.SaveKind("png", "vdraw_vgeglyph", vk.IMAGELayoutTransferSrcOptimal)
	mm.Dispose()
	vtestapp.Terminate()
}

func TestCanvas_DrawText(t *testing.T) {
	f, err := LoadFontFile("../../assets/fonts/Roboto-Regular.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}
	err = vtestapp.Init("draw_text", opt{}, vtestapp.PngSupport{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}

	mm := vtestapp.NewMainImageDesc(vk.ImageDescription{Layers: 1, MipLevels: 1, Width: 1536, Height: 1536, Depth: 1, Format: vk.FORMATR8g8b8a8Unorm},
		vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageColorAttachmentBit)
	fc := vk.NewFrameCache(vtestapp.TestApp.Dev, 2)
	c := NewCanvas(vtestapp.TestApp.Dev)
	delta := 0.0
	for idx := 0; idx < 40; idx++ {
		delta += renderText(t, fc.Instances[idx%2], c, f, mm)
	}
	t.Log("Time ", delta/1000000000.0)
	fc.Dispose()
	mm.SaveKind("png", "vdraw_text", vk.IMAGELayoutTransferSrcOptimal)
	vtestapp.Terminate()

}

func TestCanvas_DrawGlyphText(t *testing.T) {
	f, err := LoadFontFile("../../assets/fonts/Roboto-Regular.ttf")
	if err != nil {
		t.Fatal("Load font: ", err)
	}
	gs := &GlyphSet{}
	for ch := rune(33); ch < 127; ch++ {
		err = f.AddToSet(ch, gs)
		if err != nil {
			t.Fatal("Add to set ", err)
		}
	}
	err = vtestapp.Init("draw_text", opt{}, vtestapp.PngSupport{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}

	mm := vtestapp.NewMainImageDesc(vk.ImageDescription{Layers: 1, MipLevels: 1, Width: 1536, Height: 1536, Depth: 1, Format: vk.FORMATR8g8b8a8Unorm},
		vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageColorAttachmentBit)
	fc := vk.NewFrameCache(vtestapp.TestApp.Dev, 2)
	c := NewCanvas(vtestapp.TestApp.Dev)
	gs.Build(vtestapp.TestApp.Dev, image.Pt(32, 32))
	vtestapp.AddChild(gs)
	delta := 0.0
	for idx := 0; idx < 40; idx++ {
		delta += renderText(t, fc.Instances[idx%2], c, f, mm)
	}
	t.Log("Time ", delta/1000000000.0)
	fc.Dispose()
	mm.SaveKind("png", "vdraw_glyphtext", vk.IMAGELayoutTransferSrcOptimal)
	vtestapp.Terminate()

}

func renderText(t *testing.T, fi *vk.FrameInstance, c *Canvas, f *Font, mm *vtestapp.MainImage) float64 {

	brWrite := &Brush{Color: mgl32.Vec4{1, 1, 1, 1}}
	brBlack := &Brush{Color: mgl32.Vec4{0, 0, 0, 1}}

	lines := strings.Split(testLipsum, "\n")

	p := Path{}
	p.AddRect(true, mgl32.Vec2{0, 0}, mgl32.Vec2{1536, 1536})
	rect := p.Fill()
	delta := renderTo(fi, mm, c, func(rp *vk.GeneralRenderPass) *vk.DrawList {
		dl := &vk.DrawList{}
		cp := c.BeginDraw(fi, rp, dl, c.Projection(mgl32.Vec2{0, 0}, mgl32.Vec2{1536, 1536}))
		cp.Clip.To = mgl32.Vec2{1536, 1536}
		cp.Draw(rect, mgl32.Vec2{0, 0}, mgl32.Vec2{1, 1}, brWrite)
		for y, line := range lines {
			//c.Color = mgl32.Vec4{0, 0, 0, 0.5}
			//err := c.DrawText(dc, f, 14, mgl32.Vec2{21, float32(y+1)*17 + 41}, line)
			//if err != nil {
			//	t.Fatal("Draw text: ", err)
			//}
			cp.DrawText(f, 14, mgl32.Vec2{20, float32(y+1)*17 + 40}, brBlack, line)
		}
		cp.End()
		return dl
	})
	return delta
}

func renderTo(fi *vk.FrameInstance, mm *vtestapp.MainImage, c *Canvas, drawIt func(rp *vk.GeneralRenderPass) *vk.DrawList) float64 {
	fi.BeginFrame()
	c.Reserve(fi)
	fi.Commit()
	rp := vk.NewGeneralRenderPass(c.dev, false, []vk.AttachmentInfo{
		vk.AttachmentInfo{Format: mm.Desc.Format, FinalLayout: vk.IMAGELayoutTransferSrcOptimal,
			ClearColor: [4]float32{0.2, 0.2, 0.2, 1}},
	})
	defer rp.Dispose()
	fb := vk.NewFramebuffer(rp, []*vk.ImageView{mm.Image.DefaultView()})
	defer fb.Dispose()
	cmd := vk.NewCommand(c.dev, vk.QUEUEGraphicsBit, true)
	defer cmd.Dispose()
	tp := vk.NewTimerPool(c.dev, 2)
	defer tp.Dispose()
	cmd.Begin()
	cmd.WriteTimer(tp, 0, vk.PIPELINEStageTopOfPipeBit)
	cmd.BeginRenderPass(rp, fb)
	drawList := drawIt(rp)
	cmd.Draw(drawList)
	cmd.EndRenderPass()
	cmd.WriteTimer(tp, 1, vk.PIPELINEStageBottomOfPipeBit)
	fi.Freeze()
	cmd.Submit()
	cmd.Wait()
	deltas := tp.Get()
	return deltas[1] - deltas[0]
}
