package vdraw3d

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"unsafe"
)

type probe struct {
	key  vk.Key
	at   mgl32.Vec3
	desc vk.ImageDescription
	id   FrozenID

	// When recording
	viewIndex  float32
	storage    uint32
	freezeList *FreezeList
	drawView   func(*FreezeList)
	depthImage *vk.AImage
	depthViews []*vk.AImageView
	la         *vk.DescriptorLayout
	laMips     *vk.DescriptorLayout
	laSPH      *vk.DescriptorLayout
}

func (p *probe) Clone() Frozen {
	return &probe{key: p.key, at: p.at, desc: p.desc, id: p.id}
}

const sizeProbeFrame = uint64(unsafe.Sizeof(probeFrame{}))

func (p *probe) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	p.storage = storageOffset
	if p.key == 0 {
		fi.Device().ReportError(errors.New("probe needs valid key"))
		return storageOffset
	}
	fiStatic, recording := fi.GetRing(StaticRing)
	if recording {
		fiStatic.Get(p.key, func() interface{} {
			sm := &probeMapRes{}

			rFull := p.desc.FullRange()
			rRender := rFull
			rFull.ViewType = vk.CubeView
			rRender.LevelCount = 1
			viewSpecs := []vk.ImageRange{rFull, rRender}
			for layer := uint32(0); layer < 6; layer++ {
				for mips := uint32(0); mips < rFull.LevelCount; mips++ {
					ir := rRender
					ir.FirstLayer, ir.FirstMipLevel, ir.LayerCount = layer, mips, 1
					viewSpecs = append(viewSpecs, ir)
				}
			}
			fiStatic.ReserveImage(p.key, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageSampledBit|vk.IMAGEUsageStorageBit, p.desc, viewSpecs...)
			return sm
		})
		p.la = GetShadowFrameLayout(fi.Device())
		fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, sizeProbeFrame)
		fi.ReserveSlice(vk.BUFFERUsageStorageBufferBit, 4*4*9) // 4 * mgl32.vec4
		fi.ReserveDescriptor(p.la)
		dd := p.desc
		dd.MipLevels, dd.Format = 1, vk.FORMATD32Sfloat
		fi.ReserveImage(p.key, vk.IMAGEUsageDepthStencilAttachmentBit, dd, dd.FullRange())
		p.laMips = getProbeMipsLayout(fi.Device())
		p.laSPH = getProbeSPHLayout(fi.Device())
		fi.ReserveDescriptors(p.laMips, int(6*p.desc.MipLevels))
		fi.ReserveDescriptor(p.laSPH)

	}
	return storageOffset + 3
}

func (p *probe) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(UpdateFrame)
	_, ok2 := phase.(RenderColor)
	if ok || ok2 {
		return true
	}
	_, recording := fi.GetRing(StaticRing)
	if recording {
		_, ok2 = phase.(RenderMaps)
		return ok2
	}
	return false
}

func (p *probe) Render(fi *vk.FrameInstance, phase Phase) {
	if p.key == 0 {
		return
	}
	fiStatic, recording := fi.GetRing(StaticRing)
	uf, ok := phase.(UpdateFrame)
	if ok {
		pm := fiStatic.Get(p.key, func() interface{} {
			fi.Device().FatalError(errors.New("Probe structure not initialize"))
			return nil
		}).(*probeMapRes)
		if recording {
			pm.image, pm.views = fiStatic.AllocImage(p.key)
			p.freezeList = &FreezeList{BaseID: ProbeBase}
			p.drawView(p.freezeList)
			p.freezeList.RenderAll(fi, uf)
		} else {
			p.copySph(func(storagePosition uint32, index uint32, values ...float32) {
				uf.UpdateStorage(storagePosition, index, values...)
			}, pm)
		}
		sampler := vmodel.GetDefaultSampler(fi.Device())
		p.viewIndex = uf.AddView(pm.views[0], sampler)
		uf.UpdateStorage(p.storage+2, 4, p.viewIndex)

	}
	rm, ok := phase.(RenderMaps)
	if ok && recording {
		p.renderProbe(fi, fiStatic, rm)
	}
	rc, ok := phase.(RenderColor)
	if ok {
		*rc.Probe = p.storage
	}
}

