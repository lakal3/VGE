
mat4 skinMatrix() {
    float w1 = 1 - (i_weights0.y + i_weights0.z + i_weights0.w);
    return w1 * joints.jointMatrix[i_joints0.x] +
    i_weights0.y * joints.jointMatrix[i_joints0.y] +
    i_weights0.z * joints.jointMatrix[i_joints0.z] +
    i_weights0.w * joints.jointMatrix[i_joints0.w];
}
