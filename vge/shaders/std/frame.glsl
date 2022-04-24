
layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view;
    vec4 cameraPos;
    vec4 ambient;
    vec4 ambienty;
    vec4 viewPosition; // left, top, right, bottom
    uint lightPos;
    uint lights;
    uint debug;
    float dummy3;
} frame;

struct EXTRAV {
    vec4 v1;
    vec4 v2;
    vec4 v3;
    vec4 v4;
};


layout(set=0, binding=1) buffer EXTRAVA {
    EXTRAV va[];
} extrava;

layout(set=0, binding=1) buffer EXTRAMA {
    mat4 ma[];
} extrama;

layout(set=0, binding=2) uniform sampler2D frameImages2D[];
layout(set=0, binding=2) uniform samplerCube frameImagesCube[];
layout(set=0, binding=2) uniform sampler2DArray frameImagesArray[];
