# Scene handling in VGE

Main scene handling is implemented in vscene module. 

Scene in VGE will contain all elements making up a rendered image.
 
Scene is hierarchical acyclic graph of nodes starting from root node. 
Each node contains optional a NodeControl and child nodes.

Most of VGE examples shows how to construct, control and manipulate Scenes.

## NodeControl

NodeControl is something that controls rendering a Scene. 
NodeControls can for example control position, draw mesh, add light, add background etc.

Node controls are divided into multiple modules depending on what they do. 
For example env contains background and probe rendering. Shadow module contains shadow light rendering etc...


#### TransformControl

Applier 3D transform (position, scale, rotation) to current node and all child nodes.

#### AmbientLight

Adds ambient light to scene. Only one ambient light can be used in one scene. 

#### DirectionalLight

Adds directional light to scene.

#### PointLight  

Adds point light to scene.

#### PointLight (shadow) 

Adds point light to scene that will cast shadow.

_Shadow casting light are quite much more expensive (uses more GPU resources) that lights without shadow_

#### GrayBg

Gray gradient background for scene. See WebView example. Scene should only have one background.

#### EquiRectBGNode

Background rendered using  image in equirectangular projection as a background. Most examples uses EquiRectBGNode.

#### MeshNodeControl

Render mesh using given shader. Typically these nodes are constructed using existing model.
vscene.NodeFromModel. You can have multiple instances of same MeshNodeControl in single scene.

#### AnimatedNodeControl

Like MeshNodeControl but made from rigged meshes. AnimatedNodeControl has methods to update and change running animation(s).

#### MultiControl

You can place several node controls into one MultiControl. 
Sometimes it is more convenient to place for example a Transformation and a Light into one node.
 
#### Probe 

Probe will render view from probe location excluding all child nodes. Some shaders like Pbr and Std can use probe images to
render reflection of metallic surfaces. Probe also computer spherical harmonics for irradiance lightning.

_Currently only one probe per node tree is supported. This will be changed later. So you can't have a probe inside a child node which parent also has a probe._

_Probe and ambient light are exclusive. Ambient light is actually zeroth order of spherical harmonic_

#### Decal (experimental)

Only Std shader support decals. Decals can render 2D image, placed in 3D world over a 3D mesh. 

In VGE decals don't break up mesh. Updating mesh in GPU is quite expensive and not very useful for if you wan't to animate decals. 
Therefore, in VGE are decals are calculated and applied in the shader! Decals will only affect current node and it's child nodes.

See robomaze example with -oil command line switch to see oil stain decals in demo.
 
#### UIView

Places user interface into scene. See [VGE UI](vui.md)

### Custom node control

Robomaze example have quite a lot of custom node controls that handle for example animation of scene.

## Scene updates

When constructing scene you can directly change nodes and node control of scene. However, 
when you start rendering Scene it is important that you don't change Scene at same time that rendering process in using the Scene. 

Scene has Update method that takes a function. Function will be called when it is safe to update scene. 

**You should never do any I/O or other time consuming things in update. Prepare information before Update and only call Update to apply prepared changes to Scene.**

## Phases

Scene is rendered in multiple phases. Some phases may be skipped depending on rendering process.

Typical rendering of scene looks like:

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
 
PredrawPhase will handle all rendering required to support main scene rendering. 
This includes for example rendering shadow map for each light cast shadows. 
Also probes (when updates) renders them in predraw phase. 

Predraw phases submit rendering but don't necessarily wait them to complete. This allows GPU to handle multiple submit parallel. 
In order to ensure that all previous work has been done, implementation must call all pending methods of predraw phase and also
include all Needed submit infos when submitting main rendering command. See for example shadow.PointLight implementation.

Draw phases will draw different layers of scene. Layers are background, main 3D, transparent (not jet supported) and UI. 
You may add custom phases if needed. 


_You can usually use premade implementation ForwardRenderer in vapp module to handle rendering a scene_
 





 
 
                                                                                                  
  


