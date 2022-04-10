
layout(set=1, binding=0) uniform SHADOWFRAME {
// From camera to light
    vec4 plane;
    vec4 lightPos;
    float minShadow;
    float maxShadow;
    float yFactor;
    float dummy2;
} shadowFrame;