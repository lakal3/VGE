//

//go:generate glslangValidator -V testsh/testsh.vert.glsl -o testsh/testsh.vert.spv
//go:generate glslangValidator -V testsh/testsh.frag.glsl -o testsh/testsh.frag.spv

package vk

import (
	_ "embed"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestNewForwardRenderPass(t *testing.T) {
	a, err := NewApplication("Test")
	if err != nil {
		t.Fatal("NewApplication ", err)
	}
	a.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}
	a.AddValidation()
	a.AddDynamicDescriptors()
	a.Init()
	if a.hInst == 0 {
		t.Error("No instance for initialize app")
	}
	d := NewDevice(a, 0)
	if d == nil {
		t.Error("Failed to initialize application")
	}
	d.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}

	ld := NewNativeImageLoader(a)
	err = testRender(t, d, ld)
	if err != nil {
		t.Error("Test render ", err)
	}

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

func testRender(t *testing.T, d *Device, ld ImageLoader) error {
	testPattern, err := ioutil.ReadFile("../../assets/tests/uvchecker.png")
	if err != nil {
		return err
	}
	var testImageDesc ImageDescription
	err = ld.DescribeImage("png", &testImageDesc, testPattern)
	if err != nil {
		t.Fatal("Describe image", err)
	}
	fp := NewGeneralRenderPass(d, false, []AttachmentInfo{
		AttachmentInfo{FinalLayout: IMAGELayoutTransferSrcOptimal, Format: FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.1, 0.1, 0.1, 1}},
		AttachmentInfo{FinalLayout: IMAGELayoutTransferSrcOptimal, Format: FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.8, 0.8, 0.8, 1}},
	})
	if fp == nil {
		return nil
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
	mainImage := mp.ReserveImage(mi, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit)
	grayImage := mp.ReserveImage(mi, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit)
	testImage := mp.ReserveImage(mi, IMAGEUsageTransferDstBit|IMAGEUsageSampledBit)
	bImage := mp.ReserveBuffer(testImageDesc.ImageSize(), true, BUFFERUsageTransferSrcBit)
	ib := mp.ReserveBuffer(mi.ImageSize(), true, BUFFERUsageTransferDstBit)
	ibGray := mp.ReserveBuffer(mi.ImageSize(), true, BUFFERUsageTransferDstBit)
	vb := mp.ReserveBuffer(3*2*4, true, BUFFERUsageVertexBufferBit)
	ubColor := mp.ReserveBuffer(4*4, true, BUFFERUsageUniformBufferBit)
	ubWorld := mp.ReserveBuffer(MinUniformBufferOffsetAlignment*triangleCount, true, BUFFERUsageUniformBufferBit)
	mp.Allocate()
	copy(vb.Bytes(), Float32ToBytes(edges))
	copy(ubColor.Bytes(), Float32ToBytes(color))
	worldBuf := ubWorld.Bytes()
	err = ld.LoadImage("png", testPattern, bImage)
	if err != nil {
		return err
	}

	for idx := 0; idx < triangleCount; idx++ {
		m := mgl32.HomogRotate3DZ(float32(idx) * math.Pi * 2 / triangleCount)
		copy(worldBuf[idx*256:], Float32ToBytes(m[:]))
	}

	mainView := mainImage.DefaultView()
	grayView := grayImage.DefaultView()
	fb := NewFramebuffer(fp, []*ImageView{mainView, grayView})
	defer fb.Dispose()
	tp := &testPipeline{rp: fp, testImage: testImage, ubWorld: ubWorld}
	tp.copyImage(d, bImage, testImage)
	err = tp.build(d)
	if err != nil {
		return err
	}

	defer tp.Dispose()
	cmd := NewCommand(d, QUEUEGraphicsBit, true)
	if cmd == nil {
		return nil
	}
	defer cmd.Dispose()
	timer := NewTimerPool(d, 2)
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
	cmd.CopyImageToBuffer(ibGray, grayImage, &r)
	cmd.Submit()
	cmd.Wait()
	times := timer.Get()
	t.Log("Time was ", times[0], times[1], times[1]-times[0])

	defer tp.pl.Dispose()
	im := image.NewRGBA(image.Rect(0, 0, int(mainImage.Description.Width), int(mainImage.Description.Height)))
	copy(im.Pix, ib.Bytes())
	imGray := image.NewRGBA(image.Rect(0, 0, int(mainImage.Description.Width), int(mainImage.Description.Height)))
	copy(imGray.Pix, ibGray.Bytes())
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		t.Log("Unable to save test image, missing environment variable VGE_TEST_DIR")
		return nil
	}
	fOut, err := os.Create(filepath.Join(testDir, "vk.png"))
	if err != nil {
		return err
	}
	defer fOut.Close()
	err = png.Encode(fOut, im)
	if err != nil {
		return err
	}
	fOut2, err := os.Create(filepath.Join(testDir, "vk_gray.png"))
	if err != nil {
		return err
	}
	defer fOut2.Close()
	err = png.Encode(fOut2, imGray)
	if err != nil {
		return err
	}
	return nil
}

