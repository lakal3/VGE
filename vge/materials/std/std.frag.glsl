
/*
Most of PBR fomulas adapted from Filaments PBR description
https://google.github.io/filament/Filament.html
and from
https://github.com/Nadrin/PBR/blob/master/data/shaders/glsl/pbr_fs.glsl
*/
#version 450
#extension GL_GOOGLE_include_directive : require
#extension GL_EXT_nonuniform_qualifier: require

#define DYNAMIC_DESCRIPTORS 1

#include "../../forward/frame.glsl"

layout(location = 0) out vec4 o_Color;

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 2) in mat3 i_normalSpace;

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
    return normalize(i_normalSpace * bNormal);
}

#include "../../vscene/shadowfactor.glsl"

#include "../pbr/sh_helper.glsl"



// GGX/Towbridge-Reitz normal distribution function.
// Uses Disney's reparametrization of alpha = roughness^2.
float ndfGGX(float normalDHalf, float roughness)
{
    float alpha   = roughness * roughness;
    float alphaSq = alpha * alpha;

    float denom = (normalDHalf * normalDHalf) * (alphaSq - 1.0) + 1.0;
    return alphaSq / (PI * denom * denom);
}

// Single term for separable Schlick-GGX below.
float gaSchlickG1(float cosTheta, float k)
{
    return cosTheta / (cosTheta * (1.0 - k) + k);
}

// Schlick-GGX approximation of geometric attenuation function using Smith's method.
float gaSchlickGGX(float normalDView, float normalDLight, float roughness)
{
    float r = roughness + 1.0;
    float k = (r * r) / 8.0; // Epic suggests using this roughness remapping for analytic lights.
    return gaSchlickG1(normalDView, k) * gaSchlickG1(normalDLight, k);
}

vec3 F_Schlick(float viewDHalf, vec3 f0) {
    return f0 + (vec3(1.0) - f0) * pow(1.0 - viewDHalf, 5.0);
}


// NOTE: this approximation is not valid if the energy compensation term
// for multiscattering is applied. We use the DFG LUT solution to implement
// multiscattering
vec2 prefilteredDFG(float normalDView, float roughness) {
    // Karis' approximation based on Lazarov's
    const vec4 c0 = vec4(-1.0, -0.0275, -0.572,  0.022);
    const vec4 c1 = vec4( 1.0,  0.0425,  1.040, -0.040);
    vec4 r = roughness * c0 + c1;
    float a004 = min(r.x * r.x, exp2(-9.28 * normalDView)) * r.x + r.y;
    return vec2(-1.04, 1.04) * a004 + r.zw;
    // Zioma's approximation based on Karis
    // return vec2(1.0, pow(1.0 - max(roughness, NoV), 3.0));
}

vec3 calcLight(LIGHT light, vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 f0, float metallic,
  float roughness, vec2 dfg) {
    vec3 lightToPos = light.position.xyz - i_position;
    float lenToPos = length(lightToPos);
    vec3 lightDir = length(light.direction.xyz) > 0.5 ? -light.direction.xyz : normalize(lightToPos);
    vec3 halfDir = normalize(viewDir + lightDir);
    vec3 radiance = light.intensity.rgb;
    float lenToPos2 = max(0.01, light.attenuation.x + lenToPos * light.attenuation.y +
      lenToPos * lenToPos * light.attenuation.z);
    radiance = radiance / lenToPos2;

    if (length(radiance) < 0.001) {
        return vec3(0);
    }
    /* Dots */
    float normalDView = abs(dot(normal, viewDir)) + 1e-5;
    float normalDHalf = clamp(dot(normal, halfDir), 0, 1);
    float normalDLight = clamp(dot(normal, lightDir), 0, 1);
    float lightDHalf = clamp(dot(lightDir, halfDir), 0, 1);


    float D = ndfGGX(normalDHalf, roughness);
    vec3  F = F_Schlick(lightDHalf, f0);
    float V = gaSchlickGGX(normalDView, normalDLight, roughness);
    float denominator = 4.0 * normalDView * normalDLight + 0.01;

    vec3 Ks = ((D * V) * F) / denominator;

    vec3 energyCompensation = 1.0 + f0 * (1.0 / max(0.1,dfg.y) - 1.0);

    // Scale the specular lobe to account for multiscattering
    Ks *= energyCompensation;
    vec3 Kd = (vec3(1.0) - F) * (1.0 - metallic);

    return (Kd * diffuseColor / PI + Ks) * radiance * normalDLight;
}

vec3 specularIBL(vec3 r, float roughness) {
    float lod1 = floor(frame.envLoDs * roughness);
    float lod2 = ceil(frame.envLoDs * roughness);
    int tx_preEnv = int(frame.envMap);
    if (tx_preEnv == 0) {
        return vec3(0.5);
    }
    vec4 tx1 = textureLod(frameImagesCube[tx_preEnv], r, lod1);
    vec4 tx2 = textureLod(frameImagesCube[tx_preEnv], r, lod2);
    return mix(tx1.rgb, tx2.rgb, fract(roughness));
}

vec3 ibl(vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 f0, float roughness, vec2 dfg) {
    vec3 r = reflect(-viewDir, normal);
    vec3 f90 = vec3(1);
    vec3 Ld = sh(normal) * diffuseColor;
    vec3 Lld = specularIBL(r, roughness);
    vec3 Lr =  (f0 * dfg.x + f90 * dfg.y) * Lld;
    return (Ld + Lr);
}


void main() {
    vec4 diffuseAlphaColor = material.albedoColor * texture( textures[TX_ALBEDO], i_UV0);
    vec3 emissiveColor = vec3(material.emissiveColor * texture( textures[TX_EMISSIVE], i_UV0));
    float alpha =  diffuseAlphaColor.a;
    if (alpha < 0.1) {
        discard;
    }

    vec3 normal = calcNormal();
    vec3 refNormal = i_normalSpace * vec3(0,0,1.0);
    vec3 albedo = diffuseAlphaColor.rgb;
    vec3 mrColor = texture( textures[TX_METAL_ROUGHNESS], i_UV0).rgb;
    float metallic = mrColor.b * material.metallicFactor;
    float roughness = mrColor.g * material.roughnessFactor;

    // Decals
    int noDecals = int(decals.noDecals);
    for (int idx = 0; idx < noDecals; idx++) {
        calcDecal(idx, i_position, albedo, normal, metallic, roughness, refNormal);
    }

    // PBR
    vec3 viewDir = normalize(frame.cameraPos.xyz - i_position);
    float normalDView = max(dot(normal, viewDir), 0.0);
    float reflectance = 0.5;
    vec3 diffuseColor = (1.0 - metallic) * albedo;
    vec3 f0 = 0.16 * reflectance * reflectance * (1.0 - metallic) + albedo * metallic;
    vec2 dfg = prefilteredDFG(roughness, normalDView);


    vec3 lightColors = vec3(0);

    int noLight;
    for (noLight = 0; noLight < frame.noLights; noLight++) {
        LIGHT l = frame.lights[noLight];
        float shadowFactor = getShadowFactor(l, i_position);
        // o_Color = vec4(shadowFactor, 0, 0, 1);
        vec3 lightOut = shadowFactor > 0.1 ? calcLight(l, normal, viewDir, diffuseColor * shadowFactor,
        f0  * shadowFactor, metallic, roughness, dfg) : vec3(0);
        lightColors += lightOut;
    }
    vec3 iblColor = ibl(normal, viewDir, diffuseColor, f0, roughness, dfg);
    o_Color = vec4(lightColors + iblColor + emissiveColor, alpha);
}

