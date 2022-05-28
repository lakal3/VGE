package highlight

import (
	"bytes"
	_ "embed"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vdraw3d"
	"github.com/lakal3/vge/vge/vk"
	"unsafe"
)

func DrawHighlight(fl *vdraw3d.FreezeList, color mgl32.Vec4, lineWidth uint32) {

	fl.Add(flHighlight{color: color, lineWidth: lineWidth})
}

func AddPack(base *shaders.Pack) error {
	return base.Load(bytes.NewReader(highlight_bin))
}

type flHighlight struct {
	color     mgl32.Vec4
	lineWidth uint32
}

func (f flHighlight) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	return storageOffset
}

func (f flHighlight) Support(fi *vk.FrameInstance, phase vdraw3d.Phase) bool {
	_, ok := phase.(vdraw3d.RenderOverlay)
	return ok
}

func (f flHighlight) Render(fi *vk.FrameInstance, phase vdraw3d.Phase) {
	ro, ok := phase.(vdraw3d.RenderOverlay)
	if ok {
		pl := f.getPlPipeline(fi, ro.Pass, ro.Shaders)
		ptr, off := ro.DL.AllocPushConstants(uint32(unsafe.Sizeof(hlInstance{})))
		*(*hlInstance)(ptr) = hlInstance{color: f.color, lineWidth: f.lineWidth}
		ro.DL.Draw(pl, 0, 3).AddDescriptor(0, ro.DSFrame).
			AddPushConstants(uint32(unsafe.Sizeof(hlInstance{})), off)
	}
}

func (f flHighlight) Clone() vdraw3d.Frozen {
	return f
}

var kPipeline = vk.NewKey()

func (f flHighlight) getPlPipeline(fi *vk.FrameInstance, pass *vk.GeneralRenderPass, pack *shaders.Pack) *vk.GraphicsPipeline {
	return pass.Get(kPipeline, func() interface{} {
		pl := vk.NewGraphicsPipeline(fi.Device())
		sp := pack.MustGet(fi.Device(), "highlight")
		pl.AddShader(vk.SHADERStageVertexBit, sp.Vertex)
		pl.AddShader(vk.SHADERStageFragmentBit, sp.Fragment)
		pl.AddLayout(vdraw3d.GetFrameLayout(fi.Device()))
		pl.AddAlphaBlend()
		pl.AddPushConstants(vk.SHADERStageFragmentBit, uint32(unsafe.Sizeof(hlInstance{})))
		pl.Create(pass)
		return pl
	}).(*vk.GraphicsPipeline)
}

type hlInstance struct {
	color     mgl32.Vec4
	lineWidth uint32
	filler1   uint32
	filler2   uint32
	filler3   uint32
}

//go:embed highlight.bin
var highlight_bin []byte
