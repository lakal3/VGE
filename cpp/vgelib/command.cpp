#include "vgelib/vgelib.hpp"

vge::Queue::~Queue()
{
	auto dev = _dev->get_device();
	dev.destroySemaphore(_presentSem, allocator, _dev->get_dispatch());
	_queue.waitIdle(_dev->get_dispatch());
}

void vge::Queue::init()
{

	auto dev = _dev->get_device();
	_queue = dev.getQueue(_family, _index, _dev->get_dispatch());

	vk::SemaphoreCreateInfo sci;
	_presentSem = dev.createSemaphore(sci, allocator, _dev->get_dispatch());
}


void vge::Queue::submit(Command *cmd, SubmitInfo **info, size_t info_len, vk::PipelineStageFlags waitForStage, SubmitInfo*& waitFor)
{
	cmd->_cmd.end(_dev->get_dispatch());
	vk::SubmitInfo si;
	si.pCommandBuffers = &cmd->_cmd;
	si.commandBufferCount = 1;
	std::vector<vk::Semaphore> waitSems;
	std::vector<vk::PipelineStageFlags> waitAts;
	for (size_t idx = 0; idx < info_len; idx++) {
		info[idx]->prepare(si, waitSems, waitAts);
	}
	if (waitSems.size() > 0) {
		si.waitSemaphoreCount = static_cast<uint32_t>(waitSems.size());
		si.pWaitDstStageMask = waitAts.data();
		si.pWaitSemaphores = waitSems.data();
	}

	if (!!waitForStage) {
		if (cmd->_waitSem == nullptr) {
			auto ws = new WaitForCmd();
			cmd->_waitSem = ws;
			vk::SemaphoreCreateInfo sci;
			ws->semWait = _dev->get_device().createSemaphore(sci, allocator, _dev->get_dispatch());
		}
		cmd->_waitSem->at = waitForStage;
		waitFor = cmd->_waitSem;
		si.pSignalSemaphores = &(cmd->_waitSem->semWait);
		si.signalSemaphoreCount = 1;
	} else {
		waitFor = nullptr;
	}
	DISCARD(_dev->get_device().resetFences(1, &(cmd->_fence), _dev->get_dispatch()));
	DISCARD(_queue.submit(1, &si, cmd->_fence, _dev->get_dispatch()));
	
	for (size_t idx = 0; idx < info_len; idx++) {
		info[idx]->submitted(this);
	}
}

void vge::Command::Dispose()
{
	if (!!_cp) {
		DISCARD(_dev->get_device().waitForFences(1, &_fence, 1, MaxTimeout, _dev->get_dispatch()));
		_dev->get_device().destroyFence(_fence, allocator, _dev->get_dispatch());
		if (_waitSem != nullptr) {
			_dev->get_device().destroySemaphore(_waitSem->semWait, allocator, _dev->get_dispatch());
			delete _waitSem;
			_waitSem = nullptr;
		}

		_dev->get_device().destroyCommandPool(_cp, allocator, _dev->get_dispatch());
		_cmd = nullptr;
		_cp = nullptr;
	}
	delete this;
}

void vge::Command::Begin()
{

	vk::CommandBufferBeginInfo cbbi;
	if (_once) {
		cbbi.flags = vk::CommandBufferUsageFlagBits::eOneTimeSubmit;
	}
	_cmd.begin(cbbi);
}


void vge::Command::CopyBuffer(Buffer* fromBuffer, Buffer* toBuffer)
{
	vk::BufferCopy region;
	region.size = fromBuffer->_size;
	if (toBuffer->_size < region.size) {
		region.size = toBuffer->_size;
	}
	_cmd.copyBuffer(fromBuffer->_buffer, toBuffer->_buffer, 1, &region, _dev->get_dispatch());
	_cmd.copyBuffer(fromBuffer->_buffer, toBuffer->_buffer, 1, &region, _dev->get_dispatch());
}

void vge::Command::BeginRenderPass(RenderPass* rp, Framebuffer* fb)
{
	vk::RenderPassBeginInfo rpbi;
	rpbi.clearValueCount = static_cast<uint32_t>(rp->_clearValues.size());
	if (rpbi.clearValueCount > 0) {
		rpbi.pClearValues = rp->_clearValues.data();
	}
	rpbi.framebuffer = fb->get_framebuffer();
	rpbi.renderPass = rp->_renderPass;
	auto ext2d = fb->get_extent();
	rpbi.renderArea.extent = fb->get_extent();
	_cmd.beginRenderPass(rpbi, vk::SubpassContents::eInline, _dev->get_dispatch());
	vk::Viewport vp;
	vp.maxDepth = 1.0;
	vp.width = static_cast<float>(ext2d.width);
	vp.height = static_cast<float>(ext2d.height);
	_cmd.setViewport(0, 1, &vp, _dev->get_dispatch());
	vk::Rect2D rc;
	rc.extent = ext2d;
	_cmd.setScissor(0, 1, &rc, _dev->get_dispatch());
}

void vge::Command::EndRenderPass()
{
	_cmd.endRenderPass(_dev->get_dispatch());
}

