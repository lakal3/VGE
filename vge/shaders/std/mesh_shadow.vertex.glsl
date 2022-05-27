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
layout (location = 0) out vec3 o_position;
layout (location = 1) out vec2 o_uv0;

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

#include shadow_frame

#include mesh_instance

#if skinSet&jointInput
    #include skin_matrix
#endif

vec3 qtransform( vec4 q, vec3 v ){
    return v + 2.0*cross(cross(v, q.xyz ) + q.w*v, q.xyz);
}


void main() {
    mat4 world = instance.world;
    #if skinSet&jointInput
    world = world * skinMatrix();
    #endif
    o_uv0 = i_uv0;
    vec4 worldPos = world * vec4(i_position,1);
#if paraboloid
    vec3 samplePos = worldPos.xyz - shadowFrame.lightPos.xyz;
    if (shadowFrame.yFactor != 0) {
        o_position = samplePos * vec3(1, shadowFrame.yFactor, 1);
    } else {
        o_position = qtransform(shadowFrame.plane, samplePos);
    }

    // because the origin is at 0 the proj-vector
    // matches the vertex-position
    float fLength = max(length(o_position.xyz), 0.001);

    vec3 pos = o_position / vec3(fLength);

    // calc "normal" on intersection, by adding the
    // reflection-vector(0,0,1) and divide through
    // his z to get the texture coords
    float n = max(0.01,pos.y + 1);
    pos = vec3(pos.x / n, pos.z / n, (fLength - shadowFrame.minShadow) / (shadowFrame.maxShadow - shadowFrame.minShadow));
#endif
#if !paraboloid
    vec3 samplePos = worldPos.xyz - frame.cameraPos.xyz;
    o_position = qtransform(shadowFrame.plane, samplePos);
    vec3 pos = o_position;
    pos = pos * vec3(1 / shadowFrame.maxShadow) * vec3(1,1,0.5) + vec3(0,0,0.5);
#endif
    gl_Position = vec4(pos, 1);
}