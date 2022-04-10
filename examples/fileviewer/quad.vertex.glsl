#version 450

#include imageframe

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));

layout (location = 0) out vec2 o_position;
layout (location = 1) out vec2 o_uv0;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    vec2 position = vertices[gl_VertexIndex];
    mat3 mpos = mat3(imageframe.pos1.x, imageframe.pos2.x, 0,
        imageframe.pos1.y, imageframe.pos2.y, 0,
        imageframe.pos1.z, imageframe.pos2.z, 1);
    o_position = vec2(mpos * vec3(position, 1));
    o_uv0 = vec2(mpos * vec3(position, 0)) * vec2(imageframe.pos1.w);
    switch (uint(imageframe.pos2.w)) {
        case 1:
            o_uv0 = vec2(1 - o_uv0.y, o_uv0.x);
            break;
        case 2:
            o_uv0 = vec2(1 - o_uv0.x, 1- o_uv0.y);
            break;
        case 3:
            o_uv0 = vec2(o_uv0.y, 1 - o_uv0.x);
            break;
    }
    vec4 area = imageframe.area;
    vec2 vPos = vec2(o_position.x * (area.z - area.x) + area.x, o_position.y * (area.w - area.y) + area.y);
    gl_Position = vec4(vPos, 0.01, 1);
}