# VGE Architecture

This document explains how VGE is built and also, the reasons behind the choices made.


## Vulkan API

The original idea was to learn modern, multi threaded GPU API. One major goal was to select one API (to rule them all) instead of trying to implement abstract Graphics API and then try map that to different API:s as larger engines usually do. 
This allows full use of API without compromises or really complex extra layers.
 
Only two major multi threaded GPU API:s, supported on most of modern GPU:s exists, DirectX 12 and Vulkan. 
Vulkan were more interesting because it has multi platform support. Also, glsl shader language is preferences.

So, VGE uses Vulkan to access GPU. There is no plan to support other APIs in VGE.

## Go language

Initial versions of VGE tried different combinations like using C++ and TypeScript (javascript) bindings, C#. 

After some attempts I decided that language must support multi threading, strongly typed and in must have a garbage collector. Why:

- Multi threading - What's the point having multi threaded API if you language can only support single threads without some kind of isolation.
- Strongly type - I have seen what happen without it in when projects grow. And any serious graphics engine or solution will be sizable.
- Garbage collected - This goes against major graphics engine designs, but I know that manual memory management takes up resources that I simply don't have. I will take huge amount of time and resources to manual memory management event close to a level where most sophisticated GC algorithms already are.
- Multi platform - First only Windows and Linux, but support for others platforms.
 
If we include languages that have a reasonable user base, this only leave few choices that support initial requirements. Go, C# (Java). I consider C# and Java feature wise fairly close and I already knew .NET quite well, so this ruled out Java. Go came out as a winner in some areas:
- Static precompiled binary. You can make single exe. If you distribute it with VGEDll shared library that is all required to run your app.
- Non moving garbage collector. You can more easily manage object that we must send to C++ code. If is also easy to match memory layout of Go structs, C++ structs or Vulkan uniforms. Go stop the world pauses are also so short that they are unnoticeable even in normal 60 FPS render loops.
- Go´s sub second built times make development feel more like writing a script that compiling large project.
- And C# already has Unity :). In Go world they might be some other uses for graphics engine than just an academic curiosity.

*For example Kotlin/native and Nim would also fit into this category but they still have quite small user base*

## C++ 

In reality all operating system API, Vulkan API, most of graphics libraries, algorithms etc. have been written in
C/C++. All API:s use C ABI conventions. In some way you must be able to consume .h(.hpp) files of write lots and lots of things from scratch.

So the lowest level of VGE is shader library VGELib written in C++. It incorporates some really fine libraries that makes ease up graphics engine design. 
VGE will only use C/C++ libraries if no equivalent pure Go library where available (with exception of reading JPEG images. Go standard implementation was incredibly slow).

If you want to manually build C++ library see [building VGELib](build_vgelib.md).

### C++ but, no CGO (or very little in Linux)

To make using VGE easier, no C++ compiler is required when running VGE on windows. VGE only uses syscall to load VGELib shared library. 
This also means that VGE in Windows is not an CGO program. 

Unfortunately, in Linux there is no way (that i am aware of) to load shared library 
without some CGO to link in dlopen etc. functions need to load shared library.

VGE project contains own interface generation tool that will write required Go code to call C++ library. 
This tool also builds up one C++ file to implement call endpoint. 
You don't need to know anything about this tool unless you plan to change Go/C++ interface.
