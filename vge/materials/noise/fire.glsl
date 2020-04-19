
struct INSTANCE {
    mat4 world;
    float offset;
    float heat;
    float dummy1;
    float dummy2;
};

layout(constant_id = 0) const int MAX_INSTANCES = 200;

layout(set=1, binding=0) uniform INSTANCES {
    INSTANCE instances[MAX_INSTANCES];
} instance;

