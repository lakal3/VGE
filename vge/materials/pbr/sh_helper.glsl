

/*
Most of PBR fomulas are copied from Filaments PBR description
https://google.github.io/filament/Filament.html
and from
https://github.com/Nadrin/PBR/blob/master/data/shaders/glsl/pbr_fs.glsl
*/


// IBL color from spherical harmonics
vec3 sh(vec3 normal) {
    float x = normal.x;
    float y = normal.y;
    float z = normal.z;

    vec4 result = (
    frame.sph[0] +

    frame.sph[1] * -y +
    frame.sph[2] * z +
    frame.sph[3] * -x +

    frame.sph[4] * x * y +
    frame.sph[5] * -y * z +
    frame.sph[6] * (3.0 * z * z - 1.0) +
    frame.sph[7] * -x * z +
    frame.sph[8] * (x*x - y*y)
    );

    return max(vec3(result), vec3(0.0));
}
