#version 450
#extension GL_GOOGLE_include_directive : require

layout (triangles, invocations = 6) in;
layout (triangle_strip, max_vertices = 3) out;
layout(constant_id = 0) const int MAX_INSTANCES = 100;

layout(location = 0) in flat int i_index[];
layout(location = 1) in vec2 i_UV0[];

layout(location = 0) out vec3 o_position;
layout(location = 1) out vec2 o_UV0;
layout(location = 2) out flat int o_index;

#include "sframe.glsl"

void main()
{
    for (int i = 0; i < gl_in.length(); i++)
    {
        gl_Layer = gl_InvocationID  ;
        o_index = i_index[i];
        o_UV0 = i_UV0[i];
        vec4 tmpPos = gl_in[i].gl_Position;
        o_position = vec3(tmpPos);
        gl_Position = frame.projection * frame.view[gl_InvocationID ] * tmpPos;
        EmitVertex();
    }
    EndPrimitive();
}