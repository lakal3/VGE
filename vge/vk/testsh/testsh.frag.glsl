#version 450

layout(location = 0) out vec4 outColor;
layout(location = 0) in vec2 i_uv;

layout(set = 0, binding = 0) uniform sampler2D tx_color[];

void main() {
    outColor = texture(tx_color[1], i_uv);
}