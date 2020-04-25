
#include "vgelib/vgelib.hpp"
#include "renderpass.hpp"

void vge::ForwardRenderPass::Init()
{
	vk::RenderPassCreateInfo rpci;
	std::vector<vk::AttachmentDescription> attachments;
	vk::AttachmentDescription mainAttachment;
	mainAttachment.finalLayout = _endLayout;
	mainAttachment.format = _mainImageFormat;
	mainAttachment.initialLayout = vk::ImageLayout::eUndefined;
	mainAttachment.loadOp = vk::AttachmentLoadOp::eClear;
	mainAttachment.storeOp = vk::AttachmentStoreOp::eStore;
	mainAttachment.samples = vk::SampleCountFlagBits::e1;
	attachments.push_back(mainAttachment);
	vk::SubpassDescription sd;
	sd.colorAttachmentCount = 1;
	vk::AttachmentReference mainRef(0, vk::ImageLayout::eColorAttachmentOptimal);
	sd.pColorAttachments = &mainRef;
	vk::AttachmentReference depthRef(1, vk::ImageLayout::eDepthStencilAttachmentOptimal);
	if (_depthImageFormat != vk::Format::eUndefined) {
		vk::AttachmentDescription depthAttachment;
		depthAttachment.finalLayout = vk::ImageLayout::eDepthStencilAttachmentOptimal;
		depthAttachment.format = _depthImageFormat;
		depthAttachment.initialLayout = vk::ImageLayout::eUndefined;
		depthAttachment.loadOp = vk::AttachmentLoadOp::eClear;
		depthAttachment.storeOp = vk::AttachmentStoreOp::eDontCare;
		depthAttachment.samples = vk::SampleCountFlagBits::e1;
		attachments.push_back(depthAttachment);
		sd.pDepthStencilAttachment = &depthRef;
	}
	rpci.pAttachments = attachments.data();
	rpci.attachmentCount = static_cast<uint32_t>(attachments.size());
	rpci.pSubpasses = &sd;
	rpci.subpassCount = 1;
	_renderPass = _dev->get_device().createRenderPass(rpci, allocator, _dev->get_dispatch());
}

void vge::ForwardRenderPass::fillClearValues(std::vector<vk::ClearValue>& clearValues)
{
	vk::ClearColorValue ccv;
	ccv.setFloat32({ 0.2f, 0.2f, 0.2f, 1.0f });
	clearValues.push_back(ccv);
	if (_depthImageFormat != vk::Format::eUndefined) {
		vk::ClearDepthStencilValue cdps(1.0f, 0);
		clearValues.push_back(cdps);
	}
}


uint32_t vge::ForwardRenderPass::get_color_attachment_count()
{
	return 1;
}

void vge::FPlusRenderPass::Init()
{
	vk::RenderPassCreateInfo rpci;
	std::vector<vk::AttachmentDescription> attachments;
	for (uint32_t ep = 0; ep <= _extraPhases; ep++) {
		vk::AttachmentDescription mainAttachment;
		mainAttachment.finalLayout = ep == _extraPhases ? _endLayout : vk::ImageLayout::eGeneral;
		mainAttachment.format = _mainImageFormat;
		mainAttachment.initialLayout = vk::ImageLayout::eUndefined;
		mainAttachment.loadOp = vk::AttachmentLoadOp::eClear;
		mainAttachment.storeOp = ep == _extraPhases ? vk::AttachmentStoreOp::eStore : vk::AttachmentStoreOp::eDontCare;
		mainAttachment.samples = vk::SampleCountFlagBits::e1;
		attachments.push_back(mainAttachment);
	}
	vk::AttachmentDescription depthAttachment;
	depthAttachment.finalLayout = vk::ImageLayout::eUndefined;
	depthAttachment.format = _depthImageFormat;
	depthAttachment.initialLayout = vk::ImageLayout::eUndefined;
	depthAttachment.loadOp = vk::AttachmentLoadOp::eClear;
	depthAttachment.storeOp = vk::AttachmentStoreOp::eDontCare;
	depthAttachment.samples = vk::SampleCountFlagBits::e1;
	attachments.push_back(depthAttachment);

	// Passes
	auto dp = _extraPhases + 1;
	std::vector<vk::SubpassDescription> subpasses;
	std::vector<vk::SubpassDependency> dependencies;
	std::vector<vk::AttachmentReference> colRefs;
	std::vector<uint32_t> preserveRefs;

	vk::SubpassDescription sdInit;
	sdInit.colorAttachmentCount = 1;
	// Can't do layout transition for preserved image(s), so general layout is only suitable
	vk::AttachmentReference mainRef(0, vk::ImageLayout::eGeneral);
	sdInit.pColorAttachments = &mainRef;
	vk::AttachmentReference depthRef(dp, vk::ImageLayout::eGeneral);
	sdInit.pDepthStencilAttachment = &depthRef;
	subpasses.push_back(sdInit);
	// dependencies.push_back(vk::SubpassDependency(0, 1, vk::PipelineStageFlagBits::eBottomOfPipe, vk::PipelineStageFlagBits::eEarlyFragmentTests, vk::AccessFlagBits::eShaderWrite, vk::AccessFlagBits::eShaderRead));
	for (uint32_t fromEp = 0; fromEp < _extraPhases; fromEp++) {
		vk::SubpassDescription sdNext;
		sdNext.colorAttachmentCount = 1;
		vk::AttachmentReference colRef(fromEp + 1, (fromEp + 1) == _extraPhases ? vk::ImageLayout::eColorAttachmentOptimal: vk::ImageLayout::eGeneral);
		colRefs.push_back(colRef);
		sdNext.pColorAttachments = &colRefs.back();
		
		preserveRefs.push_back(fromEp);
		auto deps = &preserveRefs.back();
		preserveRefs.push_back(dp);
		sdNext.pPreserveAttachments = deps;
		sdNext.preserveAttachmentCount = 2;
		subpasses.push_back(sdNext);
		// dependencies.push_back(vk::SubpassDependency(0, 1, vk::PipelineStageFlagBits::eBottomOfPipe, vk::PipelineStageFlagBits::eEarlyFragmentTests, vk::AccessFlagBits::eShaderWrite, vk::AccessFlagBits::eShaderRead));
	}
	rpci.pAttachments = attachments.data();
	rpci.attachmentCount = static_cast<uint32_t>(attachments.size());
	rpci.pSubpasses = subpasses.data();
	rpci.subpassCount = static_cast<uint32_t>(subpasses.size());
	// rpci.dependencyCount = static_cast<uint32_t>(dependencies.size());
	// rpci.pDependencies = dependencies.data();
	_renderPass = _dev->get_device().createRenderPass(rpci, allocator, _dev->get_dispatch());
}

