#version 450
#extension GL_GOOGLE_include_directive : require

#include "../../vscene/input.glsl"
#include "../../vscene/frame.glsl"

layout(set = 1, binding = 0) uniform INSTANCE {
    mat4 world[256];
} instance;

#ifdef SKINNED
#include "../../vscene/skin.glsl"
#endif

layout(location = 0) out vec3 o_position;
layout(location = 1) out vec2 o_UV0;
layout(location = 2) out mat3 o_normalSpace;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    mat4 world = instance.world[gl_InstanceIndex];
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    o_UV0 = i_uv0;
    gl_Position = frame.projection * frame.view * world * vec4(i_position, 1.0);
    o_position = vec3(world * vec4(i_position, 1.0));
    o_normalSpace = calcNormalSpace(world);
}
