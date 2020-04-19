
// Decals (size must be 2 x mat4)
struct DECAL {
    mat4 toDecalSpace;
    vec4 albedoColor;
    vec4 dummy2;
    float metallicFactor;
    float roughnessFactor;
    float normalMap;
    float normalAttenuation;
    float tx_albedo;
    float tx_normal;
    float tx_metalRoughness;
    float dummy3;
};