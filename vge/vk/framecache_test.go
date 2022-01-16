package vk

import (
	"errors"
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

var tcInvokeCount = 0

func newFcTestImages(dev *Device, inst *FrameInstance) *fcTestImages {
	tcInvokeCount++
	if tcInvokeCount > 1 {
		dev.FatalError(errors.New("Shared allocation failed"))
	}
	_, total := inst.Index()
	ti := &fcTestImages{}
	ti.mp = NewMemoryPool(dev)
	for idx := 0; idx < total*2; idx++ {
		ti.frameImages = append(ti.frameImages, ti.mp.ReserveImage(ImageDescription{
			Width:     1024,
			Height:    768,
			Depth:     1,
			Format:    FORMATR8g8b8a8Unorm,
			Layers:    1,
			MipLevels: 1,
		}, IMAGEUsageColorAttachmentBit|IMAGEUsageTransferSrcBit))
	}
	ti.mp.Allocate()
	return ti
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

	err = test.testFrameCache()
	if err != nil {
		t.Error("Test render ", err)
	}

	test.Dispose()
	a.Dispose()
}

func (t *fcTest) testFrameCache() error {
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
		go t.renderFrames(wg, t.fc.Instances[idx])
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

func (t *fcTest) renderFrames(wg *sync.WaitGroup, instance *FrameInstance) {
	defer func() {
		wg.Done()
	}()
	fiIndex, _ := instance.Index()
	cmd := NewCommand(t.dev, QUEUEGraphicsBit, false)
	defer cmd.Dispose()
	ti := instance.GetShared(fcTestKey, func() interface{} {
		return newFcTestImages(t.dev, instance)
	}).(*fcTestImages)
	for loop := 0; loop < 10; loop++ {
		instance.BeginFrame()
		instance.ReserveSlice(BUFFERUsageVertexBufferBit, 3*2*4)
		for idx := uint32(0); idx < triangleCount; idx++ {
			instance.ReserveSlice(BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(mgl32.Mat4{})))
			instance.ReserveDescriptor(t.layout)
		}
		cmd.Begin()
		instance.Commit()
		vb := instance.AllocSlice(BUFFERUsageVertexBufferBit, 3*2*4)
		copy(vb.Bytes(), Float32ToBytes(edges))
		mainView := ti.frameImages[fiIndex*2].DefaultView()
		grayView := ti.frameImages[fiIndex*2+1].DefaultView()
		fb := NewFramebuffer(t.fp, []*ImageView{mainView, grayView})
		defer fb.Dispose()
		cmd.BeginRenderPass(t.fp, fb)
		drawList := &DrawList{}
		for idx := uint32(0); idx < triangleCount; idx++ {
			ubMatrix := instance.AllocSlice(BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(mgl32.Mat4{})))
			mat := mgl32.HomogRotate3DZ(float32(idx) * math.Pi * 2 / triangleCount)
			copy(ubMatrix.Bytes(), Float32ToBytes(mat[:]))
			dsUbf := instance.AllocDescriptor(t.layout)
			dsUbf.WriteDSSlice(0, 0, ubMatrix)
			dsUbf.WriteImage(1, 0, t.testView, t.sampler)
			drawList.Draw(t.pl, 0, 3).AddInput(0, vb).AddDescriptor(0, dsUbf)
		}
		cmd.Draw(drawList)
		cmd.EndRenderPass()
		instance.Freeze()
		cmd.Submit()
		cmd.Wait()
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
