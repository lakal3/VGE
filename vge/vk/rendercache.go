package vk

import "sync/atomic"

var lastKey uint64 = 1

type Key uint64

func NewKey() Key {
	return Key(atomic.AddUint64(&lastKey, 1))
}

func NewKeys(howMany uint64) Key {
	return Key(atomic.AddUint64(&lastKey, howMany) - howMany + 1)
}

type Constructor func() interface{}

type RenderCache struct {
	Device   *Device
	perCache Owner
	perFrame Owner
}

func (rc *RenderCache) Dispose() {
	rc.NewFrame()
	rc.perCache.Dispose()
}

func NewRenderCache(dev *Device) *RenderCache {
	return &RenderCache{Device: dev}
}

func (rc *RenderCache) Get(key Key, cons Constructor) interface{} {
	return rc.perCache.Get(key, cons)
}

func (rc *RenderCache) GetPerFrame(key Key, cons Constructor) interface{} {
	return rc.perFrame.Get(key, cons)
}

func (rc *RenderCache) SetPerFrame(key Key, val interface{}) {
	rc.perFrame.Set(key, val)

}

func (rc *RenderCache) DisposePerFrame(disp Disposable) {
	rc.perFrame.AddChild(disp)
}

func (rc *RenderCache) NewFrame() {
	rc.perFrame.Dispose()
}
