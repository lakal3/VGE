package main

import (
	"fmt"
	"image"
	"io/ioutil"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func (v *viewerApp) renderToFile() {
	memPool := vapp.Dev.NewMemoryPool()
	mainImageDesc := vk.ImageDescription{
		Width:     1024 * 2,
		Height:    768 * 2,
		Depth:     1,
		Format:    vk.FORMATR8g8b8a8Unorm,
		Layers:    1,
		MipLevels: 1,
	}
	depthDesc := mainImageDesc
	depthDesc.Format = vk.FORMATD32Sfloat
	v.owner.AddChild(memPool)
	mainImage := memPool.ReserveImage(v, mainImageDesc, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit)
	memPool.Allocate(v)
	frp := vapp.NewForwardRenderer(true)
	v.owner.AddChild(frp)
	frp.Setup(v, vapp.Dev, mainImageDesc, 1)
	rc := vk.NewRenderCache(v, vapp.Dev)
	v.owner.AddChild(rc)
	rc.Ctx = v
	pc := v.initCamera(mainImageDesc)
	pc.SetupFrame(vscene.GetFrame(rc), image.Pt(int(mainImageDesc.Width), int(mainImageDesc.Height)))
	var sc vscene.Scene
	sc.Root.Children = append(sc.Root.Children, v.nLights, v.nModel)
	sc.Init()
	frp.Render(&sc, rc, mainImage, 0, nil)
	cp := vmodel.NewCopier(v, vapp.Dev)
	defer cp.Dispose()
	content := cp.CopyFromImage(mainImage, mainImage.FullRange(), "dds", vk.IMAGELayoutUndefined)
	err := ioutil.WriteFile(config.outPath, content, 0660)
	if err != nil {
		v.SetError(err)
	}
	fmt.Println("Save output to ", config.outPath)
}

func (v *viewerApp) initCamera(desc vk.ImageDescription) *vscene.PerspectiveCamera {
	pc := vscene.NewPerspectiveCamera(1000)
	aabb := v.m.Bounds(0, mgl32.Ident4(), true)
	lMod := aabb.Max.Sub(aabb.Min).Len()
	mid := aabb.Max.Add(aabb.Min).Mul(0.5)
	pc = vscene.NewPerspectiveCamera(lMod * 10)
	pc.Target = mid
	cm := mgl32.Vec3{-0.08, 0.2, 0.5}
	pc.Position = pc.Target.Add(cm.Mul(lMod))
	return pc
}
