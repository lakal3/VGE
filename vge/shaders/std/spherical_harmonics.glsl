


/*
Most of PBR fomulas are copied from Filaments PBR description
https://google.github.io/filament/Filament.html
and from
https://github.com/Nadrin/PBR/blob/master/data/shaders/glsl/pbr_fs.glsl
*/


// IBL color from spherical harmonics
vec3 spherical_harmonics(vec3 normal) {
    float x = normal.x;
    float y = normal.y;
    float z = normal.z;

#if probe
    return frame.ambient.rgb + frame.ambienty.rgb * y;
#endif
#if !probe
    if (mat.probe == 0) {
        return frame.ambient.rgb + frame.ambienty.rgb * y;
    }
    uint probe = mat.probe;
    vec4 result = (
        extrava.va[mat.probe].v1 + // sp0

        extrava.va[mat.probe].v2 * -y +
        extrava.va[mat.probe].v3 * z +
        extrava.va[mat.probe].v4 * -x +

        extrava.va[mat.probe + 1].v1 * x * y +
        extrava.va[mat.probe + 1].v2 * -y * z +
        extrava.va[mat.probe + 1].v3 * (3.0 * z * z - 1.0) +
        extrava.va[mat.probe + 1].v4 * -x * z +
        extrava.va[mat.probe + 2].v1 * (x*x - y*y)
    );

    return max(vec3(result), vec3(0.0));
#endif
}