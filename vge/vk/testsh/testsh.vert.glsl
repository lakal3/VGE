#version 450

layout (location = 0) in vec2 i_position;
layout (location = 0) out vec2 o_uv;

layout(set = 1, binding = 0) uniform WorldUBF {
    mat4 world;
} World;

void main() {
    o_uv = i_position;
    gl_Position = World.world * vec4(i_position, 0.0, 1.0);
}