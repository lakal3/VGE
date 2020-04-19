#version 450

#extension GL_GOOGLE_include_directive : require

#include "../../vscene/frame.glsl"

layout (location = 0) in vec3 position;
layout (location = 0) out vec3 o_pos;

void main() {
    o_pos = position;
    gl_Position = frame.projection * frame.view * vec4(position * frame.far / 2, 1);
}