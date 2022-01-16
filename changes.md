# Change history

## Version 0.next (dev branch)

New features:
- Integrated glsl -> SPIR-V compiler
  - Allows dynamic shader compilation while application is running
- Support for Push Constants in DrawList
- SpinLock for very short locking needs

Breaking changes:
- APIContext is now private to vk module. No public API takes APIContext parameter


## Version 0.20.1 

This version contains several improvement and unfortunately there are also some breaking changes.

New features:
- Experimental version of deferred renderer that uses [Deferred shading](https://en.wikipedia.org/wiki/Deferred_shading) to
  render views instead of initial forward renderer. Some examples like robomaze has been upgraded to support deferred shader. 
  Use command line switch -deferred to test

- Better support different kind of renderers. Frame is now interface and not a default forward.Frame
  Renderer type is indicate by Frame that is available for all Phases through PhaseInfo. Frame is also passed to DrawContext so that material can access it more easilly. Some material like std can support multiple renderers (using different pipelines).

- More complex frames should support SimpleFrame that allows some basic types and materials to work with different Renderers

- All basic light types, spot light (**new**), directional light and point light are now supported

- All basic light types now supports shadows. 
  Shadow maps uses now Parabloid mapping that reduces number of point light shadow maps from 6 to 2 and
  allows support for spotlights using same mapping but with only one map.

- Some Frame methods have been change to interface so that for example Lights works on all renderer supporting LightPhase interface

- VGE can now create Vulkan render pass with multiple or zero outputs.

- You are now allowed to change content of model builder before loading model to GPU. 
  This allows interpreting standard model like glTF and convert it to something that is not
  directly supported by model format. (There will be an example about this later).

- Decals has beed refactor and now supports standard shader both in forward and deferred rendering mode. 
  Decal no longer require special decal builder. Instead they can use images from [model](docs/vmodel.md). See robomaze/stain.go for an example of how to use decal painter.

- GO/C++ interface generator will now make temporary copies of pointer items. This prevents Go Pointer to pointer exceptions
in CGO call in Linux (and you don't have to set GODEBUG=cgocheck=0)

Breaking changes:

- Forward renderer has been moved to own module. There will be new advanced (deferred) renderer available.

- Frame that is related to renderer is now an interface. Each node can check if renderer is supported by checking if frame can be cast to suitable type.

- Decal builder has beed removed. Use vmodel.ModelBuilder instead to upload images for decals. Decal painter API is also different from previus experimental Decal module.  (Compare file with version 0.14.1 to see API changes)

- Linux precompiled C++ .so lib has been removed. You must compile libvgelib.so yourself.
  

## Version 0.14.1

- Multiple fixes in Linux support including problems with Intel Vulkan driver that just displayed empty window.
- Dynamic descriptors now always adds BindingUpdateAfterBindBit to descriptor 
  and to all descriptors pools created from descriptor. 
  This should allow very large dynamic (~1000000) image arrays in all Windows/Linux drivers supporting
  VK_EXT_descriptor_indexing (~all new drivers)
- Fixed problem that disabled Vulkan validation. *You can now also use Vulkan Configurator from Vulkan 1.2 SDK to check for validation error(s) instead of inbuilt validator*


## Version 0.12.1

- Adopted Go 1.16 embed directive to bind compiled (spv) shaders to exe. 
  This feature makes packspv tool obsolete. Packspv will be removed soon
  
## Version 0.10.1 (2020-05-01)

Initial public version of the VGE

