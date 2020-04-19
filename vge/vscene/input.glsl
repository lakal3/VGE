

layout (location = 0) in vec3 i_position;
layout (location = 1) in vec2 i_uv0;
layout (location = 2) in vec3 i_normal;
layout (location = 3) in vec3 i_tangent;
layout (location = 4) in vec3 i_color;

#ifdef SKINNED
layout (location = 5) in vec4 i_weights0;
layout (location = 6) in uvec4 i_joints0;

#endif

mat3 calcNormalSpace(mat4 world) {
    vec3 normal = normalize(vec3(world * vec4(i_normal,0)));
    vec3 tangent = normalize(vec3(world * vec4(i_tangent,0)));
    vec3 biTangent = -normalize(cross(tangent, normal));
    return mat3(tangent, biTangent, normal);
}
