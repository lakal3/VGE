
layout(constant_id = 0) const int MAX_INSTANCES = 100;

layout(set=0, binding=0) uniform FRAME {
// From camera to light
    vec4 plane;
    vec4 lightPos;
    float minShadow;
    float maxShadow;
    float dummy1;
    float dummy2;
    mat4 [MAX_INSTANCES]instances;
} frame;
