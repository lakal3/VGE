#version 450
#extension GL_GOOGLE_include_directive : require

layout(set = 2, binding = 0) uniform sampler2D tx_noise;

layout (location = 0) out vec4 o_color;
layout (location = 0) in vec4 i_posHeat;
layout (location = 1) in vec2 i_uv;


void main() {
    float h = i_posHeat.z;
    float xOffs = sin(i_posHeat.w * 2 + i_posHeat.y) * 0.2;
    h = h * texture(tx_noise, i_uv + vec2(xOffs,0)).r;
    h -= i_posHeat.z * 1.5 * abs(i_posHeat.x);
    h -= i_posHeat.z * abs(i_posHeat.y - 0.2);
    o_color = vec4(clamp(h / 400, 0, 1), clamp(h/800, 0, 1), h / 1200, clamp(h / 300, 0, 1));
}