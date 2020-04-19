# Build VGELib.dll (libVGELib.so)

## Depencendicies

VGELib uses [GLFW](https://www.glfw.org/) library to handle native UI / event handling in platform independent manner.
GLFW is included into VGE project as a git submodule. 

You must update submodules to download actual GLTW code using git submodule command.

`git submodule init`

`git submodule update`

Other dependencies (std_image.h) are included in VGE project.

## Vulkan SDK

Download and install [Vulkan SDK](https://vulkan.lunarg.com/)

Optionally: Run vkcube example from SDK to check that your display drivers supports Vulkan.
 
## C++ version

VGE is fairly standard C++ and don't use advanced C++ concepts. However it uses some standard libraries from C++ 17 (std::string_view). 
 
## CMake

VGELib uses CMake build system. 
In addition to normal CMake parameters you must define VULKAN_SDK cache variable 
that must point to directory where you installed Vulkan SDK

`cmake -DVULKAN_SDK={Vulkan SDK install path} ...`

Installing release mode binaries should install new version of VGELib.dll (libVGELib.so)

## Linux build

Linux release build was compiled with clang 9.0 C and C++ compilers and
build has been tested only with clang compiler. 
GNU gcc/g++ compilers might still require some tweaking of compiler settings.

Linux need additional libraries to build GLWF. 
See their documentation about requirements (or just run make and install what is missing).

Also ensure that you place your shared library (libVGELib.so) where OS will be able to locate it.

## Windows build

Windows build was done with Visual Studio 2019 Community edition.

 

