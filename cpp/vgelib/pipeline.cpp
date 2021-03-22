#include "vgelib/vgelib.hpp"

void vge::GraphicsPipeline::Create(RenderPass* renderPass)
{
	vk::GraphicsPipelineCreateInfo gpci;
	
	gpci.renderPass = renderPass->get_handle();
	gpci.layout = createPipelineLayout();
	vk::PipelineVertexInputStateCreateInfo pvici;
	pvici.vertexAttributeDescriptionCount = static_cast<uint32_t>(_attrDescriptions.size());
	if (pvici.vertexAttributeDescriptionCount > 0) {
		pvici.pVertexAttributeDescriptions = _attrDescriptions.data();
	}
	pvici.vertexBindingDescriptionCount = static_cast<uint32_t>(_bindingDescriptions.size());
	if (pvici.vertexBindingDescriptionCount > 0) {
		pvici.pVertexBindingDescriptions = _bindingDescriptions.data();
	}
	gpci.pVertexInputState = &pvici;
	gpci.stageCount = static_cast<uint32_t>(_shaders.size());
	if (gpci.stageCount > 0) {
		gpci.pStages = _shaders.data();
	}
	vk::PipelineRasterizationStateCreateInfo prsci;
	prsci.cullMode = vk::CullModeFlagBits::eNone;
	prsci.lineWidth = 1.0;
	gpci.pRasterizationState = &prsci;
	vk::PipelineInputAssemblyStateCreateInfo piasci;
	piasci.topology = vk::PrimitiveTopology::eTriangleList;
	gpci.pInputAssemblyState = &piasci;
	gpci.pDepthStencilState = &_depthState;
	std::vector<vk::PipelineColorBlendAttachmentState> _colorStateAttachments;
	for (uint32_t idx = 0; idx < renderPass->get_color_attachment_count(); idx++) {
		vk::PipelineColorBlendAttachmentState st;
		st.blendEnable = false;
		switch (_blendMode)	{
		case 1:
			st.blendEnable = true;
			st.srcColorBlendFactor = vk::BlendFactor::eSrcAlpha;
			st.srcAlphaBlendFactor = vk::BlendFactor::eZero;
			st.dstColorBlendFactor = vk::BlendFactor::eOneMinusSrcAlpha;
			st.dstAlphaBlendFactor = vk::BlendFactor::eOne;
			st.colorBlendOp = vk::BlendOp::eAdd;
			st.alphaBlendOp = vk::BlendOp::eAdd;
			break;
		}
		st.colorWriteMask = vk::ColorComponentFlagBits::eA | vk::ColorComponentFlagBits::eR | vk::ColorComponentFlagBits::eG | vk::ColorComponentFlagBits::eB;
		_colorStateAttachments.push_back(st);

	}
	_colorBlendState.attachmentCount = static_cast<uint32_t>(_colorStateAttachments.size());
	if (_colorBlendState.attachmentCount > 0) {
		_colorBlendState.pAttachments = _colorStateAttachments.data();
	}
	gpci.pColorBlendState = &_colorBlendState;
	vk::PipelineDynamicStateCreateInfo pdsci;
	pdsci.dynamicStateCount = static_cast<uint32_t>(_dynStates.size());
	pdsci.pDynamicStates = _dynStates.data();
	gpci.pDynamicState = &pdsci;
	vk::PipelineViewportStateCreateInfo pvsci;
	vk::Viewport viewPort;
	viewPort.width = 1024;
	viewPort.height = 768;
	viewPort.maxDepth = 1;
	pvsci.pViewports = &viewPort;
	pvsci.viewportCount = 1;
	vk::Rect2D scissors({ 0, 0 }, { 1024, 768 });
	pvsci.scissorCount = 1;
	pvsci.pScissors = &scissors;
	gpci.pViewportState = &pvsci;
	vk::PipelineMultisampleStateCreateInfo pmsci;
	pmsci.minSampleShading = 1;
	pmsci.rasterizationSamples = vk::SampleCountFlagBits::e1;
	gpci.pMultisampleState = &pmsci;
	// If you receive error 'value': is not a member of 'vk::Pipeline' upgrade your VulkanSDK to v1.2.x.
	// ResultValue handling was changes in vulkan.hpp
	_pipeline = _dev->get_device().createGraphicsPipeline(nullptr, gpci, allocator, _dev->get_dispatch()).value;
}

