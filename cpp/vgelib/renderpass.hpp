#pragma once

namespace vge {
	class Framebuffer;

	class RenderPass : public Disposable {
		friend class Command;
	public:
		RenderPass(const Device* dev, bool depthAttachment, AttachmentInfo* attachments, size_t attachmentCount);
		void NewFrameBuffer(ImageView** attachments, size_t attachments_len, Framebuffer*& fb);
		void NewNullFrameBuffer(uint32_t width,uint32_t height, Framebuffer*& fb);
		void NewFrameBuffer2(uint32_t width, uint32_t height, uint32_t layers, void** attachments, size_t attachments_len, Framebuffer*& fb);
		void init();
		const vk::RenderPass get_handle() const {
			return _renderPass;
		}
		uint32_t get_color_attachment_count() const {
			return static_cast<uint32_t>(_attachments.size() - (_depthAttachment ? 1 : 0));
		}
	protected:
		virtual void Dispose() override;
		vk::RenderPass _renderPass;
		std::vector<vk::AttachmentDescription> _attachments;
		std::vector<vk::ClearValue> _clearValues;
		const bool _depthAttachment;

		const Device * const _dev;
	};

	/*
	class ForwardRenderPass : public RenderPass {
	public:
		ForwardRenderPass(const Device* dev, vk::ImageLayout endLayout, vk::Format mainImageFormat, vk::Format depthImageFormat ) : 
			RenderPass(dev), _endLayout(endLayout), _mainImageFormat(mainImageFormat), _depthImageFormat(depthImageFormat) {
			
		}

		virtual void Init() override;

	private:
		virtual void fillClearValues(std::vector<vk::ClearValue>& clearValues) override;
		virtual uint32_t get_color_attachment_count() override;
		vk::ImageLayout _endLayout;
		vk::Format _mainImageFormat;
		vk::Format _depthImageFormat;
	};

	class DepthRenderPass : public RenderPass {
	public:
		DepthRenderPass(const Device* dev, vk::ImageLayout endLayout, vk::Format depthImageFormat) :
			RenderPass(dev), _endLayout(endLayout), _depthImageFormat(depthImageFormat) {

		}

		virtual void Init() override;

	private:
		virtual void fillClearValues(std::vector<vk::ClearValue>& clearValues) override;
		virtual uint32_t get_color_attachment_count() override {
			return 0;
		}
		vk::ImageLayout _endLayout;
		vk::Format _depthImageFormat;
	};
	*/

	class Framebuffer: public Disposable {
		friend class RenderPass;
	public:
		vk::Framebuffer get_framebuffer() {
			return _framebuffer;
		}

		vk::Extent2D get_extent() const {
			return _extent;
		};

		
	private:
		virtual void Dispose() override;
		Framebuffer(const Device *dev, vk::Framebuffer frameBuffer, const std::vector<ImageView*> attachments) : _dev(dev), _framebuffer(frameBuffer), _attachments(attachments) {
			auto mainDesc = _attachments[0]->get_image()->get_desc();
			_extent = vk::Extent2D(mainDesc.Width, mainDesc.Height);
		}
		Framebuffer(const Device* dev, vk::Framebuffer frameBuffer, uint32_t width, uint32_t height) : _dev(dev), _framebuffer(frameBuffer) {
			_extent = vk::Extent2D(width, height);
		}
		std::vector<ImageView*> _attachments;
		vk::Framebuffer _framebuffer;
		const Device* const _dev;
		vk::Extent2D _extent;
	};
}
