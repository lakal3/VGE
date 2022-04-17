//
// Formulas from glTF 2.0 reference implementation
// https://github.com/cx20/gltf-test/blob/master/examples/khronos-gltf-loader/shaders/pbr-frag.glsl
//


// Basic Lambertian diffuse
// Implementation from Lambert's Photometria https://archive.org/details/lambertsphotome00lambgoog
// See also [1], Equation 1
vec3 diffuse()
{
    return mat.diffuse / PI;
}

// The following equation models the Fresnel reflectance term of the spec equation (aka F())
// Implementation of fresnel from [4], Equation 15
vec3 specularReflection()
{
    return mat.specular + (vec3(1) - mat.specular) * pow(1.0 - l.viewDHalf, 5);
}

// This calculates the specular geometric attenuation (aka G()),
// where rougher material will reflect less light back to the viewer.
// This implementation is based on [1] Equation 4, and we adopt their modifications to
// alphaRoughness as input as originally proposed in [2].
float geometricOcclusion()
{
    float NdotL = l.normalDLight;
    float NdotV = mat.normalDView;
    float r = mat.roughness * mat.roughness;

    float attenuationL = 2.0 * NdotL / (NdotL + sqrt(r * r + (1.0 - r * r) * (NdotL * NdotL)));
    float attenuationV = 2.0 * NdotV / (NdotV + sqrt(r * r + (1.0 - r * r) * (NdotV * NdotV)));
    return attenuationL * attenuationV;
}

// The following equation(s) model the distribution of microfacet normals across the area being drawn (aka D())
// Implementation from "Average Irregularity Representation of a Roughened Surface for Ray Reflection" by T. S. Trowbridge, and K. P. Reitz
// Follows the distribution function recommended in the SIGGRAPH 2013 course notes from EPIC Games [1], Equation 3.
float microfacetDistribution()
{
    float ar = mat.roughness * mat.roughness;
    float roughnessSq = ar * ar;
    float NdotH = l.normalDHalf;
    float f = (NdotH * roughnessSq - NdotH) * NdotH + 1.0;
    return roughnessSq / (PI * f * f);
}


void lightFormula() {
    vec3 F = specularReflection();
    float G = geometricOcclusion();
    float D = microfacetDistribution();

    // Calculation of analytical lighting contribution
    vec3 diffuseContrib = (1.0 - F) * diffuse();
    vec3 specContrib = F * G * D / (4.0 * l.normalDLight * mat.normalDView);
    // Obtain final intensity as reflectance (BRDF) scaled by the energy of the light (cosine law)
    vec3 color = l.normalDLight * l.radiance * (diffuseContrib + specContrib);

    mat.directLight = mat.directLight + color * l.shadowFactor;
}


// mat.lightSum = vec3(specFactor);
