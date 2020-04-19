#include "vgelib/vgelib.hpp"

#define GLFW_INCLUDE_NONE 
#include <GLFW/glfw3.h>
#include <chrono>
#include <atomic>

void vge::Desktop::runThread(vge::Desktop* vDesktop) {
	while (!vDesktop->stopped) {
		auto found = vDesktop->runActions();
		if (!found) {
			glfwPollEvents();
			std::this_thread::sleep_for(std::chrono::milliseconds(1));
		}
	}
}

void vge::Desktop::queueAction(std::function<void()> action)
{
	std::unique_lock<std::mutex> l(updMutex);
	actions.push_back(action);
	glfwPostEmptyEvent();
}

void vge::Desktop::pushEvent(RawEvent ev)
{
	std::unique_lock<std::mutex> l (updMutex);
	events.push_back(ev);
}

void vge::Desktop::CreateWindow(char* title, size_t title_len, WindowPos *pos, Window *&window)
{
	std::string sTitle(title, title_len);
	window = new Window(this, sTitle, *pos);
	window->init();
}

void vge::Desktop::PullEvent(RawEvent* ev)
{
	std::unique_lock<std::mutex> l(updMutex);
	bool found = events.size() > 0;
	if (found) {
		*ev = events.front();
		events.pop_front();
	} else {
		ev->eventType = EventType::Nil;
	}
}

void vge::Desktop::GetKeyName(uint32_t keyCode, uint8_t* name, size_t name_len, uint32_t &strLen)
{
	auto n = glfwGetKeyName(keyCode, 0);
	if (n == nullptr) {
		strLen = 0;
		return;
	}
	strncpy((char *) name, n, name_len);
	uint32_t l = static_cast<uint32_t>(strlen(n));
	if (l >= name_len) {
		l = static_cast<uint32_t>(name_len - 1);
	}
	strLen = l;
}

void vge::Desktop::GetMonitor(uint32_t monitor, WindowPos* info)
{
	int count = 0;
	auto monitors = glfwGetMonitors(&count);
	if (static_cast<uint32_t>(count) <= monitor) {
		info->height = 0;
		info->width = 0;
		return;
	}
	int xpos, ypos, width, height;
	glfwGetMonitorWorkarea(monitors[monitor], &xpos, &ypos, &width, &height);
	info->height = height;
	info->width = width;
	info->left = xpos;
	info->top = ypos;

}

void vge::Desktop::init()
{
	if (!glfwInit()) {
		throw std::runtime_error("GLFW init failed");
	}
	winLoop = new std::thread(runThread, this);
}

void vge::Desktop::dispose()
{
	stopped = true;
	if (winLoop != nullptr) {
		glfwPostEmptyEvent();
		winLoop->join();
		delete winLoop;
		winLoop = nullptr;
	}
	glfwTerminate();
	delete this;
}

void vge::Desktop::prepare(vk::InstanceCreateInfo& ici, std::vector<const char*>& layers, std::vector<const char*>& extensions)
{
	uint32_t count;
	const char** exts = glfwGetRequiredInstanceExtensions(&count);
	for (uint32_t idx = 0; idx < count; idx++) {
		extensions.push_back(exts[idx]);
	}
}

void vge::Desktop::attach(Instance* inst)
{
}

void vge::Desktop::detach(Instance* inst)
{
}

void vge::Desktop::isValid(Instance* inst, vk::PhysicalDevice pd, bool& valid, std::string& invalidReason)
{
	if (!DeviceExtension::checkExtension(inst, pd, "VK_KHR_swapchain")) {
		valid = false;
		invalidReason = "No VK_KHR_swapchain in device";
		return;
	}
	valid = true;
}

void vge::Desktop::prepare(vk::DeviceCreateInfo& dci, std::vector<const char*>& extensions)
{
	extensions.push_back("VK_KHR_swapchain");
}

void vge::Desktop::attach(Device* inst)
{
}

void vge::Desktop::detach(Device* inst)
{
}

bool vge::Desktop::runActions()
{
	bool found = false;
	std::function<void()> action;
	updMutex.lock();
	if (actions.size() > 0) {
		found = true;
		action = actions.front();
		actions.pop_front();
	}
	updMutex.unlock();
	if (found) {
		action();
		return true;
	}
	return false;
}

