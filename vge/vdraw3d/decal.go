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
	from      FrozenID
	to        FrozenID

	views      [4]*vk.ImageView
	sampler    [4]*vk.Sampler
	storage    uint32
	prev       float32
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
		prev := ri.AddDecal(fd.storage)
		ri.UpdateStorage(fd.storage, 0, fd.toDecal[:]...)
		ri.UpdateStorage(fd.storage+1, 0, fd.albedo[:]...)
		ri.UpdateStorage(fd.storage+1, 4, fd.emissive[:]...)
		ri.UpdateStorage(fd.storage+1, 8, fd.textures1[:]...)
		ri.UpdateStorage(fd.storage+1, 12, fd.metalness, fd.roughness, fd.normalAttn)
		ri.UpdateStorage(fd.storage+2, 0, prev, float32(fd.from), float32(fd.to))
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

func DrawDecal(fl *FreezeList, model *vmodel.Model, world mgl32.Mat4, from, to FrozenID, props vmodel.MaterialProperties) FrozenID {
	fd := &FrozenDecal{model: model, from: from, to: to}
	fd.world = world
	fd.toDecal = world.Inv()
	fd.fillProps(props)
	return fl.Add(fd)

}
