
struct INSTANCE {
    vec4 forecolor;
    vec4 backcolor;
    vec4 clip; // l, t, r, b
    vec4 position_1; //
    vec4 position_2;
    vec4 uvGlyph_1; // w = glyph index
    vec4 uvGlyph_2;
    vec4 uvMask_1; // w = foreground mask
    vec4 uvMask_2; // w = bgMask
};

layout(constant_id = 0) const int MAX_INSTANCES = 100;

layout(set=0, binding=0) uniform FRAME {
    INSTANCE instances[MAX_INSTANCES];
} frame;