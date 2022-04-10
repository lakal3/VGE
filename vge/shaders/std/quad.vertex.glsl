#version 450

#include frame

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));

layout (location = 0) out vec2 o_position;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    o_position = vertices[gl_VertexIndex];
    vec4 vp = frame.viewPosition;
    vec2 vPos = vec2(o_position.x * (vp.z - vp.x) + vp.x, o_position.y * (vp.w - vp.y) + vp.y);
    gl_Position = vec4(vPos, 0.01, 1);
}