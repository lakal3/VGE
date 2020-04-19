#pragma once
#include <thread>
#include <mutex>
#include <deque>
#include <functional>

struct GLFWwindow;

namespace vge {
	
	class Desktop : public External, public InstanceExtension, public DeviceExtension {
		friend struct Static;
	public:
		void queueAction(std::function<void()> action);
		void pushEvent(RawEvent ev);
		void CreateWindow(char* title, size_t title_len, WindowPos *pos, Window*& window);
		void PullEvent(RawEvent* ev);
		void GetKeyName(uint32_t keyCode, uint8_t* name, size_t name_len, uint32_t &strLen);
		void GetMonitor(uint32_t monitor, WindowPos* info);
	private:
		Desktop(Application *app): _app(app) {
			_app->addInstanceExtension(this);
			_app->addDeviceExtension(this);
		}
		void init();

		static void runThread(Desktop* desktop);
		bool runActions();
		// Instance extension
		virtual void dispose() override;
		virtual void prepare(vk::InstanceCreateInfo& ici, std::vector<const char*>& layers, std::vector<const char*>& extensions) override;
		virtual void attach(Instance* inst) override;
		virtual void detach(Instance* inst) override;
		// device extensions
		virtual void isValid(Instance* inst, vk::PhysicalDevice pd, bool& valid, std::string& invalidReason) override;
		virtual void prepare(vk::DeviceCreateInfo& dci, std::vector<const char*>& extensions) override;
		virtual void attach(Device* inst) override;
		virtual void detach(Device* inst) override;


		std::thread* winLoop = nullptr;
		std::mutex updMutex;
		std::deque<std::function<void()>> actions;
		std::deque<RawEvent> events;
		Application* const _app;
		bool stopped = false;
	};

	class PresentInfo : SubmitInfo {
		friend class Window;
	public:
		int imageIndex = -1;
		vk::SwapchainKHR swapchain;
	private:
		PresentInfo(Device *dev);
		~PresentInfo();
		virtual void prepare(vk::SubmitInfo& si, std::vector<vk::Semaphore>& waitFor, std::vector<vk::PipelineStageFlags>& waitAt) override;
		virtual void submitted(Queue* queue) override;


		Device* _dev;
		vk::Semaphore aquire;
		vk::Semaphore present;
		// vk::PipelineStageFlags st = vk::PipelineStageFlagBits::eTopOfPipe;
	};

	class Window : public Disposable {
		friend class Desktop;
	public:
		void PrepareSwapchain(Device *dev, ImageDescription* viewDesc, int32_t &imageCount);
		void GetNextFrame(Image*& image, SubmitInfo *&submitInfo, int32_t& viewIndex);
		void SetPos(WindowPos *position);
		void GetPos(WindowPos* position);
	private:
		Window(Desktop* desktop, const std::string title, WindowPos position) :_desktop(desktop), _title(title), _initialPosition(position) {

		}
		void init();
		void createSwapchain(Device* dev);
		void resetSwapchain();
		void setPos(const WindowPos &position);
		void getPos(WindowPos* position);
		virtual void Dispose() override;

		static void key_callback(GLFWwindow* window, int key, int scancode, int action, int mods);
		static void character_callback(GLFWwindow* window, unsigned int codepoint);
		static void mouseclick_callback(GLFWwindow*, int, int, int);
		static void mousemove_callback(GLFWwindow*, double, double);
		static void mousescroll_callback(GLFWwindow*, double, double);
		static void window_close_callback(GLFWwindow* window);
		static void window_resize_callback(GLFWwindow* window, int x, int y);
		std::string _title;
		WindowPos _initialPosition;
		Desktop* const _desktop;
		GLFWwindow* _win = nullptr;
		vk::SurfaceKHR _surface;
		vk::SwapchainKHR _swapChain;
		vk::SwapchainCreateInfoKHR _crInfo;
		std::vector<Image*> _images;
		std::vector<PresentInfo*> _presentInfos;
		int _presentIndex = 0;
		Device *_dev;
		bool _initialized = false;
	};
}
