
#version 450
#extension GL_EXT_nonuniform_qualifier: require

#include frame
#include mesh_instance

const float PI = 3.14159265;

// Output
layout(location = 0) out vec4 o_color;
#if !transparent & !probe
layout(location = 1) out vec4 o_color2;
#endif

// Input
layout(location = 0) in vec3 i_position;
layout(location = 1) in vec2 i_uv0;
layout(location = 2) in vec3 i_normal;
// Vertex color is not used by standard fragment shader
layout(location = 3) in vec3 i_color;
#if normalMap
layout(location = 4) in mat3 i_normalSpace;
#endif

#include material_def

#include light_def

#include spherical_harmonics

#include pbr_helper

void calcNormal() {
    mat.normal = i_normal;
#if normalMap
    if (instance.textures1.w > 0) {
        vec3 nDir = texture(frameImages2D[int(instance.textures1.w)], i_uv0).xyz;
        vec3 bNormal = (2 * nDir) - vec3(1.0, 1.0, 1.0);
        mat.normal = normalize(i_normalSpace * bNormal);
    }
#endif
}


#include light_base

#include indirect_light

#include decal_base

#if probe
#include probe_frame
#endif

#if withdebug
#include mat_debug
#endif

#cinclude c_fragment_def

void main() {
#if transparent
    if (texelFetch(frameImages2D[1], ivec2(gl_FragCoord.xy),0).r < gl_FragCoord.z) {
        discard;
    }
#endif
    mat.albedo = instance.albedo;
    mat.emissive = instance.emissive;
    mat.metalness = instance.metalRoughness.x;
    mat.roughness = instance.metalRoughness.y;
    mat.frozenId = float(instance.frozenId);
    calcNormal();
    mat.f0  = vec3(0.04);
    if (instance.textures1.x > 0) {
        mat.albedo = texture(frameImages2D[int(instance.textures1.x)], i_uv0) * mat.albedo;
    }
    if (instance.textures1.y > 0) {
        mat.emissive = texture(frameImages2D[int(instance.textures1.y)], i_uv0) * mat.emissive;
    }
    if (instance.textures1.z > 0) {
        vec4 mr = texture(frameImages2D[int(instance.textures1.z)], i_uv0);
        mat.metalness = mr.r;
        mat.roughness = mr.g;
    }
#cinclude c_material_calc
    addDecals();
    // Apply decals etc to change mat
    mat.diffuse = mat.albedo.rgb * (vec3(1) - mat.f0) * (1 - mat.metalness);
    mat.specular = mix(mat.f0, mat.albedo.rgb, mat.metalness);
    mat.reflectance = max(mat.specular.r, max(mat.specular.g, mat.specular.b));
    mat.directLight = vec3(0);
    mat.indirectLight = vec3(0);
#if probe
    mat.probe = 0;
    mat.viewDir = normalize(probeFrame.cameraPos.xyz - i_position);
#endif
#if !probe
    mat.probe = instance.probe;
    mat.viewDir = normalize(frame.cameraPos.xyz - i_position);
#endif
    mat.normalDView = abs(dot(mat.normal, mat.viewDir)) + 1e-5;
    dfg = pbr_prefilteredDFG();


    #if withdebug
    if (matDebug(frame.debug) != 0) {
        #if !transparent & !probe
        o_color2 = o_color;
        #endif
        return;
    }
#endif
    if (frame.lightPos != 0) {
        addLight(frame.lightPos);
    }

    addIndirect();
    if (matDebug2(frame.debug) != 0) {
        #if !transparent & !probe
        o_color2 = o_color;
        #endif
        return;
    }
    o_color = vec4(mat.emissive.rgb + mat.directLight + mat.indirectLight, mat.albedo.a);

#if !transparent & !probe
    o_color2 = o_color;
#endif
}
