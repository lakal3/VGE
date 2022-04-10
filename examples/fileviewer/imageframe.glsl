
layout(set=0, binding=0) uniform IMAGEFRAME {
    vec4 area;
    vec4 pos1;
    vec4 pos2;
} imageframe;

layout(set=0, binding=1) uniform sampler2D image;