

#version 450

#include imageframe

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));

layout (location = 0) in vec2 i_position;
layout (location = 1) in vec2 i_uv0;

layout (location = 0) out vec4 o_color;

void main() {
    if (i_position.x < 0 || i_position.x > 1 || i_position.y < 0 || i_position.y > 1) {
        discard;
    }
    o_color = texture(image, i_uv0);
}