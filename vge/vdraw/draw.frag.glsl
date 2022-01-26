#version 450
#extension GL_EXT_nonuniform_qualifier: require
#extension GL_GOOGLE_include_directive : require

layout(set=1, binding=0) buffer SEGMENTS {
    vec4 segments[];
} segments;

#include "draw.glsl"

layout(set=1, binding=1) uniform sampler2D tx_masks[];
layout(set=1, binding=1) uniform sampler2DArray tx_glyphs[];

layout(location = 0) in vec2 i_pos;
layout(location = 1) in vec2 i_wpos;
layout(location = 2) in vec2 i_uv0;
layout(location = 3) in vec2 i_uvGlyph;
layout(location = 4) in flat uvec2 i_segments;

layout(location = 0) out vec4 o_color;

float getRatio(float prevY, float nextY) {
    float dp = abs(prevY - i_pos.y);
    float dn = abs(nextY - i_pos.y);
    return dp + dn < 0.00001 ? 1 : dp / (dp + dn);
}

vec2 leftRight() {
    for (uint idx = i_segments.x; idx < i_segments.y; idx++) {
        if (segments.segments[idx].z >= i_pos.y) {
            uint prevIdx = idx > i_segments.x ? idx - 1: i_segments.x;
            vec4 sNext = segments.segments[idx];
            vec4 sPrev = segments.segments[prevIdx];
            float ratio = getRatio(sPrev.z, sNext.z);
            float xLeft = sPrev.x * (1 - ratio) + sNext.x * ratio;
            float xRight = sPrev.y * (1 - ratio) + sNext.y * ratio;
            return vec2(xLeft, xRight);
        }
    }
    return segments.segments[i_segments.y - 1].xy;
}

float checkVisible() {
    vec2 lr = leftRight();
    float maxOffset = max(lr.x - i_pos.x, i_pos.x - lr.y);
    return clamp(maxOffset >= 0 ? 0.4 - maxOffset * inst.scale.x  : 0.4 - maxOffset * inst.scale.x * 2
     , 0, 1);
}


float rangeAt() {
    vec2 lr = leftRight();
    return abs(lr.x - i_pos.x) < abs(lr.y - i_pos.x) ?
    1 - clamp(abs(lr.x - i_pos.x) * inst.scale.x, 0, 1) : 1 - clamp(abs(lr.y - i_pos.x) * inst.scale.x, 0, 1);
}

void main() {
    if (i_wpos.x < inst.clip.x || i_wpos.x > inst.clip.z || i_wpos.y < inst.clip.y || i_wpos.y > inst.clip.w) {
        discard; // Clipped
    }
    int glyphImage = int(inst.glyph.x);
    float a = 0;
#if GLYPH
    float r = texture(tx_glyphs[glyphImage], vec3(i_uvGlyph, inst.glyph.y)).r;
    a = clamp((0.4-r) * 0.75 , 0, 1);
#else
    a = checkVisible();
#endif
    vec4 c = inst.color + fract(i_uv0.x) * inst.color2;
    int textImage = int(inst.uv1.w);
    if (textImage > 0) {
        c = c * texture(tx_masks[textImage], i_uv0);
    }

    o_color = vec4(c.rgb, c.a * a);

}
