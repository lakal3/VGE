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
	pl := dc.Pass.Get(dc.Cache.Ctx, kGrayPipeline, func(ctx vk.APIContext) interface{} {
		return g.newPipeline(dc)
	}).(vk.Pipeline)
	dsFrame := vscene.BindFrame(dc.Cache)
	uc := vscene.GetUniformCache(dc.Cache)
	dsColor, slCol := uc.Alloc(dc.Cache.Ctx)
	b := *(*[unsafe.Sizeof(GrayBg{})]byte)(unsafe.Pointer(g))
	copy(slCol.Content, b[:])
	cb := getCube(dc.Cache.Ctx, dc.Cache.Device)
	dc.Draw(pl, 0, 36).AddInputs(cb.bVtx).AddDescriptors(dsFrame, dsColor)
}

func (g *GrayBg) newPipeline(dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	ctx := dc.Cache.Ctx
	gp := vk.NewGraphicsPipeline(ctx, dc.Cache.Device)
	gp.AddVextexInput(ctx, vk.VERTEXInputRateVertex, vk.FORMATR32g32b32Sfloat)
	la := vscene.GetUniformLayout(ctx, dc.Cache.Device) // Dynamic layout for colors
	laFrame := vscene.GetFrameLayout(ctx, dc.Cache.Device)
	gp.AddLayout(ctx, laFrame)
	gp.AddLayout(ctx, la)
	gp.AddShader(ctx, vk.SHADERStageVertexBit, eqrect_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, graybg_frag_spv)
	gp.Create(ctx, dc.Pass)
	return gp
}

var kGrayPipeline = vk.NewKey()
