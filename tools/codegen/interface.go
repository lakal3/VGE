package main

import "github.com/lakal3/vge/vge/vk"

type Void struct {
}

type hDisposable uintptr

type hApplication hDisposable

type hInstance hDisposable

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

type hQueryPool uintptr

type hGlslCompiler uintptr

type hAllocator hDisposable

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
	AddValidationException(struct{ msgId int32 })
	AddDynamicDescriptors(struct {
		app hApplication
	})
	NewDesktop(struct {
		app        hApplication
		imageUsage vk.ImageUsageFlags
		desktop    *hDesktop
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
		rawBuffer  *uintptr
	})
	Device_NewImage(struct {
		dev      hDevice
		usage    vk.ImageUsageFlags
		desc     *vk.ImageDescription
		image    *hImage
		rawImage *uintptr
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
	Device_NewTimestampQuery(struct {
		dev  hDevice
		size uint32
		qp   *hQueryPool
	})

	Device_NewGlslCompiler(struct {
		dev  hDevice
		comp *hGlslCompiler
	})
	Device_NewAllocator(struct {
		dev       hDevice
		allocator *hAllocator
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
	Buffer_CopyFrom(struct {
		buffer hBuffer
		offset uint64
		ptr    uintptr
		size   uint64
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
	Command_ClearImage(struct {
		cmd     hCommand
		dst     hImage
		imRange *vk.ImageRange
		layout  vk.ImageLayout
		color   float32
		alpha   float32
	})
	Command_Draw(struct {
		cmd           hCommand
		draws         []vk.DrawItem
		pushConstants []byte
	})
	Command_Transfer(struct {
		cmd      hCommand
		transfer []vk.TransferItem
	})
	Command_WriteTimer(struct {
		cmd        hCommand
		qp         hQueryPool
		stages     vk.PipelineStageFlags
		timerIndex uint32
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
	DescriptorSet_WriteDSSlice(struct {
		ds      hDescriptorSet
		binding uint32
		at      uint32
		buffer  uintptr
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
	DescriptorSet_WriteDSImageView(struct {
		ds      hDescriptorSet
		binding uint32
		at      uint32
		view    uintptr
		layout  vk.ImageLayout
		sampler hSampler
	})

	Image_NewView(struct {
		image     hImage
		imRange   *vk.ImageRange
		imageView *hImageView
		rawView   *uintptr
	})
	/*
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
	*/
	NewRenderPass(struct {
		dev             hDevice
		rp              *hRenderPass
		depthAttachment bool
		attachments     []vk.AttachmentInfo
	})

	RenderPass_NewFrameBuffer(struct {
		rp          hRenderPass
		attachments []hImageView
		fb          *hFramebuffer
	})
	RenderPass_NewNullFrameBuffer(struct {
		rp     hRenderPass
		width  uint32
		height uint32
		fb     *hFramebuffer
	})
	RenderPass_NewFrameBuffer2(struct {
		rp          hRenderPass
		width       uint32
		height      uint32
		attachments []uintptr
		fb          *hFramebuffer
	})
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
	Pipeline_AddPushConstants(struct {
		pl     hPipeline
		size   uint32
		stages vk.ShaderStageFlags
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
	GraphicsPipeline_SetTopology(struct {
		pl       hGraphicsPipeline
		topology vk.PrimitiveTopology
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
	Desktop_SetClipboard(struct {
		desktop hDesktop
		text    []byte
	})
	Desktop_GetClipboard(struct {
		desktop hDesktop
		textLen *uint64
		text    []byte
	})

	Command_Compute(struct {
		hCmd        hCommand
		hPl         hComputePipeline
		x           uint32
		y           uint32
		z           uint32
		descriptors []hDescriptorSet
	})

	QueryPool_Get(struct {
		qp              hQueryPool
		values          []uint64
		timestampPeriod *float32
	})

	GlslCompiler_Compile(struct {
		comp     hGlslCompiler
		stage    vk.ShaderStageFlags
		src      []byte
		instance *uintptr
	})

	GlslCompiler_GetOutput(struct {
		comp     hGlslCompiler
		instance uintptr
		msg      *uintptr
		msg_len  *uint64
		result   *uint32 // 0 - OK, 1 - warnings, 10 - failed
	})

	GlslCompiler_GetSpirv(struct {
		comp      hGlslCompiler
		instance  uintptr
		spirv     *uintptr
		spirv_len *uint64
	})

	GlslCompiler_Free(struct {
		comp     hGlslCompiler
		instance uintptr
	})

	Allocator_AllocBuffer(struct {
		allocator hAllocator
		usage     vk.BufferUsageFlags
		size      uint64
		hBuffer   *uintptr
		memType   *uint32
		alignment *uint32
	})

	Allocator_AllocDeviceBuffer(struct {
		allocator hAllocator
		usage     vk.BufferUsageFlags
		size      uint64
		hBuffer   *uintptr
		memType   *uint32
		alignment *uint32
	})

	Allocator_AllocMemory(struct {
		allocator  hAllocator
		size       uint64
		memType    uint32
		hostMemory bool
		hMem       *uintptr
		memPtr     *uintptr
	})

	Allocator_AllocImage(struct {
		allocator hAllocator
		usage     vk.ImageUsageFlags
		im        *vk.ImageDescription
		hImage    *uintptr
		size      *uint64
		memType   *uint32
		alignment *uint32
	})

	Allocator_AllocView(struct {
		allocator hAllocator
		hImage    uintptr
		rg        *vk.ImageRange
		im        *vk.ImageDescription
		hView     *uintptr
	})

	Allocator_FreeBuffer(struct {
		allocator hAllocator
		hBuffer   uintptr
	})

	Allocator_FreeMemory(struct {
		allocator  hAllocator
		hMem       uintptr
		hostMemory bool
	})

	Allocator_FreeImage(struct {
		allocator hAllocator
		hImage    uintptr
	})

	Allocator_FreeView(struct {
		allocator hAllocator
		hView     uintptr
	})

	Allocator_BindBuffer(struct {
		allocator hAllocator
		hMem      uintptr
		hBuffer   uintptr
		offset    uint64
	})

	Allocator_BindImage(struct {
		allocator hAllocator
		hMem      uintptr
		hBuffer   uintptr
		offset    uint64
	})
}
