package vscene

import (
	"github.com/lakal3/vge/vge/vk"
)

var UniformSize uint32 = 16384

type UniformCache struct {
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

func (uc *UniformCache) Dispose() {
	if uc.dsPool != nil {
		uc.dsPool.Dispose()
		uc.mp.Dispose()
		uc.dsPool, uc.mp = nil, nil
	}
}

func GetUniformCache(cache *vk.RenderCache) *UniformCache {
	ddc := cache.Get(ucKey, func(ctx vk.APIContext) interface{} {
		ddc := &UniformCache{cache: cache}
		ddc.dsLayout = GetUniformLayout(ctx, cache.Device)
		ddc.realloc(ctx, 10)
		return ddc
	}).(*UniformCache)
	cache.GetPerFrame(ucKey, func(ctx vk.APIContext) interface{} {
		ddc.pos = 0
		return ddc
	})
	return ddc
}

func GetUniformLayout(ctx vk.APIContext, dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(ctx, ucLayoutKey, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageAll, 1)
	}).(*vk.DescriptorLayout)
}

func (uc *UniformCache) realloc(ctx vk.APIContext, newSize int) {
	if uc.dsPool != nil {
		uc.cache.DisposePerFrame(uc.dsPool)
		uc.cache.DisposePerFrame(uc.mp)
	}
	uc.dsPool = vk.NewDescriptorPool(ctx, uc.dsLayout, newSize)
	uc.dsSets = make([]*vk.DescriptorSet, newSize)
	uc.slices = make([]*vk.Slice, newSize)
	uc.mp = vk.NewMemoryPool(uc.cache.Device)
	buffer := uc.mp.ReserveBuffer(ctx, uint64(UniformSize)*uint64(newSize), true, vk.BUFFERUsageUniformBufferBit)
	uc.mp.Allocate(ctx)
	for idx := 0; idx < newSize; idx++ {
		uc.slices[idx] = buffer.Slice(ctx,
			uint64(UniformSize)*uint64(idx), uint64(UniformSize)*uint64(idx+1))
		uc.dsSets[idx] = uc.dsPool.Alloc(ctx)
		uc.dsSets[idx].WriteSlice(ctx, 0, 0, uc.slices[idx])
	}
}

func (ddc *UniformCache) Bind(ctx vk.APIContext, dl *vk.DrawItem, set int, content []byte) *vk.DrawItem {
	ds, sl := ddc.Alloc(ctx)
	if !ctx.IsValid() {
		return dl
	}
	copy(sl.Content, content)
	dl.AddDescriptor(set, ds)
	return dl
}

func (ddc *UniformCache) Alloc(ctx vk.APIContext) (ds *vk.DescriptorSet, sl *vk.Slice) {
	if ddc.pos >= len(ddc.dsSets) {
		ddc.pos = 0
		ddc.realloc(ctx, len(ddc.dsSets)*2)
	}
	oldPos := ddc.pos
	ddc.pos++
	return ddc.dsSets[oldPos], ddc.slices[oldPos]
}
