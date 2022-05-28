#version 450

#include frame

layout(location = 0) out vec4 o_color;

layout(push_constant) uniform INSTANCE {
    vec4 color;
    int lineWidth;
} instance;

void main() {
    ivec2 pos = ivec2(gl_FragCoord.x, gl_FragCoord.y);
    int lw = 1;
    float count = 0;
    float sum = 0;
    for (int y = -lw; y <= lw; y++) {
        for (int x = -lw; x <= lw; x++) {
            sum += texelFetch(frameImages2D[2], pos + ivec2(x,y), 0).r;
            count += 1.0;
        }
    }
    float tot = sum/count;
    if (tot < 0.05 || tot > 0.95) {
        discard;
    }
    o_color = vec4(0.8,0.8,0.8,1);
}