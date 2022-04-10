

// Phong
#if phong
void lightFormula() {
    float diffFactor = max(dot(mat.normal, l.lightDir), 0.0);
    vec3 diffLight = l.radiance * mat.diffuse * vec3(diffFactor);

    #if !probe
    float specFactor = pow(max(dot(mat.normal, l.halfDir), 0.0), 25);
    vec3 specLight = l.radiance * mat.specular0 * vec3(specFactor);
    #endif
    #if probe
    vec3 specLight = l.radiance * mat.specular0 * vec3(diffFactor);
    #endif

    mat.directLight = mat.directLight + (diffLight + specLight) * l.shadowFactor;
}
#endif
#if !phong


// GGX/Towbridge-Reitz normal distribution function.
// Uses Disney's reparametrization of alpha = roughness^2.
float pbr_ndfGGX(float normalDHalf, float roughness)
{
    float alpha   = roughness * roughness;
    float alphaSq = alpha * alpha;

    float denom = (normalDHalf * normalDHalf) * (alphaSq - 1.0) + 1.0;
    return alphaSq / (PI * denom * denom);
}

// Single term for separable Schlick-GGX below.
float pbr_gaSchlickG1(float cosTheta, float k)
{
    return cosTheta / (cosTheta * (1.0 - k) + k);
}

// Schlick-GGX approximation of geometric attenuation function using Smith's method.
float pbr_gaSchlickGGX()
{
    float r = mat.roughness + 1.0;
    float k = (r * r) / 8.0; // Epic suggests using this roughness remapping for analytic lights.
    return pbr_gaSchlickG1(mat.normalDView, k) * pbr_gaSchlickG1(l.normalDLight, k);
}

vec3 pbr_F_Schlick(float viewDHalf, vec3 f0) {
    return f0 + (vec3(1.0) - f0) * pow(1.0 - viewDHalf, 5.0);
}

void lightFormula() {
    float D = pbr_ndfGGX(l.normalDHalf, mat.roughness);
    vec3  F = pbr_F_Schlick(l.lightDHalf, mat.specular0);
    float V = pbr_gaSchlickGGX();
    float denominator = 4.0 * mat.normalDView * l.normalDLight + 0.01;
    vec3 Ks = ((D * V) * F) / denominator;

    vec3 energyCompensation = 1.0 + mat.specular0 * (1.0 / max(0.1,dfg.y) - 1.0);

    // Scale the specular lobe to account for multiscattering
    Ks *= energyCompensation;
    vec3 Kd = (vec3(1.0) - F) * (1.0 - mat.metalness);

    mat.directLight = mat.directLight + (Kd * mat.diffuse / PI + Ks) * l.radiance * l.normalDLight * l.shadowFactor;
}
#endif

// mat.lightSum = vec3(specFactor);
