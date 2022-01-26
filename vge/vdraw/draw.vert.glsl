#version 450
#extension GL_GOOGLE_include_directive : require

#include "draw.glsl"

layout(location = 0) in vec2 i_corner;
layout(location = 1) in vec2 i_origCorner;
layout(location = 2) in vec2 i_segments;

layout(location = 0) out vec2 o_origCorner;
layout(location = 1) out flat uvec2 o_segments;
layout(location = 2) out vec2 o_corner;

out gl_PerVertex
{
    vec4 gl_Position;
};



void main() {
    o_origCorner = i_origCorner;
    o_corner = i_corner;
    o_segments = ivec2(uint(i_segments.x), uint(i_segments.y));
    gl_Position = frame.viewProj * vec4(i_corner, 0, 1);
}