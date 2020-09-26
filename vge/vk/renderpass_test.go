//

//go:generate glslangValidator.exe -V testsh/testsh.vert.glsl -o testsh/testsh.vert.spv
//go:generate glslangValidator.exe -V testsh/testsh.frag.glsl -o testsh/testsh.frag.spv

package vk

import (
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestNewForwardRenderPass(t *testing.T) {
	tc := &testContext{t: t}
	a := NewApplication(tc, "Test")
	a.AddValidation(tc)
	a.AddDynamicDescriptors(tc)
	a.Init(tc)
	if a.hInst == 0 {
		t.Error("No instance for initialize app")
	}
	d := NewDevice(tc, a, 0)
	if d == nil {
		t.Error("Failed to initialize application")
	}
	ld := NewNativeImageLoader(tc, a)
	testRender(tc, d, ld)

	d.Dispose()
	a.Dispose()
}

type testPipeline struct {
	pl        *GraphicsPipeline
	rp        *ForwardRenderPass
	l1        *DescriptorLayout
	dp1       *DescriptorPool
	ds1       *DescriptorSet
	iColor    *Image
	ubWorld   *Buffer
	l2        *DescriptorLayout
	dp2       *DescriptorPool
	ds2       *DescriptorSet
	s         *Sampler
	testImage *Image
}

func (tp *testPipeline) Dispose() {

}

var edges = []float32{-0.5, -0.5, 0, 0.5, 0.5, -0.5}
var color = []float32{0, 1, 0, 1}

const triangleCount = 16

func testRender(tc *testContext, d *Device, ld ImageLoader) {
	testPattern, err := ioutil.ReadFile("../../assets/tests/uvchecker.png")
	if err != nil {
		tc.SetError(err)
		return
	}
	var testImageDesc ImageDescription
	ld.DescribeImage(tc, "png", &testImageDesc, testPattern)
	fp := NewForwardRenderPass(tc, d, FORMATR8g8b8a8Unorm, IMAGELayoutTransferSrcOptimal, FORMATUndefined)
	if fp == nil {
		return
	}
	defer fp.Dispose()

	mp := d.NewMemoryPool()
	mi := ImageDescription{
		Width:     1024,
		Height:    768,
		Depth:     1,
		Format:    FORMATR8g8b8a8Unorm,
		Layers:    1,
		MipLevels: 1,
	}
	mainImage := mp.ReserveImage(tc, mi, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit)
	testImage := mp.ReserveImage(tc, mi, IMAGEUsageTransferDstBit|IMAGEUsageSampledBit)
	bImage := mp.ReserveBuffer(tc, testImageDesc.ImageSize(), true, BUFFERUsageTransferSrcBit)
	ib := mp.ReserveBuffer(tc, mi.ImageSize(), true, BUFFERUsageTransferDstBit)
	vb := mp.ReserveBuffer(tc, 3*2*4, true, BUFFERUsageVertexBufferBit)
	ubColor := mp.ReserveBuffer(tc, 4*4, true, BUFFERUsageUniformBufferBit)
	ubWorld := mp.ReserveBuffer(tc, MinUniformBufferOffsetAlignment*triangleCount, true, BUFFERUsageUniformBufferBit)
	mp.Allocate(tc)
	copy(vb.Bytes(tc), Float32ToBytes(edges))
	copy(ubColor.Bytes(tc), Float32ToBytes(color))
	worldBuf := ubWorld.Bytes(tc)
	ld.LoadImage(tc, "png", testPattern, bImage)

	for idx := 0; idx < triangleCount; idx++ {
		m := mgl32.HomogRotate3DZ(float32(idx) * math.Pi * 2 / triangleCount)
		copy(worldBuf[idx*256:], Float32ToBytes(m[:]))
	}

	mainView := mainImage.DefaultView(tc)
	fb := NewFramebuffer(tc, fp, []*ImageView{mainView})
	defer fb.Dispose()
	tp := &testPipeline{rp: fp, testImage: testImage, ubWorld: ubWorld}
	tp.copyImage(tc, d, bImage, testImage)
	tp.build(tc, d)
	defer tp.Dispose()
	cmd := NewCommand(tc, d, QUEUEGraphicsBit, true)
	if cmd == nil {
		return
	}
	defer cmd.Dispose()
	timer := NewTimerPool(tc, d, 2)
	defer timer.Dispose()
	cmd.Begin()
	cmd.WriteTimer(timer, 0, PIPELINEStageTopOfPipeBit)

	cmd.BeginRenderPass(fp, fb)
	drawList := &DrawList{}
	for idx := uint32(0); idx < triangleCount; idx++ {
		drawList.Draw(tp.pl, 0, 3).AddInput(0, vb).AddDescriptor(0, tp.ds1).
			AddDynamicDescriptor(1, tp.ds2, MinUniformBufferOffsetAlignment*idx)
	}
	cmd.Draw(drawList)
	cmd.EndRenderPass()
	cmd.WriteTimer(timer, 1, PIPELINEStageFragmentShaderBit)

	r := mainImage.FullRange()
	r.Layout = IMAGELayoutTransferSrcOptimal
	cmd.CopyImageToBuffer(ib, mainImage, &r)
	cmd.Submit()
	cmd.Wait()
	times := timer.Get(tc)
	tc.t.Log("Time was ", times[0], times[1], times[1]-times[0])

	defer tp.pl.Dispose()
	im := image.NewRGBA(image.Rect(0, 0, int(mainImage.Description.Width), int(mainImage.Description.Height)))
	copy(im.Pix, ib.Bytes(tc))
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		tc.t.Log("Unable to save test image, missing environment variable VGE_TEST_DIR")
		return
	}
	fOut, err := os.Create(filepath.Join(testDir, "vk.png"))
	if err != nil {
		tc.SetError(err)
	}
	err = png.Encode(fOut, im)
	if err != nil {
		tc.SetError(err)
	}
}

