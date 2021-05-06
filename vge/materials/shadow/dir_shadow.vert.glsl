#version 450
#extension GL_GOOGLE_include_directive : require


#include "../../vscene/input.glsl"

layout(location = 0) out vec3 o_position;
layout(location = 1) out vec2 o_uv0;
layout(location = 2) out flat int o_index;

#include "shadowframe.glsl"

#ifdef SKINNED
#define SKIN_SET 1
#include "../../vscene/skin.glsl"
#endif

vec3 qtransform( vec4 q, vec3 v ){
    return v + 2.0*cross(cross(v, q.xyz ) + q.w*v, q.xyz);
}

float centerAdjust(float f) {
    return f;
    // return f > 0 ? sqrt(f) : -sqrt(-f);
}

void main() {
    o_index = gl_InstanceIndex;
    mat4 world = instances.instances[gl_InstanceIndex].world;
    #ifdef SKINNED
    world = world * skinMatrix();
    #endif
    o_uv0 = i_uv0;
    vec4 worldPos = world * vec4(i_position, 1);
    vec3 samplePos = worldPos.xyz - frame.lightPos.xyz;
    o_position = qtransform(frame.plane, samplePos);
    vec3 pos = o_position / vec3(frame.maxShadow);
    gl_Position = vec4(centerAdjust(pos.x), centerAdjust(pos.z), pos.y, 1);
}