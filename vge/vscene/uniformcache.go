package vscene

import (
	"github.com/lakal3/vge/vge/vk"
)

// var UniformSize uint32 = 16384

type UniformCache struct {
	size     uint32
	dsLayout *vk.DescriptorLayout
	dsPool   *vk.DescriptorPool
	dsSets   []*vk.DescriptorSet
	mp       *vk.MemoryPool
	pos      int
	cache    *vk.RenderCache
	slices   []*vk.Slice
}

var ucLayoutKey = vk.NewKey()
var ucKey = vk.NewKey()
var ucSmallKey = vk.NewKey()

func (uc *UniformCache) Dispose() {
	if uc.dsPool != nil {
		uc.dsPool.Dispose()
		uc.mp.Dispose()
		uc.dsPool, uc.mp = nil, nil
	}
}

// NewUniformCache allocates cache of uniform descriptors. If cached entries runs out, UniformCache will automatically extends it's size
// size is maximum size of one uniform buffer and minEntries is initial number of uniforms
func NewUniformCache(cache *vk.RenderCache, size uint32, minEntries int) *UniformCache {
	uc := &UniformCache{cache: cache, size: size, dsLayout: GetUniformLayout(cache.Device)}
	uc.realloc(minEntries)
	return uc
}

// GetUniformCache retrieves standard uniform cache with size of 63356 bytes per entry
// 65536 is maximum limit of most NVidias GPU:s.
func GetUniformCache(cache *vk.RenderCache) *UniformCache {
	ddc := cache.Get(ucKey, func() interface{} {
		return NewUniformCache(cache, 65536, 10)
	}).(*UniformCache)
	cache.GetPerFrame(ucKey, func() interface{} {
		ddc.pos = 0
		return struct{}{}
	})
	return ddc
}

// GetSmallUniformCache retrieves standard uniform cache with size of 4096 bytes per entry.
// Use small uniform if 4k if more that you need for uniform
func GetSmallUniformCache(cache *vk.RenderCache) *UniformCache {
	ddc := cache.Get(ucSmallKey, func() interface{} {
		return NewUniformCache(cache, 4096, 10)
	}).(*UniformCache)
	cache.GetPerFrame(ucSmallKey, func() interface{} {
		ddc.pos = 0
		return struct{}{}
	})
	return ddc
}
func GetUniformLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(ucLayoutKey, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
}

func (uc *UniformCache) Size() uint32 {
	return uc.size
}

func (uc *UniformCache) realloc(newSize int) {
	if uc.dsPool != nil {
		uc.cache.DisposePerFrame(uc.dsPool)
		uc.cache.DisposePerFrame(uc.mp)
	}
	uc.dsPool = vk.NewDescriptorPool(uc.dsLayout, newSize)
	uc.dsSets = make([]*vk.DescriptorSet, newSize)
	uc.slices = make([]*vk.Slice, newSize)
	uc.mp = vk.NewMemoryPool(uc.cache.Device)
	buffer := uc.mp.ReserveBuffer(uint64(uc.size)*uint64(newSize), true, vk.BUFFERUsageUniformBufferBit)
	uc.mp.Allocate()
	for idx := 0; idx < newSize; idx++ {
		uc.slices[idx] = buffer.Slice(uint64(uc.size)*uint64(idx), uint64(uc.size)*uint64(idx+1))
		uc.dsSets[idx] = uc.dsPool.Alloc()
		uc.dsSets[idx].WriteSlice(0, 0, uc.slices[idx])
	}
}

func (ddc *UniformCache) Bind(dl *vk.DrawItem, set int, content []byte) *vk.DrawItem {
	ds, sl := ddc.Alloc()
	copy(sl.Content, content)
	dl.AddDescriptor(set, ds)
	return dl
}

func (ddc *UniformCache) Alloc() (ds *vk.DescriptorSet, sl *vk.Slice) {
	if ddc.pos >= len(ddc.dsSets) {
		ddc.pos = 0
		ddc.realloc(len(ddc.dsSets) * 2)
	}
	oldPos := ddc.pos
	ddc.pos++
	return ddc.dsSets[oldPos], ddc.slices[oldPos]
}
