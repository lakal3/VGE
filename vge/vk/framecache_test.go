package vk

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"io/ioutil"
	"math"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type fcTest struct {
	owner    Owner
	dev      *Device
	mp       *MemoryPool
	testView *ImageView
	ld       ImageLoader
	fc       *FrameCache
	fp       *GeneralRenderPass
	layout   *DescriptorLayout
	sampler  *Sampler
	pl       *GraphicsPipeline
}

var fcTestKey = NewKey()

type fcTestImages struct {
	mp          *MemoryPool
	frameImages []*Image
}

func (f *fcTestImages) Dispose() {
	if f.mp != nil {
		f.mp.Dispose()
	}
}

func (t *fcTest) Dispose() {
	t.owner.Dispose()
}

func TestNewFrameCache(t *testing.T) {
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
	var test fcTest
	test.dev = NewDevice(a, 0)
	if test.dev == nil {
		t.Error("Failed to initialize application")
	}
	test.owner.AddChild(test.dev)
	test.dev.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}
	test.ld = NewNativeImageLoader(a)

	err = test.testFrameCache(t)
	if err != nil {
		t.Error("Test render ", err)
	}

	test.Dispose()
	a.Dispose()
}

func (t *fcTest) testFrameCache(tt *testing.T) error {
	err := t.loadTestImage()
	if err != nil {
		return err
	}
	t.fc = NewFrameCache(t.dev, 3)
	t.owner.AddChild(t.fc)

	t.fp = NewGeneralRenderPass(t.dev, false, []AttachmentInfo{
		AttachmentInfo{FinalLayout: IMAGELayoutTransferSrcOptimal, Format: FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.1, 0.1, 0.1, 1}},
		AttachmentInfo{FinalLayout: IMAGELayoutTransferSrcOptimal, Format: FORMATR8g8b8a8Unorm, ClearColor: [4]float32{0.8, 0.8, 0.8, 1}},
	})
	if t.fp == nil {
		return nil
	}
	err = t.build()
	if err != nil {
		return err
	}
	t.owner.AddChild(t.fp)
	wg := &sync.WaitGroup{}
	for idx := 0; idx < len(t.fc.Instances); idx++ {
		wg.Add(1)
		go t.renderFrames(tt, wg, t.fc.Instances[idx])
		<-time.After(1 * time.Millisecond)
	}
	wg.Wait()
	return nil
}

