
#version 450

layout (local_size_x = 32, local_size_y = 32, local_size_z = 1) in;

layout (push_constant) uniform SETTINGS {
    float width; // image width
    float height; // image height
} settings;

layout (set = 0, binding = 0, r11f_g11f_b10f) uniform readonly image2D inputImage;

layout (set = 0, binding = 1, r11f_g11f_b10f) uniform writeonly image2D outputImage;

void main() {
    int w = int(settings.width);
    int h = int(settings.height);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);
    if (y >= h || x >= w) {
        return;
    }
    ivec2 pos = ivec2(x * 2, y * 2);
    vec3 c = vec3(0,0,0);
    int samples = 0;
    for (int j = -4; j <= 4; j += 2) {
        for (int i = -4; i <= 4; i += 2) {
            ivec2 pos2 = pos + ivec2(i,j);
            if (pos2.x >= 0 && pos2.y >= 0 && pos2.x < w && pos2.y < h) {
                c = c + imageLoad(inputImage, pos2).rgb;
                samples++;
            }
        }
    }
    vec4 cf = vec4(c * vec3(1.0/float(samples)),1);
    imageStore(outputImage, ivec2(x,y), cf);
}