#version 450

layout(location = 0) out vec4 outColor;
layout(location = 1) out vec4 outGray;
layout(location = 0) in vec2 i_uv;

layout(set = 0, binding = 0) uniform sampler2D tx_color[];

void main() {
    outColor = texture(tx_color[1], i_uv);
    float c = (outColor.r + outColor.g + outColor.b) / 3;
    outGray = vec4(c,c,c,1);
}