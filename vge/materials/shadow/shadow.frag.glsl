#version 450
#extension GL_GOOGLE_include_directive : require

#define MAX_INSTANCES 1

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 2) in flat int i_index;

#include "sframe.glsl"

void main() {
    float lPos = length(vec3(frame.lightPos) - i_position);
    gl_FragDepth = lPos / frame.lightPos.w;
}