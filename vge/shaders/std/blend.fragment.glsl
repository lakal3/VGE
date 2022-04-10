#version 450

#include frame

layout(location = 0) out vec4 o_color;
layout (location = 0) in vec2 i_position;

#include blendFunction

void main() {
    vec4 color = texture(frameImages2D[0], i_position);
    o_color = std_blend(color);
}
