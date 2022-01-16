#include "vgelib/vgelib.hpp"

void vge::DescriptorLayout::Dispose()
{
	if (!!_dsLayout) {
		_dev->get_device().destroyDescriptorSetLayout(_dsLayout, allocator, _dev->get_dispatch());
		_dsLayout = nullptr;
	}
}

void vge::DescriptorLayout::NewPool(uint32_t size, DescriptorPool*& pool)
{
	pool = new DescriptorPool(_dev, this, size);
	pool->init();
}

void vge::DescriptorLayout::init()
{
	vk::DescriptorSetLayoutCreateInfo dsci;
	dsci.bindingCount = _binding + 1;
	std::vector<vk::DescriptorSetLayoutBinding> bindings;
	std::vector<vk::DescriptorBindingFlagsEXT> allFlags;
	bool hasFlags = addBinding(dsci, bindings, allFlags);
	dsci.pBindings = bindings.data();
	vk::DescriptorSetLayoutBindingFlagsCreateInfoEXT addFlags;
	if (hasFlags) {
		addFlags.pBindingFlags = allFlags.data();
		addFlags.bindingCount = static_cast<uint32_t>(allFlags.size());
		dsci.pNext = &addFlags;
		dsci.flags = vk::DescriptorSetLayoutCreateFlagBits::eUpdateAfterBindPoolEXT;
		_updateAfterBind = true;
	}
	_dsLayout = _dev->get_device().createDescriptorSetLayout(dsci, allocator, _dev->get_dispatch());
}

bool vge::DescriptorLayout::addBinding(vk::DescriptorSetLayoutCreateInfo& dsci, std::vector<vk::DescriptorSetLayoutBinding>& bindings, 
	std::vector<vk::DescriptorBindingFlagsEXT>& allFlags) const
{
	bool hasFlags = false;
	if (prevLayout != nullptr) {
		hasFlags |= prevLayout->addBinding(dsci, bindings, allFlags);
	}
	vk::DescriptorSetLayoutBinding dsb;
	dsb.descriptorCount = count;
	dsb.descriptorType = type;
	dsb.stageFlags = stages;
	dsb.binding = static_cast<uint32_t>(bindings.size());
	bindings.push_back(dsb);
	allFlags.push_back(flags);
	hasFlags |= !!flags;
	return hasFlags;
}

void vge::DescriptorPool::Dispose()
{
	if (!!_pool) {
		for (auto d : sets) {
			delete d;
		}
		sets.clear();
		_dev->get_device().destroyDescriptorPool(_pool, allocator, _dev->get_dispatch());
		_pool = nullptr;
	}
}

void vge::DescriptorPool::Alloc(DescriptorSet*& desciptor)
{
	vk::DescriptorSetAllocateInfo dsai;
	dsai.descriptorPool = _pool;
	dsai.descriptorSetCount = 1;
	auto layout = _dsLayout->get_layout();
	dsai.pSetLayouts = &layout;
	auto newSets = _dev->get_device().allocateDescriptorSets(dsai, _dev->get_dispatch());
	desciptor = new DescriptorSet(_dev, newSets[0], _dsLayout);
	sets.push_back(desciptor);
}

void vge::DescriptorPool::init()
{
	vk::DescriptorPoolCreateInfo dpci;
	dpci.maxSets = _size;
	std::vector<vk::DescriptorPoolSize> poolSizes;
	fillSizes(poolSizes, _dsLayout);
	dpci.poolSizeCount = static_cast<uint32_t>(poolSizes.size());
	dpci.pPoolSizes = poolSizes.data();
	if (_dsLayout->get_updateAfterBind()) {
		dpci.flags = vk::DescriptorPoolCreateFlagBits::eUpdateAfterBindEXT;
	}
	_pool =_dev->get_device().createDescriptorPool(dpci, allocator, _dev->get_dispatch());
}

