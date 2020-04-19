#version 450

layout (set = 1, binding = 0) uniform sampler2D tx_environment;

layout (location = 0) in vec3 i_pos;
layout (location = 0) out vec4 o_color;

const float PI = 3.1415926535897932384626433832795;

vec2 getPolarPosition(vec3 sampleVector) {
    vec3 nPos = normalize(sampleVector);
    vec2 xy = vec2(atan(nPos.x, nPos.z) + PI, acos(nPos.y));
    return xy * vec2(1 / (2 * PI), 1 /  PI);
}

void main() {
    vec2 xy = getPolarPosition(i_pos);
    o_color = texture(tx_environment, xy);
    // o_color = vec4(xy, 0, 1);
}