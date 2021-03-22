# Credits

# Assets in VGE

The following open source assets and software are used in VGE.

## Software and libraries

[Vulkan SDK](https://www.lunarg.com/vulkan-sdk/)

The Vulkan headers and validation layers. Vulkan SDK's glslangValidator is used to
compile glsl files to the SPIR-V format.

[Go GL math library](https://godoc.org/github.com/go-gl/mathgl/mgl32)

The 3D math library used in VGE

[GLFW library](https://www.glfw.org/)

In VGE, GLFW multi-platform desktop support library handles all lower lever interactions with the desktop.
The library allows a nice OS independent way of handling windowing and event handling (mouse, keyboard etc.).

[STB libraries](https://github.com/nothings/stb)

Single file C/C++ libraries used for png, jpg and hdr image decompression

[Blender](https://www.blender.org/)

A superior tool to edit the 3D models. Although, not directly in the project, most of the assets models were post processed with Blender 2.8x.


## 2D and 3D assets

[HDRI heaver](https://hdrihaven.com/)

Multiple nice HDR images in assets/envhdr are from HDRI heaven.

[CC0textures](https://cc0textures.com/)

The textures used in the robomaze castle and VGE logo

[Open Game Art](https://opengameart.org/)

3D model [Goth female](https://opengameart.org/content/goth-female-fleur-du-mal)

[FreeImages](https://www.freeimages.com/)

Some textures in models.

[Mixamo](https://www.mixamo.com/#/)

Mixamo was used to rerig some test assets using it's automated character rigging tool.
Also, some sample animations are from Mixamo.


# Tutorials and samples

Articles and samples that mostly influenced building the VGE.

[github.com/SaschaWillems/Vulkan](https://github.com/SaschaWillems/Vulkan)

Excellent C++ Vulkan samples that solved many problems that were not so obvious reading the Vulkan specifications.

[Vulkan tutorial](https://vulkan-tutorial.com/)

Vulkan tutorial has a nice step-by-step description on how to setup Vulkan assets

[Intel Vulkan tutorial](https://software.intel.com/en-us/articles/api-without-secrets-introduction-to-vulkan-part-1)

Another nice tutorial used to make the first Vulkan renderings.

[Learn opengl](https://learnopengl.com)

Lots of good glsl examples. Same techniques can be used easily in Vulkan.

[glTF samples](https://github.com/KhronosGroup/glTF-Sample-Models)

A sample model to test the glTF Loader

[Filement documentation of Physical Based Rendering](https://github.com/google/filament)

In-depth explanation of Physical Based Rendering algorithms.

[Spherical harmonics](http://www.ppsloan.org/publications/StupidSH36.pdf)

Explanations and examples of spherical harmonics used to render irradiance lightning in VGE.

[Signed distance field for font rendering](https://steamcdn-a.akamaihd.net/apps/valve/2007/SIGGRAPH2007_AlphaTestedMagnification.pdf)

VGE uses this idea to bake fonts and vector graphics to renderable glyphs.

[Motion capture formats](http://www.dcs.shef.ac.uk/intranet/research/public/resmes/CS0111.pdf)

Used to build BVH parsing module, vanimation.

_There were other resources as well but the listed above were the most beneficial._








