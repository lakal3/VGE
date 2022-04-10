#version 450
#extension GL_EXT_nonuniform_qualifier: require

// Inputs
layout (location = 0) in vec3 i_position;
layout (location = 1) in vec2 i_uv0;
layout (location = 2) in vec3 i_normal;
layout (location = 3) in vec3 i_tangent;
layout (location = 4) in vec4 i_color;
#if jointInput
layout (location = #r jointInput#) in vec4 i_weights0;
layout (location = #r jointInput + 1#) in uvec4 i_joints0;
#endif

// Outputs
#if !depth_only
layout (location = 0) out vec3 o_position;
layout (location = 1) out vec2 o_uv0;
layout (location = 2) out vec3 o_normal;
layout (location = 3) out vec3 o_color;
    #if normalMap
    layout (location = 4) out mat3 o_normalSpace;
    #endif
#endif

#if skinSet
layout(set=#r skinSet#, binding=0) uniform JOINTS {
    mat4 jointMatrix[#r noJoints#];
} joints;
#endif

out gl_PerVertex
{
    vec4 gl_Position;
};

#include frame

#if probe
#include probe_frame
#endif

#include mesh_instance

#if normalMap
mat3 calcNormalSpace(mat4 world) {
    vec3 normal = normalize(vec3(world * vec4(i_normal,0)));
    vec3 tangent = normalize(vec3(world * vec4(i_tangent,0)));
    vec3 biTangent = -normalize(cross(tangent, normal));
    return mat3(tangent, biTangent, normal);
}
#endif

#if skinSet&jointInput
#include skin_matrix
#endif

void main() {
    mat4 world = instance.world;
#if skinSet&jointInput
    world = world * skinMatrix();
#endif
    vec4 vPos = world * vec4(i_position, 1.0);
#if probe
    gl_Position = vPos ;
#endif
#if !probe
    gl_Position = frame.projection * frame.view * vPos ;
#endif
#if !depth_only
    o_uv0 = i_uv0;
    o_normal = normalize(vec3(world * vec4(i_normal,0)));
    #if normalMap
        o_normalSpace = calcNormalSpace(world);
    #endif
    o_position = vec3(vPos);
#endif
}