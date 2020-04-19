
layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view[6];
    vec4 lightPos;  // w - max
    mat4 [MAX_INSTANCES]instances;
} frame;