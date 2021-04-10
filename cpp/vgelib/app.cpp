
#include "vgelib/vgelib.hpp"
#include <algorithm>
#include <iostream>

vk::Optional<const vk::AllocationCallbacks> vge::allocator(nullptr);

static thread_local vge::Exception* validationError = nullptr;
static thread_local int suppressValidation = 0;

void vge::Application::Init(Instance*& ret)
{
	
	instance = std::make_unique<Instance>(this);
	instance->init();
	ret = instance.get();
}

void vge::Application::Dispose()
{
	delete this;
}

vge::External::~External()
{
}

vge::Instance::~Instance()
{
	if (!!instance) {
		for (auto iext : app->get_instanceExtensions()) {
			iext->detach(this);
		}
		instance.destroy(allocator, dispatchLoader);
		instance = nullptr;
	}
}

inline uint32_t minSize(size_t s1, size_t s2) {
	return static_cast<uint32_t>(s1 < s2 ? s1 : s2);
}

void vge::Instance::GetPhysicalDevice(int32_t index, DeviceInfo *info)
{
	if (index >= pds.size()) {
		info->valid = 2;
		return;
	}
	auto pd = pds[index];
	auto props = pd.getProperties(get_dispatch());
	info->deviceKind = static_cast<uint32_t>(props.deviceType);
	info->nameLen = minSize(256, strlen(props.deviceName));
	strncpy(info->name, props.deviceName, info->nameLen);
	auto mps = pd.getMemoryProperties(get_dispatch());
	for (uint32_t i = 0; i < mps.memoryHeapCount; i++) {
		if (mps.memoryHeaps[i].flags == vk::MemoryHeapFlagBits::eDeviceLocal) {
			info->memorySize += mps.memoryHeaps[i].size;
		}			
	}
	info->maxSamplersPerStage = props.limits.maxPerStageDescriptorSamplers;
	info->maxImageArrayLayers = props.limits.maxImageArrayLayers;
	for (auto iext : app->get_deviceExtensions()) {
		bool isValid;
		std::string invalidReson;
		iext->isValid(this, pd, isValid, invalidReson);
		if (!isValid) {
			info->valid = 1;
			info->reasonLen = minSize(256, invalidReson.size());
			strncpy(info->reason, invalidReson.c_str(), info->reasonLen);
			return;
		} else {
			info->valid = 0;
			info->reasonLen = 0;
		}
	}
}

void vge::Instance::NewDevice(int32_t pdIndex, Device*& ret)
{
	auto dev = new Device(this, pds[pdIndex]);
	try {
		dev->init();
	} catch (...) {
		delete dev;
		throw;
	}
	ret = dev;
}


void vge::Instance::init()
{
	auto apInfo = vk::ApplicationInfo(app->getName());
	apInfo.apiVersion = VK_MAKE_VERSION(1, 1, 0);
	vk::InstanceCreateInfo ici;
	ici.pApplicationInfo = &apInfo;
	std::vector<const char*> layers;
	std::vector<const char*> extensions;
	for (auto iext : app->get_instanceExtensions()) {
		iext->prepare(ici, layers, extensions);
	}
	if (layers.size() > 0) {
		ici.ppEnabledLayerNames = layers.data();
		ici.enabledLayerCount = static_cast<uint32_t>(layers.size());
	}
	if (extensions.size() > 0) {
		ici.ppEnabledExtensionNames = extensions.data();
		ici.enabledExtensionCount = static_cast<uint32_t>(extensions.size());
	}
	instance = vk::createInstance(ici, allocator);
	dispatchLoader.init(instance, &vkGetInstanceProcAddr);
	pds = instance.enumeratePhysicalDevices(dispatchLoader);
	for (auto iext : app->get_instanceExtensions()) {
		iext->attach(this);
	}
}

bool vge::InstanceExtension::addLayer(std::vector<const char*>& layers, std::string_view layerName)
{
	auto lp = vk::enumerateInstanceLayerProperties();
	for (auto layer : lp) {
		if (layerName == layer.layerName) {
			layers.push_back(layerName.data());
			return true;
		}
	}
	return false;
}

bool vge::InstanceExtension::addExtension(std::vector<const char*>& extensions, std::string_view extName)
{
	auto exList = vk::enumerateInstanceExtensionProperties();
	for (auto ext : exList) {
		if (extName == ext.extensionName) {
			extensions.push_back(extName.data());
			return true;
		}
	}
	
	return false;
}

bool vge::DeviceExtension::checkExtension(Instance *inst, vk::PhysicalDevice pd, std::string_view extName)
{
	auto exList = pd.enumerateDeviceExtensionProperties();
	for (auto ext : exList) {
		if (extName == ext.extensionName) {
			return true;
		}
	}

	return false;
}
void vge::ValidationOption::dispose()
{
	delete this;
}

