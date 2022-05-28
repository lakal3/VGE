#version 450


vec2 vertices[3] = vec2[]( vec2(-1,-1), vec2(-1, 3), vec2(3,-1));


out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    gl_Position = vec4(vertices[gl_VertexIndex], 0.5, 1);
}