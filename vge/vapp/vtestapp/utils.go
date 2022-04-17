package vtestapp

import (
	"errors"
	"github.com/lakal3/vge/vge/vasset/pngloader"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
	"image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// PngSupport option register PNG handler that can load and save images in PNG format
type PngSupport struct {
}

func (p PngSupport) InitOption() {
	pngloader.RegisterPngLoader()
}

type MainImage struct {
	Image *vk.Image
	Desc  vk.ImageDescription
	Root  vscene.Scene
	// Camera vscene.Camera
	pool *vk.MemoryPool
}

func (m *MainImage) Dispose() {
	if m.pool != nil {
		m.pool.Dispose()
		m.pool, m.Image = nil, nil
	}
}

// NewMainImage construct new MainImage that you can use to render test scene. Image size is 1024*768
func NewMainImage() *MainImage {
	return NewMainImageDesc(vk.ImageDescription{Width: 1024, Height: 768, Depth: 1, Format: vk.FORMATR8g8b8a8Unorm, MipLevels: 1, Layers: 1},
		vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageColorAttachmentBit)
}

// NewMainImageDesc construct new MainImage in given format.
func NewMainImageDesc(desc vk.ImageDescription, usage vk.ImageUsageFlags) *MainImage {
	m := &MainImage{}
	m.pool = vk.NewMemoryPool(TestApp.Dev)
	m.Desc = desc
	m.Image = m.pool.ReserveImage(m.Desc, usage)
	m.pool.Allocate()
	m.Root.Init()
	AddChild(m)
	return m
}

func (m *MainImage) Save(testName string, layout vk.ImageLayout) {
	SaveImage(m.Image, testName, layout)
}

func (m *MainImage) SaveKind(kind string, testName string, layout vk.ImageLayout) {
	SaveImageKind(m.Image, kind, testName, layout)
}

func (m *MainImage) ForwardRender(depth bool, render func(cmd *vk.Command, dc *vmodel.DrawContext)) {
	dc := &vmodel.DrawContext{}
	df := vk.FORMATUndefined
	var att []*vk.ImageView
	att = append(att, m.Image.DefaultView())
	if depth {
		df = vk.FORMATD32Sfloat
		pool := vk.NewMemoryPool(TestApp.Dev)
		dDesc := m.Desc
		dDesc.Format = df
		di := pool.ReserveImage(dDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		pool.Allocate()
		defer pool.Dispose()
		att = append(att, di.DefaultView())
	}
	fp := vk.NewForwardRenderPass(TestApp.Dev, m.Desc.Format, vk.IMAGELayoutTransferSrcOptimal, df)
	defer fp.Dispose()
	dc.Pass = fp
	rc := vk.NewRenderCache(TestApp.Dev)
	defer rc.Dispose()
	dc.Frame = &vscene.SimpleFrame{Cache: rc}

	cmd := vk.NewCommand(TestApp.Dev, vk.QUEUEGraphicsBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	fb := vk.NewFramebuffer(fp, att)
	defer fb.Dispose()
	cmd.BeginRenderPass(fp, fb)
	render(cmd, dc)
	if dc.List != nil {
		cmd.Draw(dc.List)
	}
	cmd.EndRenderPass()
	cmd.Submit()
	cmd.Wait()

}

func (m *MainImage) RenderScene(time float64, depth bool) {
	m.RenderSceneAt(time, depth, nil)
}

func (m *MainImage) RenderSceneAt(time float64, depth bool, camera vscene.Camera) {
	rc := vk.NewRenderCache(TestApp.Dev)
	defer rc.Dispose()
	df := vk.FORMATUndefined
	var att []*vk.ImageView
	att = append(att, m.Image.DefaultView())
	if depth {
		df = vk.FORMATD32Sfloat
		pool := vk.NewMemoryPool(TestApp.Dev)
		dDesc := m.Desc
		dDesc.Format = df
		di := pool.ReserveImage(dDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		pool.Allocate()
		defer pool.Dispose()
		att = append(att, di.DefaultView())
	}
	fp := vk.NewForwardRenderPass(TestApp.Dev, m.Desc.Format, vk.IMAGELayoutTransferSrcOptimal, df)
	defer fp.Dispose()
	fb := vk.NewFramebuffer(fp, att)
	cmd := vk.NewCommand(TestApp.Dev, vk.QUEUEGraphicsBit, true)

	defer cmd.Dispose()
	cmd.Begin()
	frame := &vscene.SimpleFrame{Cache: rc}
	if camera != nil {
		frame.SSF.Projection, frame.SSF.View = camera.CameraProjection(
			image.Pt(int(m.Image.Description.Width), int(m.Image.Description.Height)))
	}
	bg := vscene.NewDrawPhase(frame, fp, vscene.LAYERBackground, cmd, func() {
		cmd.BeginRenderPass(fp, fb)
	}, nil)
	dp := vscene.NewDrawPhase(frame, fp, vscene.LAYER3D, cmd, nil, nil)
	ui := vscene.NewDrawPhase(frame, fp, vscene.LAYERUI, cmd, nil, func() {
		cmd.EndRenderPass()
	})
	m.Root.Process(time, frame, &vscene.AnimatePhase{},
		&vscene.PredrawPhase{Scene: &m.Root, Cmd: cmd}, bg, dp, ui)
	cmd.Submit()
	cmd.Wait()
}

// SaveImage image to test dir using DDS format
func SaveImage(image *vk.Image, testName string, layout vk.ImageLayout) {
	SaveImageKind(image, "dds", testName, layout)
}

// SaveImageKind saves image to test dir using kind format. You must ensure that proper image decoder has been registered
func SaveImageKind(image *vk.Image, kind string, testName string, layout vk.ImageLayout) error {
	cp := vmodel.NewCopier(TestApp.Dev)
	defer cp.Dispose()
	ir := image.FullRange()
	ir.Layout = layout
	content, err := cp.CopyFromImage(image, ir, kind, vk.IMAGELayoutTransferSrcOptimal)
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		TestApp.Dev.ReportError(errors.New("Unable to save test image, missing environment variable VGE_TEST_DIR"))
		return nil
	}
	fPath := filepath.Join(testDir, testName+"."+kind)
	err = ioutil.WriteFile(fPath, content, 0660)
	if err != nil {
		return err
	}
	// TestApp.Ctx.T.Log("Saved test image to ", fPath)
	return nil
}

type TestLoader struct {
	Path string
}

func (t TestLoader) Open(filename string) (io.ReadCloser, error) {
	testDir := os.Getenv("VGE_ASSET_DIR")
	if len(testDir) == 0 {
		return nil, errors.New("Missing environment variable VGE_ASSET_DIR")
	}
	f, err := os.Open(filepath.Join(testDir, t.Path, filename))
	return f, err
}
