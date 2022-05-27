package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"unsafe"
)

type fLight struct {
	storage     uint32
	kind        uint32
	plane       mgl32.Vec4
	intensity   mgl32.Vec3
	direction   mgl32.Vec4
	position    mgl32.Vec4
	attenuation mgl32.Vec3
	outerAngle  float32
	innerAngle  float32

	key       vk.Key
	smDesc    vk.ImageDescription
	id        FrozenID
	maxShadow float32
	minShadow float32

	image *vk.AImage
	views []*vk.AImageView
	la    *vk.DescriptorLayout
}

func (f *fLight) Clone() Frozen {
	fn := *f
	fn.image, fn.views = nil, nil
	return &fn
}

func (f *fLight) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	f.storage = storageOffset
	if f.key != 0 {
		if f.la == nil {
			f.la = GetShadowFrameLayout(fi.Device())
		}
		ranges := []vk.ImageRange{f.smDesc.FullRange()}
		for idx := uint32(0); idx < f.smDesc.Layers; idx++ {
			r := f.smDesc.FullRange()
			r.LayerCount = 1
			r.FirstLayer = idx
			ranges = append(ranges, r)
			fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, 256)
			fi.ReserveDescriptor(f.la)
		}
		fi.ReserveImage(f.key, vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageSampledBit,
			f.smDesc, ranges...)
	}
	return storageOffset + 2
}

func (f *fLight) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(UpdateFrame)
	_, ok2 := phase.(RenderMaps)
	return ok || (ok2 && f.key != 0)
}

func (f *fLight) Render(fi *vk.FrameInstance, phase Phase) {
	uf, ok := phase.(UpdateFrame)
	if ok {
		if f.key != 0 {
			f.image, f.views = fi.AllocImage(f.key)
			sIdx := uf.AddView(f.views[0], vmodel.GetDefaultSampler(fi.Device()))
			uf.UpdateStorage(f.storage+1, 2, sIdx)
			uf.UpdateStorage(f.storage+1, 8, f.plane[:]...)
		} else {
			uf.UpdateStorage(f.storage+1, 2, 0)
		}
		uf.UpdateStorage(f.storage, 0, f.intensity[:]...)
		uf.UpdateStorage(f.storage, 4, f.position[:]...)
		uf.UpdateStorage(f.storage, 8, f.direction[:]...)
		uf.UpdateStorage(f.storage, 12, f.attenuation[:]...)
		nl := uf.AddLight(f.storage)
		uf.UpdateStorage(f.storage+1, 0, nl, float32(f.kind))
		uf.UpdateStorage(f.storage+1, 4, f.innerAngle, f.outerAngle, f.minShadow, f.maxShadow)
	}
	if f.key == 0 {
		return
	}
	rm, ok := phase.(RenderMaps)
	if ok {
		f.renderShadowMap(fi, rm)
	}
}

func (f *fLight) addProps(props vmodel.MaterialProperties) {
	dz := f.direction.Vec3().Dot(zDir)
	if dz < 1 && f.direction.LenSqr() > 0 {
		angle := float32(math.Acos(float64(dz)))
		dir := f.direction.Vec3().Cross(zDir).Normalize()
		q := mgl32.QuatRotate(-angle, dir)
		f.plane = mgl32.Vec4{q.V[0], q.V[1], q.V[2], q.W}
	} else {
		f.plane = mgl32.Vec4{0, 0, 0, 1}
	}
	if props == nil {
		return
	}
	f.intensity = props.GetColor(vmodel.CIntensity, f.intensity.Vec4(0)).Vec3()
	f.attenuation[0] = props.GetFactor(vmodel.FLightAttenuation0, f.attenuation[0])
	f.attenuation[1] = props.GetFactor(vmodel.FLightAttenuation1, f.attenuation[1])
	f.attenuation[2] = props.GetFactor(vmodel.FLightAttenuation2, f.attenuation[2])
	if f.kind == 2 {
		f.outerAngle = props.GetFactor(vmodel.FOuterAngle, 0.7)
		f.innerAngle = props.GetFactor(vmodel.FInnerAngle, 0.5)
	}
	if f.key != 0 {
		f.smDesc = vk.ImageDescription{Width: props.GetUInt(vmodel.UShadowMapSize, 512),
			Depth: 1, Format: vk.FORMATD32Sfloat, Layers: 1, MipLevels: 1}
		if f.kind == 1 {
			f.smDesc.Layers = 2
		}
		if f.kind == 0 {
			f.smDesc.Layers = 3
		}
		f.smDesc.Height = f.smDesc.Width
		f.minShadow, f.maxShadow = 0.1, 10
	}
}

var zDir = mgl32.Vec3{0, 0, 1}

func (f *fLight) renderShadowMap(fi *vk.FrameInstance, rm RenderMaps) {

	rp := getDepthRenderPass(fi.Device())
	fr := shadowFrame{lightPos: f.position, minShadow: f.minShadow, maxShadow: f.maxShadow,
		plane: f.plane}

	cmd := fi.AllocCommand(vk.QUEUEGraphicsBit)
	cmd.Begin()
	for idx := uint32(0); idx < f.smDesc.Layers; idx++ {
		ubBuf := fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, 256)
		switch f.kind {
		case 1:
			fr.yFactor = -1
			if idx > 0 {
				fr.yFactor = 1
			}
		}

		*(*shadowFrame)(unsafe.Pointer(&ubBuf.Bytes()[0])) = fr
		ds := fi.AllocDescriptor(f.la)
		ds.WriteSlice(0, 0, ubBuf)
		fb := vk.NewFramebuffer2(rp, f.views[idx+1])
		fi.AddChild(fb)
		cmd.BeginRenderPass(rp, fb)
		dl := &vk.DrawList{}
		rs := RenderShadow{Render: Render{Name: "SHADOW", DSFrame: rm.DSFrame, Shaders: rm.Shaders},
			DL: dl, DSShadowFrame: ds, Pass: rp, Directional: f.kind == 0}
		rm.Static.RenderAll(fi, rs)
		rm.Dynamic.RenderAll(fi, rs)
		cmd.Draw(dl)
		cmd.EndRenderPass()
		if f.kind == 0 {
			fr.maxShadow *= 4
		}
	}
	si := cmd.SubmitForWait(0, vk.PIPELINEStageEarlyFragmentTestsBit)
	rm.AtEnd(func() []vk.SubmitInfo {
		return []vk.SubmitInfo{si}
	})
}

func DrawDirectionalLight(fl *FreezeList, shadowKey vk.Key, direction mgl32.Vec3, props vmodel.MaterialProperties) FrozenID {
	l := &fLight{intensity: mgl32.Vec3{1, 1, 1}, direction: direction.Normalize().Vec4(0),
		attenuation: mgl32.Vec3{1, 0, 0}, key: shadowKey}
	l.addProps(props)
	l.id = fl.Add(l)
	return l.id
}

func DrawPointLight(fl *FreezeList, shadowKey vk.Key, at mgl32.Vec3, props vmodel.MaterialProperties) FrozenID {
	l := &fLight{intensity: mgl32.Vec3{1, 1, 1}, position: at.Vec4(1), attenuation: mgl32.Vec3{0, 0, 1}, kind: 1, key: shadowKey}
	l.addProps(props)
	l.id = fl.Add(l)
	return l.id
}