void vge::Command::SetLayout(Image *image, vge::ImageRange *range, vk::ImageLayout layout)
{
	vk::PipelineStageFlags src;
	vk::PipelineStageFlags dst;
	vk::ImageMemoryBarrier imb;
	imb.oldLayout = vk::ImageLayout(range->Layout);
	imb.newLayout = layout;
	imb.image = image->get_handle();
	imb.subresourceRange.baseArrayLayer = range->FirstLayer;
	imb.subresourceRange.baseMipLevel = range->FirstMipLevel;
	imb.subresourceRange.layerCount = range->LayerCount;
	imb.subresourceRange.levelCount = range->LevelCount;
	imb.subresourceRange.aspectMask = image->get_aspect();
	imb.dstQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED;
	imb.srcQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED;
	switch (imb.newLayout)
	{
	case vk::ImageLayout::eShaderReadOnlyOptimal:
		imb.dstAccessMask = vk::AccessFlagBits::eMemoryRead;
		dst = vk::PipelineStageFlagBits::eBottomOfPipe;
		src = vk::PipelineStageFlagBits::eTransfer;
	case vk::ImageLayout::eColorAttachmentOptimal:
		imb.dstAccessMask = vk::AccessFlagBits::eMemoryWrite;
		dst = vk::PipelineStageFlagBits::eBottomOfPipe;
		src = vk::PipelineStageFlagBits::eTransfer;
		break;
	case vk::ImageLayout::eTransferDstOptimal:
		imb.dstAccessMask = vk::AccessFlagBits::eTransferWrite;
		src = vk::PipelineStageFlagBits::eTopOfPipe;
		dst = vk::PipelineStageFlagBits::eTransfer;
		break;
	case vk::ImageLayout::eTransferSrcOptimal:
		src = vk::PipelineStageFlagBits::eTopOfPipe;
		imb.dstAccessMask = vk::AccessFlagBits::eTransferRead;
		dst = vk::PipelineStageFlagBits::eTransfer;
		break;
	case vk::ImageLayout::eGeneral:
		src = vk::PipelineStageFlagBits::eTopOfPipe;
		dst = vk::PipelineStageFlagBits::eTopOfPipe;
	}
	vk::DependencyFlags df;
	imb.subresourceRange.aspectMask = vk::ImageAspectFlagBits::eColor;
	_cmd.pipelineBarrier(src, dst, df, 0, nullptr, 0, nullptr, 1, &imb, _dev->get_dispatch());
}



uint32_t min1(uint32_t v) {
	if (v > 1) {
		return v;
	}
	return 1;
}

void vge::Command::CopyBufferToImage(Buffer* src, Image* dst, vge::ImageRange* range, size_t offset)
{
	copyView(src, dst, range, offset, true);
}

void vge::Command::CopyImageToBuffer(Image* src, Buffer* dst, vge::ImageRange* range, size_t offset)
{
	copyView(dst, src, range, offset, false);
}

void vge::Command::Draw(DrawItem* draws, size_t draws_len)
{
	Pipeline* prevpipeline = nullptr;
	for (size_t idx = 0; idx < draws_len; idx++) {
		drawOne(draws[idx], prevpipeline);
	}
}

void vge::Command::Compute(ComputePipeline* pl, uint32_t x, uint32_t y, uint32_t z, DescriptorSet** descriptors, size_t descriptors_len)
{
	_cmd.bindPipeline(vk::PipelineBindPoint::eCompute, pl->get_handle(), _dev->get_dispatch());
	for (size_t idx = 0; idx < descriptors_len; idx++) {
		auto ds = descriptors[idx];
		if (ds != nullptr) {
			auto dss = ds->get_descriptorSet();
			_cmd.bindDescriptorSets(vk::PipelineBindPoint::eCompute, pl->get_layout(), static_cast<uint32_t>(idx), 1, &dss, 0, nullptr, _dev->get_dispatch());
		}
	}
	_cmd.dispatch(x, y, z);
}

void vge::Command::Wait()
{
	DISCARD(_dev->get_device().waitForFences(1, &_fence, true, MaxTimeout, _dev->get_dispatch()));
}


void vge::Command::copyView(Buffer* buffer, Image* image, ImageRange* range, size_t offset, bool copyToImage)
{
	std::vector<vk::BufferImageCopy> bics;
	for (uint32_t mp = range->FirstMipLevel; mp < range->FirstMipLevel + range->LevelCount; mp++) {
		auto desc = image->get_desc();
		vk::BufferImageCopy bic;
		bic.bufferOffset = offset;
		bic.imageExtent.width = min1(desc.Width >> mp);
		bic.imageExtent.height = min1(desc.Height >> mp);
		bic.imageExtent.depth = min1(desc.Depth >> mp);
		bic.imageSubresource.baseArrayLayer = range->FirstLayer;
		bic.imageSubresource.mipLevel = mp;
		bic.imageSubresource.layerCount = range->LayerCount;
		bic.imageSubresource.aspectMask = image->get_aspect();
		bics.push_back(bic);
	}
	if (copyToImage) {
		_cmd.copyBufferToImage(buffer->get_buffer(), image->get_handle(), range->get_layout(), static_cast<uint32_t>(bics.size()), bics.data(), _dev->get_dispatch());
	}
	else {
		_cmd.copyImageToBuffer(image->get_handle(), range->get_layout(), buffer->get_buffer(),  static_cast<uint32_t>(bics.size()), bics.data(), _dev->get_dispatch());
	}
}

