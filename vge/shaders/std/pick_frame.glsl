
struct PICK {
    uint id;
    uint vertex_nro;
    float depth;
    float colorR;
};


layout(set=1, binding=0) buffer PICKINFO {
    uint count;
    uint max;
    uint filler1;
    uint filler2;
    vec4 pickArea; // x,y - min fragment coord, z,w - max fragment coord
    PICK picks[];
} pickinfo;