void vge::GraphicsPipeline::AddVertexBinding(uint32_t stride, vk::VertexInputRate rate)
{
	vk::VertexInputBindingDescription bd;
	bd.binding = static_cast<uint32_t>(_bindingDescriptions.size());
	bd.inputRate = rate;
	bd.stride = stride;
	_bindingDescriptions.push_back(bd);
}

void vge::GraphicsPipeline::AddVertexFormat(vk::Format format, uint32_t offset)
{
	if (_bindingDescriptions.size() == 0) {
		throw std::runtime_error("No vertex binding defined");
	}
	vk::VertexInputAttributeDescription ad;
	ad.binding = static_cast<uint32_t>(_bindingDescriptions.size() - 1);
	ad.location = static_cast<uint32_t>(_attrDescriptions.size());
	ad.format = format;
	ad.offset = offset;
	_attrDescriptions.push_back(ad);
}

void vge::GraphicsPipeline::AddDepth(bool write, bool check)
{
	_depthState.depthWriteEnable = write;
	_depthState.depthTestEnable = check;
	_depthState.depthCompareOp = vk::CompareOp::eLessOrEqual;
}

void vge::GraphicsPipeline::AddAlphaBlend()
{
	_blendMode = 1;
}

vge::GraphicsPipeline::GraphicsPipeline(const Device* dev) :Pipeline(dev) {
	_dynStates.push_back(vk::DynamicState::eViewport);
	_dynStates.push_back(vk::DynamicState::eScissor);
	_depthState.maxDepthBounds = 1;
}

void vge::Pipeline::AddShader(vk::ShaderStageFlags stage, uint8_t* shader, size_t shaderLen)
{
	vk::ShaderModuleCreateInfo smci;
	smci.pCode = reinterpret_cast<uint32_t*>(shader);
	smci.codeSize = shaderLen;
	vk::PipelineShaderStageCreateInfo pssci;
	pssci.module = _dev->get_device().createShaderModule(smci, allocator, _dev->get_dispatch());
	VkShaderStageFlags st = VkShaderStageFlags(stage);
	pssci.stage = vk::ShaderStageFlagBits(st);
	pssci.pName = "main";
	_shaders.push_back(pssci);
}

void vge::Pipeline::AddDescriptorLayout(DescriptorLayout* layout)
{
	_layouts.push_back(layout);
}

vk::PipelineLayout vge::Pipeline::createPipelineLayout()
{
	vk::PipelineLayoutCreateInfo plci;
	std::vector<vk::DescriptorSetLayout> dsLayouts;
	for (auto l : _layouts) {
		dsLayouts.push_back(l->get_layout());
	}
	plci.setLayoutCount = static_cast<uint32_t>(dsLayouts.size());
	if (dsLayouts.size() > 0) {
		plci.pSetLayouts = dsLayouts.data();
	}
	_pipelineLayout = _dev->get_device().createPipelineLayout(plci, allocator, _dev->get_dispatch());
	return _pipelineLayout;
}

void vge::Pipeline::Dispose()
{
	if (!!_pipeline) {
		for (auto sh : _shaders) {
			_dev->get_device().destroyShaderModule(sh.module, allocator, _dev->get_dispatch());
		}
		_dev->get_device().destroyPipeline(_pipeline, allocator, _dev->get_dispatch());
	}
	if (!!_pipelineLayout) {
		_dev->get_device().destroyPipelineLayout(_pipelineLayout, allocator, _dev->get_dispatch());
	}
}

vge::ComputePipeline::ComputePipeline(const Device *dev):Pipeline(dev)
{
}

void vge::ComputePipeline::Create()
{
	vk::ComputePipelineCreateInfo cpci;
	cpci.layout = createPipelineLayout();
	if (_shaders.size() != 1) {
		throw std::runtime_error("Compute shader need 1 stage");
	}
	cpci.stage = _shaders[0];
	_pipeline = _dev->get_device().createComputePipeline(nullptr, cpci, allocator, _dev->get_dispatch()).value;
}