void vge::Command::drawOne(DrawItem &draw, Pipeline *&pipeline)
{
	if (draw.instances == 0) {
		return;
	}
	if (draw.pipeline != nullptr) {
		_cmd.bindPipeline(draw.pipeline->get_bindpoint(), draw.pipeline->get_handle(), _dev->get_dispatch());
		pipeline = draw.pipeline;
	}
	
	for (int idx = 0; idx < 8; idx++) {
		size_t offset = 0;
		if (draw.inputs[idx] != nullptr) {
			auto buf = draw.inputs[idx]->get_buffer();
			if (draw.indexed) {
				if (idx == 0) {
					_cmd.bindIndexBuffer(buf, 0, vk::IndexType::eUint32, _dev->get_dispatch());
				} else {
					_cmd.bindVertexBuffers(idx - 1, 1, &buf, &offset, _dev->get_dispatch());
				}
			} else {
				_cmd.bindVertexBuffers(idx, 1, &buf, &offset, _dev->get_dispatch());
			}
		}
	}
	for (int idx = 0; idx < 8; idx++) {
		size_t offset = 0;
		if (draw.descriptors[idx].set != nullptr) {
			auto ds = draw.descriptors[idx].set->get_descriptorSet();
			if (draw.descriptors[idx].hasOffset != 0) {
				auto offset = draw.descriptors[idx].offset;
				_cmd.bindDescriptorSets(pipeline->get_bindpoint(), pipeline->get_layout(), idx, 1, &ds, 1, &offset, _dev->get_dispatch());
			}
			else {
				_cmd.bindDescriptorSets(pipeline->get_bindpoint(), pipeline->get_layout(), idx, 1, &ds, 0, nullptr, _dev->get_dispatch());
			}
		}
	}

	if (draw.indexed) {
		_cmd.drawIndexed(draw.count, draw.instances, draw.from, 0, draw.fromInstance, _dev->get_dispatch());
	} else {
		_cmd.draw(draw.count, draw.instances, draw.from, draw.fromInstance, _dev->get_dispatch());
	}
}

void vge::Command::init()
{
	vk::CommandPoolCreateInfo cpci;
	cpci.queueFamilyIndex = _family;
	cpci.flags = _once ? vk::CommandPoolCreateFlagBits::eTransient : vk::CommandPoolCreateFlagBits::eResetCommandBuffer;
	_cp = _dev->get_device().createCommandPool(cpci, allocator, _dev->get_dispatch());
	vk::CommandBufferAllocateInfo cbai;
	cbai.commandBufferCount = 1;
	cbai.commandPool = _cp;
	auto cmds = _dev->get_device().allocateCommandBuffers(cbai, _dev->get_dispatch());
	_cmd = cmds[0];
	vk::FenceCreateInfo fci;
	_fence = _dev->get_device().createFence(fci, allocator, _dev->get_dispatch());

}

void vge::Command::WriteTimer(QueryPool* qp, vk::PipelineStageFlags stages, uint32_t timerIndex) {
	auto stage = static_cast<vk::PipelineStageFlagBits>(static_cast<VkMemoryMapFlags>(stages));
	_cmd.writeTimestamp(stage, qp->_handle, timerIndex, _dev->get_dispatch());
}

void vge::WaitForCmd::prepare(vk::SubmitInfo& si, std::vector<vk::Semaphore>& waitFor, std::vector<vk::PipelineStageFlags>& waitAt)
{
	waitFor.push_back(this->semWait);
	waitAt.push_back(this->at);
}

void vge::WaitForCmd::submitted(Queue* queue)
{
}

void vge::QueryPool::init() {
	vk::QueryPoolCreateInfo cqi;
	cqi.queryCount = _size;
	cqi.queryType = _queryType;
	_handle = _dev->get_device().createQueryPool(cqi, allocator, _dev->get_dispatch());
	Reset();
}

void vge::QueryPool::Get(uint64_t* values, size_t values_len, float &timestampPeriod)
{
	auto result =_dev->get_device().getQueryPoolResults(_handle, 0, static_cast<uint32_t>(values_len), 8 * values_len, static_cast<void*>(values), 8,
		vk::QueryResultFlagBits::e64, _dev->get_dispatch());
	timestampPeriod = _dev->get_pdProperties().limits.timestampPeriod;
	Reset();
}

void vge::QueryPool::Reset()
{
	// Not enabled by default
	// _dev->get_device().resetQueryPool(_handle, 0, _size, _dev->get_dispatch());
}

void vge::QueryPool::Dispose()
{
	_dev->get_device().destroyQueryPool(_handle, allocator, _dev->get_dispatch());
}
