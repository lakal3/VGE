package env

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type GrayBg struct {
	// Base color at horizon
	BaseColor mgl32.Vec4
	// Color up (typically most intense color). DownColor is BaseColor - (UpColor - BaseColor)
	UpColor mgl32.Vec4
}

func NewGrayBG() *GrayBg {
	return &GrayBg{BaseColor: mgl32.Vec4{0.5, 0.5, 0.5, 1}, UpColor: mgl32.Vec4{0.65, 0.65, 0.65, 1}}
}

func (g *GrayBg) Process(pi *vscene.ProcessInfo) {
	db, ok := pi.Phase.(vscene.DrawPhase)
	if ok {
		dc := db.GetDC(vscene.LAYERBackground)
		if dc != nil {
			g.Draw(dc)
		}
	}
}

func (g *GrayBg) Draw(dc *vmodel.DrawContext) {
	sfc := vscene.GetSimpleFrame(dc.Frame)
	if sfc == nil {
		return // Not supported
	}
	cache := sfc.GetCache()
	pl := dc.Pass.Get(kGrayPipeline, func() interface{} {
		return g.newPipeline(dc)
	}).(vk.Pipeline)
	dsFrame := sfc.BindFrame()
	uc := vscene.GetUniformCache(cache)
	dsColor, slCol := uc.Alloc()
	b := *(*[unsafe.Sizeof(GrayBg{})]byte)(unsafe.Pointer(g))
	copy(slCol.Content, b[:])
	cb := getCube(cache.Device)
	dc.Draw(pl, 0, 36).AddInputs(cb.bVtx).AddDescriptors(dsFrame, dsColor)
}

func (g *GrayBg) newPipeline(dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	cache := dc.Frame.GetCache()

	gp := vk.NewGraphicsPipeline(cache.Device)
	gp.AddVextexInput(vk.VERTEXInputRateVertex, vk.FORMATR32g32b32Sfloat)
	la := vscene.GetUniformLayout(cache.Device) // Dynamic layout for colors
	laFrame := vscene.GetUniformLayout(cache.Device)
	gp.AddLayout(laFrame)
	gp.AddLayout(la)
	gp.AddShader(vk.SHADERStageVertexBit, eqrect_vert_spv)
	gp.AddShader(vk.SHADERStageFragmentBit, graybg_frag_spv)
	gp.Create(dc.Pass)
	return gp
}

var kGrayPipeline = vk.NewKey()
