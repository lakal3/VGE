
#include shadow_factor

#include light_formula

void addLight(uint lightPos) {
    uint lights = frame.lights;
    while (lights > 0) {
        lights--;
        l.intensity = extrava.va[lightPos].v1.rgb;
        l.position = extrava.va[lightPos].v2;// if w = 0, directional light, 1 = point light
        l.direction = extrava.va[lightPos].v3;// direction for directional light, quoternion for spot light
        l.attenuation = extrava.va[lightPos].v4;// w = shadow map index, if any

        vec4 sizes = extrava.va[lightPos + 1].v2;
        l.innerAngle = sizes.x;
        l.outerAngle = sizes.y;
        l.minDistance = sizes.z;
        l.maxDistance = sizes.w;
        l.kind = uint(extrava.va[lightPos + 1].v1.y);
        l.tx_shadowmap = uint(extrava.va[lightPos + 1].v1.z);
        l.plane = extrava.va[lightPos + 1].v3;

        l.lightToPos = l.position.xyz - i_position;
        l.lenToPos = length(l.lightToPos);
        l.lightDir = l.position.w == 0 ? -l.direction.xyz : normalize(l.lightToPos);
        l.halfDir = normalize(mat.viewDir + l.lightDir);
        float lenToPos2 = max(0.01, l.attenuation.x + l.lenToPos * l.attenuation.y +
        l.lenToPos * l.lenToPos * l.attenuation.z);
        l.radiance = l.intensity / vec3(lenToPos2);

        if (length(l.radiance) > 0.001) {
            /* Dots */
            l.normalDHalf = clamp(dot(mat.normal, l.halfDir), 0, 1);
            l.normalDLight = clamp(dot(mat.normal, l.lightDir), 0.001, 1);
            l.lightDHalf = clamp(dot(l.lightDir, l.halfDir), 0, 1);
            l.viewDHalf = clamp(dot(mat.viewDir, l.halfDir), 0, 1);
            l.shadowFactor = 1.0;
            shadowFactorFormula();

            if (l.shadowFactor > 0.001) {
                lightFormula();
            }

        }

        // Move to next light
        lightPos = uint(extrava.va[lightPos + 1].v1.x);
    }
}