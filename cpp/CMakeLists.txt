cmake_minimum_required(VERSION 3.0)

project(VGELib C CXX)

if(NOT CMAKE_DEBUG_POSTFIX)
  set(CMAKE_DEBUG_POSTFIX d)
endif()

if (CMAKE_CXX_COMPILER_ID MATCHES "MSVC")
  add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /std:c++latest")
else()
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++17")
endif()
set(VULKAN_SDK "D:/Libs/VulkanSDK/1_1_101" CACHE PATH "Vulkan SDK path")



set(GLFW_BUILD_DOCS OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_TESTS OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
add_subdirectory("third_party/glfw")

include_directories(${PROJECT_SOURCE_DIR} "third_party/glfw/include" "${VULKAN_SDK}/include" )

if(NOT CMAKE_DEBUG_POSTFIX)
  set(CMAKE_DEBUG_POSTFIX d)
endif()

set(VGELIB_SRC  "vgelib/app.cpp" "vgelib/vgelib.hpp" "vgelib/command.cpp" "vgelib/memory.cpp" "vgelib/descriptor.cpp" "vgelib/renderpass.cpp" "vgelib/pipeline.cpp" "vgelib/desktop.cpp" "vgelib/vgelib_if.cpp")
set(VGELIB_SRCH "vgelib/app.hpp" "vgelib/command.hpp" "vgelib/memory.hpp" "vgelib/descriptor.hpp" "vgelib/renderpass.hpp" "vgelib/pipeline.hpp" "vgelib/image.hpp" "vgelib/font.hpp" "vgelib/desktop.hpp" "vgelib/vgelib_if.hpp")

if(CMAKE_SYSTEM_NAME MATCHES "Windows")
    set(VULKAN_LIB "${VULKAN_SDK}/lib/vulkan-1.lib")
    set(VGELIB_OS_SRC "vgelib/common/stb_image.cpp")
else()
    set(VULKAN_LIB "${VULKAN_SDK}/lib/libvulkan.so")
    set(VGELIB_OS_SRC "vgelib/common/stb_image.cpp" )
endif()

add_library(VGELib SHARED ${VGELIB_SRC} ${VGELIB_SRCH} ${VGELIB_OS_SRC})
target_link_libraries(VGELib ${VULKAN_LIB} glfw)

install(TARGETS VGELib DESTINATION "bin")
