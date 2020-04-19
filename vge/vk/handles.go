package vk

type hDisposable uintptr

type hApplication hDisposable

type hInstance hDisposable

type hPhysicalDevice hDisposable

type hDevice hDisposable

type hMemoryBlock hDisposable

type hMemoryObject hDisposable

type hBuffer hMemoryObject

type hBufferView hDisposable

type hCommand hDisposable

type hImage hMemoryObject

type hImageView hDisposable

type hException hDisposable

type hValidation hDisposable

type hDescriptorLayout hDisposable

type hDescriptorPool hDisposable

type hDescriptorSet hDisposable

type hImageLoader hDisposable

type hRenderPass hDisposable

type hFramebuffer hDisposable

type hPipeline hDisposable

type hGraphicsPipeline hPipeline

type hComputePipeline hPipeline

type hSampler hDisposable

type hDesktop uintptr

type hWindow hDisposable

type hSubmitInfo uintptr

type hFontLoader hDisposable

type ImageDescription struct {
	Width     uint32
	Height    uint32
	Depth     uint32
	Format    Format
	Layers    uint32
	MipLevels uint32
}

type ImageRange struct {
	FirstLayer    uint32
	LayerCount    uint32
	FirstMipLevel uint32
	LevelCount    uint32
	Layout        ImageLayout
}

type descriptorInfo struct {
	hSet      hDescriptorSet
	hasOffset uint32
	offset    uint32
}

type DrawItem struct {
	pipeline     hPipeline
	inputs       [8]hBuffer
	descriptors  [8]descriptorInfo
	from         uint32
	count        uint32
	instances    uint32
	fromInstance uint32
	indexed      uint32
	_padding     uint32
}

type RawEvent struct {
	// See vapp/win.go for list of raw event codes
	EventType uint32
	Arg1      int32
	Arg2      int32
	Arg3      int32
	hWin      hWindow
}

type CharInfo struct {
	Width   uint32
	Height  uint32
	Offsetx int32
	Offsety int32
	extra   uint64
}

type DeviceInfo struct {
	// Is device valid for given application options.
	// 0 - Valid
	// 1 - Invalid
	// 2 - No such device
	Valid               uint32
	DeviceKind          PhysicalDeviceType
	MemorySize          uint64
	MaxSamplersPerStage uint32
	MaxImageArrayLayers uint32
	NameLen             uint32
	ReasonLen           uint32
	Name                [256]byte
	Reason              [256]byte
}

type WindowState uint32

const (
	WINDOWStateNormal    = WindowState(0)
	WINDOWStateHidden    = WindowState(1)
	WINDOWStateMinimized = WindowState(2)
	WINDOWStateMaximized = WindowState(3)
	WINDOWStateIconic    = WindowState(4)
	// No resize
	WINDOWStateFixed = WindowState(16)
	// No borders, title etc.
	WINDOWStateBorderless = WindowState(32)
)

type WindowPos struct {
	Left    int32
	Top     int32
	Width   int32
	Height  int32
	State   WindowState
	Monitor uint32
}

func (di *DrawItem) SetInstances(from uint32, count uint32) *DrawItem {
	di.fromInstance, di.instances = from, count
	return di
}

func div2(v uint64) uint64 {
	if v > 1 {
		return v >> 1
	}
	return 1
}
