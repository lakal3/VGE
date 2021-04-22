// Standard frame layout

#define MAX_LIGHTS 64
// 64 is too many for Intel cards, need to reserve some images to materials
#define MAX_IMAGES 48

struct LIGHT {
    vec4 intensity;
    vec4 position;    // if w = 0, directional light, 1 = point light
    vec4 direction;   // direction for directional light, quoternion for spot light
    vec4 attenuation; // w = shadow map index, if any
    vec4 shadowPlane; // Rotation to shadow plane (quoternion)
    float innerAngle;
    float outerAngle;
    float shadowMapMethod;
    float shadowMapIndex;
};

layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view;
    vec4 cameraPos;
//Spherical harmonics
    vec4 sph[9];
    float noLights;
    float envMap;
    float envLoDs; // Level of details in envmap
    float far;
    LIGHT[MAX_LIGHTS] lights;
} frame;

#ifdef DYNAMIC_DESCRIPTORS
layout(set=0, binding=1) uniform sampler2D frameImages2D[];
layout(set=0, binding=1) uniform samplerCube frameImagesCube[];
#else
layout(set=0, binding=1) uniform sampler2D frameImages2D[MAX_IMAGES];
layout(set=0, binding=1) uniform samplerCube frameImagesCube[MAX_IMAGES];
#endif