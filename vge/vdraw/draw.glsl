
// layout(constant_id = 0) const int MAX_INSTANCES = 100;

layout(push_constant) uniform INSTANCE {
    vec4 clip;  // Clip in world coordinates
    vec4 color; // Color
    vec4 color2; // Color
    vec4 uv1; // first row of uv matrix from pos, w - image index
    vec4 uv2;
    vec2 scale;
    vec2 glyph;// Glyph image if x > 0, image layer in y
} inst;


layout(set = 0, binding = 0) uniform FRAME {
    mat4 viewProj;
    // INSTANCE instances[MAX_INSTANCES];
} frame;

