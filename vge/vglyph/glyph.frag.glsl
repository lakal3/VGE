#version 450
#extension GL_GOOGLE_include_directive : require

layout(constant_id = 1) const int MAX_GLYPH_SETS =  16;

layout(location = 0) out vec4 o_Color;

layout (location = 1) in vec4 i_uv0_1;
layout (location = 0) in vec2 i_pos;
layout (location = 2) in flat int i_index;

#include "glyph.glsl"

layout(set = 1, binding = 0) uniform sampler2D tx_glyph[MAX_GLYPH_SETS];
layout(set = 1, binding = 1) uniform sampler2DArray tx_mask;

void main() {
    INSTANCE inst = frame.instances[i_index];
    vec4 clip = inst.clip;
    if (clip.x > i_pos.x || clip.y > i_pos.y || clip.z < i_pos.x || clip.w < i_pos.y) {
        discard;
    }
    vec2 uvMask = i_uv0_1.zw;
    vec2 uvGlyph = i_uv0_1.xy;
    vec4 fc = inst.forecolor * texture(tx_mask, vec3(uvMask, inst.uvMask_1.w));
    vec4 bc = inst.backcolor; // * texture(tx_mask, vec3(uvMask, inst.uvMask_2.w));
    int glyphIndex = int(inst.uvGlyph_1.w);
    int kind = int(inst.uvGlyph_2.w);
    if (kind == 2) {
        o_Color = texture(tx_glyph[glyphIndex], uvGlyph);
    } else if (kind == 1) {
        vec2 col = texture(tx_glyph[glyphIndex], uvGlyph).rg;
        float ratio = col.r;
        fc.a = col.g * fc.a;
        bc.a = col.g * bc.a;
        o_Color = ratio * fc + (1 - ratio) * bc;
    } else {
        float ratio = 1.0 - smoothstep(0.45, 0.65, texture(tx_glyph[glyphIndex], uvGlyph).r);
        o_Color = (1 - ratio) * bc + ratio * fc;
    }
}