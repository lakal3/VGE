package vscene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

type Layer uint32

const (
	LAYERBackground Layer = 1000
	LAYER3D         Layer = 2000
	// Render 3D shaders for probe. There will be only simple frame
	LAYER3DProbe     Layer = 2050
	LAYERTransparent Layer = 3000
	LAYERUI          Layer = 4000
)

type Phase interface {
	Begin() (atEnd func())
	GetCache() *vk.RenderCache
}

type DrawPhase interface {
	Phase
	GetDC(layer Layer) *vmodel.DrawContext
}

type ShadowPhase interface {
	Phase
	DrawShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex)
	DrawSkinnedShadow(mesh vmodel.Mesh, world mgl32.Mat4, albedoTexture vmodel.ImageIndex, aniMatrix []mgl32.Mat4)
}

func NewDrawPhase(rc *vk.RenderCache, pass vk.RenderPass, layer Layer, cmd *vk.Command, begin func(), commit func()) DrawPhase {
	return &BasicDrawPhase{DrawContext: vmodel.DrawContext{Cache: rc, Pass: pass}, Layer: layer, Cmd: cmd, begin: begin, commit: commit}
}

type BasicDrawPhase struct {
	vmodel.DrawContext
	Cmd    *vk.Command
	Layer  Layer
	begin  func()
	commit func()
	// FP  *vk.Framebuffer
}

func (d *BasicDrawPhase) GetCache() *vk.RenderCache {
	return d.DrawContext.Cache
}

func (d *BasicDrawPhase) GetDC(layer Layer) *vmodel.DrawContext {
	if d.Layer == layer {
		return &d.DrawContext
	}
	return nil
}

func (d *BasicDrawPhase) Begin() (atEnd func()) {

	if d.begin != nil {
		d.begin()
	}
	return func() {
		if d.List != nil {
			d.Cmd.Draw(d.List)
		}
		if d.commit != nil {
			d.commit()
		}
	}
}

type PredrawPhase struct {
	Scene   *Scene
	Cmd     *vk.Command
	Cache   *vk.RenderCache
	Frame   Frame
	Needeed []vk.SubmitInfo
	Pending []func()
}

func (p *PredrawPhase) GetCache() *vk.RenderCache {
	return p.Cache
}

func (p *PredrawPhase) Begin() (atEnd func()) {
	return nil
}

type AnimatePhase struct {
	Cache *vk.RenderCache
}

func (a *AnimatePhase) GetCache() *vk.RenderCache {
	return nil
}

func (a *AnimatePhase) Begin() (atEnd func()) {
	return nil
}

type BoudingBox struct {
	box   vmodel.AABB
	first bool
}

func (b *BoudingBox) GetCache() *vk.RenderCache {
	return nil
}

func (b *BoudingBox) Begin() (atEnd func()) {
	return nil
}

func (b *BoudingBox) Add(aabb vmodel.AABB) {
	if b.first {
		b.box.Add(true, aabb.Min)
		b.first = false
	} else {
		b.box.Add(false, aabb.Min)
	}
	b.box.Add(false, aabb.Max)
}

func (b *BoudingBox) Get() (aabb vmodel.AABB, empty bool) {
	return b.box, b.first
}

type LightPhase interface {
	// Add light to scene
	AddLight(standard Light, shadowMap *vk.ImageView, samples *vk.Sampler)

	// Add Special light. Light phase return true if it can handle given special light
	AddSpecialLight(special interface{}, shadowMap *vk.ImageView, samples *vk.Sampler) bool
}
