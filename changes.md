# Change history

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

