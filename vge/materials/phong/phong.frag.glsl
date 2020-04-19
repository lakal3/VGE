#version 450
#extension GL_GOOGLE_include_directive : require

#include "../../vscene/frame.glsl"
layout(location = 0) out vec4 o_Color;

layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_UV0;
layout(location = 2) in mat3 i_normalSpace;

layout(set = 2, binding = 0) uniform MATERIAL {
    vec4 diffuseFactor;
    vec4 specularFactor;
    vec4 emissiveFactor;
    float normalMap;
    float specularPower;
} material;

#define TX_DIFFUSE 0
#define TX_NORMAL 1
#define TX_SPECULAR 2
#define TX_EMISSIVE 3

// Diffuse, Specular, Emissive, BumpMap
layout(set = 2, binding = 1) uniform sampler2D textures[4];

vec3 calcNormal() {
    vec3 bNormal = vec3(0, 0, 1);
    if (material.normalMap > 0.5) {
        bNormal = (2 * texture( textures[TX_NORMAL], i_UV0).xyz) - vec3(1.0, 1.0, 1.0);
    }
    return normalize(i_normalSpace * bNormal);
}

float getShadowFactor(LIGHT l, vec3 position) {
    return 1.0;
}

vec3 calcLight(LIGHT l, vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 specularColor) {
    vec3 lightDir;
    float attn;
    if (l.position.w < 0.5) {
        // Directional light
        lightDir = -normalize(vec3(l.direction));
        attn = 1;
    } else {
        lightDir = normalize(vec3(l.position - frame.cameraPos));
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
    vec3 specularColor = vec3(material.specularFactor * texture( textures[TX_SPECULAR], i_UV0));
    vec3 emissiveColor = vec3(material.emissiveFactor * texture( textures[TX_EMISSIVE], i_UV0));
    float alpha =  diffuseAlphaColor.a;
    vec3 diffuseColor = vec3(diffuseAlphaColor);
    vec3 normal = calcNormal();
    vec3 viewDir = normalize(frame.cameraPos.xyz - i_position);
    float normalDView = max(dot(normal, viewDir), 0.0);

    vec3 lightColors = vec3(0);

    int noLight;
    for (noLight = 0; noLight < frame.noLights; noLight++) {
        vec3 lightOut = calcLight(frame.lights[noLight], normal, viewDir, diffuseColor, specularColor);
        lightColors += lightOut;
    }
    vec3 iblColor = ibl(normal, viewDir, diffuseColor, specularColor);
    o_Color = vec4(lightColors + iblColor + emissiveColor, alpha);
}