#version 450
#extension GL_EXT_nonuniform_qualifier: require
#extension GL_KHR_shader_subgroup_basic: require

const uint groupSize = 32;
layout (local_size_x = 32, local_size_y = 1, local_size_z = 1) in;

const float PI = 3.1415926535897932384626433832795;


layout (push_constant) uniform SETTINGS {
    float width; // image width
    float height; // image height
    float tx_index; // Image index
    float lod; // Lod from view
} settings;

#include frame

layout (set = 1, binding = 0) buffer SPH {
    vec4 vals[9];
} sph;

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
    vec3 vals[9 * groupSize];
    float weights[groupSize];
    uint sums[groupSize];
};

shared COEFF coeff;

void main()
{
    float dy = PI / float(groupSize);
    float y = -PI/2;
    float xStart = 2 * PI * float(gl_LocalInvocationID.x) / float(groupSize);;
    float xEnd = 2 * PI * float(gl_LocalInvocationID.x + 1) /groupSize;
    float dx = PI / (groupSize * 4);
    float pixelArea = dx * dy;
    for (uint i = 0; i < 9; i++) {
        coeff.vals[i + 9 * gl_LocalInvocationID.x] = vec3(0);
    }
    float weightSum = 0;
    uint sum = 0;
    uint tx_index = uint(settings.tx_index);
    for (; y < PI/2; y += dy) {
        float weight = cos(y + dy/2) * pixelArea;

        for (float x = xStart; x < xEnd; x += dx) {
            sum ++;
            weightSum += weight;
            vec3 dir = toDir(x, y);
            vec3 rgb = textureLod(frameImagesCube[tx_index], dir, settings.lod).rgb;
            // rgb = vec3(0.5, 0.5, 0.5);
            // rgb = pow(rgb / vec3(settings.white), vec3(1 / settings.gamma));
            for (uint i = 0; i < 9; i++) {
                coeff.vals[i + 9 * gl_LocalInvocationID.x] += sphFactor(i, dir) * rgb * weight;
            }
        }
    }
    coeff.weights[gl_LocalInvocationID.x] = weightSum;
    coeff.sums[gl_LocalInvocationID.x] = sum;
    subgroupBarrier();
    if (subgroupElect()) {
        float w = 0;
        uint s = 0;
        for (uint i = 0; i < groupSize; i++) {
            w += coeff.weights[i];
            s += coeff.sums[i];
        }
        for (uint i = 0; i < groupSize; i++) {
            for (uint j = 0; j < 9; j++) {
                float debug = 0;
                switch (j) {
                case 0:
                    debug = w;
                    break;
                case 1:
                    debug = float(s);
                    break;
                case 2:
                    debug = dx;
                    break;
                case 3:
                    debug = dy;
                    break;
                case 4:
                    debug = xStart;
                    break;
                case 5:
                    debug = xEnd;
                    break;
                case 6:
                    debug = float(groupSize);
                    break;
                case 7:
                    debug = float(gl_SubgroupSize);
                    break;
                }
                sph.vals[j] = (i == 0 ? vec4(0) :sph.vals[j]) + vec4(coeff.vals[i * 9 + j]/w, i > 0 ? 0.0 :debug);
            }
        }
    }
    // harmonics.sph[pos] = vec4(dx, dy, pixelArea, 0);
    // x = 2 * PI * float(gl_GlobalInvocationID.x) / float(gl_NumWorkGroups.x);
    // harmonics.sph[pos + 1] = vec4(x, xEnd, 0, 0);
    // harmonics.sph[pos] =texture(inputImage, vec3(0.5, 0.5, 0.5));
}
