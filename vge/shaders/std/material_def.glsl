
struct MATERIAL {
    vec4 albedo;
    vec4 emissive;
    float metalness;
    float roughness;
    float opacity;
    uint probe;
    uint meshID;
    vec3 f0;
    float reflectance;
    vec3 specular;
    vec3 diffuse;
    vec3 normal;
    vec3 viewDir;
    float normalDView;
    vec3 directLight;
    vec3 indirectLight;
} mat;

