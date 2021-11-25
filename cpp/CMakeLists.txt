cmake_minimum_required(VERSION 3.0)

project(vgelib C CXX)

if(NOT CMAKE_DEBUG_POSTFIX)
  set(CMAKE_DEBUG_POSTFIX d)
endif()

if (CMAKE_CXX_COMPILER_ID MATCHES "MSVC")
  add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /std:c++latest")
else()
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++17")
endif()
set(VULKAN_SDK "" CACHE PATH "Vulkan SDK path")
set(NO_GLSLCOMPILER false CACHE BOOLEAN "No embedded GLSL->SPIR-V compiler")

set(GLFW_BUILD_DOCS OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_TESTS OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)

add_subdirectory("third_party/glfw")

include_directories(${PROJECT_SOURCE_DIR} "third_party/glfw/include")

if(NOT NO_GLSLCOMPILER) 
    add_subdirectory("third_party/glslang")
    include_directories("third_party/glslang/glslang/Include")
else()
    add_compile_definitions("NO_GLSLCOMPILER") 
endif()

if(NOT VULKAN_SDK STREQUAL "")
    include_directories("${VULKAN_SDK}/include")
endif()

if(NOT CMAKE_DEBUG_POSTFIX)
  set(CMAKE_DEBUG_POSTFIX d)
endif()

set(VGELIB_SRC  "vgelib/app.cpp" "vgelib/vgelib.hpp" "vgelib/command.cpp" "vgelib/memory.cpp" "vgelib/descriptor.cpp" "vgelib/renderpass.cpp" "vgelib/pipeline.cpp" "vgelib/desktop.cpp" "vgelib/vgelib_if.cpp")
set(VGELIB_SRCH "vgelib/app.hpp" "vgelib/command.hpp" "vgelib/memory.hpp" "vgelib/descriptor.hpp" "vgelib/renderpass.hpp" "vgelib/pipeline.hpp" "vgelib/image.hpp" "vgelib/font.hpp" "vgelib/desktop.hpp" "vgelib/vgelib_if.hpp")

if(NOT NO_GLSLCOMPILER) 
    set(VGELIB_SRC_GLSLCOMPILER "vgelib/glslcompiler.cpp" "vgelib/glslcompiler.hpp")
endif()

if(CMAKE_SYSTEM_NAME MATCHES "Windows")
    set(VULKAN_LIB "${VULKAN_SDK}/lib/vulkan-1.lib")
    set(VGELIB_OS_SRC "vgelib/common/stb_image.cpp")
else()
    if(NOT VULKAN_SDK STREQUAL "")
        set(VULKAN_LIB "${VULKAN_SDK}/lib/libvulkan.so")
    else()
        set(VULKAN_LIB "/usr/lib/libvulkan.so")
    endif()
    set(VGELIB_OS_SRC "vgelib/common/stb_image.cpp" )
endif()

add_library(vgelib SHARED ${VGELIB_SRC} ${VGELIB_SRCH} ${VGELIB_OS_SRC} ${VGELIB_SRC_GLSLCOMPILER})
target_link_libraries(vgelib ${VULKAN_LIB} glfw)
if(NOT NO_GLSLCOMPILER) 
    target_link_libraries(vgelib glslang SPIRV)
endif()

install(TARGETS vgelib DESTINATION "bin")
