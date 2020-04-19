#version 450

layout (set = 1, binding = 0) uniform COLORS {
    vec4 baseColor;
    vec4 upColor;
} colors;

layout (location = 0) in vec3 i_pos;
layout (location = 0) out vec4 o_color;

const float PI = 3.1415926535897932384626433832795;

void main() {
    float y = normalize(i_pos).y;
    vec4 difColor = colors.upColor - colors.baseColor;
    o_color = colors.baseColor + difColor * y;
}