# VGE Architecture

This document explains how VGE is built and also, the reasons behind the choices made.


## Vulkan API

The original idea was to learn a modern, multi threaded GPU API. One major goal was to select one API (to rule them all) instead of trying to implement an abstract Graphics API and then trying to map that to different APIs as larger engines usually do.
This allows the full use of the selected API without compromises or really complex extra layers.

Only two major multi threaded GPU APIs that are supported on most of modern GPU:s exist - DirectX 12 and Vulkan.
Vulkan was more interesting because it has multi-platform support. Also, glsl shader language is a personal preference.

Thus VGE uses the Vulkan to access the GPU. There is no plan to support other APIs in VGE.

## Go language

Initial versions of VGE tried different combinations like using C++ and TypeScript (javascript) bindings and C#.

After some attempts I decided that the language must support multithreading, be strongly typed and it must have a garbage collector.

The reasons for that being:
- Multi threading - What is the point having a multithreaded API if your language can only support single threads without some kind of isolation?
- Strongly typed - I have seen what happens without it in when projects grow. And any serious graphics engine or a solution for that purpose will be a large project.
- Garbage collected - This goes against major graphics engine designs, but I know that manual memory management takes up resources that I simply don't have. It will take a huge amount of time and resources to manually manage memory to be even close to the level where most sophisticated GC algorithms already operate.
- 	Multi-platform - First only Windows and Linux, but support for other platforms is a possibility.

If we consider languages that have a reasonable user base, this only leaves few choices that support the initial requirements. Go and C# (Java). I consider C# and Java languages to be feature wise fairly close and I already knew .NET quite well, so this ruled out Java. Go seemed better than C# in the following areas:
- Static precompiled binary. You can make a single exe-file. If you distribute it with the VGEDll shared library, you have all that is required to run your app.
- Non-moving garbage collector. You can more easily manage objects that must be sent to C++ code. It is also easy to match memory layout of Go structs, C++ structs and Vulkan uniforms. Go stop-the-world pauses are also so short that they are unnoticeable even in normal 60 FPS render loops.
- Go´s sub-second build times make the development feel more like writing a script than compiling a large project.
- And C# already has Unity :). In the Go world there might be some other use for a graphics engine than just academic curiosity.

*For example Kotlin/native and Nim would also fit into this category but they still have quite small user base*

## C++

In reality all operating system APIs, the Vulkan API, and most graphics libraries, algorithms etc. have been written in C/C++. All API:s use C ABI conventions. In some way, you must be able to consume .h(.hpp) files or write lots and lots of things from scratch.

So the lowest level of the VGE is a shader library VGELib written in C++. It incorporates some really fine libraries that makes the graphics engine design easier.
VGE only uses C/C++ libraries if no equivalent pure Go library is available (with exception of reading JPEG images. The Go standard implementation was incredibly slow).

If you want to manually build the C++ library see [building VGELib](build_vgelib.md).

### C++ but, no CGO (or very little in Linux)

To make using VGE easier, no C++ compiler is required when running VGE on Windows. VGE only uses syscall to load VGELib shared library.
This also means that VGE in Windows is not an CGO program.

Unfortunately, in Linux there is no way (that I am aware of) to load a shared library
without some kind of CGO to link to dlopen etc. functions that are needed to load a shared library.

The VGE project contains its own interface generation tool that will write the required Go code to call the C++ library.
This tool also builds one C++ file to implement the call endpoint.
You do not to know anything about this tool unless you plan to change the Go/C++ interface.

### GODEBUG=cgocheck=0

For performance reason VGE will pass pointers to Go memory. VGE support library VGELib.dll (libVGELib.so) is aware of Golangs carbage collector and will not hold pointer to memory after call to library completed. As long as Golang don't implement moving carbage collector this is safe! 

In Windows this works out of box. However in Linux we must use short CGO module to invoke libVGELib.so. 
Unfortunately current implementation of CGO will check that we will not send pointers from Go's heap to CGO calls :(. cgocheck=0 will disable this check.

**Go 1.17 will most like support new Handle type that solves this problem!**

VGE will support handles when they are available. See more from Golang issue 37033.







