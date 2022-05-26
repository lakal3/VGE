#version 450
#extension GL_EXT_samplerless_texture_functions: enable

layout(location = 0) out vec4 o_color;
layout (set=0, binding=0) uniform texture2D src;

void main() {
    ivec2 pos = ivec2(gl_FragCoord.x, gl_FragCoord.y) * ivec2(2,2);
    vec4 c = texelFetch(src, pos,0) + texelFetch(src, pos + ivec2(0,1),0) +
      texelFetch(src, pos + ivec2(1,0),0) + texelFetch(src, pos + ivec2(1,1),0);
    o_color = vec4(vec3(c) * vec3(0.25), 1.0);
}