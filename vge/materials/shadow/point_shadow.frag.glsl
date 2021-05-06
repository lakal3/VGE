#version 450
#extension GL_GOOGLE_include_directive : require
#extension GL_EXT_nonuniform_qualifier: require

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 2) in flat int i_index;

#include "shadowframe.glsl"



void main() {
    if (i_position.y < 0) {
        discard;
    }
#if DYNAMIC_DESCRIPTORS
    INSTANCE inst = instances.instances[i_index];
    if (inst.alphaCutoff > 0 && inst.tx_albedo > 0) {
        float a = texture(frameImages2D[int(inst.tx_albedo)], i_UV0).a;
        if (a < inst.alphaCutoff) {
            discard;
        }
    }
#endif
}