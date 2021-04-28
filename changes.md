# Change history

## Version 0.20.1 (beta)
This version is bigger improvement and unfortunately there are some breaking changes.

New features:
- Experimental version of deferred renderer that uses [Deferred shading](https://en.wikipedia.org/wiki/Deferred_shading) to
  render views instead of initial forward renderer.

- Better support different kind of renderers. Frame is now interface and not a default forward.Frame
  Renderer type is indicate by Frame that is available for all Phases through PhaseInfo. Frame is allow passed to DrawContext so that material can access it more easilly

- More complex frames should support SimpleFrame that allows some node types and materials to work with different Renderers

- All basic light types, spot light (**new**), directional light and point light are now supported

- All basic light types now supports shadows. 
  Shadow maps uses now Parabloid mapping that reduces number of point light shadow maps from 6 to 2 and
  allows support for spotlights using same mapping but with only one map.

- Some forward Frame methods have been change to interface so that for example Lights works on all renderer supporting LightPhase interface

- VGE can now create Vulkan render pass with multiple or zero outputs.

- You are now allowed to change content of model builder before loading model to GPU. 
  This allows interpreting standard model like glTF and convert it to something that is not
  directly supported by model format. (There will be an example about this later).

Breaking changes:

- Forward renderer has been moved to own module. There will be new advanced (deferred) renderer available.

- Frame that is related to renderer is now an interface. Each node can check if renderer is supported by checking if frame can be cast to suitable type.

TODO: Add example from cube

  

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

