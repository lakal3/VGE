#include "vgelib/vgelib.hpp"

void vge::Buffer::GetPtr(void*& boundMemory)
{
	char* memPtr = nullptr;
	if (_owner != nullptr) {
		memPtr = static_cast<char*>(_owner->get_memPtr());
	}
	if (memPtr == nullptr) {
		throw std::runtime_error("Buffer memory not bound");
	}
	boundMemory = memPtr + _offset;
}

void vge::Buffer::CopyFrom(size_t offset, void* ptr, size_t size)
{
	char* memPtr = nullptr;
	if (_owner != nullptr) {
		memPtr = static_cast<char*>(_owner->get_memPtr());
	}
	if (memPtr == nullptr) {
		throw std::runtime_error("Buffer memory not bound");
	}
	if (_offset + offset + size > _size) {
		throw std::out_of_range("Size + offset larger that buffer size");
	}
	std::memcpy(memPtr + _offset + offset, ptr, size);
}


void vge::Buffer::NewView(vk::Format format, uint64_t offset, uint64_t size, BufferView*& view)
{
	view = new BufferView(_dev, _buffer, format);
	view->init(size, offset);
}

void vge::Buffer::Dispose()
{
	if (!!_buffer) {
		_dev->get_device().destroyBuffer(_buffer, allocator, _dev->get_dispatch());
		_buffer = nullptr;
		_owner = nullptr;
	}
}

vk::MemoryRequirements vge::Buffer::getMemoryRequirements(bool& hostMemory)
{
	hostMemory = _hostMemory;
	return _dev->get_device().getBufferMemoryRequirements(_buffer, _dev->get_dispatch());
}

void vge::Buffer::bind()
{
	_dev->get_device().bindBufferMemory(_buffer, _owner->get_mem(), _offset, _dev->get_dispatch());
}

void vge::Buffer::init()
{
	vk::BufferCreateInfo bci;
	bci.size = _size;
	bci.usage = _usage;
	_buffer = _dev->get_device().createBuffer(bci, allocator, _dev->get_dispatch());
}


void vge::Image::Dispose()
{
	if (_swapchainImage) {
		return;
	}
	if (!!_image) {
		_dev->get_device().destroyImage(_image, allocator, _dev->get_dispatch());
		_image = nullptr;
	}
}

vk::MemoryRequirements vge::Image::getMemoryRequirements(bool& hostMemory)
{
	hostMemory = false;
	return _dev->get_device().getImageMemoryRequirements(_image, _dev->get_dispatch());
}

void vge::Image::bind()
{
	_dev->get_device().bindImageMemory(_image, _owner->get_mem(), _offset, _dev->get_dispatch());
}

void vge::Image::NewView(vge::ImageRange *range, ImageView*& view, bool cubeView)
{
	auto ivci = cvInfo(0, 0);
	ivci.subresourceRange.layerCount = range->LayerCount;
	ivci.subresourceRange.baseArrayLayer = range->FirstLayer;
	ivci.subresourceRange.levelCount = range->LevelCount;
	ivci.subresourceRange.baseMipLevel = range->FirstMipLevel;
	if (cubeView) {
		if (ivci.viewType == vk::ImageViewType::e2D) {
			ivci.viewType = vk::ImageViewType::eCube;
		}
	} else {
		if (range->LayerCount > 1 && ivci.viewType == vk::ImageViewType::e2D) {
			ivci.viewType = vk::ImageViewType::e2DArray;
		}
	}
	auto vh = _dev->get_device().createImageView(ivci, allocator, _dev->get_dispatch());
	view = new ImageView(_dev, vh, this, *range);
}

void vge::Image::init()
{
	vk::ImageCreateInfo ici;
	ici.arrayLayers = _description.Layers;
	ici.format = _description.Format;
	ici.samples = vk::SampleCountFlagBits::e1;
	ici.imageType = vk::ImageType::e2D;
	ici.extent.width = _description.Width;
	ici.extent.height = _description.Height;
	if (_description.Depth > 1) {
		ici.imageType = vk::ImageType::e3D;
		ici.extent.depth = _description.Depth;
	}
	else {
		ici.extent.depth = 1;
	}
	ici.mipLevels = _description.MipLevels;
	ici.usage = _usage;
	ici.sharingMode = vk::SharingMode::eExclusive;
	if (ici.arrayLayers == 6) {
		ici.flags = vk::ImageCreateFlagBits::eCubeCompatible;
	}
	_image = _dev->get_device().createImage(ici, allocator, _dev->get_dispatch());
}

