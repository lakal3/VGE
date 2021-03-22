# Building and loading a model in VGE

Before you can render any complex 3D object in Vulkan, you must build a vertex buffer containing input(s) for the Vextex shader <sup name="anchor1">[1](#footnote1)</sup>.
Of course, you can manually assemble a vertex buffer and other assets required to render 3D models, however, this is really quite tedious work.
(See vge/materials/env/erbgmap.go cubeVertex for an example)

To ease up this process the VGE includes the vmodel module that handles modeling 3D models much easier.

## ModelBuilder

vmodel includes the ModelBuilder that will be used to bake the actual model. In the ModelBuilder you can add meshes, materials, images and nodes.

Nodes in model build up a hierarchy. There is one root node and
each node can have multiple child nodes. Each node can also have a transformation matrix that
can scale, rotate and translate itself and all child nodes.

Nodes may additionally have a mesh and material. Rigged meshes will also have a skin.

Materials typically need textures (images) to represent color, normal changes, metallicity, roughness etc.. You can add these images to ModelBuilder.

### Materials

ModelBuilder needs a ShaderFactory. The VGE contains several standard shader factories (unlit, Phong, pbr, std).

ShaderFactory is responsible for converting material properties to a Shader that can be used to actually render 3D meshes.

See model example on how to setup the ModelBuilder. Unlit shader is a simple shader

### Meshes

3D Meshes are built with MeshBuilder. Mesh builder allows you to add all vertex information that makes up a mesh.
Only position is mandatory. Other values (normal, uv, tangent) are calculated automatically but for best results you should provide all values.

_Meshes also support vertex color. Color is not used in existing shaders, and you may use this in any way you like if you implement custom loader / shader_

## Loaders

Typically, you do not call AddNode, AddMesh, ... methods in ModelBuilder directly. Instead, you use model loader that can
parse model files in different formats and convert them to ModelBuilder assets.

VGE has two prebuilt loaders, one to handle Wavefront OBJ format and other to import glTF 2.0 files.

OBJ loader can only handle unrigged meshes. OBJ color properties are best matched with Phong shader.

glTF loader supports rigged meshed (Skin) and animations.  glTF loader also supports model hierarchy.

glTFviewer example can be used to view different kind of glTF files from a sample set (The sample set must be loaded separately)

**It is possible to load multiple model files into a single ModelBuilder and make just one Model out of these files**


## Bake
When all assets have been uploaded to ModelBuilder you can bake a model using ToModel method.

Bake will upload all assets to Model structure and:
- Rearrange all vertex data (position, normal, tangent, uv, color) into two continuous buffers.
One buffer is for unrigged meshes and the other is for rigged meshes. Rigged meshes contain two extra inputs for weights and joists.
 VGE standard shader assumes that all vertex data is packed this way.
- All images are uploaded into GPU and optionally MIP mapped during upload.
- Allocates DescriptorPool(s) and DescriptorSets to hold material static data. Each material can have a different layout for material descriptor. For an example, see the 'unlit' module.

Arranging vertexes into a single buffer allows the VGE to draw multiple meshes from single model without rebinding vertex buffers between each draw call.
If you draw the same mesh several times (to different position) without drawing other meshes in between, the VGE shaders can draw meshes in a single instanced draw.

Bake will try to minimize GPU memory allocation when uploading a model to GPU. It uses single MemoryPool to contain all Model assets.

When a model is baked all the images are uploaded to the GPU. The ModelBuilder can optionally mipmap all uploaded images.

----

<b id="footnote1">1</b> Assuming you are using a vertex shader. You could use the very latest real time rendering extensions or compute shaders to render 3D images. [↩](#anchor1)
