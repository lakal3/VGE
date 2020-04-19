#version 450

#define WG_SIZE  16

layout (local_size_x = WG_SIZE, local_size_y = WG_SIZE, local_size_z = 1) in;

layout (set = 0, binding = 0) uniform SETTINGS {
    float width; // target image width
    float height; // target image height
} settings;

layout(set = 0, binding = 1, rgba8) uniform readonly image2D inputImage;
layout(set = 0, binding = 2, rgba8) uniform writeonly image2D outputImage;

void main() {
    int w = int(settings.width);
    int h = int(settings.height);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);
    if (y >= h || x >= w) {
        return;
    }
    vec4 v0 = imageLoad(inputImage, ivec2(x * 2,y * 2));
    vec4 v1 = imageLoad(inputImage, ivec2(x * 2 + 1,y * 2));
    vec4 v2 = imageLoad(inputImage, ivec2(x * 2,y * 2 + 1));
    vec4 v3 = imageLoad(inputImage, ivec2(x * 2 + 1,y * 2 + 1));
    vec4 v = (v0 + v1 + v2 + v3) / 4;
    imageStore(outputImage, ivec2(x,y), v);
}