package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/materialicons"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

var kImageLayout = vk.NewKeys(2)
var kImagePipeline = vk.NewKey()

func (iv *imageViewer) Reserve(fi *vk.FrameInstance) {
	if iv.err != nil {
		return
	}
	if iv.state == 1 {
		iv.state = 2
		iv.rm()
		return
	}
	la := iv.getLayout()
	fi.ReserveDescriptor(la)
	fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, 256)
}

func (iv *imageViewer) PreRender(fi *vk.FrameInstance) (done vapp.Completed) {
	return nil
}

func (iv *imageViewer) PostRender(fi *vk.FrameInstance) {
}

var kImageFrameView = vk.NewKey()

func (iv *imageViewer) Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass) {
	if iv.err != nil {
		return
	}

	if iv.state == 2 {
		return
	}
	pl := iv.getPipeline(rp)
	ds := fi.AllocDescriptor(iv.getLayout())
	sl := fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, 256)
	fr := (*imageframe)(unsafe.Pointer(&sl.Bytes()[0]))
	desc := fi.Output.Describe()
	fr.area = mgl32.Vec4{-0.5, (StatHeight+1)/float32(desc.Height)*2 - 1, 1, 1}
	iv.area.From = mgl32.Vec2{float32(desc.Width / 4), StatHeight}
	iv.area.To = mgl32.Vec2{float32(desc.Width), float32(desc.Height)}
	var translate mgl32.Mat3
	if iv.rot%2 == 0 {
		translate = mgl32.Scale2D(iv.scale, iv.scale/iv.aspect).Mul3(mgl32.Translate2D(iv.offset[0], iv.offset[1]))
	} else {
		translate = mgl32.Scale2D(iv.scale/iv.aspect, iv.scale).Mul3(mgl32.Translate2D(iv.offset[0], iv.offset[1]))
	}
	fr.uv1 = translate.Row(0).Vec4(1 / iv.scale)
	fr.uv2 = translate.Row(1).Vec4(float32(iv.rot))
	ds.WriteSlice(0, 0, sl)
	view := fi.Get(kImageFrameView, func() interface{} {
		r := vk.ImageRange{LayerCount: 1, LevelCount: 1, FirstLayer: iv.layer, FirstMipLevel: iv.mip}
		return vk.NewImageView(iv.img, &r)
	}).(*vk.ImageView)
	ds.WriteView(1, 0, view, vk.IMAGELayoutShaderReadOnlyOptimal, vmodel.GetDefaultSampler(iv.dev))
	var dl vk.DrawList
	dl.Draw(pl, 0, 6).AddDescriptors(ds)
	cmd.Draw(&dl)
}

type imageframe struct {
	area mgl32.Vec4
	uv1  mgl32.Vec4
	uv2  mgl32.Vec4
}

func (iv *imageViewer) getLayout() *vk.DescriptorLayout {
	la1 := iv.dev.Get(kImageLayout, func() interface{} {
		return vk.NewDescriptorLayout(iv.dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageFragmentBit|vk.SHADERStageVertexBit, 1)
	}).(*vk.DescriptorLayout)
	la2 := iv.dev.Get(kImageLayout+1, func() interface{} {
		return la1.AddBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 1)
	}).(*vk.DescriptorLayout)
	return la2
}

func (iv *imageViewer) getPipeline(rp *vk.GeneralRenderPass) *vk.GraphicsPipeline {
	return rp.Get(kImagePipeline, func() interface{} {
		pl := vk.NewGraphicsPipeline(iv.dev)
		pl.AddLayout(iv.getLayout())
		sp := imageShaders.Get("imageview")
		pl.AddShader(vk.SHADERStageFragmentBit, sp.Fragment)
		pl.AddShader(vk.SHADERStageVertexBit, sp.Vertex)
		pl.Create(rp)
		return pl
	}).(*vk.GraphicsPipeline)
}

var imageShaders *shaders.Pack

type imageViewer struct {
	path  string
	info  os.FileInfo
	err   error
	kind  string
	rm    func()
	mb    *vk.MemoryPool
	img   *vk.Image
	dev   *vk.Device
	desc  vk.ImageDescription
	area  vdraw.Area
	state uint32

	offset mgl32.Vec2
	scale  float32
	aspect float32
	rot    int
	layer  uint32
	mip    uint32

	movePos mgl32.Vec2
	drag    bool
}

func (iv *imageViewer) HandleEvent(event vapp.Event) {
	md, ok := event.(*vapp.MouseDownEvent)
	if ok {
		pos := mgl32.Vec2{float32(md.MousePos.X), float32(md.MousePos.Y)}
		if !iv.area.Contains(pos) {
			return
		}
		if md.Button == 0 {
			iv.movePos = pos
			iv.drag = true
			md.SetHandled()
		}
	}
	mm, ok := event.(*vapp.MouseMoveEvent)
	if ok {
		if !mm.HasMods(vapp.MODMouseButton1) || !iv.drag {
			iv.drag = false
			return
		}
		pos := mgl32.Vec2{float32(mm.MousePos.X), float32(mm.MousePos.Y)}
		dPos := pos.Sub(iv.movePos).Mul(1 / iv.scale).Mul(2 / float32(iv.area.Width()))
		iv.offset = iv.offset.Add(dPos)
		iv.movePos = pos
		mm.SetHandled()
	}
	se, ok := event.(*vapp.ScrollEvent)
	if ok {
		pos := mgl32.Vec2{float32(se.MousePos.X), float32(se.MousePos.Y)}
		if !iv.area.Contains(pos) {
			return
		}
		r := se.Range.Y
		iv.scale *= 1 + float32(r)/20
		se.SetHandled()
	}
	kd, ok := event.(*vapp.KeyDownEvent)
	if ok {
		pos := mgl32.Vec2{float32(kd.MousePos.X), float32(kd.MousePos.Y)}
		if !iv.area.Contains(pos) {
			return
		}
		if kd.KeyCode == 'E' {
			iv.rot--
			if iv.rot < 0 {
				iv.rot = 3
			}
			kd.SetHandled()
		}
		if kd.KeyCode == 'Q' {
			iv.rot = (iv.rot + 1) % 4
			kd.SetHandled()
		}
	}
}

