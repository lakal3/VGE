// Standard deferred frame layout

layout(set=0, binding=0) uniform FRAME {
    mat4 projection;
    mat4 view;
    vec4 cameraPos;
    float divisor;
    float filler1;
    float filler2;
    float filler3;
} frame;

layout(set=0, binding=1) uniform sampler2D frameImages2D[];
