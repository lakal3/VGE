# Build VGELib.dll (libVGELib.so)

## Depencendicies

VGELib uses [GLFW](https://www.glfw.org/) library to handle native UI / event handling in a platform independent manner.
GLFW is included in the VGE project as a git submodule.

You must update submodules to download the actual GLTW code using the git submodule command.

`git submodule init`

`git submodule update`

Other dependencies (std_image.h) are included in the VGE project.

## Vulkan SDK

Download and install [Vulkan SDK](https://vulkan.lunarg.com/)

Optionally: Run vkcube example from SDK to check that your display drivers support Vulkan.

## C++ version

VGE is fairly standard C++ code and does not use advanced C++ concepts. However, it uses some standard libraries from C++ 17 (std::string_view).

## CMake

VGELib uses CMake build system.
In addition to the normal CMake parameters you must define the VULKAN_SDK cache variable
that must point to the directory where you installed Vulkan SDK

`cmake -DVULKAN_SDK={Vulkan SDK install path} ...`

Installing release mode binaries should install a new version of VGELib.dll (libVGELib.so)

## Linux build

Linux release build was compiled with clang 9.0 C and C++ compilers and the build has been tested only with the clang compiler.
GNU gcc/g++ compilers should also work if you prefer them.

Linux needs additional libraries to build GLWF. You should have at least installed:
- libx11dev
- libxrandr-dev
- libxinerama-dev
- libxcursor-dev
- libxi-dev


See their documentation about the requirements (or just run make and install what is missing).

Also, make sure that you place your shared library (libVGELib.so) where the OS will be able to locate it.

## Windows build

The Windows build was made with the Visual Studio 2019 Community edition.



