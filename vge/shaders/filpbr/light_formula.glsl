


//
// PBR formulas from
// https://google.github.io/filament/Filament.html
// See actual document for description and references to original implementations

// Approximation of geometric shadowing (V). Filament 4.4.2
float pbr_V_Filament() {
    float NoV = mat.normalDView;
    float NoL = l.normalDLight;
    float perpentualRougnress = mat.roughness * mat.roughness;
    float a2 = perpentualRougnress * perpentualRougnress;
    float GGXV = NoL * sqrt(NoV * NoV * (1.0 - a2) + a2);
    float GGXL = NoV * sqrt(NoL * NoL * (1.0 - a2) + a2);
    return 0.5 / (GGXV + GGXL);
}

// Normal distribution function (specular D). Filement 4.4.1
float pbr_D_Filament() {
    float NoH = l.normalDHalf;
    float a = NoH * mat.roughness;
    float k = mat.roughness / (1.0 - NoH * NoH + a * a);
    return k * k * (1.0 / PI);
}

// Fresnel (specular F). Filament 4.4.3
vec3 pbr_F_Filament() {
    vec3 f0 = mat.specular0;
    float f90 = 1;
    return f0 + (vec3(f90) - f0) * pow(1.0 - l.viewDHalf, 5);
}

void lightFormula() {
    float D = pbr_D_Filament();
    vec3  F = pbr_F_Filament();
    float V = pbr_V_Filament();
    // Specular BRDF
    vec3 Ks = (D * V) * F;
    // Energy compensation (Filement 4.7.2)
    vec3 energyCompensation = 1.0 + mat.specular0 * (1.0 / max(0.1,dfg.y) - 1.0);
    Ks *= energyCompensation;

    // Diffuse BRDF
    vec3 Kd = mat.diffuse / PI;

    mat.directLight = mat.directLight + (Kd + Ks) * l.normalDLight * l.radiance * l.shadowFactor;
}


// mat.lightSum = vec3(specFactor);
