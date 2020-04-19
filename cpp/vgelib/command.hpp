#pragma once

#include <mutex>

namespace vge {
	
	struct SubmitInfo {
		virtual void prepare(vk::SubmitInfo& si, std::vector<vk::Semaphore> &waitFor, std::vector<vk::PipelineStageFlags> &waitAt) = 0;
		virtual void submitted(Queue* queue) = 0;
	};

	struct WaitForCmd : public SubmitInfo {
		vk::Semaphore semWait;
		vk::PipelineStageFlags at;
		virtual void prepare(vk::SubmitInfo& si, std::vector<vk::Semaphore>& waitFor, std::vector<vk::PipelineStageFlags>& waitAt) override;
		virtual void submitted(Queue* queue) override;
	};

	class Command : Disposable {
		friend class Device;
		friend class Queue;
	public:
		virtual void Dispose() override;
		void Begin();
		void CopyBuffer(Buffer* fromBuffer, Buffer* toBuffer);
		void BeginRenderPass(RenderPass* rp, Framebuffer* fb);
		void EndRenderPass();
		void SetLayout(Image* view, vge::ImageRange* range, vk::ImageLayout layout);
		void CopyBufferToImage(Buffer* src, Image* dst, vge::ImageRange* range, size_t offset);
		void CopyImageToBuffer(Image* src, Buffer* dst, vge::ImageRange* range, size_t offset);
		void Draw(DrawItem* draws, size_t draws_len);
		void Compute(ComputePipeline* pl, uint32_t x, uint32_t y, uint32_t z, DescriptorSet** descriptors, size_t descriptors_len);
		void Wait();
	private:
		Command(Device* dev, uint32_t family, bool once) : _dev(dev), _family(family), _once(once) {

		}
		void copyView(Buffer* buffer, Image* image, ImageRange* range, size_t offset, bool copyToImage);
		void drawOne(DrawItem &draw, Pipeline*& pipeline);
		void init();
		Device* _dev;
		WaitForCmd* _waitSem = nullptr;
		vk::Fence _fence;
		const uint32_t _family;
		vk::CommandPool _cp;
		vk::CommandBuffer _cmd;
		const bool _once;
	};

	

	class Queue : External {
		friend class Device;
	public:
		vk::Queue get_queue() const {
			return _queue;
		}
	private:
		Queue(Device* dev, vk::QueueFlags flags, uint32_t family, uint32_t index) :_dev(dev), _flags(flags), _family(family), _index(index) {

		}
		~Queue();
	
		void init();

		
		void submit(Command *cmd, SubmitInfo** info, size_t info_len, vk::PipelineStageFlags waitForStage, SubmitInfo*& waitFor);

		Device* _dev;
		const vk::QueueFlags _flags;
		const uint32_t _family;
		const uint32_t _index;
		vk::Queue _queue;

		vk::Semaphore _presentSem;
		
	};
}