void vge::DescriptorPool::fillSizes(std::vector<vk::DescriptorPoolSize>& poolSizes, const DescriptorLayout* dsLayout)
{
	bool found = false;
	for (int idx = 0; idx < poolSizes.size(); idx++) {
		if (poolSizes[idx].type == dsLayout->type) {
			poolSizes[idx].descriptorCount += _size * dsLayout->count;
			found = true;
			break;
		}
	}
	if (!found) {
		auto dps = vk::DescriptorPoolSize(dsLayout->type, dsLayout->count * _size);
		poolSizes.push_back(dps);
	}
	if (dsLayout->prevLayout != nullptr) {
		fillSizes(poolSizes, dsLayout->prevLayout);
	}
}


void vge::Sampler::Dispose()
{
	if (!!_sampler) {
		_dev->get_device().destroySampler(_sampler, allocator, _dev->get_dispatch());
		_sampler = nullptr;
	}
}

void vge::Sampler::init()
{
	vk::SamplerCreateInfo sci;
	sci.addressModeU = _mode;
	sci.addressModeV = _mode;
	sci.addressModeW = _mode;
	sci.mipmapMode = vk::SamplerMipmapMode::eLinear;
	sci.magFilter = vk::Filter::eLinear;
	sci.minFilter = vk::Filter::eLinear;
	sci.maxLod = 8;
	_sampler = _dev->get_device().createSampler(sci, allocator, _dev->get_dispatch());
}

void vge::DescriptorSet::WriteBuffer(uint32_t binding, uint32_t at, Buffer* content, uint64_t from, uint64_t size)
{
	vk::DescriptorBufferInfo dib;
	dib.buffer = content->get_buffer();
	if (from == 0 && size == 0) {
		dib.range = VK_WHOLE_SIZE;
	} else {
		dib.range = size;
		dib.offset = from;
	}

	vk::WriteDescriptorSet wds;
	wds.descriptorCount = 1;
	wds.descriptorType = getType(static_cast<uint32_t>(binding), _dsLayout);
	wds.pBufferInfo = &dib;
	wds.dstSet = _ds;
	wds.dstBinding = binding;
	wds.dstArrayElement = at;
	_dev->get_device().updateDescriptorSets(1, &wds, 0, nullptr, _dev->get_dispatch());
}

void vge::DescriptorSet::WriteImage(uint32_t binding, uint32_t at, ImageView* view, Sampler* sampler)
{
	vk::DescriptorImageInfo dii;
	dii.imageView = view->get_view();
	dii.imageLayout = view->range.get_layout();
	if (sampler != nullptr) {
		dii.sampler = sampler->get_sampler();
	}
	vk::WriteDescriptorSet wds;
	wds.descriptorCount = 1;
	wds.descriptorType = getType(static_cast<uint32_t>(binding), _dsLayout);
	wds.pImageInfo = &dii;
	wds.dstSet = _ds;
	wds.dstBinding = binding;
	wds.dstArrayElement = at;
	_dev->get_device().updateDescriptorSets(1, &wds, 0, nullptr, _dev->get_dispatch());
}

void vge::DescriptorSet::WriteDSImageView(uint32_t binding, uint32_t at, void* view, vk::ImageLayout layout, Sampler* sampler)
{
	vk::DescriptorImageInfo dii;
	dii.imageView = static_cast<VkImageView>(view);
	dii.imageLayout = layout;
	if (sampler != nullptr) {
		dii.sampler = sampler->get_sampler();
	}
	vk::WriteDescriptorSet wds;
	wds.descriptorCount = 1;
	wds.descriptorType = getType(static_cast<uint32_t>(binding), _dsLayout);
	wds.pImageInfo = &dii;
	wds.dstSet = _ds;
	wds.dstBinding = binding;
	wds.dstArrayElement = at;
	_dev->get_device().updateDescriptorSets(1, &wds, 0, nullptr, _dev->get_dispatch());
}

