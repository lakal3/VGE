#version 450
#extension GL_GOOGLE_include_directive : require


#include "../../vscene/input.glsl"

layout(location = 0) out vec3 o_position;
layout(location = 1) out vec2 o_uv0;

#include "shadowframe.glsl"

#ifdef SKINNED
#define SKIN_SET 1
#include "../../vscene/skin.glsl"
#endif

vec3 qtransform( vec4 q, vec3 v ){
    return v + 2.0*cross(cross(v, q.xyz ) + q.w*v, q.xyz);
}

void main() {
    mat4 world = frame.instances[gl_InstanceIndex];
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    o_uv0 = i_uv0;
    vec4 worldPos = world * vec4(i_position, 1);
    vec3 samplePos = worldPos.xyz - frame.lightPos.xyz;
    o_position = qtransform(frame.plane, samplePos);

    // because the origin is at 0 the proj-vector
    // matches the vertex-position
    float fLength = max(length(o_position.xyz), 0.001);

    vec3 pos = o_position / vec3(fLength);

    // calc "normal" on intersection, by adding the
    // reflection-vector(0,0,1) and divide through
    // his z to get the texture coords
    float n = pos.y + 1;
    float n2 = -pos.y + 1;
    pos = pos.y > 0 ? vec3(pos.x / n, pos.z / n, (fLength - frame.minShadow) / (frame.maxShadow - frame.minShadow)) :
        vec3(pos.x / n2, pos.z / n2, -(fLength - frame.minShadow) / (frame.maxShadow - frame.minShadow));
    gl_Position = vec4(pos, 1);
}