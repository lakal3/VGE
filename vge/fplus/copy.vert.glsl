#version 450

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));

layout(location = 0) out vec2 o_pos;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    o_pos = vertices[gl_VertexIndex];
    gl_Position = vec4(o_pos, 0.01, 1);
}