func (t *fcTest) loadTestImage() error {
	testPattern, err := ioutil.ReadFile("../../assets/tests/uvchecker.png")
	if err != nil {
		return err
	}
	var testImageDesc ImageDescription
	err = t.ld.DescribeImage("png", &testImageDesc, testPattern)
	if err != nil {
		return fmt.Errorf("Describe image %v", err)
	}

	t.mp = NewMemoryPool(t.dev)
	t.owner.AddChild(t.mp)
	img := t.mp.ReserveImage(testImageDesc,
		IMAGEUsageTransferDstBit|IMAGEUsageSampledBit)
	b := t.mp.ReserveBuffer(testImageDesc.ImageSize(), true, BUFFERUsageTransferSrcBit)
	t.mp.Allocate()
	err = t.ld.LoadImage("png", testPattern, b)
	if err != nil {
		return err
	}
	cmd := NewCommand(t.dev, QUEUETransferBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	r := img.FullRange()
	cmd.SetLayout(img, &r, IMAGELayoutTransferDstOptimal)
	cmd.CopyBufferToImage(img, b, &r)
	cmd.SetLayout(img, &r, IMAGELayoutShaderReadOnlyOptimal)
	cmd.Submit()
	cmd.Wait()
	t.testView = img.DefaultView()
	return nil
}

var fbDesc = ImageDescription{
	Width:     1024,
	Height:    768,
	Depth:     1,
	Format:    FORMATR8g8b8a8Unorm,
	Layers:    1,
	MipLevels: 1,
}

func (t *fcTest) renderFrames(tt *testing.T, wg *sync.WaitGroup, instance *FrameInstance) {
	defer func() {
		wg.Done()
	}()
	kFb := NewKeys(2)

	cmd := NewCommand(t.dev, QUEUEGraphicsBit, false)
	defer cmd.Dispose()
	for loop := 0; loop < 10; loop++ {
		last := loop == 9
		instance.BeginFrame()
		instance.ReserveSlice(BUFFERUsageVertexBufferBit, 3*2*4)
		instance.ReserveImage(kFb, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit, fbDesc, ImageRange{LayerCount: 1, LevelCount: 1})
		instance.ReserveImage(kFb+1, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit, fbDesc, ImageRange{LayerCount: 1, LevelCount: 1})
		for idx := uint32(0); idx < triangleCount; idx++ {
			instance.ReserveSlice(BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(mgl32.Mat4{})))
			instance.ReserveDescriptor(t.layout)
		}
		if last {
			instance.ReserveSlice(BUFFERUsageTransferDstBit, uint64(fbDesc.Width*fbDesc.Height*4))
			instance.ReserveSlice(BUFFERUsageTransferDstBit, uint64(fbDesc.Width*fbDesc.Height*4))
		}
		cmd.Begin()
		instance.Commit()
		vb := instance.AllocSlice(BUFFERUsageVertexBufferBit, 3*2*4)
		copy(vb.Bytes(), Float32ToBytes(edges))
		mainImage, mainViews := instance.AllocImage(kFb)
		grayImage, grayViews := instance.AllocImage(kFb + 1)
		fb := NewFramebuffer2(t.fp, mainViews[0], grayViews[0])
		defer fb.Dispose()
		cmd.BeginRenderPass(t.fp, fb)
		drawList := &DrawList{}
		for idx := uint32(0); idx < triangleCount; idx++ {
			ubMatrix := instance.AllocSlice(BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(mgl32.Mat4{})))
			mat := mgl32.HomogRotate3DZ(float32(idx) * math.Pi * 2 / triangleCount)
			copy(ubMatrix.Bytes(), Float32ToBytes(mat[:]))
			dsUbf := instance.AllocDescriptor(t.layout)
			dsUbf.WriteSlice(0, 0, ubMatrix)
			dsUbf.WriteImage(1, 1, t.testView, t.sampler)
			drawList.Draw(t.pl, 0, 3).AddInput(0, vb).AddDescriptor(0, dsUbf)
		}
		cmd.Draw(drawList)
		cmd.EndRenderPass()
		var im, imGray *ASlice
		if last {
			im = instance.AllocSlice(BUFFERUsageTransferDstBit, uint64(fbDesc.Width*fbDesc.Height*4))
			imGray = instance.AllocSlice(BUFFERUsageTransferDstBit, uint64(fbDesc.Width*fbDesc.Height*4))
			tl := TransferList{}
			tl.CopyTo(im, mainImage, IMAGELayoutTransferSrcOptimal, IMAGELayoutTransferSrcOptimal, 0, 0)
			tl.CopyTo(imGray, grayImage, IMAGELayoutTransferSrcOptimal, IMAGELayoutTransferSrcOptimal, 0, 0)
			cmd.Transfer(tl)
		}
		instance.Freeze()
		cmd.Submit()
		cmd.Wait()
		if last {
			savePng(tt, "vkframe.png", im.Bytes(), fbDesc)
			savePng(tt, "vkframe_gray.png", imGray.Bytes(), fbDesc)
		}
	}

}

func (t *fcTest) build() error {
	la := t.dev.NewDescriptorLayout(DESCRIPTORTypeUniformBuffer, SHADERStageVertexBit, 1)
	t.layout = la.AddDynamicBinding(DESCRIPTORTypeCombinedImageSampler, SHADERStageFragmentBit,
		8, DESCRIPTORBindingPartiallyBoundBitExt|DESCRIPTORBindingUpdateUnusedWhilePendingBitExt)
	// tp.l2 = dev.NewDescriptorLayout(DESCRIPTORTypeUniformBufferDynamic, SHADERStageVertexBit, 1)
	t.sampler = t.dev.NewSampler(SAMPLERAddressModeRepeat)
	t.owner.AddChild(t.sampler)
	t.pl = NewGraphicsPipeline(t.dev)
	t.pl.AddLayout(t.layout)
	t.pl.AddVextexInput(VERTEXInputRateVertex, FORMATR32g32Sfloat)
	comp := NewCompiler(t.dev)
	defer comp.Dispose()
	spir_frag, _, err := comp.Compile(SHADERStageFragmentBit, testsh_frag_ubf)
	if err != nil {
		return err
	}
	spir_vert, _, err := comp.Compile(SHADERStageVertexBit, testsh_vert_ubf)
	if err != nil {
		return err
	}

	t.pl.AddShader(SHADERStageFragmentBit, spir_frag)
	t.pl.AddShader(SHADERStageVertexBit, spir_vert)
	t.pl.Create(t.fp)
	return nil
}
