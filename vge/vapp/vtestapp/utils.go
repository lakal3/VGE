package vtestapp

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type MainImage struct {
	Image  *vk.Image
	Desc   vk.ImageDescription
	Root   vscene.Scene
	Camera vscene.Camera
	pool   *vk.MemoryPool
}

func (m *MainImage) Dispose() {
	if m.pool != nil {
		m.pool.Dispose()
		m.pool, m.Image = nil, nil
	}
}

func NewMainImage() *MainImage {
	m := &MainImage{}
	m.pool = vk.NewMemoryPool(TestApp.Dev)
	m.Desc = vk.ImageDescription{Width: 1024, Height: 768, Depth: 1, Format: vk.FORMATR8g8b8a8Unorm, MipLevels: 1, Layers: 1}
	m.Image = m.pool.ReserveImage(TestApp.Ctx, m.Desc, vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageColorAttachmentBit)
	m.pool.Allocate(TestApp.Ctx)
	m.Root.Init()
	return m
}

func (m *MainImage) Save(testName string, layout vk.ImageLayout) {
	SaveImage(m.Image, testName, layout)
}

func (m *MainImage) ForwardRender(depth bool, render func(cmd *vk.Command, dc *vmodel.DrawContext)) {
	dc := &vmodel.DrawContext{}
	df := vk.FORMATUndefined
	var att []*vk.ImageView
	att = append(att, m.Image.DefaultView(TestApp.Ctx))
	if depth {
		df = vk.FORMATD32Sfloat
		pool := vk.NewMemoryPool(TestApp.Dev)
		dDesc := m.Desc
		dDesc.Format = df
		di := pool.ReserveImage(TestApp.Ctx, dDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		pool.Allocate(TestApp.Ctx)
		defer pool.Dispose()
		att = append(att, di.DefaultView(TestApp.Ctx))
	}
	fp := vk.NewForwardRenderPass(TestApp.Ctx, TestApp.Dev, m.Desc.Format, vk.IMAGELayoutTransferSrcOptimal, df)
	defer fp.Dispose()
	dc.Pass = fp
	dc.Cache = vk.NewRenderCache(TestApp.Ctx, TestApp.Dev)
	cmd := vk.NewCommand(TestApp.Ctx, TestApp.Dev, vk.QUEUEGraphicsBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	fb := vk.NewFramebuffer(TestApp.Ctx, fp, att)
	cmd.BeginRenderPass(fp, fb)
	render(cmd, dc)
	if dc.List != nil {
		cmd.Draw(dc.List)
	}
	cmd.EndRenderPass()
	cmd.Submit()
	cmd.Wait()
}

var kFpTest = vk.NewKey()

func (m *MainImage) RenderScene(time float64, depth bool) {
	rc := vk.NewRenderCache(TestApp.Ctx, TestApp.Dev)
	defer rc.Dispose()
	df := vk.FORMATUndefined
	var att []*vk.ImageView
	att = append(att, m.Image.DefaultView(TestApp.Ctx))
	if depth {
		df = vk.FORMATD32Sfloat
		pool := vk.NewMemoryPool(TestApp.Dev)
		dDesc := m.Desc
		dDesc.Format = df
		di := pool.ReserveImage(TestApp.Ctx, dDesc, vk.IMAGEUsageDepthStencilAttachmentBit)
		pool.Allocate(TestApp.Ctx)
		defer pool.Dispose()
		att = append(att, di.DefaultView(TestApp.Ctx))
	}
	fp := vk.NewForwardRenderPass(TestApp.Ctx, TestApp.Dev, m.Desc.Format, vk.IMAGELayoutTransferSrcOptimal, df)
	defer fp.Dispose()
	fb := vk.NewFramebuffer(TestApp.Ctx, fp, att)
	cmd := vk.NewCommand(TestApp.Ctx, TestApp.Dev, vk.QUEUEGraphicsBit, true)

	defer cmd.Dispose()
	cmd.Begin()
	bg := vscene.NewDrawPhase(rc, fp, vscene.LAYERBackground, cmd, func() {
		cmd.BeginRenderPass(fp, fb)
	}, nil)
	dp := vscene.NewDrawPhase(rc, fp, vscene.LAYER3D, cmd, nil, nil)
	ui := vscene.NewDrawPhase(rc, fp, vscene.LAYERUI, cmd, nil, func() {
		cmd.EndRenderPass()
	})
	frame := vscene.GetFrame(rc)
	m.Root.Process(time, &vscene.AnimatePhase{}, &vscene.PredrawPhase{F: frame, Scene: &m.Root, Cache: rc, Cmd: cmd}, bg, dp, ui)
	cmd.Submit()
	cmd.Wait()
}

func SaveImage(image *vk.Image, testName string, layout vk.ImageLayout) {
	cp := vmodel.NewCopier(TestApp.Ctx, TestApp.Dev)
	defer cp.Dispose()
	ir := image.FullRange()
	ir.Layout = layout
	content := cp.CopyFromImage(image, ir, "dds", vk.IMAGELayoutTransferSrcOptimal)
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		TestApp.Ctx.T.Log("Unable to save test image, missing environment variable VGE_TEST_DIR")
		return
	}
	fPath := filepath.Join(testDir, testName+".dds")
	err := ioutil.WriteFile(fPath, content, 0660)
	if err != nil {
		TestApp.Ctx.SetError(err)
	}
	TestApp.Ctx.T.Log("Saved test image to ", fPath)
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
