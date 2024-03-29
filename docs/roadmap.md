# Roadmap

Planned VGE features (not in priority order)

### Simpler, near-term goals
- [ ] Support for multiple probes in a scene.
- [ ] Forward+ render pass that supports post processing effects like the depth fog.
   - Some shaders like fire should also use postprocessing so that they could sample an already rendered scene
   - Depth effects like the fog
- [x] Shadows for directional lights (in version 0.20.1)
- [x] Spot lights (with shadows) (in version 0.20.1)
- [x] Deferred renderer (experimental version available in 0.20.1)
- [ ] Improved decals in deferred shader
- [ ] Basic dialogs like yes/no
- [ ] Water shader (Needs Forward+ render pass)
- [ ] Asset packing
   - Currently the VGE processes all raw assets at the start of each run, like: Rendering fonts, uncompressing images and parsing models files and converting them to GPU renderable assets.
   This process is quite fast on modern GPUs, but it would still be nice to store the results once the assets have been processed
   to a format that only need loading to GPU.


### Complex, longer term goals
- [ ] Use SDFs (signed distance fields) for shadow calculation (replaces shadow maps)
- [ ] Example of a large open world scene (most likely a separate project)
- [ ] Real time ray tracing using NVidia's Vulkan extensions (Standard extensions in Vulkan 1.2)
- [ ] Less 'game' UI support including:
  - Standard dialogs (file open, color picker)
  - Clipboard handling
  - Better text editing
- [ ] Integration with some physics engine (most likely a separate project)

And of course, the never ending project to improve documentation, API reference and examples.
