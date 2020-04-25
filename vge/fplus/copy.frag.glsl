#version 450

layout(location = 0) in vec2 i_pos;
layout(location = 0) out vec4 o_Color;

layout(set=0, binding=0) uniform sampler2D tx_image[2];

void main() {
    o_Color = texture(tx_image[0], i_pos);
}
