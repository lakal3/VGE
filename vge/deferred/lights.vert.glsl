#version 450

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));


layout (location = 0) out vec2 o_position;
layout (location = 1) out flat int o_index;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    o_index = gl_InstanceIndex;
    o_position = vertices[gl_VertexIndex];
    gl_Position = vec4(o_position * vec2(2) + vec2(-1), 0.01, 1);
}