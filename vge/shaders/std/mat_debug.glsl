

uint matDebug(uint debug) {
    switch (debug) {
        case 1:
            o_color = mat.albedo;
            return 1;
        case 2:
            o_color = mat.emissive;
            return 1;
        case 3:
            o_color = vec4(mat.normal * vec3(0.5) + vec3(0.5), 1);
            return 1;
        case 4:
            o_color = vec4(mat.metalness, mat.roughness, 0, 1);
            return 1;
    }
    return 0;
}

uint matDebug2(uint debug) {
    switch (debug) {
    case 11:
        o_color = vec4(mat.directLight, 1);
        return 1;
    case 12:
        o_color = vec4(mat.indirectLight, 1);
        return 1;
    }
    return 0;
}