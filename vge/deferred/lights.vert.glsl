#version 450

vec2 vertices[3] = vec2[]( vec2(-1,-1), vec2(-1, 3), vec2(3,-1));


layout (location = 0) out vec2 o_position;
layout (location = 1) out flat int o_index;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    o_index = gl_InstanceIndex;
    o_position = vertices[gl_VertexIndex];
    gl_Position = vec4(o_position, 0.01, 1);
}