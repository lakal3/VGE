#version 450

#define WG_SIZE  16

layout (local_size_x = WG_SIZE, local_size_y = WG_SIZE, local_size_z = 1) in;

layout (set = 0, binding = 0) uniform SETTINGS {
    float width; // image width
    float height; // image height
    float xOffset; // X offset in target image
    float yOffset; // Y offset in target image
    float fromSegment;
} settings;


struct LINELEN {
    float dist;
    int vn;
};

struct SEGMENT {
    float dim;
    float filler;
    vec2 p1;
    vec2 p2;
    vec2 p3;
};

layout(set = 0, binding = 1) buffer SEGMENTS {
    SEGMENT segment[];
} segments;

layout(set = 1, binding = 0, r32f) uniform writeonly image2D outputImage;

float isLeft(vec2 pos, vec2 p1, vec2 p2)  {
    return (p2.x - p1.x)*(pos.y-p1.y) - (pos.x - p1.x)*(p2.y - p1.y);
    // return (p2[0]-p1[0])*(pos[1]-p1[1]) - (pos[0]-p1[0])*(p2[1]-p1[1])
}

LINELEN  lineLen(vec2 pos, vec2 p1 , vec2 p2) {
    LINELEN ll;
    ll.vn = 0;
    vec2 a = pos - p1;
    vec2 v = p2 - p1;
    float l2 = v.x * v.x + v.y * v.y;
    if (l2 == 0) {
        ll.dist = length(a);
        return ll;
    }
    if (p1.y <= pos.y) {
        if (p2.y > pos.y) { // Upwards
            if (isLeft(pos, p1, p2) > 0) {
                ll.vn = 1;
            }
        }
    } else {
        if (p2.y <= pos.y) { // Downwards
            if (isLeft(pos, p1, p2) < 0) {
                ll.vn = -1;
            }
        }
    }

    float t = dot(a,v) / l2;
    if (t < 0) {
        ll.dist = length(a);
    } else if (t > 1) {
        ll.dist = length(p2 - pos);
    } else {
        ll.dist = length(p1 + (v * t) - pos);
    }
    return ll;
}

LINELEN quadLength(vec2 pos, vec2 p0, vec2 pMid, vec2 p1) {
    LINELEN ll;
    ll.dist = 1e10;
    ll.vn = 0;
    vec2 segPrev = p0;
    float t;
    for (t = 0.125; t <= 1; t += 0.125) {
        vec2 segNext = p0 * ((1-t) * (1-t)) + pMid * (2 * t * (1-t)) + p1 * (t * t);
        LINELEN lSeg = lineLen(pos, segPrev, segNext);
        ll.vn += lSeg.vn;
        if (ll.dist > lSeg.dist) {
            ll.dist = lSeg.dist;
        }
        segPrev = segNext;
    }
    return ll;
}

const float MaxDistance = 3.0;

void main() {
    int w = int(settings.width);
    int h = int(settings.height);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);
    if (y >= h || x >= w) {
        return;
    }
    vec2 pos = vec2(x,y);
    int fs = int(settings.fromSegment);
    int dims = 1;
    float dist = 1e10;
    int vn = 0;
    while (dims > 0) {
        SEGMENT s = segments.segment[fs];
        dims = int(s.dim);
        if (dims == 1) {
            LINELEN ll = lineLen(pos, s.p1, s.p2);
            vn += ll.vn;
            if (ll.dist < dist) {
                dist = ll.dist;
            }
        } else if (dims == 2) {
            LINELEN ll = quadLength(pos, s.p1, s.p2, s.p3);
            vn += ll.vn;
            if (ll.dist < dist) {
                dist = ll.dist;
            }
        } else {
            dims = 0;
        }
        fs += 1;
    }
    if (dist > MaxDistance) {
        dist = MaxDistance;
    }
    float f;
    if (vn == 0) { // Outside
        f = 0.5 + 0.5*dist/MaxDistance;
    } else {
        f = 0.5 - 0.5*dist/MaxDistance;
    }
    imageStore(outputImage, ivec2(x + settings.xOffset,y + settings.yOffset), vec4(f,0,0,1));
}