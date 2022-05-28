#version 450
#extension GL_EXT_nonuniform_qualifier: require

#include pick_frame

#include mesh_instance

layout(location = 0) in vec4 i_color;
layout(location = 1) in flat uvec2 i_ids;
layout(location = 0) out vec4 o_color;

void main() {
    switch(instance.idMethod) {
    case 0:
        o_color = vec4(0);
        break;
    case 1:
        o_color = vec4(instance.idValue, 0.0, 0.0, 1.0);
        break;
    case 2:
        o_color = vec4(gl_FragCoord.z, 0.0, 0.0, 1.0);
        break;
    }

    if (gl_FragCoord.x < pickinfo.pickArea.x || gl_FragCoord.y < pickinfo.pickArea.y) {
        return;
    }
    if (gl_FragCoord.x > pickinfo.pickArea.z || gl_FragCoord.y > pickinfo.pickArea.w) {
        return;
    }
    uint pos = atomicAdd(pickinfo.count, 1);
    if (pos >= pickinfo.max) {
        return;
    }

    pickinfo.picks[pos].id = i_ids.x;
    pickinfo.picks[pos].vertex_nro = i_ids.y;
    pickinfo.picks[pos].depth = gl_FragCoord.z;
    pickinfo.picks[pos].idValue = instance.idValue;
    pickinfo.picks[pos].color = i_color;
}