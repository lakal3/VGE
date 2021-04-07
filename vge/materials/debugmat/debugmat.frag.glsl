#version 450
#extension GL_GOOGLE_include_directive : require
#extension GL_EXT_nonuniform_qualifier: require

#define DYNAMIC_DESCRIPTORS 1

#include "../../forward/frame.glsl"
layout(location = 0) out vec4 o_Color;

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 4) in mat3 i_normalSpace;
layout(location = 2) in flat vec4 i_modes;
layout(location = 3) in flat ivec2 i_decalIndex;

layout(set = 2, binding = 0) uniform MATERIAL {
    vec4 diffuseFactor;
    float normalMap;
} material;

#define TX_DIFFUSE 0
#define TX_NORMAL 1
#define TX_DEBUG 2

#include "../pbr/sh_helper.glsl"

// Diffuse, Specular, Emissive, BumpMap
layout(set = 2, binding = 1) uniform sampler2D textures[8];

#include "../decal/decal.glsl"

layout(set=3, binding=0) uniform DECALS {
    DECAL decals[256];
} decals;

vec3 lerp3(vec3 base, vec3 target, float factor) {
    return base * (1 - factor) + target * factor;
}

void calcDecal(int idx, vec3 pos, inout vec3 albedo, inout vec3 normal, vec3 refNormal) {
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
    albedo = lerp3(albedo, vec3(aBase), factor);
    normal = lerp3(normal, aNormal, factor);
}

vec3 calcNormal() {
    vec3 bNormal = vec3(0, 0, 1);
    if (material.normalMap > 0.5) {
        bNormal = (2 * texture( textures[TX_NORMAL], i_UV0).xyz) - vec3(1.0, 1.0, 1.0);
    }
    return normalize(i_normalSpace * bNormal);
}

#include "../../vscene/shadowfactor.glsl"


vec3 calcLight(LIGHT l, vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 specularColor) {
    vec3 lightDir;
    float attn;
    if (l.position.w < 0.5) {
        // Directional light
        lightDir = -normalize(vec3(l.direction));
        attn = 1;
    } else {
        lightDir = normalize(vec3(l.position) - i_position);
        float dist = length(vec3(l.position) - i_position);
        if (dist > l.attenuation.w) { // Too far away
            return vec3(0);
        }
        float dist2 = max(0.01, l.attenuation.x + dist * l.attenuation.y +
        dist * dist * l.attenuation.z);
        attn = 1 / dist2;
    }
    float shadowFactor = getShadowFactor(l, vec3(i_position));
    if (shadowFactor < 0.05) {
        return vec3(0);
    }
    float diffFactor = max(dot(normal, lightDir), 0.0) * attn;
    vec3 finalLight = vec3(l.intensity) * diffuseColor * diffFactor;

    vec3 halfDir = normalize(lightDir + viewDir);
    float specFactor = pow(max(dot(normal,halfDir), 0.0), 25) * attn;
    finalLight = finalLight + vec3(l.intensity) * specularColor * specFactor;
    return finalLight * shadowFactor;
}

vec3 ibl(vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 specularColor) {
    return diffuseColor * vec3(frame.sph[0]);
}

void main() {
    vec4 diffuseAlphaColor = material.diffuseFactor * texture( textures[TX_DIFFUSE], i_UV0);
    int m = int(i_modes.w + 0.1);

    if (m == 1) {
        o_Color = material.diffuseFactor * texture( textures[int(i_modes.x)], i_UV0);
        return;
    }

    float alpha =  diffuseAlphaColor.a;
    vec3 diffuseColor = vec3(diffuseAlphaColor);
    vec3 normal = calcNormal();
    vec3 refNormal = normalize(i_normalSpace * vec3(0,0,1));
    // Calc decals
    for (int idx = i_decalIndex.x; idx < i_decalIndex.y; idx++) {
        calcDecal(idx, i_position, diffuseColor, normal, refNormal);
    }
    if (m == 2) {
        o_Color = vec4((normal * vec3(0.5) + vec3(0.5)) * i_modes.xyz,1);
        return;
    }
    if (m == 3) {
        o_Color = vec4(i_modes.rgb,1);
        return;
    }
    if (m == 4) {
        vec3 tangent = i_normalSpace * vec3(1,0,0);
        o_Color = vec4((tangent * vec3(0.5) + vec3(0.5)) * i_modes.xyz,1);
        return;
    }
    int em = int(frame.envMap);
    if (m == 5 && em > 0) {
        vec3 iDir = i_position - vec3(frame.cameraPos);
        o_Color = texture(frameImagesCube[em],reflect(iDir, normal)) * vec4(i_modes.xyz,1);
        // o_Color = texture(frameImagesCube[em],normal) * vec4(i_modes.xyz,1);
        return;
    }
    if (m == 6) {
        vec3 iDir = i_position - vec3(frame.cameraPos);
        o_Color = vec4(sh(normalize(reflect(iDir, normal))) * i_modes.xyz,1);
        return;
    }
    vec3 viewDir = normalize(frame.cameraPos.xyz - i_position);
    float normalDView = max(dot(normal, viewDir), 0.0);

    vec3 lightColors = vec3(0);
    vec3 specularColor = vec3(0.5, 0.5, 0.5);
    diffuseColor = diffuseColor * i_modes.rgb;
    int noLight;
    for (noLight = 0; noLight < frame.noLights; noLight++) {

        vec3 lightOut = calcLight(frame.lights[noLight], normal, viewDir, diffuseColor, specularColor );
        lightColors += lightOut;
    }
    vec3 iblColor = ibl(normal, viewDir, diffuseColor, specularColor);
    o_Color = vec4(lightColors + iblColor, alpha);
}
