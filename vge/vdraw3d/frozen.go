package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
)

type FrozenID int

const (
	Main FrozenID = -1
)

type Phase interface {
	TargetID() FrozenID
}

type Frozen interface {
	Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32)
	Support(fi *vk.FrameInstance, phase Phase) bool
	Render(fi *vk.FrameInstance, phase Phase)
	Clone() Frozen
}

type Excluder interface {
	Frozen
	// -1 include, 0 no change, 1 exclude
	Exclude(phase Phase) int
}

type FreezeList struct {
	BaseID FrozenID
	items  []Frozen
}

func (fl *FreezeList) GetID() FrozenID {
	return FrozenID(len(fl.items)) + fl.BaseID
}

func (fl *FreezeList) Add(fr Frozen) FrozenID {
	id := fl.GetID()
	fl.items = append(fl.items, fr)
	return id
}

func (fl *FreezeList) RenderAll(fi *vk.FrameInstance, phase Phase) {
	exclude := false
	for _, item := range fl.items {
		el, ok := item.(Excluder)
		if ok {
			switch el.Exclude(phase) {
			case -1:
				exclude = false
			case 1:
				exclude = true
			}
		}
		if !exclude && item.Support(fi, phase) {
			item.Render(fi, phase)
		}
	}
}

func (fl *FreezeList) Clone() {
	for idx, fr := range fl.items {
		fl.items[idx] = fr.Clone()
	}
}

type excluder struct {
	from   FrozenID
	to     FrozenID
	method int
}

func (e excluder) Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32) {
	return storageOffset
}

func (e excluder) Support(fi *vk.FrameInstance, phase Phase) bool {
	return false
}

func (e excluder) Render(fi *vk.FrameInstance, phase Phase) {

}

func (e excluder) Clone() Frozen {
	return e
}

func (e excluder) Exclude(phase Phase) int {
	tid := phase.TargetID()
	if tid >= e.from && tid <= e.to {
		return e.method
	}
	return 0
}

func (fl *FreezeList) Exclude(from FrozenID, to FrozenID) {
	fl.Add(excluder{from: from, to: to, method: 1})
}

func (fl *FreezeList) Include(from FrozenID, to FrozenID) {
	fl.Add(excluder{from: from, to: to, method: -1})
}

type Render struct {
	Target  FrozenID
	Shaders *shaders.Pack
	DSFrame *vk.DescriptorSet
}

func (r Render) TargetID() FrozenID {
	return r.Target
}

type RenderDepth struct {
	Render
	DL   *vk.DrawList
	Pass *vk.GeneralRenderPass
}

type RenderPick struct {
	Render
	DL     *vk.DrawList
	Pass   *vk.GeneralRenderPass
	DSPick *vk.DescriptorSet
}

type RenderColor struct {
	Render
	DL                *vk.DrawList
	Pass              *vk.GeneralRenderPass
	Probe             *uint32
	ViewTransform     mgl32.Mat4
	RenderTransparent func(priority float32, render func(dl *vk.DrawList, pass *vk.GeneralRenderPass))
}

type RenderMaps struct {
	Render
	Static        *FreezeList
	Dynamic       *FreezeList
	AtEnd         func(end func() []vk.SubmitInfo)
	UpdateStorage func(storagePosition uint32, index uint32, values ...float32)
}

type RenderShadow struct {
	Render
	DL            *vk.DrawList
	Pass          *vk.GeneralRenderPass
	DSShadowFrame *vk.DescriptorSet
}

type RenderProbe struct {
	Render
	DL           *vk.DrawList
	Pass         *vk.GeneralRenderPass
	DSProbeFrame *vk.DescriptorSet
}

type UpdateFrame interface {
	Phase
	IsStatic() bool
	AddView(view vk.VImageView, sampler *vk.Sampler) float32
	AddLight(storagePosition uint32) (prev float32)
	AddDecal(storagePosition uint32) (prev float32)
	UpdateStorage(storagePosition uint32, index uint32, values ...float32)
}
