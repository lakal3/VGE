#version 450
#extension GL_GOOGLE_include_directive : require

#include "../../vscene/input.glsl"
#include "../../forward/frame.glsl"

layout(set = 1, binding = 0) uniform INSTANCE {
    mat4 world[256];
} instance;

#ifdef SKINNED
layout (location = 3) in vec4 i_weights0;
layout (location = 4) in uvec4 i_joints0;

layout(set=3, binding=0) uniform JOINTS {
    mat4 jointMatrix[768];
} joints;
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
    mat4 skinMatrix =
    i_weights0.x * joints.jointMatrix[int(i_joints0.x + inst.jointOffset)] +
    i_weights0.y * joints.jointMatrix[int(i_joints0.y + inst.jointOffset)] +
    i_weights0.z * joints.jointMatrix[int(i_joints0.z + inst.jointOffset)] +
    i_weights0.w * joints.jointMatrix[int(i_joints0.w + inst.jointOffset)];
    world = world * skinMatrix;
    #endif
    o_UV0 = i_uv0;
    gl_Position = frame.projection * frame.view * world * vec4(i_position, 1.0);
    o_position = vec3(world * vec4(i_position, 1.0));
    o_normalSpace = calcNormalSpace(world);
}
