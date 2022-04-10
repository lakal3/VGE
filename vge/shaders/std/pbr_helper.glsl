
// NOTE: this approximation is not valid if the energy compensation term
// for multiscattering is applied. We use the DFG LUT solution to implement
// multiscattering
vec2 pbr_prefilteredDFG() {
    // Karis' approximation based on Lazarov's
    const vec4 c0 = vec4(-1.0, -0.0275, -0.572,  0.022);
    const vec4 c1 = vec4( 1.0,  0.0425,  1.040, -0.040);
    vec4 r = mat.roughness * c0 + c1;
    float a004 = min(r.x * r.x, exp2(-9.28 * mat.normalDView)) * r.x + r.y;
    return vec2(-1.04, 1.04) * a004 + r.zw;
    // Zioma's approximation based on Karis
    // return vec2(1.0, pow(1.0 - max(roughness, NoV), 3.0));
}

vec2 dfg;