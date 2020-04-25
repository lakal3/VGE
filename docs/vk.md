# VGE core object (Module vk)

This module contains lowers level of VGE objects written in Go. 
Vk module is also the only module that directly calls C++ library VGELib.dll (libVGELib.so)

Main objects in code module:

## Application

Application object is root object in Vulkan. You must create an application object before you
can access actual Vulkan devices.

Vulkan has a lot of extensions and extra layers. Required extensions must be loaded when we create
application or device. In VGE all supported extensions must be added before initializing application.

From Application you can list available Physical devices (DeviceInfo). 
VGE will also indicate in DeviceInfo if some of available device is not supported.

Cube and WebView examples shows how to create Application and Device using vk module.
 
## Device

Device is created using device index (0 you have only one device).
 
On most laptops with dedicated GPU you also have an integrated GPU. 
On those devices you should implement some logic to select proper device. 
In most of VGE example you can give device index from command line.

Device is used to create nearly all of other Vulkan resources. You must have a Device to create to create those resources.

## Ownership and resource lifetime   

Nearly all VGE resources are backed up by Vulkan resources. Those resources don't enjoy luxury of automatic reclamation when
resource is no longer used (Garbage collection). In stead you must manually dispose them when no longer used. 
All struct implementing Dispose method must be disposed manually.

To support disposing VGE library have some helpers methods:

#### Get(ctx APIContext, key Key, cons Constructor) interface{}

Several struct implement Get method. Get method access resource by a key. 
If resource is already available, it is just returned. If resource is not available, 
it will be created using Constructor function. 
If constructed object is a Disposable, object implementing Get will own this new object and automatically dispose object when it is disposed.

#### something.NewXX

If any object implements NewXXX method (like dev.NewSampler), that object is owner new resource and will new object when it is disposed.

#### Owner struct

You can use existing struct Owner to implement Get method in your own classes that wan't to support ownership.

## APIContext

Typically error handling in Go uses return values. However, in Vulkan every function could fail be these failures are not expected. 
There are typically errors in program, drivers or hardware and should not normally occur. 

Therefore using an error return value is not preferable option. 
Always panicing inside library is also bit problematics. Therefore VGE uses APIContent interface for most of API calls. 
If we encounter an error we will call APIContext SetError and context itself can decide how to handle error. Typically VGE examples will
use log.Fatal to terminate but yours solutions might used different approach.

_WebView example uses panic if web request fails, so it can try to serve an other request after failed on._

## Allocating memory in VGE

Memory allocation in Vulkan is a bit tricky. First you must indicate what kind of memory (image or buffer) you require, 
how you are going to use it and should it be placed in device or host memory.
Then you must query what kind of memory pools there are and allocate memory slices from appropriate memory pools.

Unfortunately memory pools are limited resource and you should allocate as much as possible from one pool otherwise you will run out of memory pool handles.

In VGE allocating works like this:
1. You create a memory Pool
2. You register all allocations using ReserveBuffer or ReserveImage
3. You allocate all reserved buffers and images.
 VGE will allocate minimum number of required memory pools to satisfy all reservations.
 
Also note that you cannot dispose individual image or buffer. You must dispose whole pool at once.

## RenderPasses

Vulkan handles rendering in render passes. See [https://vulkan-tutorial.com/Drawing_a_triangle/Graphics_pipeline_basics/Render_passes].

Render passes are fairly complex to setup. So VGE offers few prebuilt render passes with some options. 
If you need a new kind of render pass, that must be implemented in VGELib (C++ part).

#### Standard render passes

- ForwardRenderPass with or without depth buffer. Forward render pass support one output (main image)
- DepthRenderPass support rendering depth buffer only. Used in shadow rendering.  

## Pipelines

Pipelines are used to combine shaders (vertex, frag), inputs, descriptors sets into single "runnable" unit.
See for examples different materials (vge/materils/...) on how to build up pipelines.

_Pipelines need compiled SPIR-V shader modules. It is possible to load those from a hard drive. 
However, VGE has also tool called packspv that will convert SPIR-V files to Go code that compiles SPIR-V code
directly into binary. All VGE material use this tool to combine SPIR-V code to module itself._
   
 
## Other resources in vk

See API documentation for more details about other resources:
- Sampler are used to sample image 
- Command are used to record and submit command to Vulkan. In Vulkan API we have command pools and commands. VGE combines those are one entity.
- DescriptorLayout describe a layout of a shader biding slots. See [https://vulkan.lunarg.com/doc/view/1.0.49.0/windows/tutorial/html/08-init_pipeline_layout.html]
- DescriptorPool is used to allocate DescriptorSets. In VGE each DescriptorPool only support one kind of descriptors.
- DescriptorSet is an individual descriptor you can write values (uniforms, sampled images etc..) to and then bind descriptors to recorded commands.
Shaders access values and images through these sets.
- Desktop and Window handles OS Window creating and representing rendered image on OS Window
- ImageView and ImageRange allows Vulkan to use subrange of Image. Most Vulkan API calls uses ImageView, not Image
- RenderCache is helper to handle per instance and per frame resources need during rendering

Nice overview of Vulkan objects [https://gpuopen.com/understanding-vulkan-objects/] might make it easier to understand how Vulkan objects are related. 

## Concurrent access

Vulkan in multi threaded. However, most of vk and Vulkan objects are not thread safe. You must use mutex or other mechanism to synchronize access.

So you can render two separate image concurrently without synchronization using different set of Vulkan objects. 
You can't use same object in multiple coroutines without external synchronization.
 
API document will tell if some method is safe for concurrent access.

All constructor functions, NewXXX, are safe for concurrent calls.



 
   

   

   