func (p *probe) renderProbe(fi *vk.FrameInstance, fiStatic *vk.FrameInstance, rm RenderMaps) {
	pm := fiStatic.Get(p.key, func() interface{} {
		fi.Device().FatalError(errors.New("Probe structure not initialize"))
		return nil
	}).(*probeMapRes)
	p.depthImage, p.depthViews = fi.AllocImage(p.key)
	rp := getProbeRenderPass(fi.Device())
	fr := probeFrame{}
	fr.projection = mgl32.Perspective(math.Pi/2, 1, 0.01, 1000)
	pos := p.at
	fr.views[0] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{1, 0, 0}), mgl32.Vec3{0, -1, 0})
	fr.views[1] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{-1, 0, 0}), mgl32.Vec3{0, -1, 0})
	fr.views[2] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 1, 0}), mgl32.Vec3{0, 0, 1})
	fr.views[3] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, -1, 0}), mgl32.Vec3{0, 0, -1})
	fr.views[4] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, 1}), mgl32.Vec3{0, -1, 0})
	fr.views[5] = mgl32.LookAtV(pos, pos.Add(mgl32.Vec3{0, 0, -1}), mgl32.Vec3{0, -1, 0})
	fr.cameraPos = p.at.Vec4(1)
	cmd := fi.AllocCommand(vk.QUEUEGraphicsBit)
	cmd.Begin()
	ubBuf := fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, sizeProbeFrame)
	*(*probeFrame)(unsafe.Pointer(&ubBuf.Bytes()[0])) = fr
	ds := fi.AllocDescriptor(p.la)
	ds.WriteSlice(0, 0, ubBuf)
	fb := vk.NewFramebuffer2(rp, pm.views[1], p.depthViews[0])
	cmd.BeginRenderPass(rp, fb)
	dl := &vk.DrawList{}
	p.freezeList.RenderAll(fi, RenderProbe{Render: Render{Name: "PROBE", DSFrame: rm.DSFrame, Shaders: rm.Shaders},
		DL: dl, DSProbeFrame: ds, Pass: rp})
	cmd.Draw(dl)
	cmd.EndRenderPass()
	rp2 := getProbeMipPipeline(fi.Device(), rm.Shaders)

	tl := vk.TransferList{}
	for mips := uint32(0); mips < p.desc.MipLevels-1; mips++ {
		for layer := uint32(0); layer < 6; layer++ {
			tl.Transfer(pm.image, vk.IMAGELayoutUndefined, vk.IMAGELayoutGeneral, layer, mips+1)
		}
	}
	cmd.Transfer(tl)
	for mips := uint32(0); mips < p.desc.MipLevels-1; mips++ {
		sw := p.desc.Width >> mips
		sh := p.desc.Height >> mips
		size := mgl32.Vec2{float32(sw), float32(sh)}
		bSize := vk.Float32ToBytes(size[:])
		for layer := uint32(0); layer < 6; layer++ {
			viewOffset := layer*p.desc.MipLevels + mips + 2
			vFrom := pm.views[viewOffset]
			vTo := pm.views[viewOffset+1]
			dsMips := fi.AllocDescriptor(p.laMips)
			dsMips.WriteView(0, 0, vFrom, vk.IMAGELayoutGeneral, nil)
			dsMips.WriteView(0, 1, vTo, vk.IMAGELayoutGeneral, nil)
			cmd.ComputeWith(rp2, sw/32+1, sh/32+1, 1, bSize, dsMips)
		}
	}
	tl = vk.TransferList{}
	for mips := uint32(0); mips < p.desc.MipLevels; mips++ {
		for layer := uint32(0); layer < 6; layer++ {
			tl.Transfer(pm.image, vk.IMAGELayoutGeneral, vk.IMAGELayoutShaderReadOnlyOptimal, layer, mips)
		}
	}
	cmd.Transfer(tl)
	sbSPH := fi.AllocSlice(vk.BUFFERUsageStorageBufferBit, 4*4*9)
	dsSPH := fi.AllocDescriptor(p.laSPH)
	dsSPH.WriteSlice(0, 0, sbSPH)
	rp3 := getProbeSPHPipeline(fi.Device(), rm.Shaders)
	mips := p.desc.MipLevels - 2
	sw := p.desc.Width >> mips
	sh := p.desc.Height >> mips
	params := []float32{float32(sw), float32(sh), p.viewIndex, float32(mips), float32(p.storage)}
	bParams := vk.Float32ToBytes(params)
	cmd.ComputeWith(rp3, 1, 1, 1, bParams, rm.DSFrame, dsSPH)
	cmd.Submit()
	cmd.Wait()
	sphFactors := unsafe.Slice((*mgl32.Vec4)(unsafe.Pointer(&sbSPH.Bytes()[0])), 9)
	copy(pm.sphFactors[:], sphFactors)
	p.copySph(rm.UpdateStorage, pm)
}

