#version 450

#extension GL_GOOGLE_include_directive : require

layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view;
} frame;

layout (location = 0) in vec3 position;
layout (location = 0) out vec3 o_pos;

void main() {
    o_pos = position;
    float c = frame.projection[2][2];
    float d = frame.projection[3][2];
    float far = d / ( c + 1);
    gl_Position = frame.projection * frame.view * vec4(position * far / 2, 1);
}