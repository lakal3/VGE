# Building and loading model in VGE

Before you can render any complex 3D object in Vulkan you must build vertex buffer containing input(s) for Vextex shader <sup>1</sup>.
Of cause you can manually assemble vertex buffer and other assets required to render 3D models, however this is really quite tedious work. 
(See vge/materials/env/erbgmap.go cubeVertex for an example) 

To ease up this process VGE include vmodel module that handles model 3D models much easier.

## ModelBuilder

vmodel includes ModelBuilder that will be used to bake actual model. In ModelBuilder you can add
meshes, materials, images and nodes.

Nodes in model build up a hierarchy. There is one root node and 
each node can have multiple child nodes. Each node also have a transformation matrix that 
can scale, rotate and translate itself and all child nodes.

Nodes may additionally have a mesh and material. Rigged meshed will also have a skin.

Materials typically need textures (images) to represent color, normal changes, metalness, roughness etc.. You can add these images to ModelBuilder.

### Materials

ModelBuilder needs a ShaderFactory. VGE contains several standard shader factories (unlit, Phong, pbr, std). 

ShadedFactory is responsible to convert material properties to a Shader that can be used to actually render 3D meshes.

See model example on how to setup ModelBuilder. Unlit shader is simple shader

### Meshes

3D Meshes are built with MeshBuilder. Mesh builder allows you to add all vertex information that makes up a mesh. 
Only position is mandatory. Other values (normal, uv, tangent) are calculated automatically but for best results you should provide all values.

_Meshes also support vertex color. Color is not used in existing shaders, and you may use this however you like if you implement custom loader / shader_

## Loaders

Typically you don't call AddNode, AddMesh, ... methods in ModelBuilder directly. Instead, you use model loader that can
parse model files in different formats and convert them to ModelBuilder assets. 

VGE has two prebuilt loaders, one to handle Wavefront OBJ format and other to import glTF 2.0 files.

OBJ loader can only handle unrigged meshes. OBJ color properties are best matched with Phong shader.

glTF loader support or rigged meshed (Skin) and animations.  glTF loader also support model hierarchy.

glTFviewer example can be used to view different kind of glTF files from sample set (Sample set must be loaded separately)

**It is possible to load multiple model files into single ModelBuilder and make just one Model out of these files**
  

## Bake
When all assets have been uploaded to ModelBuilder you can bake a model using ToModel method. 

Bake will upload all assets to Model structure and:
- Rearrange all vertex data (position, normal, tangent, uv, color) into two continuous buffer. 
One buffer is for unrigged meshes and other is for rigged meshes. Rigged meshes contains two extra inputs for weights and joists.
 VGE standard shader assumes that all vertex data is packed this way. 
- All images are uploaded into GPU and optionally MIP mapped during upload.
- Allocate DescriptorPool(s) and DescriptorSets to hold material static data. Each material can have have different layout for material descriptor. See for example unlit for sample. 

Arranging vertexes into single buffer allows VGE to draw multiple meshed from single model without rebinding vertex buffers between each draw call. 
If you draw same mesh several times (to different position) without drawing other meshed in between, VGE's shaders can draw meshes in single instanced draw. 

Bake will try to minimize GPU memory allocations when uploading model to GPU. It is using single MemoryPool to contain all Model assets.

When model is baked all images are uploaded to GPU. ModelBuilder can optionally mipmap all uploaded images.    

----

<sup>1</sup> Assuming you are using vertex shader. You could use very latest real time rendering extensions or compute shaders to render 3D images. 