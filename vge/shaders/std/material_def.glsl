
struct MATERIAL {
    vec4 albedo;
    vec4 emissive;
    float metalness;
    float roughness;
    float opacity;
    uint probe;
    float frozenId;
    vec3 reflectance;
    vec3 specular0;
    vec3 diffuse;
    vec3 normal;
    vec3 viewDir;
    float normalDView;
    vec3 directLight;
    vec3 indirectLight;
} mat;

