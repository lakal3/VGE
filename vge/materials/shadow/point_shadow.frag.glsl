#version 450
#extension GL_GOOGLE_include_directive : require

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;

#include "shadowframe.glsl"



void main() {
    if (i_position.y < 0) {
        discard;
    }
}