void vge::ValidationOption::prepare(vk::InstanceCreateInfo& ici, std::vector<const char*>& layers, std::vector<const char*>& extensions)
{
	if (!InstanceExtension::addExtension(extensions, "VK_EXT_debug_utils") || !InstanceExtension::addExtension(extensions, "VK_EXT_debug_report")) {
		return;
	}
	if (!InstanceExtension::addLayer(layers, "VK_LAYER_LUNARG_standard_validation")) {
		return;
	}
	active = true;
}


class VulkanValidationError : public std::exception {
public:
	VulkanValidationError(const char* msg) : error("Vulkan validation error: ") {
		error.append(msg);
	}

#ifdef __GNUC__
	const char* what() const noexcept override  {
#else 		
	const char* what() const override {	
#endif		
		return error.c_str();
	}
private:
	std::string error;
};

VkBool32 debugCallback(vk::DebugUtilsMessageSeverityFlagBitsEXT messageSeverity,
	vk::DebugUtilsMessageTypeFlagsEXT                  messageTypes,
	const vk::DebugUtilsMessengerCallbackDataEXT* pCallbackData,
	void* pUserData) {
	auto cb = reinterpret_cast<vge::DebugCallback>(pUserData);
	if (cb == nullptr) {
		
		if (messageSeverity == vk::DebugUtilsMessageSeverityFlagBitsEXT::eError) {
			// std::cout << "Vulkan error: " << pCallbackData->pMessage << "\n";
			if (validationError == nullptr) {
				std::string error("Vulkan validation error: ");
				error.append(pCallbackData->pMessage);
				if (suppressValidation > 0) {
					return 0;
				}
				validationError = new vge::Exception(error);
			}
		}
		return 0;
	}
	cb(static_cast<int32_t>(messageSeverity), pCallbackData->pMessage);
	return 0;
}

void vge::ValidationOption::attach(Instance* inst)
{
	if (!active) {
		return;
	}
	auto dg = vk::DebugUtilsMessengerCreateInfoEXT();
	dg.messageSeverity = vk::DebugUtilsMessageSeverityFlagBitsEXT::eWarning | vk::DebugUtilsMessageSeverityFlagBitsEXT::eError;
	dg.pUserData = reinterpret_cast<void *>(callback);
	dg.pfnUserCallback = reinterpret_cast<PFN_vkDebugUtilsMessengerCallbackEXT> (&debugCallback);
	dg.messageType = vk::DebugUtilsMessageTypeFlagBitsEXT::eGeneral | vk::DebugUtilsMessageTypeFlagBitsEXT::ePerformance | vk::DebugUtilsMessageTypeFlagBitsEXT::eValidation;
	messanger = inst->get_instance().createDebugUtilsMessengerEXT(dg, allocator, inst->get_dispatch());
}

void vge::ValidationOption::detach(Instance* inst) {
	if (!!messanger) {
		inst->get_instance().destroyDebugUtilsMessengerEXT(messanger, allocator, inst->get_dispatch());
		messanger = nullptr;
	}
}

void vge::Device::NewBuffer(uint64_t size, bool hostMemory, vk::BufferUsageFlags usage, Buffer*& ret)
{
	auto b = new Buffer(this, hostMemory, usage, size);
	b->init();
	ret = b;
	static_assert(sizeof(usage) == 4);
}



void vge::Device::NewImage(vk::ImageUsageFlags usage, const ImageDescription *imageDescription, Image*& ret)
{
	auto img = new Image(this, usage, *imageDescription);
	img->init();
	ret = img;
}

void vge::Device::NewCommand(vk::QueueFlags queueType, bool once, Command*& ret)
{
	auto pos = std::find_if(queues.begin(), queues.end(), [=](Queue* q) {
		return q->_flags == queueType;
		});
	if (pos == queues.end()) {
		pos = std::find_if(queues.begin(), queues.end(), [=](Queue* q) {
			return (q->_flags & queueType) == queueType;
			});
	}
	if (pos == queues.end()) {
		throw std::runtime_error("No suitable queue found");
	}
	auto cmd = new Command(this, (*pos)->_family, once);
	cmd->init();
	ret = cmd;
}

void vge::Device::NewMemoryBlock(MemoryBlock*& memBlock)
{
	memBlock = new MemoryBlock(this);
}

void vge::Device::NewDescriptorLayout(vk::DescriptorType descType, vk::ShaderStageFlags stages, uint32_t count, vk::DescriptorBindingFlagsEXT flags, DescriptorLayout* prevBinding, DescriptorLayout*& descriptorLayout)
{
	descriptorLayout = new DescriptorLayout(this, descType, stages, count, flags, prevBinding);
	descriptorLayout->init();

}

void vge::Device::NewGraphicsPipeline(GraphicsPipeline*& gp)
{
	gp = new GraphicsPipeline(this);
}

void vge::Device::NewComputePipeline(ComputePipeline*& cp)
{
	cp = new ComputePipeline(this);
}

void vge::Device::NewSampler(vk::SamplerAddressMode mode, Sampler*& sampler)
{
	sampler = new Sampler(this, mode);
	sampler->init();

}

void vge::Device::NewTimestampQuery(uint32_t size, QueryPool*& qp)
{
	qp = new QueryPool(this, vk::QueryType::eTimestamp, size);
	qp->init();
}

void vge::Device::Submit(Command* command, uint32_t priority, SubmitInfo **info, size_t info_len, vk::PipelineStageFlags waitForStage, SubmitInfo* &waitFor)
{
	Queue* found = nullptr;
	for (auto q : queues) {
		if (q->_family == command->_family) {
			found = q;
			if (priority == 0) {
				break;
			}
			priority--;
		}
	}
	if (found != nullptr) {
		found->submit(command, info, info_len, waitForStage, waitFor);
	}
}


void vge::Device::Dispose()
{
	if (!!_dev) {
		for (auto q : queues) {
			delete q;
		}
		queues.clear();
		_dev.waitIdle(_dispatchLoader);
		_dev.destroy(allocator, _dispatchLoader);
		_dev = nullptr;
	}
}

vge::Device::~Device()
{
	std::vector<DeviceExtension*> deviceExtensions;
	for (auto dext : _inst->get_app()->get_deviceExtensions()) {
		dext->detach(this);
	}
	for (auto q : queues) {
		delete q;
	}
	queues.clear();
	_dev.destroy(allocator);
}

void vge::Device::init()
{
	std::vector<const char *> extensions;
	
	vk::DeviceCreateInfo dci;
	vk::PhysicalDeviceFeatures pf;
	dci.pEnabledFeatures = &pf;
	// Available on all Windows / Linux drivers
	pf.shaderStorageImageExtendedFormats = 1;
	pf.geometryShader = 1;

	std::vector<vk::DeviceQueueCreateInfo> crqs;
	float priorities[3] = { 1, 0.5, 0.25 };
	auto qfs = _pd.getQueueFamilyProperties();
	uint32_t qIndex = 0;
	for (auto qf : qfs) {
		if ((qf.queueFlags & vk::QueueFlagBits::eGraphics) == vk::QueueFlagBits::eGraphics && _graphicsQueueIdx > qfs.size()) {
			_graphicsQueueIdx = qIndex;
		}
		uint32_t qc = 3;
		if (qf.queueCount < qc) {
			qc = qf.queueCount;
		}
		vk::DeviceQueueCreateInfo dcqi;
		dcqi.pQueuePriorities = priorities;
		dcqi.queueFamilyIndex = qIndex;
		dcqi.queueCount = qc;
		crqs.push_back(dcqi);
		qIndex++;
	}
	dci.queueCreateInfoCount = qIndex;
	dci.pQueueCreateInfos = crqs.data();
	_memProps = _pd.getMemoryProperties(_inst->get_dispatch());
	_properties = _pd.getProperties();
	for (auto dext : _inst->get_app()->get_deviceExtensions()) {
		dext->prepare(dci, extensions);
	}
	if (extensions.size() > 0) {
		dci.enabledExtensionCount = static_cast<uint32_t>(extensions.size());
		dci.ppEnabledExtensionNames = extensions.data();
	}
	_dev = _pd.createDevice(dci, allocator, _inst->get_dispatch());
	_dispatchLoader.init(_inst->get_instance(), &vkGetInstanceProcAddr, _dev, _inst->get_dispatch().vkGetDeviceProcAddr);
	for (uint32_t familyIndex = 0; familyIndex < qIndex; familyIndex++) {
		for (uint32_t index = 0; index < crqs[familyIndex].queueCount; index++) {
			auto q = new vge::Queue(this, qfs[familyIndex].queueFlags, familyIndex, index);
			q->init();
			queues.push_back(q);
		}
	}
	for (auto dext : _inst->get_app()->get_deviceExtensions()) {
		dext->attach(this);
	}
}

void vge::Exception::GetError(char* msg_, size_t msg_len, int32_t& msgLen)
{
	msgLen = static_cast<int32_t>(msg.size());
	size_t cl = msg.size();
	if (cl > msg_len) {
		cl = msg_len;
	}
	if (cl > 0) {
		std::memcpy(msg_, msg.data(), cl);
	}
}


vge::Exception* vge::Exception::getValidationError()
{
	auto tmp = validationError;
	validationError = nullptr;
	return tmp;
}

void vge::Static::NewApplication(char* name, size_t name_len, Application*& app)
{
	app = new Application();
	app->name = std::string(name, name_len);

}

void vge::Static::NewDesktop(Application *app, Desktop*& desktop)
{
	desktop = new Desktop(app);
	desktop->init();
}

void vge::Static::AddValidation(Application* app)
{
	new ValidationOption(app);
}

void vge::Static::AddDynamicDescriptors(Application* app)
{
	new DynamicDescriptorOption(app);
}


void vge::Static::NewImageLoader(ImageLoader*& loader)
{
	loader = new ImageLoader();
}

void vge::Static::DebugPoint(const char* point, size_t len_point)
{
	#ifdef _WIN32
	DebugBreak();
	#endif
}


vge::SuppressValidation::SuppressValidation()
{
	suppressValidation++;
}

vge::SuppressValidation::~SuppressValidation()
{
	suppressValidation--;
}
