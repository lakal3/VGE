#version 450

layout(location = 0) out vec4 outColor;
layout(location = 0) in vec2 i_uv;

layout(set = 2, binding = 0) uniform MaterialUBF {
    vec4 color;
    float textured;
} Material;

layout(set = 2, binding = 1) uniform sampler2D textures[1];

void main() {
   if (Material.textured > 0) {
       outColor = texture(textures[0], i_uv);
   } else {
       outColor = Material.color;
   }
}