#version 450

#include frame

vec3 vertices[4] = vec3[]( vec3(-1, -0.7011, -0.7011), vec3(1, -0.7011, -0.7011), vec3(0, -0.7011, 1), vec3(0, 1, 0));
uint indices[12] = uint[](0, 1, 2, 0, 1, 3, 1, 2, 3, 2, 0, 3);

layout (location = 0) out vec3 o_position;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    o_position = vertices[indices[gl_VertexIndex]];
#if probe
    gl_Position = vec4(o_position, 1);
#endif
#if !probe
    vec3 dir = vec3(frame.view * vec4(o_position,0));
    gl_Position = frame.projection * vec4(dir, 1);
#endif
}