void vge::Window::PrepareSwapchain(Device* dev, ImageDescription* imageDesc, int32_t& imageCount)
{
	if (_win == nullptr) {
		imageCount = 0;
		return;
	}
	if (!_swapChain) {
		createSwapchain(dev);
	}
	imageDesc->Depth = 1;
	imageDesc->Height = _crInfo.imageExtent.height;
	imageDesc->Width = _crInfo.imageExtent.width;
	imageDesc->Layers = 1;
	imageDesc->Format = _crInfo.imageFormat;
	imageDesc->MipLevels = 1;
	imageCount = static_cast<int32_t>(_images.size());
}

void vge::Window::GetNextFrame(Image*& image, SubmitInfo*& submitInfo, int32_t& viewIndex)
{
	if (_images.size() == 0) {
		viewIndex = -1;
		return;
	}
	auto pr = _presentInfos[_presentIndex];
	if (++_presentIndex >= _presentInfos.size()) {
		_presentIndex = 0;
	}
	try {
		auto val = _dev->get_device().acquireNextImageKHR(_swapChain, MaxTimeout, pr->aquire, nullptr, _dev->get_dispatch());

		if (val.result == vk::Result::eSuboptimalKHR || val.result == vk::Result::eErrorOutOfDateKHR) {
			resetSwapchain();
			viewIndex = -1;
			return;
		}
		viewIndex = val.value;
	} catch (const vk::OutOfDateKHRError&) {
		resetSwapchain();
		viewIndex = -1;
		return;
	}
	
	image = _images[viewIndex];
	pr->swapchain = _swapChain;
	pr->imageIndex = viewIndex;
	submitInfo = pr;
}

void vge::Window::SetPos(WindowPos* position)
{
	std::atomic_flag done;
	done.test_and_set();
	_desktop->queueAction([this, position, &done] {
		setPos(*position);
		done.clear();
	});

	while (done.test_and_set()) {
		std::this_thread::yield();
	}
}

void vge::Window::GetPos(WindowPos* position)
{
	std::atomic_flag done;
	done.test_and_set();
	_desktop->queueAction([this, position, &done] {
		getPos(position);
		done.clear();
	});


	while (done.test_and_set()) {
		std::this_thread::yield();
	}
}


void vge::Window::init()
{
	std::atomic_flag done;
	done.test_and_set();
	_desktop->queueAction([this,&done] {
		_initialized = true;
		glfwWindowHint(GLFW_CLIENT_API, GLFW_NO_API);
		glfwWindowHint(GLFW_DECORATED, (_initialPosition.state & WindowState::Borderless) != 0 ? GLFW_FALSE : GLFW_TRUE);
		glfwWindowHint(GLFW_RESIZABLE, (_initialPosition.state & WindowState::Fixed) != 0 ? GLFW_FALSE : GLFW_TRUE);
		_win = glfwCreateWindow(_initialPosition.width, _initialPosition.height, _title.c_str(), nullptr, nullptr);
		glfwSetWindowUserPointer(_win, this);
		glfwSetKeyCallback(_win, key_callback);
		glfwSetCharCallback(_win, character_callback);
		glfwSetScrollCallback(_win, mousescroll_callback);
		glfwSetCursorPosCallback(_win, mousemove_callback);
		glfwSetMouseButtonCallback(_win, mouseclick_callback);
		glfwSetWindowCloseCallback(_win, window_close_callback);
		glfwSetWindowSizeCallback(_win, window_resize_callback);
		setPos(_initialPosition);
		done.clear();
	});
	while (done.test_and_set()) {
		std::this_thread::yield();
	}
}

