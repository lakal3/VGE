#version 450
#extension GL_EXT_nonuniform_qualifier: require

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_uv0;

#include frame

#include shadow_frame

#include mesh_instance

void main() {
    if (i_position.y < 0) {
        discard;
    }
        #if DYNAMIC_DESCRIPTORS
    if (instance.alphaCutoff > 0 && instance.tx_albedo > 0) {
        float a = texture(frameImages2D[int(inst.tx_albedo)], i_uv0).a;
        if (a < inst.alphaCutoff) {
            discard;
        }
    }
        #endif
}