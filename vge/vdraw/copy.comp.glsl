#version 450
#extension GL_EXT_nonuniform_qualifier: require

layout (local_size_x = 16, local_size_y = 16, local_size_z = 1) in;

layout (set = 0, binding = 0) buffer SETTINGS {
    float width; // image width
    float height; // image height
    float dummy1;
    float dummy2;
    float pixels[];
} settings;

layout(set = 1, binding = 0, r8_snorm) uniform writeonly image2D outputImage;

void main() {
    int w = int(settings.width);
    int h = int(settings.height);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);
    if (y > h || x > w) {
        return;
    }
    vec4 c = vec4(clamp(settings.pixels[w * y + x] / 5.0, -1, 1));
    imageStore(outputImage, ivec2(x,y),c);
}
