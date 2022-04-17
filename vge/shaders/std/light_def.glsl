
struct LIGHT {
    vec3 intensity;
    vec4 position; // if w = 0, directional light, 1 = point light
    vec4 direction; // direction for directional light, quoternion for spot light
    vec4 attenuation; //
    float innerAngle;
    float outerAngle;
    float minDistance;
    float maxDistance;
    uint kind;
    uint tx_shadowmap;
    vec3 lightToPos; // vector from light source to current position
    float lenToPos;
    vec3 lightDir;
    vec3 halfDir; // Middle vector between view direction and light direction

    vec3 radiance;
    float shadowFactor;
    // Dot products
    float normalDHalf;
    float normalDLight;
    float lightDHalf;
    float viewDHalf;
} l;