vk::ImageViewCreateInfo vge::Image::cvInfo(int32_t layer, int32_t mipLevel)
{
	vk::ImageViewCreateInfo ivci;
	ivci.image = _image;
	ivci.format = _description.Format;
	ivci.viewType = get_viewType();
	ivci.components.a = vk::ComponentSwizzle::eA;
	ivci.components.r = vk::ComponentSwizzle::eR;
	ivci.components.g = vk::ComponentSwizzle::eG;
	ivci.components.b = vk::ComponentSwizzle::eB;
	ivci.subresourceRange.aspectMask = vk::ImageAspectFlagBits::eColor;
	if (ivci.format == vk::Format::eD32Sfloat || ivci.format == vk::Format::eD32SfloatS8Uint || ivci.format == vk::Format::eD16Unorm ||
		ivci.format == vk::Format::eD16UnormS8Uint || ivci.format == vk::Format::eD24UnormS8Uint) {
		ivci.subresourceRange.aspectMask = vk::ImageAspectFlagBits::eDepth;
	}
	ivci.subresourceRange.baseArrayLayer = layer;
	ivci.subresourceRange.baseMipLevel = mipLevel;
	return ivci;
}

void vge::MemoryBlock::Reserve(MemoryObject* obj, bool &ok)
{
	bool hostMemory;
	auto mr = obj->getMemoryRequirements(hostMemory);
	auto mi = findMemIndex(mr, hostMemory);
	size_t offset = 0;
	if (_objects.size() == 0) {
		_memIndex = mi;
	} else {
		if (_memIndex != mi) {
			ok = 0;
			return;
		}
		offset = _objects[_objects.size() - 1]->_offset + _objects[_objects.size() - 1]->_allocSize;
	}
	ok = 1;
	_hostMem |= hostMemory;
	size_t rem = mr.size % mr.alignment;
	if (rem > 0) {
		obj->_allocSize = mr.size + (mr.alignment - rem);
	} else {
		obj->_allocSize = mr.size;
	}
	rem = offset % mr.alignment;
	if (rem != 0) {
		offset += mr.alignment - rem;
	}

	obj->_offset = offset;
	_objects.push_back(obj);
}

void vge::MemoryBlock::Allocate()
{
	if (_objects.size() == 0) {
		throw std::runtime_error("Can't allocate empty memory block!");
	}
	size_t size = _objects[_objects.size() - 1]->_offset + _objects[_objects.size() - 1]->_allocSize;
	vk::MemoryAllocateInfo mai;
	mai.allocationSize = size;
	mai.memoryTypeIndex = _memIndex;
	_mem = _dev->get_device().allocateMemory(mai, allocator, _dev->get_dispatch());
	if (_hostMem) {
		_memPtr = _dev->get_device().mapMemory(_mem, 0, size, vk::MemoryMapFlagBits(), _dev->get_dispatch());
	}

	size_t offset = 0;
	for (auto obj : _objects) {
		obj->_owner = this;
		obj->bind();
	}
}

void vge::MemoryBlock::Dispose()
{
	if (_memPtr != nullptr) {
		_dev->get_device().unmapMemory(_mem, _dev->get_dispatch());
		_memPtr = nullptr;
	}
	if (!!_mem) {
		for (auto obj : _objects) {
			obj->Dispose();
		}
		_objects.clear();
		_dev->get_device().freeMemory(_mem, allocator, _dev->get_dispatch());
		_mem = nullptr;
	}
}

uint32_t vge::MemoryBlock::findMemIndex(vk::MemoryRequirements mr, bool hostMemory)
{
	vk::MemoryPropertyFlags hmFlags = vk::MemoryPropertyFlagBits::eHostVisible | vk::MemoryPropertyFlagBits::eHostCoherent;
	auto mps = _dev->get_memoryProps();
	for (uint32_t memIndex = 0; memIndex < mps.memoryTypeCount; memIndex++) {
		if ((mr.memoryTypeBits & (1 << memIndex)) == 0) {
			continue;
		}
		if (!hostMemory && (mps.memoryTypes[memIndex].propertyFlags & vk::MemoryPropertyFlagBits::eDeviceLocal) == vk::MemoryPropertyFlagBits::eDeviceLocal) {
			return memIndex;
		}
		if (hostMemory && (mps.memoryTypes[memIndex].propertyFlags & hmFlags) == hmFlags) {
			return memIndex;
		}
	}
	throw std::runtime_error("No suitable memory found!");
}

void vge::ImageView::Dispose()
{
	if (!!_view) {
		
		_dev->get_device().destroyImageView(_view, allocator, _dev->get_dispatch());
		_view = nullptr;
	}
}

void vge::BufferView::init(size_t size, size_t offset)
{
	vk::BufferViewCreateInfo bvci;
	bvci.buffer = _buffer;
	bvci.format = _format;
	bvci.offset = offset;
	if (size > 0) {
		bvci.range = size;
	} else {
		bvci.range = VK_WHOLE_SIZE;
	}
	_view = _dev->get_device().createBufferView(bvci, allocator, _dev->get_dispatch());
}

void vge::BufferView::Dispose()
{
	_dev->get_device().destroyBufferView(_view, allocator, _dev->get_dispatch());
	delete this;
}
