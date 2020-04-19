#version 450
#extension GL_GOOGLE_include_directive : require

#include "../../vscene/input.glsl"

#ifdef SKINNED
#include "../../vscene/skin.glsl"
#endif

layout(location = 0) out vec2 o_uv;

layout(set = 0, binding = 0) uniform FRAME {
    mat4 projection;
    mat4 view;
} frame;

layout(set = 1, binding = 0) uniform INSTANCE {
    mat4 world[256];
} instance;

void main() {
    mat4 world = instance.world[gl_InstanceIndex];
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    vec4 pos = world * vec4(i_position, 1.0);
    gl_Position = frame.projection * frame.view * pos;
    o_uv = i_uv0;
}