void vge::Window::createSwapchain(Device *dev)
{
	vk::SurfaceCapabilitiesKHR sfCap;
	if (_crInfo.imageFormat == vk::Format::eUndefined) {
		_dev = dev;
		VkSurfaceKHR surface;
		glfwCreateWindowSurface(_dev->get_instance()->get_instance(), _win, nullptr, &surface);
		_surface = surface;
		auto supported = dev->get_pd().getSurfaceSupportKHR(_dev->get_graphicQueueFamily(), surface, _dev->get_dispatch());
		if (!supported) {
			throw std::runtime_error("Surface not supported by this device!");
		}
		sfCap = dev->get_pd().getSurfaceCapabilitiesKHR(_surface, dev->get_dispatch());
		uint32_t sfCount = 0;
		dev->get_pd().getSurfaceFormatsKHR(_surface, &sfCount, nullptr, dev->get_dispatch());
		std::vector<vk::SurfaceFormatKHR> formats(sfCount);
		dev->get_pd().getSurfaceFormatsKHR(_surface, &sfCount, formats.data(), dev->get_dispatch());
		if (formats[0].format == vk::Format::eUndefined) {
			_crInfo.imageFormat = vk::Format::eR8G8B8A8Unorm;			
		} else {
			_crInfo.imageFormat = formats[0].format;
		}
		_crInfo.imageColorSpace = formats[0].colorSpace;
		_crInfo.minImageCount = sfCap.minImageCount + 1;
		if (_crInfo.minImageCount > sfCap.maxImageCount) {
			_crInfo.minImageCount = sfCap.maxImageCount;
		}
		_crInfo.imageArrayLayers = 1;
		_crInfo.compositeAlpha = vk::CompositeAlphaFlagBitsKHR::eOpaque;
		_crInfo.presentMode = vk::PresentModeKHR::eFifo;
		_crInfo.imageUsage = vk::ImageUsageFlagBits::eColorAttachment | vk::ImageUsageFlagBits::eTransferSrc;
		_crInfo.clipped = 1;
		_crInfo.surface = _surface;
		_crInfo.preTransform = sfCap.currentTransform;
		for (uint32_t idx = 0; idx < _crInfo.minImageCount + 1; idx++) {
			_presentInfos.push_back(new PresentInfo(dev));
		}
	} else {
		sfCap = dev->get_pd().getSurfaceCapabilitiesKHR(_surface, dev->get_dispatch());
	}
	// int w, h;
	_crInfo.imageExtent.height = static_cast<uint32_t>(sfCap.currentExtent.height);
	_crInfo.imageExtent.width = static_cast<uint32_t>(sfCap.currentExtent.width);
	_crInfo.imageSharingMode = vk::SharingMode::eExclusive;
	{
		SuppressValidation sv;
		_swapChain = dev->get_device().createSwapchainKHR(_crInfo, allocator, dev->get_dispatch());
	}
	uint32_t imCount = 0;
	dev->get_device().getSwapchainImagesKHR(_swapChain, &imCount, nullptr, dev->get_dispatch());
	std::vector<vk::Image> images(imCount);
	dev->get_device().getSwapchainImagesKHR(_swapChain, &imCount, images.data(), dev->get_dispatch());

	ImageDescription desc;
	desc.Depth = 1;
	desc.Height = _crInfo.imageExtent.height;
	desc.Width = _crInfo.imageExtent.width;
	desc.Layers = 1;
	desc.Format = _crInfo.imageFormat;
	desc.MipLevels = 1;

	for (uint32_t idx = 0; idx < imCount; idx++) {
		_images.push_back(new Image(dev, images[idx], _crInfo.imageUsage, desc));
	}
	_presentIndex = 0;
}

void  vge::Window::resetSwapchain() {
	if (!!_swapChain) {
		for (auto img : _images) {
			img->Dispose();
		}
		_images.clear();
		_dev->get_device().destroySwapchainKHR(_swapChain, allocator, _dev->get_dispatch());
		_swapChain = nullptr;
		
	}
}

void vge::Window::setPos(const WindowPos& position)
{
	
	if (_win == nullptr) {
		return;
	}
	if (position.left >= 0 && position.top >= 0) {
		glfwSetWindowPos(_win, position.left, position.top);
	}
	if (position.height > 0 && position.width > 0) {
		glfwSetWindowSize(_win, position.width, position.height);
	}
	switch (position.state & WindowState::Modes) {
	case WindowState::Normal:
		glfwShowWindow(_win);
		glfwRestoreWindow(_win);
		break;
	case WindowState::Hidden:
		glfwHideWindow(_win);
		break;
	case WindowState::Minimized:
		glfwShowWindow(_win);
		glfwIconifyWindow(_win);
		break;
	case WindowState::Maximized:
		glfwShowWindow(_win);
		glfwMaximizeWindow(_win);
		break;
	default:
		break;
	}
}