// go:embed testsh/testsh.frag.spv
// var testsh_frag []byte

const testsh_frag = `
#version 450

layout(location = 0) out vec4 outColor;
layout(location = 1) out vec4 outGray;
layout(location = 0) in vec2 i_uv;

layout(set = 0, binding = 0) uniform sampler2D tx_color[];

void main() {
    outColor = texture(tx_color[1], i_uv);
    float c = (outColor.r + outColor.g + outColor.b) / 3;
    outGray = vec4(c,c,c,1);
}
`

// go:embed testsh/testsh.vert.spv
// var testsh_vert []byte

const testsh_vert = `
#version 450

layout (location = 0) in vec2 i_position;
layout (location = 0) out vec2 o_uv;

layout(set = 1, binding = 0) uniform WorldUBF {
    mat4 world;
} World;

void main() {
    o_uv = i_position;
    gl_Position = World.world * vec4(i_position, 0.0, 1.0);
}
`

func (tp *testPipeline) build(dev *Device) error {
	tp.l1 = dev.NewDynamicDescriptorLayout(DESCRIPTORTypeCombinedImageSampler, SHADERStageFragmentBit,
		8, DESCRIPTORBindingPartiallyBoundBitExt|DESCRIPTORBindingUpdateUnusedWhilePendingBitExt)
	tp.l2 = dev.NewDescriptorLayout(DESCRIPTORTypeUniformBufferDynamic, SHADERStageVertexBit, 1)
	tp.dp1 = NewDescriptorPool(tp.l1, 1)
	dev.Get(NewKey(), func() interface{} {
		return tp.dp1
	})
	tp.dp2 = NewDescriptorPool(tp.l2, 1)
	dev.Get(NewKey(), func() interface{} {
		return tp.dp2
	})
	tp.ds1 = tp.dp1.Alloc()
	tp.ds2 = tp.dp2.Alloc()
	tp.s = dev.NewSampler(SAMPLERAddressModeRepeat)
	tp.ds1.WriteImage(0, 1, tp.testImage.DefaultView(), tp.s)
	tp.ds2.WriteBuffer(0, 0, tp.ubWorld)
	tp.pl = NewGraphicsPipeline(dev)
	tp.pl.AddLayout(tp.l1)
	tp.pl.AddLayout(tp.l2)
	tp.pl.AddVextexInput(VERTEXInputRateVertex, FORMATR32g32Sfloat)

	comp := NewCompiler(dev)
	defer comp.Dispose()
	spir_frag, _, err := comp.Compile(SHADERStageFragmentBit, testsh_frag)
	if err != nil {
		return err
	}
	spir_vert, _, err := comp.Compile(SHADERStageVertexBit, testsh_vert)
	if err != nil {
		return err
	}

	tp.pl.AddShader(SHADERStageFragmentBit, spir_frag)
	tp.pl.AddShader(SHADERStageVertexBit, spir_vert)
	tp.pl.Create(tp.rp)
	return nil
}

func (tp *testPipeline) copyImage(dev *Device, b *Buffer, img *Image) {
	cmd := NewCommand(dev, QUEUETransferBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	r := img.FullRange()
	cmd.SetLayout(img, &r, IMAGELayoutTransferDstOptimal)
	cmd.CopyBufferToImage(img, b, &r)
	cmd.SetLayout(img, &r, IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
}
