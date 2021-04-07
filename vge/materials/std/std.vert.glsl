#version 450
#extension GL_GOOGLE_include_directive : require
#extension GL_EXT_nonuniform_qualifier: require

#define DYNAMIC_DESCRIPTORS 1

#define MAX_INSTANCES 200

#include "../../vscene/input.glsl"
#include "../../forward/frame.glsl"

struct INSTANCE {
    mat4 world;
    vec2 decalIndex;
    vec2 dummy;
};

layout(set = 1, binding = 0) uniform INSTANCES {
    INSTANCE inst[MAX_INSTANCES];
} instances;

#ifdef SKINNED
#include "../../vscene/skin.glsl"
#endif

layout(location = 0) out vec3 o_position;
layout(location = 1) out vec2 o_UV0;
layout(location = 2) out flat ivec2 o_decalIndex;
layout(location = 3) out mat3 o_normalSpace;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    mat4 world = instances.inst[gl_InstanceIndex].world;
    vec2 decalIndex = instances.inst[gl_InstanceIndex].decalIndex;
    o_decalIndex = ivec2(int(decalIndex.x), int(decalIndex.y));

    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    o_UV0 = i_uv0;
    gl_Position = frame.projection * frame.view * world * vec4(i_position, 1.0);
    o_position = vec3(world * vec4(i_position, 1.0));
    o_normalSpace = calcNormalSpace(world);
}
