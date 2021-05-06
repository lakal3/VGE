
layout(constant_id = 0) const int MAX_INSTANCES = 800;

struct INSTANCE {
    mat4 world;
    float tx_albedo;
    float alphaCutoff;
    float filler1;
    float filler2;
};

layout(set=0, binding=0) uniform FRAME {
// From camera to light
    vec4 plane;
    vec4 lightPos;
    float minShadow;
    float maxShadow;
    float yFactor;
    float dummy2;
} frame;

#ifdef DYNAMIC_DESCRIPTORS
layout(set=0, binding=1) uniform sampler2D frameImages2D[];
#endif

layout(set=1, binding=0) uniform INSTANCES {
    INSTANCE [MAX_INSTANCES]instances;
} instances;

