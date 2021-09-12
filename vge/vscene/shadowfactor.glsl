
#define SAMPLES 5
#define PCF 1
const float shadowBias = 0.002;

vec2 sampleOffsetDirections[5] = vec2[] (
vec2( 0, 0), vec2( 1,  1), vec2( 1, -1), vec2(-1, -1), vec2(-1,  1)
);

// From: https://community.khronos.org/t/quaternion-functions-for-glsl/50140/2
// Transform (rotate) vector by quoternion
vec3 qtransform( vec4 q, vec3 v ){
    return v + 2.0*cross(cross(v, q.xyz ) + q.w*v, q.xyz);
}


float getShadowFactor(LIGHT l, vec3 worldPos) {
    vec4 lightPos = l.position;
    float maxDist = l.attenuation.w;
    vec3 samplePos = worldPos - lightPos.xyz;
    float fLength = length(samplePos.xyz);
    float spFactor = 1.0;
    if (l.position.w == 2) { // Spotlight visibility check
        // inner and outer angle are cos(given angle)
        float dotAngle = dot(normalize(samplePos), l.direction.xyz);
        if (dotAngle < l.outerAngle) {
            return 0;
        }
        if (dotAngle < l.innerAngle) {
            // Linear fade from inner angle to outer angle
            spFactor = (dotAngle - l.outerAngle) / (l.innerAngle - l.outerAngle);
        }
    }
    if (fLength > maxDist) {
        return l.attenuation.x > 0 ? spFactor: 0;  // If we have constant attenuation, light is visible outside shadow range
    }
    int shadowMethod = int(l.shadowMapMethod);
    int shadowMapIdx = int(l.shadowMapIndex);
    #ifndef DYNAMIC_DESCRIPTORS
    if (shadowMapIdx >= MAX_IMAGES) {
        return spFactor;
    }
    #endif

    // Method 1, older point light shadow cubemap
    if (shadowMethod == 1) {
        #if PCF
        float sum = 0;
        vec3 l2f = normalize(samplePos);
        vec3 bt = vec3(0, 1, 0);
        if (abs(l2f.y) > 0.7) {
            bt = vec3(0, 0, 1);
        }
        vec3 tn = cross(l2f, bt);
        bt = cross(l2f, tn);
        for (int i = 0; i < SAMPLES; i++) {
            float z = texture(frameImagesCube[shadowMapIdx], l2f + sampleOffsetDirections[i].x * tn * 0.002 +
            + sampleOffsetDirections[i].y * bt * 0.002).r;
            sum += (fLength < (z + shadowBias) * maxDist) ? 1 : 0;
        }
        return sum / SAMPLES * spFactor;
        #else
        float z = texture(frameImagesCube[shadowMapIdx], l2f).r;
        return (lr < (z + shadowBias) * maxDist) ? 1 : 0;
        #endif
        // return clamp((z * maxDist - lr) * 5  + 0.5, 0, 1);
        // return z;

    }

    // Method 2 intentionally missing

    // Point light (3) or Spot light 4.
    // Spot light have plane transform in quoternion l.shadowPlane.
    // Point lights have always 2 maps with 0,-1,0 and 0,1,0 axis
    if (shadowMethod == 3 || shadowMethod == 4) {
        if (shadowMethod == 4) {
            samplePos = qtransform(l.shadowPlane, samplePos);
        }
        samplePos = samplePos / vec3(fLength);
        if (fLength > maxDist) {
            return 0;
        }
        // calc "normal" on intersection, by adding the
        // reflection-vector(0,0,1) and divide through
        // his z to get the texture coords
        float d = 0;
        float sum = 0;
        #if PCF
        for (int idx = 0; idx < SAMPLES; idx++) {
            vec2 offset = sampleOffsetDirections[idx] * vec2(0.001);
            #else
            vec2 offset = vec2(0);
            #endif

            if (shadowMethod == 4) {
                float n = samplePos.y + 1;
                vec2 pos = vec2(samplePos.x / n, samplePos.z / n) * vec2(0.5, 0.5) + vec2(0.5, 0.5) + offset;
                d = texture(frameImages2D[shadowMapIdx], pos).x;
            } else if (samplePos.y < 0 ) {
                float n = -samplePos.y + 1;
                vec3 pos = vec3(samplePos.x / n, samplePos.z / n, 0) * vec3(0.5, 0.5, 1) + vec3(0.5, 0.5, 0) + vec3(offset,0);
                // pos.y = 1 - pos.y;
                d = texture(frameImagesArray[shadowMapIdx], pos).x;
            } else {
                float n = samplePos.y + 1;
                vec3 pos = vec3(samplePos.x / n, samplePos.z / n, 1) * vec3(0.5, 0.5, 1) + vec3(0.5, 0.5, 0) + vec3(offset,0);
                // pos.y = 1 - pos.y;
                d = texture(frameImagesArray[shadowMapIdx + 1], pos).x;
            }
            sum += d + shadowBias > (fLength / maxDist) ? spFactor : 0;
            #if PCF
        }
        return sum / SAMPLES * spFactor;
        #else
        returm sum * spFactor;
        #endif

    }
    if (shadowMethod == 5) {
        samplePos = qtransform(l.shadowPlane, samplePos);
        float sum = 0;
        #if PCF
        for (int idx = 0; idx < SAMPLES; idx++) {
            vec2 offset = sampleOffsetDirections[idx] * vec2(0.001);
            #else
            vec2 offset = vec2(0);
            #endif
            vec2 pos = samplePos.xz / vec2(maxDist) * vec2(0.5, 0.5) + vec2(0.5, 0.5) + offset;
            float d = texture(frameImages2D[shadowMapIdx], pos).x;
            sum += d + shadowBias > (samplePos.y / maxDist) ? spFactor : 0;
            #if PCF
        }
        return sum / SAMPLES * spFactor;
        #else
        returm sum * spFactor;
        #endif
    }
    // TODO: Directional light
    return spFactor;
}
/* OLD version
float getShadowFactor(LIGHT l, vec3 position) {
    int shadowMethod = int(l.shadowMapMethod);
    int shadowMapIdx = int(l.shadowMapIndex);
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
*/