func (p *probe) copySph(storage func(storagePosition uint32, index uint32, values ...float32), pm *probeMapRes) {
	for idx := uint32(0); idx < 9; idx++ {
		storage(p.storage+(idx/4), (idx%4)*4, pm.sphFactors[idx][:]...)
	}
}

var kMipCompute = vk.NewKey()

func getProbeMipPipeline(dev *vk.Device, st *shaders.Pack) *vk.ComputePipeline {
	return dev.Get(kMipCompute, func() interface{} {
		cl := vk.NewComputePipeline(dev)
		code := st.MustGet(dev, "probe_mips")
		cl.AddLayout(getProbeMipsLayout(dev))
		cl.AddShader(code.Compute)
		cl.AddPushConstants(vk.SHADERStageComputeBit, 8)
		cl.Create()
		return cl
	}).(*vk.ComputePipeline)

}

var kSphCompute = vk.NewKey()

func getProbeSPHPipeline(dev *vk.Device, st *shaders.Pack) *vk.ComputePipeline {
	return dev.Get(kSphCompute, func() interface{} {
		cl := vk.NewComputePipeline(dev)
		code := st.MustGet(dev, "probe_spherical")
		cl.AddLayout(GetFrameLayout(dev))
		cl.AddLayout(getProbeSPHLayout(dev))
		cl.AddShader(code.Compute)
		cl.AddPushConstants(vk.SHADERStageComputeBit, 4*5)
		cl.Create()
		return cl
	}).(*vk.ComputePipeline)

}

func DrawProbe(fl *FreezeList, key vk.Key, at mgl32.Vec3, drawView func(fl *FreezeList)) FrozenID {
	pr := &probe{key: key, at: at}
	pr.desc = vk.ImageDescription{Layers: 6, Width: 512, Height: 512, Depth: 1, MipLevels: 6, Format: vk.FORMATB10g11r11UfloatPack32}
	pr.drawView = drawView
	pr.id = fl.Add(pr)
	return pr.id
}

type probeMapRes struct {
	la         *vk.DescriptorLayout
	image      *vk.AImage
	views      []*vk.AImageView
	sphFactors [9]mgl32.Vec4
}

var kMipsLayouts = vk.NewKeys(3)

func getProbeMipsLayout(dev *vk.Device) *vk.DescriptorLayout {
	la1 := dev.Get(kMipsLayouts, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	la2 := dev.Get(kMipsLayouts+1, func() interface{} {
		return la1.AddBinding(vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return la2
}

func getProbeSPHLayout(dev *vk.Device) *vk.DescriptorLayout {
	la1 := dev.Get(kMipsLayouts+2, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	return la1
}
