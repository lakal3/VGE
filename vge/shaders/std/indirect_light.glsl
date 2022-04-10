


vec3 specularIBL() {
    if (mat.probe == 0) {
        return frame.ambient.rgb + frame.ambienty.rgb * mat.normal.y;
    }
    uint probe = mat.probe;
    vec3 r = reflect(-mat.viewDir, mat.normal);
    uint tx_env = uint(extrava.va[mat.probe + 2].v2.x);
    if (tx_env == 0) {
        return frame.ambient.rgb + frame.ambienty.rgb * mat.normal.y;
    }
    float r2 = 4.99 * mat.roughness;
    float lod1 = floor(r2);
    float lod2 = ceil(r2);
    vec4 tx1 = textureLod(frameImagesCube[tx_env], r, lod1);
    vec4 tx2 = textureLod(frameImagesCube[tx_env], r, lod2);
    return mix(tx1.rgb, tx2.rgb, fract(r2));
}

void addIndirect() {
    mat.indirectLight = spherical_harmonics(mat.normal) * mat.diffuse;
    vec3 specular90 = vec3(1);
    vec3 Lld = specularIBL();
    mat.indirectLight += (mat.specular0 * dfg.x + specular90 * dfg.y) * Lld;
}