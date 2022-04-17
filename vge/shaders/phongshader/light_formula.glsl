
void lightFormula() {
    float diffFactor = max(l.normalDLight, 0.0);
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