func (tp *testPipeline) build(ctx APIContext, dev *Device) {
	tp.l1 = dev.NewDynamicDescriptorLayout(ctx, DESCRIPTORTypeCombinedImageSampler, SHADERStageFragmentBit,
		8, DESCRIPTORBindingPartiallyBoundBitExt|DESCRIPTORBindingUpdateUnusedWhilePendingBitExt)
	tp.l2 = dev.NewDescriptorLayout(ctx, DESCRIPTORTypeUniformBufferDynamic, SHADERStageVertexBit, 1)
	tp.dp1 = NewDescriptorPool(ctx, tp.l1, 1)
	dev.Get(ctx, NewKey(), func(ctx APIContext) interface{} {
		return tp.dp1
	})
	tp.dp2 = NewDescriptorPool(ctx, tp.l2, 1)
	dev.Get(ctx, NewKey(), func(ctx APIContext) interface{} {
		return tp.dp2
	})
	tp.ds1 = tp.dp1.Alloc(ctx)
	tp.ds2 = tp.dp2.Alloc(ctx)
	tp.s = dev.NewSampler(ctx, SAMPLERAddressModeRepeat)
	tp.ds1.WriteImage(ctx, 0, 1, tp.testImage.DefaultView(ctx), tp.s)
	tp.ds2.WriteBuffer(ctx, 0, 0, tp.ubWorld)
	tp.pl = NewGraphicsPipeline(ctx, dev)
	tp.pl.AddLayout(ctx, tp.l1)
	tp.pl.AddLayout(ctx, tp.l2)
	tp.pl.AddVextexInput(ctx, VERTEXInputRateVertex, FORMATR32g32Sfloat)
	code, err := ioutil.ReadFile("testsh/testsh.frag.spv")
	if err != nil {
		ctx.SetError(err)
		return
	}
	tp.pl.AddShader(ctx, SHADERStageFragmentBit, code)
	code, err = ioutil.ReadFile("testsh/testsh.vert.spv")
	if err != nil {
		ctx.SetError(err)
		return
	}
	tp.pl.AddShader(ctx, SHADERStageVertexBit, code)
	tp.pl.Create(ctx, tp.rp)
}

func (tp *testPipeline) copyImage(tc *testContext, dev *Device, b *Buffer, img *Image) {
	cmd := NewCommand(tc, dev, QUEUETransferBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	r := img.FullRange()
	cmd.SetLayout(img, &r, IMAGELayoutTransferDstOptimal)
	cmd.CopyBufferToImage(img, b, &r)
	cmd.SetLayout(img, &r, IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
}
