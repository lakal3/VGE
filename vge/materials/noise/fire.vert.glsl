#version 450
#extension GL_GOOGLE_include_directive : require

vec2 vertices[6] = vec2[]( vec2(-0.5,0), vec2(-0.5, 1), vec2(0.5,1), vec2(-0.5,0), vec2(0.5,1), vec2(0.5, 0));

layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view;
} frame;

#include "fire.glsl"

layout (location = 0) out vec4 o_posHeat;
layout (location = 1) out vec2 o_uv;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    INSTANCE inst = instance.instances[gl_InstanceIndex];
    vec2 basePos = vertices[gl_VertexIndex];
    o_posHeat = vec4(basePos, inst.heat, inst.heat);
    o_uv = vec2(basePos.x * 0.6 + 0.5, basePos.y * 0.72 - inst.offset);
    // Align to camera
    vec4 align = frame.view * inst.world* vec4(1, 0, 0, 0);
    float angle = -atan(align.z, align.x);
    vec4 pos = vec4(basePos.x * cos(angle), basePos.y, basePos.x * sin(angle), 1);

    vec4 viewPos = frame.view * inst.world * pos;
    gl_Position = frame.projection * viewPos;
}
