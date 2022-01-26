#version 450
#extension GL_GOOGLE_include_directive : require

#include "draw.glsl"

layout (lines, invocations = 1) in;
layout (triangle_strip, max_vertices = 6) out;

layout(location = 0) in vec2 i_origCorner[];
layout(location = 1) in flat uvec2 i_segments[];
layout(location = 2) in vec2 i_corner[];

layout(location = 0) out vec2 o_pos;
layout(location = 1) out vec2 o_wpos;
layout(location = 2) out vec2 o_uv0;
layout(location = 3) out vec2 o_uvGlyph;
layout(location = 4) out flat uvec2 o_segments;

out gl_PerVertex
{
    vec4 gl_Position;
};



void setConst() {
    o_segments = i_segments[0];
    // o_mask_index = uint(inst.uv1.w);
    gl_Position = frame.viewProj * vec4(o_wpos, 0, 1);
}

void main()
{
    mat3 mUV = mat3(inst.uv1.x, inst.uv2.x, 0,
    inst.uv1.y, inst.uv2.y, 0,
    inst.uv1.z, inst.uv2.z, 1);

    o_pos = vec2(i_origCorner[1].x, i_origCorner[0].y);
    o_uv0 = vec2(mUV * vec3(o_pos, 1));
    o_wpos = vec2(i_corner[1].x, i_corner[0].y);
    o_uvGlyph = vec2(1,0);
    setConst();
    EmitVertex();

    o_pos = i_origCorner[0];
    o_uv0 = vec2(mUV * vec3(o_pos, 1));
    o_wpos = i_corner[0];
    o_uvGlyph = vec2(0,0);
    setConst();
    EmitVertex();

    o_pos = i_origCorner[1];
    o_uv0 = vec2(mUV * vec3(o_pos, 1));
    o_wpos = i_corner[1];
    o_uvGlyph = vec2(1,1);
    setConst();
    EmitVertex();

    o_pos = vec2(i_origCorner[0].x, i_origCorner[1].y);
    o_uv0 = vec2(mUV * vec3(o_pos, 1));
    o_wpos = vec2(i_corner[0].x, i_corner[1].y);
    o_uvGlyph = vec2(0,1);
    setConst();
    EmitVertex();

    EndPrimitive();
}