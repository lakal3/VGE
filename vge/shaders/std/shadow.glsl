#version 450
#extension GL_EXT_nonuniform_qualifier: require

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_uv0;

#include frame

#include shadow_frame

#include mesh_instance

void main() {
#if parabloid
    if (i_position.y < 0) {
        discard;
    }
#endif
    if (instance.alphaCutoff > 0 && instance.textures1.x > 0) {
        float a = texture(frameImages2D[int(instance.textures1.x)], i_uv0).a;
        if (a < instance.alphaCutoff) {
            discard;
        }
    }
}