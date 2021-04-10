#pragma once

#include <vector>
#include <string>

#ifdef _WIN32 
#include <Windows.h>
#pragma warning( disable : 26495)
#endif

#include <vulkan/vulkan.hpp>

extern thread_local std::string lastError;

const uint64_t MaxTimeout = 2000000000;

namespace vge {
    class Instance;
    class PhysicalDevice;
    class Device;
    class Queue;
    class Buffer;
    class Image;
    class ImageView;
    class MemoryBlock;
    class Command;
    class DescriptorLayout;
    class DescriptorSet;
    class RenderPass;
    class ImageLoader;
    struct FontLoader;
    class Pipeline;
    class GraphicsPipeline;
    class ComputePipeline;
    class Sampler;
    class Desktop;
    class Window;
    class QueryPool;
    struct SubmitInfo;
}

#ifdef CreateWindow
#undef CreateWindow
#endif

#ifndef DISCARD
#define DISCARD(x) static_cast<void>(x)
#endif

#include "vgelib/vgelib_if.hpp"
#include "vgelib/app.hpp"
#include "vgelib/memory.hpp"
#include "vgelib/descriptor.hpp"
#include "vgelib/renderpass.hpp"
#include "vgelib/image.hpp"
#include "vgelib/pipeline.hpp"
#include "vgelib/command.hpp"
#include "vgelib/desktop.hpp"

#ifndef DLLEXPORT
#ifdef _WIN32
#define DLLEXPORT __declspec(dllexport)
#else
#define DLLEXPORT
#endif
#endif



