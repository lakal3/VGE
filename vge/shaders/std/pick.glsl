#version 450
#extension GL_EXT_nonuniform_qualifier: require

#include pick_frame

layout(location = 0) in vec4 i_color;
layout(location = 1) in flat uvec2 i_ids;

void main() {
    if (gl_FragCoord.x < pickinfo.pickArea.x || gl_FragCoord.y < pickinfo.pickArea.y) {
        discard;
    }
    if (gl_FragCoord.x > pickinfo.pickArea.z || gl_FragCoord.y > pickinfo.pickArea.w) {
        discard;
    }
    uint pos = atomicAdd(pickinfo.count, 1);
    if (pos >= pickinfo.max) {
        discard;
    }

    pickinfo.picks[pos].id = i_ids.x;
    pickinfo.picks[pos].vertex_nro = i_ids.y;
    pickinfo.picks[pos].depth = gl_FragCoord.z;
    pickinfo.picks[pos].colorR = i_color.r;
}