func (iv *imageViewer) View() vapp.View {
	if iv.img != nil {
		return iv
	}
	return nil
}

func (iv *imageViewer) Close() {
	if iv.state == 0 {
		iv.state = 1
	}
}

var kImageView = vk.NewKeys(10)

var layerView *vimgui.View

func (iv *imageViewer) Stat(fr *vimgui.UIFrame) {
	if iv.err != nil {
		fr.NewLine(-100, 20, 5)
		fr.WithTags("error")
		vimgui.Label(fr, "Failed to load file: "+iv.err.Error())
		return
	}
	DrawFileInfo(fr, iv.path, iv.info)
	fr.NewLine(120, 20, 5)
	vimgui.Label(fr, "Image kind: "+iv.kind)
	fr.NewColumn(120, 5)
	vimgui.Label(fr, fmt.Sprintf("Image size: %d", iv.info.Size()))
	fr.NewLine(120, 30, 5)
	if vimgui.IconButton(fr, kImageView+1, materialicons.GetRunes("home")[0], "Home") {
		iv.toHome()
	}
	fr.NewColumn(140, 10)
	if vimgui.IconButton(fr, kImageView+2, materialicons.GetRunes("arrow_left")[0], "Rotate left") {
		iv.rot = (iv.rot + 1) % 4
	}
	fr.NewColumn(140, 10)
	if vimgui.IconButton(fr, kImageView+3, materialicons.GetRunes("arrow_right")[0], "Rotate right") {
		iv.rot--
		if iv.rot < 0 {
			iv.rot = 3
		}
	}
	if iv.desc.Layers > 1 {
		fr.NewColumn(100, 10)
		vimgui.Label(fr, fmt.Sprintf("Layers %d/%d", iv.layer, iv.desc.Layers))
		fr.NewColumn(60, 5)
		f := float64(iv.layer)
		if vimgui.Number(fr, kImageView+3, 0, &f) {
			if f >= 0 {
				fi := uint32(f)
				if fi < iv.desc.Layers {
					iv.layer = fi
				}
			}
		}
	}
	if iv.desc.MipLevels > 1 {
		fr.NewColumn(150, 5)
		mip := int(iv.mip)
		if vimgui.Increment(fr, kImageView+4, 0, int(iv.desc.MipLevels), iv.mipName, &mip) {
			iv.mip = uint32(mip)
		}
	}

}

func newImageViewer(path string, info os.FileInfo, content []byte) (isImage bool) {
	kind := ""
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		kind = "png"
	case ".jpeg":
		fallthrough
	case ".jpg":
		kind = "jpg"
	case ".dds":
		kind = "dds"
	case ".hdr":
		kind = "hdr"
	}
	if len(kind) == 0 {
		return false
	}
	iv := &imageViewer{path: path, kind: kind, info: info, dev: vapp.Dev}
	if imageShaders == nil {
		st := &shaders.Pack{}
		iv.err = st.Load(bytes.NewReader(imageshader_bin))
		if iv.err != nil {
			iv.rm = SetViewer(iv, iv)
			return true
		}
		imageShaders = st
	}

	iv.err = vasset.DescribeImage(kind, &iv.desc, content)
	if iv.err != nil {
		iv.rm = SetViewer(iv, iv)
		return true
	}
	iv.mb = vk.NewMemoryPool(iv.dev)
	iv.img = iv.mb.ReserveImage(iv.desc, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	iv.mb.Allocate()
	cp := vmodel.NewCopier(iv.dev)
	cp.CopyToImage(iv.img, kind, content, iv.desc.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	defer cp.Dispose()

	iv.toHome()
	iv.rm = SetViewer(iv, iv)
	return true

}

func (iv *imageViewer) toHome() {
	iv.aspect = float32(iv.desc.Width) / float32(iv.desc.Height)
	iv.rot = 0
	if iv.aspect > 1 {
		iv.offset = mgl32.Vec2{0, (iv.aspect - 1) / (2 * iv.aspect)}
		iv.scale = 1
	} else {
		iv.offset = mgl32.Vec2{(1 - iv.aspect) / 2, 0}
		iv.scale = 1 / iv.aspect
	}
}

func (iv *imageViewer) mipName(val int) string {
	w := iv.desc.Width >> val
	h := iv.desc.Height >> val
	return fmt.Sprintf("Mip %d (%d x %d)", val, w, h)
}

//go:embed imageshader.bin
var imageshader_bin []byte
