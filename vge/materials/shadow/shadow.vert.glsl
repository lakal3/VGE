#version 450
#extension GL_GOOGLE_include_directive : require

layout(constant_id = 0) const int MAX_INSTANCES = 100;

#include "../../vscene/input.glsl"

layout(location = 0) out flat int o_index;
layout(location = 1) out vec2 o_uv0;

#include "sframe.glsl"

#ifdef SKINNED
#define SKIN_SET 1
#include "../../vscene/skin.glsl"
#endif

void main() {
    o_index = gl_InstanceIndex;
    mat4 world = frame.instances[gl_InstanceIndex];
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    o_uv0 = i_uv0;
    gl_Position = world * vec4(i_position, 1);
}