# 3D view and drawing

vdraw3d is new module in VGE that will eventually replace older vscene. 
vdraw3d has already nearly all features of vscene and associated modules and some
new features like picking, shader customization, transparency support.

vscene will remain in this project but all development efforts goes to vdraw3d.

## Using 3D view

### Render image

New 3D drawing module vdraw3d implements 3D drawing with vdraw3d.View. To draw an 3D image, view
will invoke one function that should record all 3D drawing commands to immutable FreezeList.

View will then use FreezeList to:
- Allocate all resources deeded to render frame (in coordination with ViewWindow)
- Render each phase of frame

Most common rendering commands are:
- DrawMesh - draws single mesh from vmodel.Model
- DrawAnimated - draws single mesh with animation from vmodel.Model
- DrawDecal - draws decals on meshes that support decals
- DrawDirectionalLight - draws directional light with or without shadow
- DrawPointLight - draws point light with or without shadow
- DrawNodes - draw node tree from vmodel.Model
- DrawProbe - draws a probe. DrawMesh and DrawAnimated will use probe to estimate indirect lighting and reflections
- DrawBackground - draws Equirectangular image to background of view

### Static and dynamic draws

Actually there are 2 draw functions for each View. One is used to draw static parts of image. Static parts will not change between frames.
Dynamic draw contains everything that can move or animate. Even light that move or change intercity must be part of dynamic chain.

VGE can use static drawn primitives to optimize some costly operations. Currently, only probe
uses static optimization to draw probe only when static chain is draw. 
But in future for example static light shadow maps could predraw all static meshes during static draw phase.


### Display image

To display view on UI, you must attach View to vapp.ViewWindow that will coordinate
rendering all views attached to it. View can overlap and one ViewWindow can contain several 
views. This way you can attach multiple 3D and UI views to a single OS Window. 
ViewWindow supports also anti aliasing witch will improve image quality with cost of some GPU resources.

You can also implement your own view like ImageViewer in examples/fileviewer module.

### Picking

View support picking of objects from 3D world using Pick method. Picking is actually done we render FreezeList. 
During normal rendering of frame, VGE will use specific shader that records each triangle that is draw over specified area. 
Picking will get all, not only foremost hit to given area. Pick result will be available after next full frame has been drawn. 

### Customizing shaders

vdraw3d uses forward+ rendering pipeline that has been split to several smaller glsl fragments.
vdraw3d shader fragments are compiled with new vcompile tool included in this project. 
vcompile uses programs.json file to configure each individual variant of shaders we need to compile.
Currently there are >30 variations of standard shaders. With customized materials like shaders/mixshader,
we can easily reach hundreds of variations.

Forward+ shading will require several, slightly different version of same shader code that could not been easily managed without vcompile or similar tool.

You can create custom version of any of these fragments and compile it with other fragments to make custom ShaderPack. 
See for example shaders/phongshader on how to change light calculation of std shader.
Constructor of vdraw3d allows you to specify which shaderpack we use.

vcompile uses standard [spir-v compiler](https://github.com/KhronosGroup/SPIRV-Tools) from Khronos group to compile final shader. 
This has been compiled into vgelib(.dll|.so)
vcompile implements its own directives (starting with #). Some directive works similar to ones used in glslangValidator. 
*TODO: Explain custom directives in glsl fragments*

It is also possible to invoke vcompile interface during runtime. This allows dynamic shader compilation on run.
However, compiling larger chunks of glsl will take time, so it is preferable to use statically compiled
shaderpack that is bound to executable or loaded from external file.


### Customize draw commands 

You can add your own draw commands to record Frozen entities to FreezeList. 

View will invoke Frozen entity in several phases that are:
- RenderMaps - This phase should be implemented if you need to render some other view of FreeseList like shadowmap
- RenderProbe - This is called when we render probe. Probe rendering uses simplified calculations for indirect lighting and reflections (otherwise rendering probe would be recursive operation)
- RenderShadow - Is called every time we draw shadow-map. This can be called several times during one frame.
- RenderDepth - This phase record depth of each mesh to depth buffer. Semitransparent objects don't draw themselves to depth buffer.
- RenderColor - This phase is used to render actual color of fragments with all light calculations.
Transparent entities will record themselves during this phase so that View can draw them later in priority order.
- RenderPick - This phase will render special pick shader and ID image.
- RenderOverlay - This phase is called after we have rendered everything. Overlay can draw special overlays using for example depth information or ID image

Finally, image is copied to ViewWindows image using std_blend function. You can override this by customizing this function.

*TODO: Describe ID image*

## Implementation

  


