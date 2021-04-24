#version 450
#extension GL_EXT_nonuniform_qualifier: require
#extension GL_GOOGLE_include_directive: require

#define TX_WHITE 0
#define TX_ALBEDO 1
#define TX_NORMAL 2
#define TX_MATERIAL 3
#define TX_DEPTH 4

#define MAX_PROBES 16
#define MAX_LIGHTS 64

#define DYNAMIC_DESCRIPTORS 1

const float PI = 3.14159265;

layout(location = 0) out vec4 o_Color;

layout(location = 0) in vec2 i_position;
layout(location = 1) in flat int i_index;


struct LIGHT {
    vec4 intensity;
    vec4 position;    // if w = 0, directional light
    vec4 direction;   // if w > 0, shadowmap index = w - 1
    vec4 attenuation; // w = shadow map index, if any
    vec4 shadowPlane; // Rotation to shadow plane (quoternion)
    float innerAngle;
    float outerAngle;
    float shadowMapMethod;
    float shadowMapIndex;

};

struct PROBE {
    vec4 sph[9];
    float envImage;
    float filler1;
    float filler2;
    float filler3;
};

layout(set=0, binding=0) uniform FRAME {
    float noProbes;
    float noLights;
    float debugMode;
    float debugIndex;
    mat4 invProjection;
    mat4 invView;
    mat4 view;
    vec4 eyePos;
    PROBE probes[MAX_PROBES];
    LIGHT lights[MAX_LIGHTS];
} frame;

layout(set=0, binding=1) uniform sampler2D frameImages2D[];

layout(set=0, binding=1) uniform samplerCube frameImagesCube[];

layout(set=0, binding=1) uniform sampler2DArray frameImagesArray[];

layout(set=0, binding=1) uniform usampler2D frameImagesU2D[];

/**************** Helpers *****************************************************/
#include "../vscene/shadowfactor.glsl"

