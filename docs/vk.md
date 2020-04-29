# VGE core object (Module vk)

This module contains the lower levels of VGE objects written in Go. 
The Vk module is also the only module that directly calls C++ library VGELib.dll (libVGELib.so)

Main objects in code module:

## Application

The Application object is the root object in Vulkan. You must create an application object before you
can access actual Vulkan devices.

Vulkan has a lot of extensions and extra layers. The required extensions must be loaded when we create
an application or a device. In VGE, all supported extensions must be added before initializing application.

From Application you can list the available Physical devices (DeviceInfo). 
The VGE will also indicate in DeviceInfo if some of the available devices are not supported.

Cube and WebView examples shows how to create an Application and Device using vk module.
 
## Device

Device is created using device index (0 when you have only one device).
 
On most laptops with dedicated GPU, you also have an integrated GPU. 
On those devices, you should implement some logic to select the proper device. 
In most VGE examples, you can select the device index from the command line.

The Device is used to create nearly all the other Vulkan resources. You must have a Device to create those resources.

## Ownership and resource lifetime   

Nearly all VGE resources are backed up by Vulkan resources. Those resources do not enjoy the luxury of automatic reclamation when
the resource is no longer used (Garbage collection). Instead, you must manually dispose them when no longer used. 
All structs implementing the Dispose method must be disposed manually.

To support disposing the VGE library, there are some helpers methods:

#### Get(ctx APIContext, key Key, cons Constructor) interface{}

Several structs that implement the Get method. The Get method allows access to a resource by a key. 
If the resource is already available, it is just returned. If the resource is not available, 
it will be created using the Constructor function. 
If the constructed object is a Disposable, the object implementing Get will own this new object and automatically dispose the object when it is disposed.

#### something.NewXX

If any object implements the NewXXX method (like dev.NewSampler), that object is the owner of the newly created object and will dispose the new object when the object itself is disposed.

#### Owner struct

You can use the existing struct Owner to implement a Get method in your own classes that want to support ownership.

## APIContext

Typically, error handling in Go uses return values. However, in Vulkan, every function could fail in an unexpected way. 
There are typically errors in the program, drivers or hardware and these errors should not normally occur. 

Therefore, using an error return value is not the preferable option. 
As always, panicing inside library is also bit problematic. Therefore, VGE uses the APIContent interface for most of the API calls. 
If we encounter an error we will call the APIContext SetError and the context itself can decide how to handle the error. Typically, VGE examples will
use log.Fatal to terminate but your solutions might use a different approach.

_The WebView example uses panic if the web request fails, so it can try to serve another request after the failure._

## Allocating memory in VGE

Memory allocation in Vulkan is a bit tricky. First, you must indicate what kind of memory (image or buffer) you require, 
how you are going to use it and should it be placed in the device or host memory.
Then you must query what kind of memory pools there are and allocate memory slices from the appropriate memory pools.

Unfortunately, memory pools are a limited resource and you should allocate as much as possible from a single pool. Otherwise you will run out of memory pool handles.

In VGE allocating works like this:
1. You create a memory Pool
2. You register all allocations using ReserveBuffer or ReserveImage
3. You allocate all reserved buffers and images.
 VGE will allocate the minimum number of required memory pools to satisfy all reservations.
 
Also note that you cannot dispose individual images or buffers. You must dispose the whole pool at once.

## RenderPasses

Vulkan handles rendering in render passes. See [https://vulkan-tutorial.com/Drawing_a_triangle/Graphics_pipeline_basics/Render_passes].

Render passes are fairly complex to setup. So the VGE offers a few prebuilt render passes with some options. 
If you need a new kind of render pass, you must implement that in VGELib (C++ part).

#### Standard render passes

- ForwardRenderPass with or without depth buffer. Forward render pass supports one output (main image)
- DepthRenderPass supports rendering with depth buffer only. Used in shadow rendering.  

## Pipelines

Pipelines are used to combine shaders (vertex, frag), inputs and descriptors sets into single "runnable" unit. 
For examples, see different materials (vge/materials/...) on how to build up pipelines.

_Pipelines need compiled SPIR-V shader modules. It is possible to load those from a hard drive. 
However, the VGE has also tool called packspv that will convert SPIR-V files to Go code that compiles the SPIR-V code
directly into binary. All VGE materials use this tool to combine SPIR-V code to the module itself._
   
 
## Other resources in vk

See API documentation for more details about other resources:
- Sampler is used to sample images 
- Command are used to record and submit command to Vulkan. In Vulkan API we have command pools and commands. VGE combines those into one entity. 
- DescriptorLayout describes a layout of a shader binding slots. See [https://vulkan.lunarg.com/doc/view/1.0.49.0/windows/tutorial/html/08-init_pipeline_layout.html]
- DescriptorPool is used to allocate DescriptorSets. In VGE, each DescriptorPool supports only one kind of descriptors.
- DescriptorSet is a collection of individual descriptors that you can write values to (uniforms, sampled images etc..). Later you can bind descriptors to recorded commands.
Shaders access values and images through these sets.
- Desktop and Window handle the operating system window creation and representing the rendered image on an operating system window
- ImageView and ImageRange allow Vulkan to use a subrange of Image. Most Vulkan API calls use ImageView, not Image
- RenderCache is a helper to handle per instance and per frame resources needs during rendering

A nice overview of Vulkan objects [https://gpuopen.com/understanding-vulkan-objects/] might make it easier to understand how Vulkan objects are related. 

## Concurrent access

Vulkan is multi-threaded. However, most of the vk and Vulkan objects are not thread safe. You must use mutex or other mechanism to synchronize access.

This means that you can render two separate images concurrently without synchronization using different set of Vulkan objects. 
You cannot use the same object in multiple coroutines without external synchronization.
 
The VGE API documentation will indicate if some method is safe for concurrent access.

All constructor functions (such as NewXXX) are safe for concurrent calls.



 
   

   

   