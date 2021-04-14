//

package noise

import (
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type Fire struct {
	Size mgl32.Vec2
	Heat float32
}

func NewFire(sizeX, sizeY float32) *Fire {
	return &Fire{Size: mgl32.Vec2{sizeX, sizeY}, Heat: 1200}
}

func (f *Fire) Process(pi *vscene.ProcessInfo) {
	pd, ok := pi.Phase.(vscene.DrawPhase)
	if ok {
		dc := pd.GetDC(vscene.LAYERTransparent)
		if dc != nil {
			w := pi.World.Mul4(mgl32.Scale3D(f.Size[0], f.Size[1], f.Size[0]))
			f.draw(dc, w, pi.Time)
		}
	}
}

const maxInstances = 200

func (f *Fire) draw(dc *vmodel.DrawContext, w mgl32.Mat4, time float64) {
	sfc := vscene.GetSimpleFrame(dc.Frame)
	if sfc == nil {
		return // Not supported
	}
	rc := dc.Frame.GetCache()
	gp := dc.Pass.Get(rc.Ctx, kFirePipeline, func(ctx vk.APIContext) interface{} {
		return f.newPipeline(ctx, dc)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(rc)
	dsFrame := sfc.BindFrame()
	fis := rc.GetPerFrame(kFireInstance, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		return &fireInstances{ds: ds, sl: sl}
	}).(*fireInstances)
	dsFire := f.getFireTexture(rc.Ctx, rc.Device)
	fi := fireInstance{world: w, heat: f.Heat, offset: frag(float32(time) / 2)}
	lInst := uint32(unsafe.Sizeof(fireInstance{}))
	b := *(*[unsafe.Sizeof(fireInstance{})]byte)(unsafe.Pointer(&fi))
	copy(fis.sl.Content[fis.count*lInst:(fis.count+1)*lInst], b[:])
	dc.Draw(gp, 0, 6).AddDescriptors(dsFrame, fis.ds, dsFire).SetInstances(fis.count, 1)
	fis.count++
	if fis.count >= maxInstances {
		rc.SetPerFrame(kFireInstance, nil)
	}
}

func frag(f float32) float32 {
	return f - 2*float32(int(f/2))
}

type fireInstances struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	count uint32
}

type fireInstance struct {
	world  mgl32.Mat4
	offset float32
	heat   float32
	dummy1 float32
	dummy2 float32
}

type fireTexture struct {
	pool    *vk.MemoryPool
	dsPool  *vk.DescriptorPool
	ds      *vk.DescriptorSet
	sampler *vk.Sampler
}

func (f *fireTexture) Dispose() {
	if f.sampler != nil {
		f.sampler.Dispose()
		f.sampler = nil
	}
	if f.dsPool != nil {
		f.dsPool.Dispose()
		f.ds, f.dsPool = nil, nil
	}
	if f.pool != nil {
		f.pool.Dispose()
		f.pool = nil
	}
}

func (f *Fire) getFireTexture(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorSet {
	ft := dev.Get(ctx, kFireTexture, func(ctx vk.APIContext) interface{} {
		return f.buildFireTexture(ctx, dev)
	}).(*fireTexture)
	return ft.ds
}

func (f *Fire) newPipeline(ctx vk.APIContext, dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	rc := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	la := vscene.GetUniformLayout(ctx, rc.Device)
	laFrame := vscene.GetUniformLayout(ctx, rc.Device)
	laFire := getFireLayout(ctx, rc.Device)
	gp.AddLayout(ctx, laFrame)
	gp.AddLayout(ctx, la)
	gp.AddLayout(ctx, laFire)
	// gp.AddLayout(ctx, la2)
	gp.AddShader(ctx, vk.SHADERStageVertexBit, fire_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, fire_frag_spv)
	gp.AddAlphaBlend(ctx)
	gp.AddDepth(ctx, false, true)
	gp.Create(ctx, dc.Pass)
	return gp
}

func (f *Fire) buildFireTexture(ctx vk.APIContext, dev *vk.Device) *fireTexture {
	ft := &fireTexture{}
	ft.pool = vk.NewMemoryPool(dev)
	pn := NewPerlinNoise(256)
	pn.Add(1, 35.7)
	pn.Add(0.4, 15.7)
	desc := vk.ImageDescription{Width: 256, Depth: 1, Height: 256, MipLevels: 1, Layers: 1, Format: vk.FORMATR8Unorm}
	img := ft.pool.ReserveImage(ctx, desc, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	ft.pool.Allocate(ctx)
	cp := vmodel.NewCopier(ctx, dev)
	defer cp.Dispose()
	cp.CopyToImage(img, "raw", pn.ToBytes(), img.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	la := getFireLayout(ctx, dev)
	ft.dsPool = vk.NewDescriptorPool(ctx, la, 1)
	ft.ds = ft.dsPool.Alloc(ctx)
	ft.sampler = vk.NewSampler(ctx, dev, vk.SAMPLERAddressModeMirroredRepeat)
	ft.ds.WriteImage(ctx, 0, 0, img.DefaultView(ctx), ft.sampler)
	return ft
}

func getFireLayout(ctx vk.APIContext, device *vk.Device) *vk.DescriptorLayout {
	return device.Get(ctx, kFireLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, device, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 1)
	}).(*vk.DescriptorLayout)
}

var kFirePipeline = vk.NewKey()
var kFireInstance = vk.NewKey()
var kFireLayout = vk.NewKey()
var kFireTexture = vk.NewKey()
