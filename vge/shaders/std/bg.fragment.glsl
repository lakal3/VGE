#version 450
#extension GL_EXT_nonuniform_qualifier: require

#include frame

layout(location = 0) out vec4 o_color;
#if !probe
layout(location = 1) out vec4 o_color2;
#endif
layout (location = 0) in vec3 i_position;

const float PI = 3.1415926535897932384626433832795;

vec2 getPolarPosition(vec3 sampleVector) {
    vec3 nPos = normalize(sampleVector);
    vec2 xy = vec2(atan(nPos.x, nPos.z) + PI, acos(nPos.y));
    return xy * vec2(1 / (2 * PI), 1 /  PI);
}

layout(push_constant) uniform INSTANCE {
    vec4 baseColor;
    vec4 upColor;
    float txIndex;
} inst;

void main() {
    vec2 xy = getPolarPosition(i_position);
    int tx = int(inst.txIndex);
    if (tx > 0) {
        o_color = texture(frameImages2D[tx], xy);
    } else {
        vec4 difColor = inst.upColor - inst.baseColor;
        o_color = inst.baseColor + difColor * xy.y;
    }
#if !probe
    o_color2 = o_color;
#endif
}