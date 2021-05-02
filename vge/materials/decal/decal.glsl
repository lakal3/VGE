
#define MAX_DECALS 256

// Decals
struct DECAL {
    mat4 toDecalSpace;
    vec4 albedoColor;
    float metallicFactor;
    float roughnessFactor;
    float normalAttenuation;
    float tx_albedo;
    float tx_normal;
    float tx_metalRoughness;
    float filler1;
    float filler2;
};

layout(set=3, binding=0) uniform DECALS {
    float noDecals;
    float filler1;
    float filler2;
    float filler3;
    DECAL decals[MAX_DECALS];
} decals;

vec3 lerp3(vec3 base, vec3 target, float factor) {
    return base * (1 - factor) + target * factor;
}

float lerp1(float base, float target, float factor) {
    return base * (1 - factor) + target * factor;
}



void calcDecal(int idx, vec3 pos, inout vec3 albedo, inout vec3 normal, inout float metallic, inout float roughness, vec3 refNormal) {
    DECAL d = decals.decals[idx];
    vec4 dp = d.toDecalSpace * vec4(pos, 1);
    if (dp.x < -1 || dp.x > 1 || dp.y < -1 || dp.y > 1 || dp.z < -1 || dp.z > 1) {
        return;
    }
    vec3 dNormal = vec3(d.toDecalSpace * vec4(normal, 0));
    vec4 aBase = d.albedoColor;
    vec2 samplePoint = dp.xz * vec2(0.5) + vec2(0.5);
    if (d.tx_albedo > 0) {
        aBase = aBase * texture(frameImages2D[int(d.tx_albedo)], samplePoint);
    }
    float factor = aBase.a;
    vec3 aNormal = vec3(0, 1, 0);
    factor = dot(aNormal, refNormal) * d.normalAttenuation * factor + (1 - d.normalAttenuation) * factor;
    if (factor < 0.01) {
        return;
    }
    if (d.tx_normal > 0) {
        aNormal = vec3(texture(frameImages2D[int(d.tx_normal)], samplePoint)) * vec3(2) - vec3(1);
        aNormal = normalize(aNormal.xzy);   // Convert to Y - up
    }
    float aMetallic = d.metallicFactor;
    float aRoughness = d.roughnessFactor;
    if (d.tx_metalRoughness > 0) {
        vec3 mrColor = texture(frameImages2D[int(d.tx_normal)], samplePoint).rgb;
        aMetallic = mrColor.b * aMetallic;
        aRoughness = mrColor.g * aRoughness;
    }
    albedo = lerp3(albedo, vec3(aBase), factor);
    // adjust normal if decal has bump map
    if (d.tx_normal > 0) {
        vec3 wNormal = vec3(inverse(d.toDecalSpace) * vec4(aNormal, 0));
        normal = lerp3(normal, aNormal, factor);
    }
    metallic = lerp1(metallic,aMetallic,factor);
    roughness = lerp1(roughness,aRoughness,factor);
}
