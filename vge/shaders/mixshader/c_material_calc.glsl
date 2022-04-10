
// void main() {
    vec4 c1 = texture(frameImages2D[int(instance.cTextures1.x)], i_uv0);
    vec4 c2 = texture(frameImages2D[int(instance.cTextures1.y)], i_uv0);
    mat.albedo = mix(c1, c2, mat.albedo.r);
// }