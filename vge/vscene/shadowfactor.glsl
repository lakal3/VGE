
#define SAMPLES 5
#define PCF 1
const float shadowBias = 0.002;

vec2 sampleOffsetDirections[5] = vec2[] (
vec2( 0, 0), vec2( 1,  1), vec2( 1, -1), vec2(-1, -1), vec2(-1,  1)
);

float getShadowFactor(LIGHT l, vec3 position) {
    int shadowMethod = int(l.shadowMapMethod);
    int shadowMapIdx = int(l.direction.w);
#ifndef DYNAMIC_DESCRIPTORS
    if (shadowMapIdx >= MAX_IMAGES) {
        return 1;
    }
#endif
    vec4 lightPos = l.position;
    float maxDist = l.attenuation.w;
    if (shadowMethod == 1) {   // Point light shadow cubemap
        vec3 l2f = position - vec3(lightPos);
        float lr = length(l2f);

        #if PCF
        float sum = 0;
        l2f = normalize(l2f);
        vec3 bt = vec3(0, 1, 0);
        if (abs(l2f.y) > 0.7 * lr) {
            bt = vec3(0, 0, 1);
        }
        vec3 tn = cross(l2f, bt);
        bt = cross(l2f, tn);
        for (int i = 0; i < SAMPLES; i++) {
            float z = texture(frameImagesCube[shadowMapIdx], l2f + sampleOffsetDirections[i].x * tn * 0.002 +
            + sampleOffsetDirections[i].y * bt * 0.002).r;
            sum += (lr < (z + shadowBias) * maxDist) ? 1 : 0;
        }
        return sum / SAMPLES;
        #else
        float z = texture(frameImagesCube[shadowMapIdx], l2f).r;
        return (lr < (z + shadowBias) * maxDist) ? 1 : 0;
        #endif
        // return clamp((z * maxDist - lr) * 5  + 0.5, 0, 1);
        // return z;

    }
    // TODO: Directional light
    return 0;
}
