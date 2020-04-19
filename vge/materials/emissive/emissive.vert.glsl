#version 450
#extension GL_GOOGLE_include_directive : require

#include "../../vscene/input.glsl"

#ifdef SKINNED
#define SKIN_SET 2
#include "../../vscene/skin.glsl"
#endif

layout (constant_id = 0) const int MaxInstances = 200;

layout(set = 1, binding = 0) uniform INSTANCE {
    mat4 world[MaxInstances];
    vec4 color[MaxInstances];
} instances;

layout(location = 0) out vec4 o_color;

layout(set = 0, binding = 0) uniform FRAME {
    mat4 projection;
    mat4 view;
} frame;



void main() {
    mat4 world = instances.world[gl_InstanceIndex];
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    vec4 pos = world * vec4(i_position, 1.0);
    gl_Position = frame.projection * frame.view * pos;
    o_color = instances.color[gl_InstanceIndex];
}
