package vdraw3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
)

type FrozenID uint32

type Phase interface {
	PhaseName() string
}

type Frozen interface {
	Reserve(fi *vk.FrameInstance, storageOffset uint32) (newOffset uint32)
	Support(fi *vk.FrameInstance, phase Phase) bool
	Render(fi *vk.FrameInstance, phase Phase)
	Clone() Frozen
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
	for _, item := range fl.items {
		item.Render(fi, phase)
	}
}

func (fl *FreezeList) Clone() {
	for idx, fr := range fl.items {
		fl.items[idx] = fr.Clone()
	}
}

type Render struct {
	Name    string
	Shaders *shaders.Pack
	DSFrame *vk.DescriptorSet
}

func (r Render) PhaseName() string {
	return r.Name
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
	Decal             *uint32
	ViewTransform     mgl32.Mat4
	RenderTransparent func(priority float32, render func(dl *vk.DrawList, pass *vk.GeneralRenderPass))
}

type RenderOverlay struct {
	Render
	DL   *vk.DrawList
	Pass *vk.GeneralRenderPass
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
	Directional   bool
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
	AddDecal(storagePosition uint32) (prev uint32)
	PopDecal(oldPos uint32)
	UpdateStorage(storagePosition uint32, index uint32, values ...float32)
}
