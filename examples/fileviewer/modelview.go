package main

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vdraw3d"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vmodel/objloader"
	"github.com/lakal3/vge/vge/vscene"
	"os"
	"path/filepath"
	"strings"
)

type modelViewer struct {
	path string
	info os.FileInfo

	state   uint32
	rm      func()
	view    *vdraw3d.View
	md      *vmodel.Model
	err     error
	dev     *vk.Device
	mbImage vmodel.ImageIndex
	probe   vdraw3d.FrozenID
	pc      *vscene.PerspectiveCamera
}

func (mv *modelViewer) HandleEvent(event vapp.Event) {
	mv.view.HandleEvent(event)
}

func (mv *modelViewer) Reserve(fi *vk.FrameInstance) {
	if mv.state > 0 {
		// View closed
		mv.state = 2
		mv.rm()
		fi.AddChild(closeModelView{md: mv.md})
	} else {
		mv.view.Reserve(fi)
	}
}

func (mv *modelViewer) PreRender(fi *vk.FrameInstance) (done vapp.Completed) {
	if mv.state != 2 {
		return mv.view.PreRender(fi)
	}
	return nil
}

func (mv *modelViewer) Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass) {
	if mv.state != 2 {
		mv.view.Render(fi, cmd, rp)
	}
}

type closeModelView struct {
	md *vmodel.Model
}

func (c closeModelView) Dispose() {
	c.md.Dispose()
}

func (mv *modelViewer) Close() {
	if mv.state == 0 {
		mv.state = 1
	}
}

func (mv *modelViewer) Stat(fr *vimgui.UIFrame) {
	if mv.err != nil {
		fr.NewLine(-100, 20, 5)
		fr.WithTags("error")
		vimgui.Label(fr, "Failed to load file: "+mv.err.Error())
		return
	}
	DrawFileInfo(fr, mv.path, mv.info)
}

func (mv *modelViewer) objLoader(content []byte) (err error) {
	dl := vasset.DirectoryLoader{Directory: filepath.Dir(mv.path)}
	b := &vmodel.ModelBuilder{MipLevels: 5}
	err = mv.loadBg(b)
	if err != nil {
		return err
	}
	ol := &objloader.ObjLoader{Loader: dl, Builder: b}
	err = ol.Load(bytes.NewReader(content))
	if err != nil {
		return err
	}
	mv.md, err = b.ToModel(mv.dev)
	if err != nil {
		return err
	}

	return
}

func (mv *modelViewer) gltfLoader(content []byte) (err error) {
	dl := vasset.DirectoryLoader{Directory: filepath.Dir(mv.path)}
	b := &vmodel.ModelBuilder{MipLevels: 5}
	b.ShaderFactory = func(dev *vk.Device, propSet vmodel.MaterialProperties) (sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
		return nil, nil, nil, nil
	}
	err2 := mv.loadBg(b)
	if err2 != nil {
		return err2
	}
	ol := &gltf2loader.GLTF2Loader{Loader: dl, Builder: b}

	err = ol.ParseGltf(content)
	if err != nil {
		return err
	}
	err = ol.Convert(0)
	if err != nil {
		return err
	}
	mv.md, err = b.ToModel(mv.dev)
	if err != nil {
		return err
	}

	return
}

