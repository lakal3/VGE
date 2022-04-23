

mat3 caclNormatRotation(vec3 axis, float c)
{
    float s = sqrt(1 - c * c);
    float oc = 1.0 - c;

    return mat3(oc * axis.x * axis.x + c,           oc * axis.x * axis.y - axis.z * s,  oc * axis.z * axis.x + axis.y * s,
        oc * axis.x * axis.y + axis.z * s,  oc * axis.y * axis.y + c,           oc * axis.y * axis.z - axis.x * s,
        oc * axis.z * axis.x - axis.y * s,  oc * axis.y * axis.z + axis.x * s,  oc * axis.z * axis.z + c);
}

void addDecal(uint pos) {
    vec2 limits = extrava.va[pos + 2].v1.yz;
    if (limits.y > limits.x && (mat.meshID < limits.x || mat.meshID >= limits.y)) {
        return;
    }
    mat4 toDecalSpace = extrama.ma[pos];
    vec4 dp = toDecalSpace * vec4(i_position, 1);
    if (dp.x < -1 || dp.x > 1 || dp.y < -1 || dp.y > 1 || dp.z < -1 || dp.z > 1) {
        return;
    }
    vec2 samplePoint = dp.xz * vec2(0.5) + vec2(0.5);
    vec3 aNormal = vec3(0, 1, 0);
    vec3 dNormal = normalize(vec3(toDecalSpace * vec4(mat.normal, 0)));
    float normalAttenuation = extrava.va[pos + 1].v4.z;
    float attn = dot(aNormal, dNormal) * normalAttenuation  + (1 - normalAttenuation);
    vec4 eBase = extrava.va[pos + 1].v2;
    uint tx_emissive = uint(extrava.va[pos + 1].v3.y);
    if (tx_emissive > 0) {
        eBase = eBase * texture(frameImages2D[tx_emissive], samplePoint);
    }
    float eFactor = attn * eBase.a;
    if (eFactor > 0.001) {
        mat.emissive += vec4(mix(vec3(0), vec3(eBase), eFactor),0);
        mat.emissive.w = 1.0;
    }
    vec4 aBase = extrava.va[pos + 1].v1;
    uint tx_albedo = uint(extrava.va[pos + 1].v3.x);
    if (tx_albedo > 0) {
        aBase = aBase * texture(frameImages2D[tx_albedo], samplePoint);
    }

    float factor = attn * aBase.a;
    if (factor > 0.01) {
        mat.albedo = vec4(mix(mat.albedo.rgb, vec3(aBase), factor), mat.albedo.a);
    }

    float aMetallic = extrava.va[pos + 1].v4.x;
    float aRoughness = extrava.va[pos + 1].v4.y;
    uint tx_metalRoughness = uint(extrava.va[pos + 1].v3.z);
    if (tx_metalRoughness > 0) {
        vec3 mrColor = texture(frameImages2D[int(tx_metalRoughness)], samplePoint).rgb;
        aMetallic = mrColor.b * aMetallic;
        aRoughness = mrColor.g * aRoughness;
    }
    mat.metalness = mix(mat.metalness,aMetallic,factor);
    mat.roughness = mix(mat.roughness,aRoughness,factor);

    // adjust normal if decal has bump map
    uint tx_normal = uint(extrava.va[pos + 1].v3.w);
    if (tx_normal > 0 && factor > 0.001) {
        vec3 normDif = texture(frameImages2D[int(tx_normal)], samplePoint).rgb;
        normDif = normDif * vec3(2) - vec3(1);
        vec3 z = vec3(0,0,1);
        float c = dot(normDif,z);
        if (c > 0.999) {
            return;
        }
        vec3 axis = normalize(cross(normDif, z));
        mat3 rot = caclNormatRotation(axis, c);
        mat.normal = rot * mat.normal;
    }

}

void addDecals() {
    uint decalPos = frame.decalPos;
    uint decals = frame.decals;
    while (decals-- > 0) {
        addDecal(decalPos);
        decalPos =  uint(extrava.va[decalPos + 2].v1.x);
    }
}