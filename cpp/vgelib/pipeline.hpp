#pragma once


namespace vge {
	class Pipeline : public Disposable {
	public:
		void AddShader(vk::ShaderStageFlags stage, uint8_t* shader, size_t shaderLen);
		void AddDescriptorLayout(DescriptorLayout* layout);
		const vk::Pipeline get_handle() const {
			return _pipeline;
		}
		virtual const vk::PipelineBindPoint get_bindpoint() const = 0;
		vk::PipelineLayout get_layout() const {
			return _pipelineLayout;
		}
		
	protected:
		Pipeline(const Device* dev) :_dev(dev) {

		}
		vk::PipelineLayout createPipelineLayout();
		virtual void Dispose() override;
		std::vector<vk::PipelineShaderStageCreateInfo> _shaders;
		std::vector<DescriptorLayout*> _layouts;
		vk::Pipeline _pipeline;
		vk::PipelineLayout _pipelineLayout;
		int _blendMode = 0;
		const Device* const _dev;
	};

	class GraphicsPipeline: public Pipeline {
		friend class Device;
	public:
		void Create(RenderPass *pr);
		
		void AddVertexBinding(uint32_t stride, vk::VertexInputRate rate);
		void AddVertexFormat(vk::Format format, uint32_t offset);
		void AddDepth(bool write, bool check);
		void AddAlphaBlend();
		void SetTopology(vk::PrimitiveTopology topology) {
			_topology = topology;
		}
	private:
		GraphicsPipeline(const Device* dev);
		virtual const vk::PipelineBindPoint get_bindpoint() const override {
			return vk::PipelineBindPoint::eGraphics;
		}
		std::vector<vk::VertexInputAttributeDescription> _attrDescriptions;
		std::vector<vk::VertexInputBindingDescription> _bindingDescriptions;
		vk::PipelineColorBlendStateCreateInfo _colorBlendState;
		vk::PipelineDepthStencilStateCreateInfo _depthState;
		std::vector<vk::DynamicState> _dynStates;
		vk::PrimitiveTopology _topology;
	};

	class ComputePipeline : public Pipeline {
		friend class Device;
	public:
		void Create();
	private:
		virtual const vk::PipelineBindPoint get_bindpoint() const override {
			return vk::PipelineBindPoint::eCompute;
		}
		ComputePipeline(const Device* dev);
	};
}