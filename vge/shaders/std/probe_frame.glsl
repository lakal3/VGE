

layout(set=1, binding=0) uniform PROBEFRAME {
// From camera to light
   mat4 projection;
   mat4 views[6];
   vec4 cameraPos;

} probeFrame;