void vge::Window::getPos(WindowPos* position)
{
	if (_win == nullptr) {
		return;
	}
	int x, y;
	glfwGetWindowPos(_win, &x, &y);
	position->left = x;
	position->top = y;
	glfwGetWindowSize(_win, &x, &y);
	position->width = x;
	position->height = y;
	if (glfwGetWindowAttrib(_win, GLFW_VISIBLE)) {
		position->state = WindowState::Normal;
		if (glfwGetWindowAttrib(_win, GLFW_ICONIFIED)) {
			position->state = WindowState::Minimized;
		}
		if (glfwGetWindowAttrib(_win, GLFW_MAXIMIZED)) {
			position->state = WindowState::Maximized;
		}
	} else {
		position->state = WindowState::Hidden;
	}
}

void vge::Window::Dispose()
{
	if (_win != nullptr) {
		resetSwapchain();
		for (auto pre : _presentInfos) {
			delete pre;
		}
		auto wTmp = _win;
		_presentInfos.clear();
		_desktop->queueAction([=] {
			glfwDestroyWindow(wTmp);
			});
		_win = nullptr;
		_dev->get_instance()->get_instance().destroySurfaceKHR(_surface, allocator, _dev->get_instance()->get_dispatch());
	}
}

void vge::Window::key_callback(GLFWwindow* window, int key, int scancode, int action, int mods)
{
	Window *win = reinterpret_cast<Window *>(glfwGetWindowUserPointer(window));
	RawEvent ev = { action == GLFW_RELEASE ? EventType::KeyUp : EventType::KeyDown, scancode, key };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

void vge::Window::character_callback(GLFWwindow* window, unsigned int codepoint)
{
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { EventType::Char, static_cast<int32_t>(codepoint) };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

void vge::Window::mouseclick_callback(GLFWwindow* window, int button, int action, int mods)
{
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { action == GLFW_RELEASE ? EventType::MouseUp : EventType::MouseDown, button };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

void vge::Window::mousemove_callback(GLFWwindow* window, double xpos, double ypos)
{
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { EventType::MouseMove, static_cast<int32_t>(xpos), static_cast<int32_t>(ypos) };
	ev.win = win;
	win->_desktop->pushEvent(ev);

}

void vge::Window::mousescroll_callback(GLFWwindow* window, double xpos, double ypos)
{
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { EventType::MouseScroll, static_cast<int32_t>(xpos), static_cast<int32_t>(ypos) };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

void vge::Window::window_close_callback(GLFWwindow* window)
{
	glfwSetWindowShouldClose(window, GLFW_FALSE);
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { EventType::CloseWindow };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

void vge::Window::window_resize_callback(GLFWwindow* window, int x, int y)
{
	Window* win = reinterpret_cast<Window*>(glfwGetWindowUserPointer(window));
	RawEvent ev = { EventType::ResizeWindow,  static_cast<int32_t>(x), static_cast<int32_t>(y) };
	ev.win = win;
	win->_desktop->pushEvent(ev);
}

vge::PresentInfo::PresentInfo(Device* dev): _dev(dev) {
	vk::SemaphoreCreateInfo sci;
	aquire = dev->get_device().createSemaphore(sci, allocator, _dev->get_dispatch());
	present = dev->get_device().createSemaphore(sci, allocator, _dev->get_dispatch());
}

vge::PresentInfo::~PresentInfo()
{
	_dev->get_device().destroySemaphore(aquire, allocator, _dev->get_dispatch());
	_dev->get_device().destroySemaphore(present, allocator, _dev->get_dispatch());
}

void vge::PresentInfo::prepare(vk::SubmitInfo& si, std::vector<vk::Semaphore>& waitFor, std::vector<vk::PipelineStageFlags>& waitAt)
{
	waitFor.push_back(aquire);
	waitAt.push_back(vk::PipelineStageFlagBits::eTopOfPipe);
	si.signalSemaphoreCount = 1;
	si.pSignalSemaphores = &present;
}

void vge::PresentInfo::submitted(Queue* queue)
{
	vk::PresentInfoKHR pi;
	pi.waitSemaphoreCount = 1;
	pi.pWaitSemaphores = &present;
	pi.swapchainCount = 1;
	pi.pSwapchains = &swapchain;
	uint32_t ii = static_cast<uint32_t>(imageIndex);
	pi.pImageIndices = &ii;
	try {
		auto val = queue->get_queue().presentKHR(pi, _dev->get_dispatch());
	} catch (const vk::OutOfDateKHRError &) {
		// Ignore this
	}

}
