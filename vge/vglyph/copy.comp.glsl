#version 450

layout (local_size_x = 16, local_size_y = 16, local_size_z = 1) in;

layout (set = 0, binding = 0) uniform SETTINGS {
    float width; // image width
    float height; // image height
    float xOffset; // X offset in target image
    float yOffset; // Y offset in target image
    float colorIndex; // 0 - red, 1 - green, 2 - blue
    float margin;
} settings;

layout(set = 0, binding = 1) uniform samplerBuffer tx_source;

layout(set = 1, binding = 0, rg32f) uniform writeonly image2D outputImage;

vec2 getSource(int dx, int dy, int w, int h, int m, int colorIndex) {
    if (dx < m || dx >= w - 2 * m || dy < m || dy >= h - 2 * m) {
        return vec2(0);
    }
    vec4 col = texelFetch(tx_source, (dy - m) * (w - 2 * m) + dx - m);
    switch(colorIndex) {
        case 0:
        return col.ra;
        case 1:
        return col.ga;
        case 2:
        return col.ba;
        case 3:
        return col.rr;
        case 4:
        if (col.g > 0.75) {
            return vec2(0,0);   // Transparent
        }
        return vec2(col.r, 1);
    }
    return vec2(0);
}

void main() {
    int w = int(settings.width);
    int h = int(settings.height);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);

    int c = int(settings.colorIndex);
    int m = int(settings.margin);
    if (y >= h || x >= w) {
        return;
    }
    vec2 comp = getSource(x,y,w,h,m,c);
    imageStore(outputImage, ivec2(x + settings.xOffset,y + settings.yOffset), vec4(comp,0,0));
}