void vge::DescriptorSet::WriteDSSlice(uint32_t binding, uint32_t at, void* buffer, uint64_t from, uint64_t size)
{
	vk::DescriptorBufferInfo dib;
	dib.buffer = static_cast<VkBuffer>(buffer);
	dib.range = size;
	dib.offset = from;

	vk::WriteDescriptorSet wds;
	wds.descriptorCount = 1;
	wds.descriptorType = getType(static_cast<uint32_t>(binding), _dsLayout);
	wds.pBufferInfo = &dib;
	wds.dstSet = _ds;
	wds.dstBinding = binding;
	wds.dstArrayElement = at;
	_dev->get_device().updateDescriptorSets(1, &wds, 0, nullptr, _dev->get_dispatch());
}

void vge::DescriptorSet::WriteBufferView(uint32_t binding, uint32_t at, BufferView* content)
{
	auto view = content->get_view();
	vk::WriteDescriptorSet wds;
	wds.descriptorCount = 1;
	wds.descriptorType = getType(static_cast<uint32_t>(binding), _dsLayout);
	wds.pTexelBufferView = &view;
	wds.dstSet = _ds;
	wds.dstBinding = binding;
	wds.dstArrayElement = at;
	_dev->get_device().updateDescriptorSets(1, &wds, 0, nullptr, _dev->get_dispatch());
}

vk::DescriptorType vge::DescriptorSet::getType(uint32_t binding, const DescriptorLayout* dsLayout)
{
	if (binding > dsLayout->get_binding()) {
		throw std::runtime_error("No such binding in descriptor set");
	}
	if (binding == dsLayout->get_binding()) {
		return dsLayout->type;
	}
	return getType(binding, dsLayout->prevLayout);
}

const char* EXT_descriptor_indexing = "VK_EXT_descriptor_indexing";

static char validateInfo[128];

void vge::DynamicDescriptorOption::isValid(Instance* inst, vk::PhysicalDevice pd, bool& valid, std::string& invalidReason)
{
	if (!DynamicDescriptorOption::checkExtension(inst, pd, EXT_descriptor_indexing)) {
		valid = false;
		invalidReason = "Hardware must support VK_EXT_descriptor_indexing";
		return;
	}
	auto features = pd.getFeatures2<vk::PhysicalDeviceFeatures2, vk::PhysicalDeviceDescriptorIndexingFeaturesEXT>(inst->get_dispatch());
	auto extF = features.get< vk::PhysicalDeviceDescriptorIndexingFeaturesEXT>();

	if (extF.descriptorBindingPartiallyBound && extF.descriptorBindingUpdateUnusedWhilePending && extF.descriptorBindingVariableDescriptorCount &&
		extF.runtimeDescriptorArray && extF.descriptorBindingSampledImageUpdateAfterBind && extF.descriptorBindingStorageImageUpdateAfterBind) {		
		valid = true;
		invalidReason = "";
		return;
	}

	valid = false;
	invalidReason = "Hardware must support partiallyBoundDescriptors, runtimeDescriptorArray, descriptorBindingUpdateUnusedWhilePending, descriptorBindingSampledImageUpdateAfterBind, descriptorBindingStorageImageUpdateAfterBind and descriptorBindingVariableDescriptorCount";
}

void vge::DynamicDescriptorOption::prepare(vk::DeviceCreateInfo& dci, std::vector<const char*>& extensions)
{
	
	_extOptions.pNext = const_cast<void*>(dci.pNext);
	_extOptions.descriptorBindingPartiallyBound = 1;
	_extOptions.descriptorBindingUpdateUnusedWhilePending = 1;
	_extOptions.descriptorBindingVariableDescriptorCount = 1;
	_extOptions.descriptorBindingSampledImageUpdateAfterBind = 1;
	_extOptions.descriptorBindingStorageImageUpdateAfterBind = 1;
	_extOptions.runtimeDescriptorArray = 1;
	dci.pNext = &_extOptions;
	extensions.push_back(EXT_descriptor_indexing);
}

void vge::DynamicDescriptorOption::dispose()
{
	delete this;
}
