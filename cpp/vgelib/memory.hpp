#pragma once

namespace vge {

	class MemoryObject : public Disposable {
		friend class MemoryBlock;
	protected:
		MemoryObject(const Device* dev):_dev(dev), _owner(nullptr), _allocSize(0), _offset(0) {

		}

		virtual vk::MemoryRequirements getMemoryRequirements(bool &hostMemory) = 0;
		virtual void bind() = 0;
		
		const Device* const _dev;
		MemoryBlock* _owner;
		size_t _offset;
		size_t _allocSize;
	};

	class BufferView;

	class Buffer: public MemoryObject {
		friend class Device;
		friend class Command;
	public:
		void GetPtr(void*& boundMemory);
		void CopyFrom(size_t offset, void* ptr, size_t size);
		void NewView(vk::Format format, uint64_t offset, uint64_t size, BufferView*& view);
		vk::Buffer get_buffer() const {
			return _buffer;
		}
		size_t getSize() {
			return _size;
		}
	protected:
		virtual void Dispose() override;
		virtual vk::MemoryRequirements getMemoryRequirements(bool& hostMemory) override;
		virtual void bind() override;
		
	private:
		Buffer(const Device* dev, bool hostMemory, vk::BufferUsageFlags usage, size_t size):MemoryObject(dev), _size(size), _usage(usage), _hostMemory(hostMemory) {

		}

		void init();
		vk::Buffer _buffer;
		const size_t _size;
		const vk::BufferUsageFlags _usage;
		const bool _hostMemory;

	};

	class BufferView : Disposable {
		friend class Buffer;
	public:
		vk::BufferView get_view() {
			return _view;
		}
	protected:
		BufferView(const Device* dev, vk::Buffer buffer, vk::Format format) :_buffer(buffer), _dev(dev), _format(format) {

		}
		void init(size_t size, size_t offset);
		virtual void Dispose() override;
		vk::BufferView _view;
		const vk::Buffer _buffer;
		const Device* const _dev;
		const vk::Format _format;
	};

	class ImageView;
	class Image: public MemoryObject {
		friend class Device;
	public:
		
		void NewView(ImageRange *range, ImageView*& view, bool cube);

		
		const ImageDescription& get_desc() const {
			return _description;
		}

		const vk::Image get_handle() const {
			return _image;
		}

		const vk::ImageViewType get_viewType() const {
			if (_description.Depth > 2) {
				return vk::ImageViewType::e3D;
			}
			return vk::ImageViewType::e2D;
		}

		const vk::ImageAspectFlags get_aspect() const {
			if (_description.Format >= vk::Format::eD16Unorm && _description.Format <= vk::Format::eD32SfloatS8Uint) {
				return vk::ImageAspectFlagBits::eDepth;
			}
			return vk::ImageAspectFlagBits::eColor;
		}

		Image(const Device* dev, vk::Image image, vk::ImageUsageFlags usage, const ImageDescription& description) :
			MemoryObject(dev), _usage(usage), _description(description), _image(image) {
			_swapchainImage = true;
		}
		virtual void Dispose() override;
	protected:
		
		virtual vk::MemoryRequirements getMemoryRequirements(bool& hostMemory) override;
		virtual void bind() override;
		
	private:
		Image(const Device* dev, vk::ImageUsageFlags usage, const ImageDescription &description) :MemoryObject(dev), _usage(usage), _description(description) {
			
		}

		void init();
		vk::ImageViewCreateInfo cvInfo(int32_t layer, int32_t mipLevel);

		vk::Image _image = nullptr;
		// size_t _size = 0;
		const vk::ImageUsageFlags _usage;
		const ImageDescription _description;
		bool _swapchainImage = false;
	};

	class ImageView : Disposable {
		friend class Image;
	public:
		virtual void Dispose() override;
		Image* get_image() {
			return _image;
		}
		vk::ImageView get_view() const {
			return _view;
		}
		ImageRange range;
	private:
		ImageView(const Device* dev, vk::ImageView view, Image* image, ImageRange range_):_dev(dev), _view(view), _image(image), range(range_) {

		}
		vk::ImageView _view;
		Image* const _image;

		const Device* const _dev;
	};

	class MemoryBlock: public Disposable {
		friend class Device;
	public:
		void Reserve(MemoryObject* obj, bool & ok);
		void Allocate();

		vk::DeviceMemory get_mem() const {
			return _mem;
		}

		Device* get_dev() const {
			return _dev;
		}

		void* get_memPtr() {
			return _memPtr;
		}
	protected:
		virtual void Dispose() override;
	private:
		MemoryBlock(Device *dev) : _dev(dev), _memIndex(0) {

		}
		uint32_t findMemIndex(vk::MemoryRequirements, bool hostMemory);
		Device* _dev;
		vk::DeviceMemory _mem = nullptr;
		void* _memPtr = nullptr;
		std::vector<MemoryObject*> _objects;
		uint32_t _memIndex;
		bool _hostMem = false;
	};
}