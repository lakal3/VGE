#version 450
#extension GL_EXT_nonuniform_qualifier: require

layout (triangles, invocations = 6) in;
layout (triangle_strip, max_vertices = 3) out;

in gl_PerVertex
{
    vec4 gl_Position;
} gl_in[3];

layout (location = 0) in vec3 i_position[];
#if !background
layout (location = 1) in vec2 i_uv0[];
layout (location = 2) in vec3 i_normal[];
layout (location = 3) in vec3 i_color[];
#if normalMap
layout (location = 4) in mat3 i_normalSpace[];
#endif
#endif

layout(location = 0) out vec3 o_position;
#if !background
layout(location = 1) out vec2 o_uv0;
layout (location = 2) out vec3 o_normal;
layout (location = 3) out vec3 o_color;
#if normalMap
layout (location = 4) out mat3 o_normalSpace;
#endif
#endif

#include probe_frame

void main()
{
    for (int i = 0; i < gl_in.length(); i++)
    {
        gl_Layer = gl_InvocationID;
        o_position = i_position[i];
        #if background
        vec3 dir = vec3(probeFrame.views[gl_InvocationID ] * vec4(i_position[i],0));
        gl_Position = probeFrame.projection * vec4(dir, 1);
        #endif
        #if !background
        o_uv0 = i_uv0[i];
        o_normal = i_normal[i];
        o_color = i_color[i];
        #if normalMap
        o_normalSpace = i_normalSpace[i];
        #endif
        gl_Position = probeFrame.projection * probeFrame.views[gl_InvocationID ] * vec4(o_position, 1);
        #endif


        EmitVertex();
    }
    EndPrimitive();
}