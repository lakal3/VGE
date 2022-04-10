
layout(push_constant) uniform INSTANCE {
    mat4 world;
    vec4 albedo;
    vec4 emissive;
    vec4 metalRoughness;
    vec4 cColor; // Custom color or other settings. Not used by standard shader
    vec4 textures1; // x - albedo, y - emissive, z - normal, w - metallRoughness
    vec4 cTextures1; // Custom texture indexes. Not used by standard shader
    float alphaCutoff;
    uint probe;
    uint frozenId;
    float dummy3;
} instance;
