package deferred

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

const (
	MAX_PROBES = 16
	MAX_LIGHTS = 64
)

// DrawFrame are frame settings sent to first Phase (draw geomety)
type DrawFrame struct {
	Projection mgl32.Mat4
	View       mgl32.Mat4
	EyePos     mgl32.Vec4
}

type Probe struct {
	SPH      [9]mgl32.Vec4
	EnvImage float32
	Filler1  float32
	Filler2  float32
	Filler3  float32
}

type LightsFrame struct {
	NoProbes      float32
	NoLight       float32
	Debug         float32
	Index         float32
	InvProjection mgl32.Mat4
	InvView       mgl32.Mat4
	View          mgl32.Mat4
	EyePos        mgl32.Vec4
	Probes        [MAX_PROBES]Probe
	Lights        [MAX_LIGHTS]vscene.Light
}

type DeferredFrame struct {
	DrawPhase     DrawFrame
	LightsFrame   LightsFrame
	dsLight       *vk.DescriptorSet
	dsDraw        *vk.DescriptorSet
	bfLightsFrame *vk.Buffer
	cache         *vk.RenderCache
	imagesUsed    uint32
	probesUsed    int
	lightUsed     int
	sf            *vscene.SimpleFrame
	renderer      *Renderer
}

func (d *DeferredFrame) GetRenderer() vmodel.Renderer {
	return d.renderer
}

func (d *DeferredFrame) GetCache() *vk.RenderCache {
	return d.cache
}

func (d *DeferredFrame) BindFrame() *vk.DescriptorSet {
	if d.dsDraw == nil {
		d.writeDrawFrame()
	}
	return d.dsDraw
}

func (d *DeferredFrame) AddFrameImage(view *vk.ImageView, sampler *vk.Sampler) (imageIndex vmodel.ImageIndex) {
	d.imagesUsed++
	d.dsLight.WriteImage(d.cache.Ctx, 1, d.imagesUsed, view, sampler)
	return vmodel.ImageIndex(d.imagesUsed)
}

var _ vscene.ImageFrame = &DeferredFrame{}

func (d *DeferredFrame) AddEnvironment(SPH [9]mgl32.Vec4, ubfImage vmodel.ImageIndex, pi *vscene.ProcessInfo) {
	if d.probesUsed >= MAX_PROBES {
		return
	}
	d.LightsFrame.Probes[d.probesUsed] = Probe{SPH: SPH, EnvImage: float32(ubfImage)}
	d.probesUsed++
	return
}

func (f *DeferredFrame) GetSimpleFrame() *vscene.SimpleFrame {
	if f.sf == nil {
		f.sf = &vscene.SimpleFrame{SSF: vscene.SimpleShaderFrame{Projection: f.DrawPhase.Projection, View: f.DrawPhase.View}, Cache: f.cache}
	}
	return f.sf
}

var kBoundDrawFrame = vk.NewKey()

func (f *DeferredFrame) writeDrawFrame() *vk.DescriptorSet {
	cache := vscene.GetUniformCache(f.cache)
	ds, sl := cache.Alloc(f.cache.Ctx)
	b := *(*[unsafe.Sizeof(DrawFrame{})]byte)(unsafe.Pointer(&f.DrawPhase))
	copy(sl.Content, b[:])
	f.dsDraw = ds
	return f.dsDraw
}

func (d *DeferredFrame) writeLightsFrame() {
	d.LightsFrame.NoProbes = float32(d.probesUsed)
	d.LightsFrame.NoLight = float32(d.lightUsed)
	d.LightsFrame.InvProjection = d.DrawPhase.Projection.Inv()
	d.LightsFrame.InvView = d.DrawPhase.View.Inv()
	d.LightsFrame.View = d.DrawPhase.View
	d.LightsFrame.EyePos = d.DrawPhase.EyePos
	b := *(*[unsafe.Sizeof(LightsFrame{})]byte)(unsafe.Pointer(&d.LightsFrame))
	copy(d.bfLightsFrame.Bytes(d.cache.Ctx), b[:])
	d.dsLight.WriteBuffer(d.cache.Ctx, 0, 0, d.bfLightsFrame)
}