// IBL color from spherical harmonics
vec3 sh(vec3 normal, PROBE p) {
    float x = normal.x;
    float y = normal.y;
    float z = normal.z;
    vec4 result = (
        p.sph[0] +

        p.sph[1] * -y +
        p.sph[2] * z +
        p.sph[3] * -x +

        p.sph[4] * x * y +
        p.sph[5] * -y * z +
        p.sph[6] * (3.0 * z * z - 1.0) +
        p.sph[7] * -x * z +
        p.sph[8] * (x*x - y*y)
    );

    return max(vec3(result), vec3(0.0));
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


vec3 specularIBL(vec3 reflectDir, float roughness, uint tx_envMap) {
    if (tx_envMap == 0) {
        return vec3(0.5);
    }
    int lods = textureQueryLevels(frameImagesCube[tx_envMap]);
    float lod1 = floor(lods * roughness);
    float lod2 = ceil(lods * roughness);
    vec4 tx1 = textureLod(frameImagesCube[tx_envMap], reflectDir, lod1);
    vec4 tx2 = textureLod(frameImagesCube[tx_envMap], reflectDir, lod2);
    return mix(tx1.rgb, tx2.rgb, fract(roughness));
}

vec3 ibl(vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 f0, float roughness, vec2 dfg, uint probe) {
    if (probe == 0) {
        return diffuseColor;
    }
    PROBE p = frame.probes[probe - 1];
    vec3 r = reflect(-viewDir, normal);
    vec3 f90 = vec3(1);
    vec3 Ld = sh(normal, p) * diffuseColor;
    vec3 Lld = specularIBL(r, roughness, int(p.envImage));
    vec3 Lr =  (f0 * dfg.x + f90 * dfg.y) * Lld;
    return (Ld + Lr);
}

vec3 getScreenPosition(vec2 position, ivec2 texelPosition) {
    float z = texelFetch(frameImages2D[TX_DEPTH], texelPosition, 0).x;
    vec2 pointPos = position * vec2(2.0, 2.0) + vec2(-1.0, -1.0);
    return vec3(pointPos, z);
}

vec3 getViewPosition(vec3 screenPos) {
    vec4 viewPos = frame.invProjection * vec4(screenPos, 1);
    viewPos = viewPos / vec4(viewPos.w);
    return vec3(viewPos);
}

vec3 getWorldPosition(vec3 viewPos) {
    return vec3(frame.invView * vec4(viewPos, 1)).xyz;
}

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

vec3 calcLight(LIGHT light, vec3 worldPos, vec3 normal, vec3 viewDir, vec3 diffuseColor, vec3 f0, float metalness,
  float roughness, vec2 dfg) {
    vec3 lightToPos = light.position.xyz - worldPos;
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
    vec3 Kd = (vec3(1.0) - F) * (1.0 - metalness);

    return (Kd * diffuseColor / PI + Ks) * radiance * normalDLight;
}

// Debug helpers
vec3 debugSH(vec3 normal, uint probe) {
    if (probe == 0) {
        return vec3(0.2);
    }
    PROBE p = frame.probes[probe - 1];
    return sh(normal, p);
}

vec3 debugSpecularIBL(float roughness, vec3 normal, vec3 viewDir, uint probe) {
    if (probe == 0) {
        return vec3(0.2);
    }
    PROBE p = frame.probes[probe - 1];
    vec3 reflectDir = reflect(-viewDir, normal);
    return specularIBL(reflectDir, roughness, uint(p.envImage));
}

/**************** main *****************************************************/


void main() {
    ivec2 texelPosition = ivec2(gl_FragCoord.xy);
    uvec4 i_material = texelFetch(frameImagesU2D[TX_MATERIAL], texelPosition, 0);
    vec4 rawNormal = texelFetch(frameImages2D[TX_NORMAL], texelPosition, 0);
    vec4 albedo = texelFetch(frameImages2D[TX_ALBEDO], texelPosition, 0);
    float metalness = float(i_material.r) / 255;
    float roughness = float(i_material.g) / 255;
    uint probe = i_material.b;
    if (frame.debugMode == 1) {
        o_Color = albedo;
        return;
    }
    if (frame.debugMode == 2) {
        o_Color = rawNormal;
        return;
    }
    if (frame.debugMode == 3) {
        o_Color = i_material / vec4(255, 255, 255, 255);
        o_Color.a = 1;
        return;
    }
    if (frame.debugMode == 4) {
        float z = texture(frameImages2D[TX_DEPTH], i_position).x;
        o_Color = vec4((1 - z) * 5.0, i_position, 1);
        return;
    }
    o_Color = albedo;
    if (rawNormal.a < 0.5) {
        // Emissive color
        return;
    }
    vec3 screenPos =  getScreenPosition(i_position, texelPosition);
    vec3 viewPos = getViewPosition(screenPos);
    vec3 worldPosition = getWorldPosition(viewPos);
    vec3 normal = vec3(rawNormal) * vec3(2.0, 2.0, 2.0) + vec3(-1.0, -1.0, -1.0);
    vec3 eyePos = vec3(frame.eyePos);
    vec3 viewDir = normalize(eyePos - worldPosition);
    float normalDView = max(dot(normal, viewDir), 0.0);

    if (frame.debugMode == 5) {
        o_Color = vec4((worldPosition + vec3(5,0,5)) * vec3(0.1, 0.3, 0.1),1);
        return;
    }
    if (frame.debugMode == 6) {
        o_Color = vec4(viewDir * vec3(0.5, 0.5, 0.5) + vec3(0.5, 0.5, 0.5),1);
        return;
    }

    if (frame.debugMode == 7) {
        o_Color = vec4(debugSH(normal, probe), 1);
        return;
    }
    if (frame.debugMode == 8) {
        o_Color = vec4(debugSpecularIBL(roughness, normal, viewDir, probe), 1);
        return;
    }
    float reflectance = 0.5;
    vec3 diffuseColor = (1.0 - metalness) * albedo.rgb;
    vec3 f0 = 0.16 * reflectance * reflectance * (1.0 - metalness) + albedo.rgb * metalness;
    vec2 dfg = prefilteredDFG(roughness, normalDView);
    vec3 iblColor = ibl(normal, viewDir, diffuseColor, f0, roughness, dfg, probe);

    // Calculate lights
    vec3 lightColors = vec3(0);

    uint noLight;
    if (frame.debugMode == 10) {
        for (noLight = 0; noLight < frame.noLights; noLight++) {
            LIGHT l = frame.lights[noLight];
            float shadowFactor = getShadowFactor(l, worldPosition);
            lightColors = lightColors + vec3(shadowFactor /frame.noLights);
        }
        o_Color = vec4(lightColors, 1);
        return;
    }
    for (noLight = 0; noLight < frame.noLights; noLight++) {
        LIGHT l = frame.lights[noLight];
        float shadowFactor = getShadowFactor(l, worldPosition);
        vec3 lightOut = shadowFactor > 0.1 ? calcLight(l, worldPosition, normal, viewDir, diffuseColor * shadowFactor,
        f0  * shadowFactor, metalness, roughness, dfg) : vec3(0);
        lightColors += lightOut;
    }
    if (frame.debugMode == 9) {
        o_Color = vec4(lightColors, 1);
        return;
    }
    o_Color = vec4(iblColor + lightColors, 1);
}
