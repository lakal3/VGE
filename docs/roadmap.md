# Roadmap

Planned VGE features (not in priority order)

### Simpler, near-term goals
- [x] Support for glsl compilation with Khronos glslang (in dev branch)
    - This allows compiling shaders when application is already running
- [x] Shadows for directional lights (in version 0.20.1)
- [x] Spot lights (with shadows) (in version 0.20.1)
- [x] Deferred renderer (experimental version available in 0.20.1)
- [x] Basic dialogs like yes/no (in dev branch)
- [x] Vector drawing support (in dev branch)
- [ ] 3D drawing support. Draw 3D like 2D without any scene
  - Like immediate mode 3D
- [x] Less 'game' like UI using vector graphics and immediate mode UI principles (in dev branch)
  - Clipboard handling 
  - Better text editing
- [ ] Standard dialogs (file open, color picker)
- [ ] Support for multiple probes in a scene.
- [ ] Forward+ render pass that supports post processing effects like the depth fog.
    - Some shaders like fire should also use postprocessing so that they could sample an already rendered scene
    - Depth effects like the fog
- [ ] Water shader (Needs Forward+ render pass)
- [ ] Asset packing
   - Currently the VGE processes all raw assets at the start of each run, like: Rendering fonts, uncompressing images and parsing models files and converting them to GPU renderable assets.
   This process is quite fast on modern GPUs, but it would still be nice to store the results once the assets have been processed
   to a format that only need loading to GPU.


### Complex, longer term goals
- [ ] Example of a large open world scene (most likely a separate project)
- [ ] Real time ray tracing using NVidia's Vulkan extensions (Standard extensions in Vulkan 1.2)

- [ ] Integration with some physics engine (most likely a separate project)

And of course, the never ending project to improve documentation, API reference and examples.
