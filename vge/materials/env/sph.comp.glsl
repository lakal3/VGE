#version 450

#define  LOCAL_SIZE 32
layout (local_size_x = LOCAL_SIZE, local_size_y = 1, local_size_z = 1) in;

const float PI = 3.1415926535897932384626433832795;

layout (set = 0, binding = 0) uniform SETTINGS {
    float gamma;
    float white;
    float width; // image width
    float height; // image height
} settings;

layout (set = 0, binding = 1) uniform samplerCube inputImage;

layout (set = 0, binding = 2) writeonly buffer HARMONICS
{
    vec4 sph[];
} harmonics;

vec3 toDir(float phi, float theta) {
    float r = cos(theta);
    return vec3(r * cos(phi - PI), sin(theta), r * sin(phi - PI));
}

float sphFactor(uint order, const vec3 dir) {
    switch (order) {
        case 0:
        // 0.5 * sqrt(1/pi)
        return 0.282095;
        case 1:
        // -sqrt(3/(4pi)) * y
        return -0.488603 * dir.y;
        case 2:
        // sqrt(3/(4pi)) * z
        return 0.488603 * dir.z;
        case 3:
        // -sqrt(3/(4pi)) * x
        return -0.488603 * dir.x;
        case 4:
        // 0.5 * sqrt(15/pi) * x * y
        return 1.092548 * dir.x * dir.y;
        case 5:
        // -0.5 * sqrt(15/pi) * y * z
        return -1.092548 * dir.y * dir.z;
        case 6:
        // 0.25 * sqrt(5/pi) * (-x^2-y^2+2z^2)
        // Changed to 3z^2 - 1
        return 0.315392 * (3.0 * dir.z * dir.z - 1);
        case 7:
        // -0.5 * sqrt(15/pi) * x * z
        return -1.092548 * dir.x * dir.z;
        case 8:
        // 0.25 * sqrt(15/pi) * (x^2 - y^2)
        return 0.546274 * (dir.x * dir.x - dir.y * dir.y);
    }
    return 0.0;
}

struct COEFF
{
    vec3 vals[9];
} coeff;

void main()
{
    uint pos = gl_GlobalInvocationID.x * 9;
    float dy = PI / settings.height;
    float y = -PI/2;
    float xStart = 2 * PI * float(gl_GlobalInvocationID.x) / float(gl_NumWorkGroups.x * LOCAL_SIZE);
    float xEnd = 2 * PI * float(gl_GlobalInvocationID.x + 1) / float(gl_NumWorkGroups.x * LOCAL_SIZE);
    float dx = 2 * PI / settings.width;
    float pixelArea = dx * dy;
    for (uint i = 0; i < 9; i++) {
        coeff.vals[i] = vec3(0);
    }
    float weightSum = 0;
    for (; y < PI/2; y += dy) {
        float weight = cos(y + dy/2) * pixelArea;

        for (float x = xStart; x < xEnd; x += dx) {
            weightSum += weight;
            vec3 dir = toDir(x, y);
            vec3 rgb = texture(inputImage, dir).rgb;
            // rgb = vec3(0.5, 0.5, 0.5);
            rgb = pow(rgb / vec3(settings.white), vec3(1 / settings.gamma));
            for (uint i = 0; i < 9; i++) {
                coeff.vals[i] += sphFactor(i, dir) * rgb * weight;
            }
        }
    }
    for (uint i = 0; i < 9; i++) {
        harmonics.sph[pos + i] = vec4(coeff.vals[i], i == 0 ? weightSum: 0);
    }
    // harmonics.sph[pos] = vec4(dx, dy, pixelArea, 0);
    // x = 2 * PI * float(gl_GlobalInvocationID.x) / float(gl_NumWorkGroups.x);
    // harmonics.sph[pos + 1] = vec4(x, xEnd, 0, 0);
    // harmonics.sph[pos] =texture(inputImage, vec3(0.5, 0.5, 0.5));
}

