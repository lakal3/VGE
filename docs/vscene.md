# Scene handling in VGE

The main scene handling is implemented in the vscene module. 

A Scene in the VGE will contain all the elements making up a rendered image.
 
A Scene is a hierarchical acyclic graph of nodes starting from root node. 
Each node contains optionally a NodeControl and child nodes.

Most of the VGE examples show how to construct, control and manipulate Scenes.

## NodeControl

NodeControl is something that controls rendering a Scene. 
NodeControls can, for example, control the position, draw mesh, add light, add background etc.

Node controls are divided into multiple modules depending on what they do. 
For example, the Env module contains background and probe rendering. The Shadow module contains shadow light rendering etc...


#### TransformControl

Applies a 3D transform (position, scale, rotation) to the current node and all child nodes.

#### AmbientLight

Adds an ambient light to a scene. Only one ambient light can be used in one scene. 

#### DirectionalLight

Adds a directional light to scene.

#### PointLight  

Adds a point light to scene.

#### PointLight (shadow) 

Adds a point light to scene that will cast shadow.

_Shadow casting lights are much more expensive (uses more GPU resources) than lights without a shadow_

#### GrayBg

Gray gradient background for a scene. See the WebView example. A scene should only have one background.

#### EquiRectBGNode

Background rendering using an image in equirectangular projection as a background. Most examples uses EquiRectBGNode.

#### MeshNodeControl

Renders a mesh using the given shader. Typically, these nodes are constructed using existing model.
vscene.NodeFromModel. You can have multiple instances of same MeshNodeControl in single scene.

#### AnimatedNodeControl

Like MeshNodeControl but made from rigged meshes. AnimatedNodeControl has methods to update and change a running animation(s).

#### MultiControl

You can place several node controls into one MultiControl. 
Sometimes it is more convenient to place, for example, a Transformation and a Light into one node.
 
#### Probe 

Probe will render a view from the probe location excluding all child nodes. Some shaders like Pbr and Std can use probe images to
render the reflection of metallic surfaces. Probes also compute spherical harmonics for irradiance lightning.

_Currently, only one probe per node tree is supported. This will be changed later. So you cannot have a probe inside a child node when the parent also has a probe._

_Probe and ambient light are exclusive. Ambient light is actually zeroth order of spherical harmonic_

#### Decal (experimental)

Only Std shader supports decals. Decals can render a 2D image, placed in 3D world over a 3D mesh. 

In VGE, the decals do not break up mesh. Updating a mesh in GPU is quite expensive and not very pragmatic when you desire to animate decals. 
Therefore, in VGE the decals are calculated and applied in the shader! Decals will only affect the current node and it's child nodes.

See the robomaze example with -oil command line switch to see oil stain decals in demo.
 
#### UIView

Places an user interface into a scene. See [VGE UI](vui.md)

### Custom node control

Robomaze examples have quite a lot of custom node controls that handle, for example, the animation of scene.

## Scene updates

When constructing a scene you can directly change the nodes and node controls of scene. However,
when you start rendering a scene it is important that you do not change the scene at same time that the rendering process in using the scene. 

A Scene has an Update method that takes a function. The function will be called when it is safe to the update scene. 

**You should never do any I/O or other time consuming things in an update. Prepare data before the Update and only call Update to apply prepared changes to a scene.**

## Phases

A scene is rendered in multiple phases. Some phases may be skipped depending on the rendering process.

A typical rendering of scene looks like the following:

```go
    bg := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERBackground, cmd, func() {
        if !f.depthPrePass {
            cmd.BeginRenderPass(f.frp, fb)
        }
    }, nil)
    dp := vscene.NewDrawPhase(rc, f.frp, vscene.LAYER3D, cmd, nil, nil)
    dt := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERTransparent, cmd, nil, nil)
    ui := vscene.NewDrawPhase(rc, f.frp, vscene.LAYERUI, cmd, nil, func() {
        cmd.EndRenderPass()
    })
    frame := vscene.GetFrame(rc)
    ppPhase := &vscene.PredrawPhase{Scene: sc, F: frame, Cache: rc, Cmd: cmd}
        
    sc.Process(sc.Time, ppPhase, bg, dp, dt, ui)
    // Complete pendings from predraw phase
    for _, pd := range ppPhase.Pending {
        pd()
    }
    cmd.Submit(append(infos, ppPhase.Needeed...)...)
    cmd.Wait()
```
 
The PredrawPhase will handle all rendering required to support main scene rendering. 
This includes, for example, rendering a shadow map for each light that casts shadows. 
Also probes (when updating) renders them in predraw phase.

Predraw phases submit rendering but do not necessarily wait for them to complete. This allows GPU to handle multiple submits in parallel. 
In order to ensure that all previous work has been done, the implementation must call all pending methods of the predraw phase and also
include all necessary submit infos when submitting the main rendering command. For example, see the  shadow.PointLight implementation.

Draw phases will draw the different layers of scene. Layers are background, main 3D, transparent (not yet supported) and UI. 
You may add a custom phases if necessary. 


_You can usually use the premade implementation ForwardRenderer in vapp module to handle rendering a scene_
 





 
 
                                                                                                  
  


