#pragma once

namespace vge {


	extern vk::Optional<const vk::AllocationCallbacks> allocator;
	

	class External {
	public:
		virtual ~External();
	};

	class Disposable : public External {
	public:
		virtual void Dispose() = 0;
	};

	class Exception : Disposable {
	public:
		Exception(const std::exception& ex) {
			msg = ex.what();
		}
		Exception(const std::string msg_) {
			msg = msg_;
		}
		void GetError(char* msg, size_t msg_len, int32_t& msgLen);
		static Exception* getValidationError();
	private:
		virtual void Dispose() override {
			delete this;
		}
		std::string msg;
	};




	struct InstanceExtension {
		virtual void prepare(vk::InstanceCreateInfo& ici, std::vector<const char *> &layers, std::vector<const char*>& extensions) = 0;
		virtual void attach(Instance* inst) = 0;
		virtual void detach(Instance* inst) = 0;
		virtual void dispose() = 0;
		static bool addLayer(std::vector<const char*>& layers, std::string_view layerName);
		static bool addExtension(std::vector<const char*>& extensions, std::string_view extName);
	};

	struct DeviceExtension  {
		virtual void isValid(Instance* inst, vk::PhysicalDevice pd, bool& valid, std::string& invalidReason) = 0;
		virtual void prepare(vk::DeviceCreateInfo &dci, std::vector<const char*>& extensions) = 0;
		virtual void attach(Device* inst) = 0;
		virtual void detach(Device* inst) = 0;
		virtual void dispose() = 0;
		static bool checkExtension(Instance *inst, vk::PhysicalDevice pd, std::string_view extName);
	};

	class Application : public Disposable {
		friend class Instance;
		friend struct Static;
	public:
		void Init(Instance*& instance);
		virtual void Dispose() override;


		const char* getName() {
			return name.c_str();
		}

		void addInstanceExtension(InstanceExtension* ext) {
			instanceExtensions.push_back(ext);
		}

		const std::vector< InstanceExtension*> &get_instanceExtensions() const {
			return instanceExtensions;
		}

		void addDeviceExtension(DeviceExtension* ext) {
			deviceExtensions.push_back(ext);
		}

		const std::vector< DeviceExtension*> &get_deviceExtensions() const {
			return deviceExtensions;
		}

	private:
		std::vector<InstanceExtension*> instanceExtensions;
		std::vector<DeviceExtension*> deviceExtensions;
		std::string name;
		std::unique_ptr<Instance> instance;
	};

	

	class Instance : public External {
		friend class Application;
	public:
		~Instance();
		Instance(Application* app_) : app(app_) {

		}

		void GetPhysicalDevice(int32_t index, DeviceInfo *info);

		void NewDevice(int32_t pdIndex, Device*& pd);

		vk::Instance get_instance() {
			return instance;
		}
		vk::DispatchLoaderDynamic& get_dispatch() {
			return dispatchLoader;
		}

		Application* get_app() {
			return app;
		}

		std::vector<vk::PhysicalDevice> pds;
	private:
		
		void init();
		vk::Instance instance = nullptr;
		vk::DispatchLoaderDynamic dispatchLoader;
		Application* app;
	};

	
	class Device: public Disposable {
		friend class Instance;
	public:
		
		void NewBuffer(uint64_t size, bool hostMemory, vk::BufferUsageFlags usage, Buffer*& buffer);
		void NewImage(vk::ImageUsageFlags usage, const ImageDescription *imageDescription, Image*& image);
		void NewCommand(vk::QueueFlags queueType, bool once, Command*& command);
		void NewMemoryBlock(MemoryBlock*& memBlock);
		void NewDescriptorLayout(vk::DescriptorType descType, vk::ShaderStageFlags stages, uint32_t size, vk::DescriptorBindingFlagsEXT flags, 
			DescriptorLayout* prevBinding, DescriptorLayout*& descriptorLayout);
		void NewGraphicsPipeline(GraphicsPipeline*& gp);
		void NewComputePipeline(ComputePipeline*& cp);
		void NewSampler(vk::SamplerAddressMode mode, Sampler*& sampler);
		void Submit(Command* command, uint32_t priority, SubmitInfo **info, size_t info_len, vk::PipelineStageFlags waitForStage, SubmitInfo*& waitFor);

		const vk::DispatchLoaderDynamic& get_dispatch() const {
			return _dispatchLoader;
		}

		vk::Device get_device() const {
			return _dev;
		}

		vk::PhysicalDevice get_pd() const {
			return _pd;
		}

		Instance* get_instance() const {
			return _inst;
		}

		const vk::PhysicalDeviceMemoryProperties &get_memoryProps() const {
			return _memProps;
		}
		uint32_t get_graphicQueueFamily() {
			return _graphicsQueueIdx;
		}
	protected:
		virtual void Dispose() override;
		
	private:
		Device(Instance* inst, vk::PhysicalDevice pd) : _pd(pd), _inst(inst) {
			_graphicsQueueIdx = 0xffffffff;
		}
		~Device();
		void init();

		uint32_t _graphicsQueueIdx;
		std::vector<Queue*> queues;
		vk::PhysicalDeviceMemoryProperties _memProps;
		vk::PhysicalDeviceProperties _properties;
		vk::PhysicalDevice _pd;
		vk::Device _dev;
		vk::DispatchLoaderDynamic _dispatchLoader;
		Instance* _inst;

	};

	using DebugCallback = void (*)(int32_t severity, const char* msg);

	class ValidationOption : public InstanceExtension {
		friend struct Static;
	private:
		ValidationOption(DebugCallback callback_, Application* app_) : callback(callback_), app(app_) {
			app->addInstanceExtension(this);
		}
		ValidationOption(Application* app_) : callback(nullptr), app(app_) {
			app->addInstanceExtension(this);
		}

		DebugCallback callback;
		Application* app;
		bool active = false;
		vk::DebugUtilsMessengerEXT messanger = nullptr;

		virtual void dispose() override;
		virtual void prepare(vk::InstanceCreateInfo& ici, std::vector<const char*>& layers, std::vector<const char*>& extensions) override;
		virtual void attach(Instance* inst) override;
		virtual void detach(Instance* inst) override;

	};

	struct Static {
		static void NewApplication(char* name, size_t name_len, Application*& app);
		static void NewDesktop(Application* app, Desktop*& desktop);
		static void AddValidation(Application *app);
		static void NewForwardRenderPass(Device* dev, vk::ImageLayout finalLayout, vk::Format mainImageFormat, vk::Format depthImageFormat, RenderPass*& rp);
		static void NewDepthRenderPass(Device* dev, vk::ImageLayout finalLayout, vk::Format depthImageFormat, RenderPass*& rp);
		static void NewImageLoader(ImageLoader*& loader);
		static void DebugPoint(const char* point, size_t len_point);
		static void AddDynamicDescriptors(Application* app);
	};

	class SuppressValidation {
	public:
		SuppressValidation();
		~SuppressValidation();
	};
}