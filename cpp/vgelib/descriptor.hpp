#pragma once

namespace vge {
	class DescriptorSet;
	class DescriptorPool;

	class DynamicDescriptorOption : public DeviceExtension {
		friend struct Static;
		DynamicDescriptorOption(Application *app) {
			app->addDeviceExtension(this);
		}
		virtual void isValid(Instance* inst, vk::PhysicalDevice pd, bool& valid, std::string& invalidReason) override;
		virtual void prepare(vk::DeviceCreateInfo& dci, std::vector<const char*>& extensions) override;
		virtual void attach(Device* inst) override {

		}
		virtual void detach(Device* inst) override {

		}
		virtual void dispose() override;
		vk::PhysicalDeviceDescriptorIndexingFeaturesEXT _extOptions;
	};

	class DescriptorLayout : public Disposable {
		friend class Device;
	public:
		virtual void Dispose() override;
		void NewPool(uint32_t size, DescriptorPool*& pool);

		const vk::DescriptorType type;
		const uint32_t count;
		DescriptorLayout* const prevLayout;
		const vk::ShaderStageFlags stages;
		const vk::DescriptorBindingFlagsEXT flags;
		vk::DescriptorSetLayout get_layout() const {
			return _dsLayout;
		}
		uint32_t get_binding() const {
			return _binding;
		}
	protected:
		DescriptorLayout(Device *dev, vk::DescriptorType type_, vk::ShaderStageFlags stages_, uint32_t count_, vk::DescriptorBindingFlagsEXT flags_, DescriptorLayout *prevLayout_) :
			_dev(dev), type(type_), stages(stages_), count(count_), prevLayout(prevLayout_), flags(flags_) {
			if (prevLayout == nullptr) {
				_binding = 0;
			} else {
				_binding = prevLayout->_binding + 1;
			}
		}
		virtual void init();
		virtual bool addBinding(vk::DescriptorSetLayoutCreateInfo &dsci, std::vector<vk::DescriptorSetLayoutBinding> &bindings, 
			std::vector<vk::DescriptorBindingFlagsEXT> &flags) const;
		vk::DescriptorSetLayout _dsLayout;
		const Device* const _dev;
	private:
		uint32_t _binding;
	};

	class Sampler : public Disposable {
		friend class Device;
	public:
		virtual void Dispose() override;
		vk::Sampler get_sampler() {
			return _sampler;
		}
	private:
		Sampler(Device* dev, vk::SamplerAddressMode mode) : _dev(dev), _mode(mode) {

		}
		void init();
		const Device* const _dev;
		const vk::SamplerAddressMode _mode;
		vk::Sampler _sampler;
	};

	class DescriptorPool : public Disposable {
		friend class DescriptorLayout;
	public:
		virtual void Dispose() override;
		void Alloc(DescriptorSet*& ds);
	private:
		DescriptorPool(const Device* dev, DescriptorLayout* dsLayout, uint32_t size): _dsLayout(dsLayout), _dev(dev), _size(size) {

		}

		void init();

		void fillSizes(std::vector<vk::DescriptorPoolSize>& poolSizes, const DescriptorLayout* dsLayout);
		std::vector<DescriptorSet*> sets;
		const DescriptorLayout* const _dsLayout;
		const Device* const _dev;
		const uint32_t _size;
		vk::DescriptorPool _pool;
	};

	class DescriptorSet  {
		friend class DescriptorPool;
	public:
		void WriteBuffer(uint32_t binding, uint32_t at, Buffer* content, uint64_t from, uint64_t size);
		void WriteImage(uint32_t binding, uint32_t at, ImageView* content, Sampler *sampler);
		void WriteBufferView(uint32_t binding, uint32_t at, BufferView * content);
		vk::DescriptorSet get_descriptorSet() {
			return _ds;
		}
	private:
		DescriptorSet(const Device* dev, vk::DescriptorSet ds, const DescriptorLayout *dsLayout):_dev(dev), _ds(ds), _dsLayout(dsLayout) {

		}

		vk::DescriptorType getType(uint32_t binding, const DescriptorLayout* dsLayout);
		const Device* const _dev;
		const vk::DescriptorSet _ds;
		const DescriptorLayout* const _dsLayout;
		
	};
}