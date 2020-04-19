#pragma once

namespace vge {
	class Framebuffer;

	class RenderPass : public Disposable {
		friend class Command;
	public:
		void NewFrameBuffer(ImageView** attachments, size_t attachments_len, Framebuffer*& fb);
		virtual void Init() = 0;
		const vk::RenderPass get_handle() const {
			return _renderPass;
		}
		virtual uint32_t get_color_attachment_count() = 0;
	protected:
		RenderPass(const Device* dev) :_dev(dev) {

		}
		virtual void fillClearValues(std::vector<vk::ClearValue> &clearValues) = 0;
		virtual void Dispose() override;
		vk::RenderPass _renderPass;
		const Device * const _dev;
	};

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

	class Framebuffer: public Disposable {
		friend class RenderPass;
	public:
		vk::Framebuffer get_framebuffer() {
			return _framebuffer;
		}

		vk::Extent2D get_extent() const {
			auto mainDesc = _attachments[0]->get_image()->get_desc();
			return vk::Extent2D(mainDesc.Width, mainDesc.Height);
		};

		const std::vector<ImageView*> &get_attachments() const {
			return _attachments;
		};
	private:
		virtual void Dispose() override;
		Framebuffer(const Device *dev, vk::Framebuffer frameBuffer, const std::vector<ImageView*> attachments) : _dev(dev), _framebuffer(frameBuffer), _attachments(attachments) {

		}
		std::vector<ImageView*> _attachments;
		vk::Framebuffer _framebuffer;
		const Device* const _dev;
	};
}
