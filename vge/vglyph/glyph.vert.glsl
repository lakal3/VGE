#version 450
#extension GL_GOOGLE_include_directive : require

vec2 vertices[6] = vec2[]( vec2(0,0), vec2(0, 1), vec2(1,1), vec2(0,0), vec2(1,1), vec2(1, 0));

#include "glyph.glsl"

layout (location = 1) out vec4 o_uv0_1;
layout (location = 0) out vec2 o_pos;
layout (location = 2) out flat int o_index;

out gl_PerVertex
{
    vec4 gl_Position;
};

void main() {
    INSTANCE inst = frame.instances[gl_InstanceIndex];
    o_index = gl_InstanceIndex;
    mat3 mPos = mat3(inst.position_1.x, inst.position_2.x, 0,
        inst.position_1.y, inst.position_2.y, 0,
        inst.position_1.z, inst.position_2.z, 1);
    vec3 pos = vec3(vertices[gl_VertexIndex], 1);
    o_pos = vec2(mPos * pos);
    gl_Position = vec4(o_pos, 0.01, 1);
    mat3 mGlyph = mat3(inst.uvGlyph_1.x, inst.uvGlyph_2.x, 0,
        inst.uvGlyph_1.y, inst.uvGlyph_2.y, 0,
        inst.uvGlyph_1.z, inst.uvGlyph_2.z, 1);
    mat3 mMask = mat3(inst.uvMask_1.x, inst.uvMask_2.x, 0,
        inst.uvMask_1.y, inst.uvMask_2.y, 0,
        inst.uvMask_1.z, inst.uvMask_2.z, 1);
    vec3 uvGlyph = mGlyph * pos;
    vec3 uvMask = mMask * pos;
    o_uv0_1 = vec4(uvGlyph.xy, uvMask.xy);
}