func (mv *modelViewer) loadBg(b *vmodel.ModelBuilder) error {
	if len(settings.bgImage) != 0 {
		kind := filepath.Ext(settings.bgImage)[1:]
		content, err := os.ReadFile(settings.bgImage)
		if err != nil {
			return err
		}
		mv.mbImage = b.AddImage(kind, content, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	}
	return nil
}

func (mv *modelViewer) glbLoader(content []byte) (err error) {
	dl := vasset.DirectoryLoader{Directory: filepath.Dir(mv.path)}
	b := &vmodel.ModelBuilder{MipLevels: 5}
	b.ShaderFactory = func(dev *vk.Device, propSet vmodel.MaterialProperties) (sh vmodel.Shader, layout *vk.DescriptorLayout, ubf []byte, images []vmodel.ImageIndex) {
		return nil, nil, nil, nil
	}
	err = mv.loadBg(b)
	if err != nil {
		return err
	}
	ol := &gltf2loader.GLTF2Loader{Loader: dl, Builder: b}
	err = ol.ParseGLB(content)
	if err != nil {
		return err
	}
	err = ol.Convert(0)
	if err != nil {
		return err
	}

	mv.md, err = b.ToModel(mv.dev)
	if err != nil {
		return err
	}

	return
}

var probeKey = vk.NewKey()

func (mv *modelViewer) drawStatic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	p := vmodel.NewMaterialProperties()
	p.SetColor(vmodel.CIntensity, mgl32.Vec4{0.7, 0.7, 0.7, 1})
	vdraw3d.DrawDirectionalLight(dl, mgl32.Vec3{0, -1, 0.1}.Normalize(), p)
	if mv.mbImage != 0 {
		dl.Exclude(vdraw3d.Main, vdraw3d.Main)
		vdraw3d.DrawBackground(dl, mv.md, mv.mbImage)
		dl.Include(vdraw3d.Main, vdraw3d.Main)
		mv.probe = vdraw3d.DrawProbe(dl, probeKey, mv.pc.Target)
		dl.Exclude(mv.probe, mv.probe)
	}
	n := mv.md.GetNode(0)
	n.Enum(mgl32.Ident4(), func(local mgl32.Mat4, n vmodel.Node) {
		if n.Mesh >= 0 {
			if n.Skin == 0 {
				vdraw3d.DrawMesh(dl, mv.md.GetMesh(n.Mesh), local, mv.md.GetMaterial(n.Material).Props)
			}
		}
	})
}

func (mv *modelViewer) drawDynamic(v *vdraw3d.View, dl *vdraw3d.FreezeList) {
	if mv.probe != 0 {
		dl.Exclude(mv.probe, mv.probe)
	}
	n := mv.md.GetNode(0)
	n.Enum(mgl32.Ident4(), func(local mgl32.Mat4, n vmodel.Node) {
		if n.Mesh >= 0 {
			if n.Skin != 0 {
				sk := mv.md.GetSkin(n.Skin)
				vdraw3d.DrawAnimated(dl, mv.md.GetMesh(n.Mesh), sk, sk.Animations[0], mv.view.Elapsed,
					local, mv.md.GetMaterial(n.Material).Props)
			}
		}
	})
}

func (mv *modelViewer) toHome() {
	aabb := mv.md.Bounds(0, mgl32.Ident4(), true)
	mv.pc = vscene.NewPerspectiveCamera(aabb.Len() * 10)
	if mv.pc.Far < 10 {
		mv.pc.Far = 10
	}
	mv.pc.Near = mv.pc.Far / 10000.0
	c := vapp.OrbitControlFrom(0, nil, mv.pc)
	c.ZoomTo(aabb.Center(), aabb.Len())
	mv.view.Camera = c

}

func newModelViewer(path string, info os.FileInfo, content []byte) (isModel bool) {
	var loader func(content []byte) error
	mv := &modelViewer{path: path, info: info, dev: vapp.Dev}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".obj":
		loader = mv.objLoader
	case ".glb":
		loader = mv.glbLoader
	case ".gltf":
		loader = mv.gltfLoader
	}
	if loader == nil {
		return false
	}
	mv.err = loader(content)
	if mv.err != nil {
		return true
	}
	if settings.phong {
		mv.view = vdraw3d.NewCustomView(mv.dev, app.phongshaders, mv.drawStatic, mv.drawDynamic)
	} else {
		mv.view = vdraw3d.NewView(mv.dev, mv.drawStatic, mv.drawDynamic)
	}
	mv.view.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		return vdraw.Area{From: mgl32.Vec2{float32(desc.Width)/4 + 1, StatHeight + 1}, To: mgl32.Vec2{float32(desc.Width), float32(desc.Height)}}
	}
	mv.toHome()
	mv.rm = SetViewer(mv, mv)
	return true
}
