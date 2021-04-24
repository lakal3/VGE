#version 450
#extension GL_GOOGLE_include_directive : require
#extension GL_EXT_nonuniform_qualifier: require

layout(location = 0) out vec4 o_Color;
layout(location = 1) out vec4 o_Normal;
layout(location = 2) out uvec4 o_Material;

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 2) in mat3 i_normalSpace;

#include "../../deferred/frame.glsl"

layout(set = 2, binding = 0) uniform MATERIAL {
    vec4 albedoColor;
    vec4 emissiveColor;
    float metallicFactor;
    float roughnessFactor;
    float normalMap;
} material;

#define TX_ALBEDO 0
#define TX_NORMAL 1
#define TX_METAL_ROUGHNESS 2
#define TX_EMISSIVE 3


layout(set=2, binding=1) uniform sampler2D textures[4];

#include "../decal/decal.glsl"

const float PI = 3.14159265;
vec3 calcNormal() {
    vec3 bNormal = vec3(0, 0, 1);
    if (material.normalMap > 0.5) {
        bNormal = (2 * texture( textures[TX_NORMAL], i_UV0).xyz) - vec3(1.0, 1.0, 1.0);
    }
    return normalize(i_normalSpace * bNormal) * vec3(0.5, 0.5, 0.5) + vec3(0.5, 0.5, 0.5);
}



void main() {
    vec4 albedoColor = material.albedoColor * texture( textures[TX_ALBEDO], i_UV0);

    vec4 emissiveColor = vec4(material.emissiveColor * texture( textures[TX_EMISSIVE], i_UV0));
    vec3 normal = calcNormal();
    vec3 refNormal = i_normalSpace * vec3(0,0,1.0);
    vec3 albedo = albedoColor.rgb;
    vec3 mrColor = texture( textures[TX_METAL_ROUGHNESS], i_UV0).rgb;
    float metallic = mrColor.b * material.metallicFactor;
    float roughness = mrColor.g * material.roughnessFactor;

    // Decals
    int noDecals = int(decals.noDecals);
    for (int idx = 0; idx < noDecals; idx++) {
        calcDecal(idx, i_position, albedo, normal, metallic, roughness, refNormal);
    }

    if (emissiveColor.a > 0.1) {
        // We can only handle emissive or normal color
        o_Color = emissiveColor;
        o_Normal = vec4(0);
        o_Material = uvec4(0,0,0,0);
    } else {
        o_Color = vec4(albedo, 1);
        o_Normal = vec4(normal, 1.0);
        o_Material = uvec4(metallic * 255, roughness * 255, 1, 0);
    }
}
