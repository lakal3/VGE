package decal

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"unsafe"
)

type DecalInstance struct {
	Model        *vmodel.Model
	Material     vmodel.MaterialIndex
	AlbedoFactor mgl32.Vec4
	Location     mgl32.Mat4

	// Tag is extra information for Instance user
	Tag interface{}
}

type LocalPainter struct {
	Decals []DecalInstance
}

func (l *LocalPainter) AddDecal(model *vmodel.Model, material vmodel.MaterialIndex, at mgl32.Mat4, albedoFactor mgl32.Vec4) int {
	idx := len(l.Decals)
	l.Decals = append(l.Decals, DecalInstance{Model: model, Material: material, AlbedoFactor: albedoFactor, Location: at})
	return idx
}

var kPainter = vk.NewKey()
var kNullDescriptor = vk.NewKey()

func (l *LocalPainter) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(vscene.DrawPhase)
	if ok {
		// ?Layer
		pi.Set(kPainter, pi.Frame.GetCache().GetPerFrame(kPainter, func(ctx vk.APIContext) interface{} {
			return l.buildPhaseInfo(pi.Frame, pi.World)
		}).(phaseInfo))
	}
}

func (l *LocalPainter) buildPhaseInfo(frame vmodel.Frame, world mgl32.Mat4) phaseInfo {
	var decals GPUDecals
	for idx, inst := range l.Decals {
		gInst := inst.BuildGPUInstance(frame, world)
		if idx < MAX_DECALS {
			decals.Instances[idx] = gInst
			decals.NoDecals = float32(idx + 1)
		}
	}
	uc := vscene.GetUniformCache(frame.GetCache())
	ds, sl := uc.Alloc(frame.GetCache().Ctx)
	b := *(*[unsafe.Sizeof(GPUDecals{})]byte)(unsafe.Pointer(&decals))
	copy(sl.Content, b[:])
	return phaseInfo{ds: ds}
}

func (i DecalInstance) BuildGPUInstance(frame vmodel.Frame, world mgl32.Mat4) GPUInstance {
	mat := i.Model.GetMaterial(i.Material)
	return GPUInstance{
		ToDecalSpace: world.Mul4(i.Location).Inv(),
		AlbedoFactor: mat.Props.GetColor(vmodel.Color, mgl32.Vec4{1, 1, 1, 1}),
		TxAlbedo:     i.getImage(frame, mat, vmodel.TxAlbedo), TxNormal: i.getImage(frame, mat, vmodel.TxBump),
		TxMetalRoughness:  i.getImage(frame, mat, vmodel.TxMetallicRoughness),
		MetalnessFactor:   mat.Props.GetFactor(vmodel.FMetalness, 1),
		RoughnessFactor:   mat.Props.GetFactor(vmodel.FRoughness, 1),
		NormalAttenuation: mat.Props.GetFactor(vmodel.FNormalAttenuation, 1),
	}
}

func (i DecalInstance) getImage(frame vmodel.Frame, mat vmodel.Material, tx vmodel.Property) float32 {
	imageIndex := mat.Props.GetImage(tx)
	if imageIndex == 0 {
		return 0
	}
	imf, ok := frame.(vscene.ImageFrame)
	if !ok {
		return 0
	}
	cache := frame.GetCache()
	view := i.Model.GetImage(mat.Props.GetImage(tx)).DefaultView(cache.Ctx)
	sampler := vmodel.GetDefaultSampler(cache.Ctx, cache.Device)
	return float32(imf.AddFrameImage(view, sampler))
}

func BindPainter(rc *vk.RenderCache, extra vmodel.ShaderExtra) *vk.DescriptorSet {
	v := extra.Get(kPainter)
	if v == nil {
		return buildNullDescriptor(rc)
	}
	return v.(phaseInfo).ds
}

func buildNullDescriptor(rc *vk.RenderCache) *vk.DescriptorSet {
	return rc.GetPerFrame(kNullDescriptor, func(ctx vk.APIContext) interface{} {
		var decals GPUDecals
		uc := vscene.GetUniformCache(rc)
		ds, sl := uc.Alloc(ctx)
		b := *(*[unsafe.Sizeof(GPUDecals{})]byte)(unsafe.Pointer(&decals))
		copy(sl.Content, b[:])
		return ds
	}).(*vk.DescriptorSet)
}

type phaseInfo struct {
	ds *vk.DescriptorSet
}