void vge::FPlusRenderPass::fillClearValues(std::vector<vk::ClearValue>& clearValues)
{
	vk::ClearColorValue ccv;
	ccv.setFloat32({ 0.2f, 0.2f, 0.2f, 1.0f });
	for (uint32_t ep = 0; ep <= _extraPhases; ep++) {
		clearValues.push_back(ccv);
	}
	vk::ClearDepthStencilValue cdps(1.0f, 0);
	clearValues.push_back(cdps);
}


uint32_t vge::FPlusRenderPass::get_color_attachment_count()
{
	return 1;
}

void vge::RenderPass::NewFrameBuffer(ImageView** attachments, size_t attachments_len, Framebuffer*& fb)
{
	vk::FramebufferCreateInfo fbci;
	std::vector<vk::ImageView> atList;
	for (int idx = 0; idx < attachments_len; idx++) {
		atList.push_back(attachments[idx]->get_view());
	}
	fbci.layers = attachments[0]->range.LayerCount;
	fbci.attachmentCount = static_cast<int32_t>(attachments_len);
	fbci.pAttachments = atList.data();
	auto desc = attachments[0]->get_image()->get_desc();
	fbci.width = desc.Width;
	fbci.height = desc.Height;
	fbci.renderPass = _renderPass;
	fb = new Framebuffer(_dev, _dev->get_device().createFramebuffer(fbci, allocator, _dev->get_dispatch()), std::vector<ImageView *>(attachments, attachments+ attachments_len));
	
}

void vge::RenderPass::Dispose()
{
	_dev->get_device().destroyRenderPass(_renderPass, allocator, _dev->get_dispatch());
	delete this;
}


void vge::Static::NewForwardRenderPass(Device* dev, vk::ImageLayout finalLayout, vk::Format mainImageFormat, vk::Format depthImageFormat, RenderPass*& rp)
{
	rp = new vge::ForwardRenderPass(dev, finalLayout, mainImageFormat, depthImageFormat);
}

void vge::Static::NewFPlusRenderPass(Device* dev, uint32_t extraPhases, vk::ImageLayout finalLayout, vk::Format mainImageFormat, vk::Format depthImageFormat, RenderPass*& rp)
{
	rp = new vge::FPlusRenderPass(dev, extraPhases, finalLayout, mainImageFormat, depthImageFormat);
}


void vge::Framebuffer::Dispose()
{
	_dev->get_device().destroyFramebuffer(_framebuffer, allocator, _dev->get_dispatch());
	delete this;
}

void vge::Static::NewDepthRenderPass(Device* dev, vk::ImageLayout finalLayout, vk::Format depthImageFormat, RenderPass*& rp)
{
	rp = new vge::DepthRenderPass(dev, finalLayout, depthImageFormat);
}

void vge::DepthRenderPass::Init()
{
	vk::RenderPassCreateInfo rpci;
	std::vector<vk::AttachmentDescription> attachments;
	vk::SubpassDescription sd;
	vk::AttachmentReference depthRef(0, vk::ImageLayout::eDepthStencilAttachmentOptimal);
	vk::AttachmentDescription depthAttachment;
	depthAttachment.finalLayout = _endLayout;
	depthAttachment.format = _depthImageFormat;
	depthAttachment.initialLayout = vk::ImageLayout::eUndefined;
	depthAttachment.loadOp = vk::AttachmentLoadOp::eClear;
	depthAttachment.storeOp = vk::AttachmentStoreOp::eStore;
	depthAttachment.samples = vk::SampleCountFlagBits::e1;
	attachments.push_back(depthAttachment);
	sd.pDepthStencilAttachment = &depthRef;
	
	rpci.pAttachments = attachments.data();
	rpci.attachmentCount = static_cast<uint32_t>(attachments.size());
	rpci.pSubpasses = &sd;
	rpci.subpassCount = 1;
	_renderPass = _dev->get_device().createRenderPass(rpci, allocator, _dev->get_dispatch());
}

void vge::DepthRenderPass::fillClearValues(std::vector<vk::ClearValue>& clearValues)
{
	vk::ClearDepthStencilValue cdps(1.0f, 0);
	clearValues.push_back(cdps);
}
