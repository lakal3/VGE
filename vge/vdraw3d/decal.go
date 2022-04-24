package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type FrozenDecal struct {
	model     *vmodel.Model
	world     mgl32.Mat4
	toDecal   mgl32.Mat4
	albedo    mgl32.Vec4
	emissive  mgl32.Vec4
	textures1 mgl32.Vec4
	from      uint32
	to        uint32

	views      [4]*vk.ImageView
	sampler    [4]*vk.Sampler
	storage    uint32
	prev       uint32
	metalness  float32
	roughness  float32
	normalAttn float32
}

func (fd *FrozenDecal) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	fd.storage = storageOffset
	return storageOffset + 3
}

func (fd *FrozenDecal) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(UpdateFrame)
	if ok {
		return true
	}
	return false
}

func (fd *FrozenDecal) Clone() Frozen {
	fd.prev, fd.storage = 0, 0
	return fd
}

func (fd *FrozenDecal) Render(fi *vk.FrameInstance, phase Phase) {
	ri, ok := phase.(UpdateFrame)
	if ok {
		for idx := 0; idx < 4; idx++ {
			if fd.views[idx] != nil {
				fd.textures1[idx] = ri.AddView(fd.views[idx], fd.sampler[idx])
			}
		}
		fd.prev = ri.AddDecal(fd.storage)
		ri.UpdateStorage(fd.storage, 0, fd.toDecal[:]...)
		ri.UpdateStorage(fd.storage+1, 0, fd.albedo[:]...)
		ri.UpdateStorage(fd.storage+1, 4, fd.emissive[:]...)
		ri.UpdateStorage(fd.storage+1, 8, fd.textures1[:]...)
		ri.UpdateStorage(fd.storage+1, 12, fd.metalness, fd.roughness, fd.normalAttn)
		ri.UpdateStorage(fd.storage+2, 0, float32(fd.prev), float32(fd.from), float32(fd.to))
	}
	rc, ok := phase.(RenderColor)
	if ok && fd.storage != 0 {
		fd.prev = *rc.Decal
		*rc.Decal = fd.storage
	}
}

func (fd *FrozenDecal) fillProps(props vmodel.MaterialProperties) {
	fd.albedo = props.GetColor(vmodel.CAlbedo, mgl32.Vec4{0, 0, 0, 1})
	fd.emissive = props.GetColor(vmodel.CEmissive, mgl32.Vec4{0, 0, 0, 0})
	fd.metalness = props.GetFactor(vmodel.FMetalness, 0)
	fd.roughness = props.GetFactor(vmodel.FRoughness, 1)
	fd.normalAttn = props.GetFactor(vmodel.FNormalAttenuation, 1)
	txIdx := props.GetImage(vmodel.TxAlbedo)
	if txIdx != 0 {
		fd.textures1[0] = float32(txIdx)
		fd.views[0], fd.sampler[0] = fd.model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxEmissive)
	if txIdx != 0 {
		fd.textures1[1] = float32(txIdx)
		fd.views[1], fd.sampler[1] = fd.model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxMetallicRoughness)
	if txIdx != 0 {
		fd.textures1[2] = float32(txIdx)
		fd.views[2], fd.sampler[2] = fd.model.GetImageView(txIdx)
	}
	txIdx = props.GetImage(vmodel.TxBump)
	if txIdx != 0 {
		fd.textures1[3] = float32(txIdx)
		fd.views[3], fd.sampler[3] = fd.model.GetImageView(txIdx)
	}

}

type popDecal struct {
	fd *FrozenDecal
}

func (p popDecal) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	return storageOffset
}

func (p popDecal) Support(fi *vk.FrameInstance, phase Phase) bool {
	_, ok := phase.(RenderColor)
	return ok
}

func (p popDecal) Render(fi *vk.FrameInstance, phase Phase) {
	ri, ok := phase.(UpdateFrame)
	if ok {
		ri.PopDecal(p.fd.prev)
		return
	}
	rc, ok := phase.(RenderColor)
	if ok {
		*rc.Decal = p.fd.prev
		p.fd.storage = 0
	}
}

func (p popDecal) Clone() Frozen {
	return p
}

func (fd *FrozenDecal) pop(fl *FreezeList) {
	fl.Add(popDecal{fd: fd})
}

func DrawDecal(fl *FreezeList, model *vmodel.Model, world mgl32.Mat4, props vmodel.MaterialProperties) (pop func(fl *FreezeList), id FrozenID) {
	return DrawDecalOn(fl, model, world, 0, 0, props)
}

func DrawDecalOn(fl *FreezeList, model *vmodel.Model, world mgl32.Mat4, fromMaterialID, toMaterialID uint32, props vmodel.MaterialProperties) (pop func(fl *FreezeList), id FrozenID) {
	fd := &FrozenDecal{model: model, from: fromMaterialID, to: toMaterialID}
	fd.world = world
	fd.toDecal = world.Inv()
	fd.fillProps(props)
	return fd.pop, fl.Add(fd)

}
