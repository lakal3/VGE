package main

import "github.com/lakal3/vge/vge/vk"

type Void struct {
}

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

type hDescriptorLayout hDisposable

type hDescriptorPool hDisposable

type hDescriptorSet hDisposable

type hRenderPass hDisposable

type hFramebuffer hDisposable

type hImageLoader hDisposable

type hPipeline hDisposable

type hGraphicsPipeline hPipeline

type hComputePipeline hPipeline

type hSampler hDisposable

type hDesktop uintptr

type hWindow hDisposable

type hSubmitInfo uintptr

type MainLib interface {
	Exception_GetError(struct {
		ex     hException
		msg    string
		msgLen *int32
	}) Void
	NewApplication(struct {
		name string
		app  *hApplication
	})
	AddValidation(struct{ app hApplication })
	AddDynamicDescriptors(struct {
		app hApplication
	})
	NewDesktop(struct {
		app     hApplication
		desktop *hDesktop
	})
	DebugPoint(struct{ point string }) Void
	Desktop_CreateWindow(struct {
		desktop hDesktop
		title   string
		pos     *vk.WindowPos
		win     *hWindow
	})
	Desktop_PullEvent(struct {
		desktop hDesktop
		ev      *vk.RawEvent
	})
	Desktop_GetKeyName(struct {
		desktop hDesktop
		keyCode uint32
		name    []byte
		strLen  *uint32
	})
	Desktop_GetMonitor(struct {
		desktop hDesktop
		monitor uint32
		info    *vk.WindowPos
	})
	Application_Init(struct {
		app  hApplication
		inst *hInstance
	})
	Disposable_Dispose(struct{ disp hDisposable }) Void
	Instance_GetPhysicalDevice(struct {
		instance hInstance
		index    int32
		info     *vk.DeviceInfo
	})
	Instance_NewDevice(struct {
		instance hInstance
		index    int32
		pd       *hDevice
	})

	Device_NewBuffer(struct {
		dev        hDevice
		size       uint64
		hostMemory bool
		usage      vk.BufferUsageFlags
		buffer     *hBuffer
	})
	Device_NewImage(struct {
		dev   hDevice
		usage vk.ImageUsageFlags
		desc  *vk.ImageDescription
		image *hImage
	})
	Device_NewCommand(struct {
		dev       hDevice
		queueType vk.QueueFlags
		once      bool
		command   *hCommand
	})
	Device_NewMemoryBlock(struct {
		dev      hDevice
		memBlock *hMemoryBlock
	})
	Device_NewDescriptorLayout(struct {
		dev            hDevice
		descriptorType vk.DescriptorType
		stages         vk.ShaderStageFlags
		element        uint32
		flags          vk.DescriptorBindingFlagBitsEXT
		prevLayout     hDescriptorLayout
		dsLayout       *hDescriptorLayout
	})
	MemoryBlock_Reserve(struct {
		memBlock  hMemoryBlock
		memObject hMemoryObject
		suitable  *bool
	})
	MemoryBlock_Allocate(struct {
		memBlock hMemoryBlock
	})
	Buffer_GetPtr(struct {
		buffer hBuffer
		ptr    *uintptr
	})
	Buffer_NewView(struct {
		buffer hBuffer
		format vk.Format
		offset uint64
		size   uint64
		view   *hBufferView
	})
	Command_Begin(struct{ cmd hCommand })
	Command_Wait(struct {
		cmd hCommand
	})
	Command_CopyBuffer(struct {
		cmd hCommand
		src hBuffer
		dst hBuffer
	})
	Command_BeginRenderPass(struct {
		cmd hCommand
		rp  hRenderPass
		fb  hFramebuffer
	})
	Command_EndRenderPass(struct{ cmd hCommand })
	Command_SetLayout(struct {
		cmd       hCommand
		image     hImage
		imRange   *vk.ImageRange
		newLayout vk.ImageLayout
	})
	Command_CopyBufferToImage(struct {
		cmd     hCommand
		src     hBuffer
		dst     hImage
		imRange *vk.ImageRange
		offset  uint64
	})
	Command_CopyImageToBuffer(struct {
		cmd     hCommand
		src     hImage
		dst     hBuffer
		imRange *vk.ImageRange
		offset  uint64
	})
	Command_Draw(struct {
		cmd   hCommand
		draws []vk.DrawItem
	})
	Device_Submit(struct {
		dev       hDevice
		cmd       hCommand
		priority  uint32
		info      []hSubmitInfo
		waitStage vk.PipelineStageFlags
		waitInfo  *hSubmitInfo
	})

	DescriptorLayout_NewPool(struct {
		layout hDescriptorLayout
		size   uint32
		pool   *hDescriptorPool
	})
	DescriptorPool_Alloc(struct {
		pool hDescriptorPool
		ds   *hDescriptorSet
	})
	DescriptorSet_WriteBuffer(struct {
		ds      hDescriptorSet
		binding uint32
		at      uint32
		buffer  hBuffer
		from    uint64
		size    uint64
	})
	DescriptorSet_WriteBufferView(struct {
		ds         hDescriptorSet
		binding    uint32
		at         uint32
		bufferView hBufferView
	})
	DescriptorSet_WriteImage(struct {
		ds      hDescriptorSet
		binding uint32
		at      uint32
		view    hImageView
		sampler hSampler
	})

	Image_NewView(struct {
		image     hImage
		imRange   *vk.ImageRange
		imageView *hImageView
		cube      bool
	})
	NewForwardRenderPass(struct {
		dev              hDevice
		finalLayout      vk.ImageLayout
		mainImageFormat  vk.Format
		depthImageFormat vk.Format
		rp               *hRenderPass
	})
	NewDepthRenderPass(struct {
		dev              hDevice
		finalLayout      vk.ImageLayout
		depthImageFormat vk.Format
		rp               *hRenderPass
	})
	RenderPass_NewFrameBuffer(struct {
		rp          hRenderPass
		attachments []hImageView
		fb          *hFramebuffer
	})
	RenderPass_Init(struct{ rp hRenderPass })
	Device_NewSampler(struct {
		dev        hDevice
		repeatMode vk.SamplerAddressMode
		sampler    *hSampler
	})
	Device_NewGraphicsPipeline(struct {
		dev hDevice
		gp  *hGraphicsPipeline
	})
	Device_NewComputePipeline(struct {
		dev hDevice
		cp  *hComputePipeline
	})
	ComputePipeline_Create(struct {
		cp hComputePipeline
	})
	Pipeline_AddShader(struct {
		pl    hPipeline
		stage vk.ShaderStageFlags
		code  []byte
	})
	Pipeline_AddDescriptorLayout(struct {
		pl       hPipeline
		dsLayout hDescriptorLayout
	})
	GraphicsPipeline_Create(struct {
		pipeline   hGraphicsPipeline
		renderPass hRenderPass
	})
	GraphicsPipeline_AddVertexBinding(struct {
		pl     hGraphicsPipeline
		stride uint32
		rate   vk.VertexInputRate
	})
	GraphicsPipeline_AddVertexFormat(struct {
		pl     hGraphicsPipeline
		format vk.Format
		offset uint32
	})
	GraphicsPipeline_AddDepth(struct {
		pl    hGraphicsPipeline
		write bool
		check bool
	})
	GraphicsPipeline_AddAlphaBlend(struct {
		pl hGraphicsPipeline
	})
	NewImageLoader(struct{ loader *hImageLoader })
	ImageLoader_Supported(struct {
		loader hImageLoader
		kind   string
		read   *bool
		write  *bool
	})
	ImageLoader_Describe(struct {
		loader  hImageLoader
		kind    string
		desc    *vk.ImageDescription
		content []byte
	})
	ImageLoader_Load(struct {
		loader  hImageLoader
		kind    string
		content []byte
		buf     hBuffer
	})
	ImageLoader_Save(struct {
		loader  hImageLoader
		kind    string
		desc    *vk.ImageDescription
		buf     hBuffer
		content []byte
		reqSize *uint64
	})
	//void PrepareSwapchain(Device *dev, ImageDescription* viewDesc, int32_t &imageCount);
	//		void GetNextFrame(Image*& image, SubmitInfo *&submitInfo, int32_t& viewIndex);
	Window_PrepareSwapchain(struct {
		win        hWindow
		dev        hDevice
		imageDesc  *vk.ImageDescription
		imageCount *int32
	})
	Window_GetNextFrame(struct {
		win        hWindow
		image      *hImage
		submitInfo *hSubmitInfo
		viewIndex  *int32
	})
	Window_GetPos(struct {
		win hWindow
		pos *vk.WindowPos
	})
	Window_SetPos(struct {
		win hWindow
		pos *vk.WindowPos
	})

	Command_Compute(struct {
		hCmd        hCommand
		hPl         hComputePipeline
		x           uint32
		y           uint32
		z           uint32
		descriptors []hDescriptorSet
	})
}
