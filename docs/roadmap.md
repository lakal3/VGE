# Roadmap

Planned VGE features (not in any priority order)

### Simpler, near term goals
- [ ] Support multiple probes support in scene.
- [ ] Forward+ render pass that supports post processing effect like depth fog. 
   - Some shaders like fire should also use postprocessing so that they could sample already rendered scene
   - Depth effect like fog  
- [ ] Shadows for directional lights
- [ ] Spot lights
- [ ] Basic dialogs like yes/no
- [ ] Water shader (Needs Forward+ render pass)
- [ ] Asset packing
   - Currently VGE process all raw assets at start of each run, like: Rendering fonts, uncompressing image and parsing models files and convert them to GPU renderable assets.
   This process is quite fast on modern GPUs but it would still be nice to store once processed assets
   to format that only need loading to GPU.


### Complex, longer term goals
- [ ] Example of large open world scene (most like a separate project)
- [ ] Real time ray tracing using NVidia's Vulkan extensions
- [ ] Business UI support including standard dialogs (file open, color picker)
- [ ] Integration with some physics engine (most likely a separate project)

And of cause never ending project to improve documentation, API references and examples.
