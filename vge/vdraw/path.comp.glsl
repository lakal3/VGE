#version 450

#extension GL_EXT_nonuniform_qualifier: require

layout (local_size_x = 16, local_size_y = 16, local_size_z = 1) in;

struct SEGMENT {
    uint deg;
    uint id;
    vec2 from;
    vec2 to;
    vec2 mid;
} segment;

layout(set=0, binding=0) buffer SEGMENTS {
    SEGMENT segments[];
} segments;

layout(set = 1, binding = 0, r8_snorm) uniform writeonly image2D outputImage;

const float epsilon = 0.000001;

vec2 rootsY(SEGMENT sg, float y0) {
    if (sg.deg != 2) {
        float d = sg.to.y - sg.from.y;
        if (abs(d) < epsilon) {
            return vec2(-1, -1);
        }
        float t = (y0 -  sg.from.y) / d;
        return vec2(t, -1);
    }
    float a = sg.from.y - 2 * sg.mid.y +sg.to.y;
    float b = -2*sg.from.y + 2 * sg.mid.y;
    float c = sg.from.y - y0;
    float r = b*b - 4*a*c;
    if (r < 0) {
        // Imaginary
        return vec2(-1, -1);
    }
    r = sqrt(r);
    return vec2((-b + r) / (2 * a), (-b - r) / (2 * a));
}

vec2 at(SEGMENT sg, float u) {
    if (sg.deg == 1) {
        vec2 d = sg.to - sg.from;
        return sg.from + d * u;
    }
    float u1 = 1 - u;
    return sg.to * (u * u) + sg.mid * (2 * u * u1) + sg.from * (u1 * u1);
}

float len2(vec2 v) {
    return v.x * v.x + v.y * v.y;
}

float ptToVector(vec2 from, vec2 to, vec2 pos) {
    vec2 v = to - from;
    return dot(pos - from, v) / len2(v);
}

const float step = 0.125;

vec3 closestPoint(uint segIndex, vec2 pos) {
    SEGMENT sg = segments.segments[segIndex];
    if (sg.deg == 1) {
        float u = ptToVector(sg.from, sg.to, pos);
        if (u < 0) {
            return vec3(sg.from, 0);
        }
        if (u > 1) {
            return vec3(sg.to, 1);
        }
        return vec3(at(sg, u), u);
    }
    vec2 prev = sg.from;
    float bestU = 0;
    float bestLen = len2(sg.from - pos);
    for (float u0 = step; u0 <= 1.0; u0 += step) {
        vec2 next = at(sg, u0);
        float u = ptToVector(prev, next, pos);
        if (u >= 0 && u <= 1) {
            vec2 at = at(sg, u0 + u * step - step);
            float nLen = len2(at - pos);
            if (nLen < bestLen) {
                bestLen = nLen;
                bestU = u0 + u * step - step;
            }
        } else {
            vec2 at = at(sg, u0);
            float nLen = len2(at - pos);
            if (nLen < bestLen) {
                bestLen = nLen;
                bestU = u0;
            }
        }
        prev = next;
    }
    return vec3(at(sg, bestU), bestU);
}

float minDistance(uint to, vec2 pos) {
    // return length(pos - segments.segments[1].from) / 10;
    uint bestSeg = 1;
    vec3 best = closestPoint(1, pos);
    float bestLen = len2(best.xy - pos);
    for (uint i = 2; i <= to; i++) {
        vec3 next = closestPoint(i, pos);
        float nextLen = len2(next.xy - pos);
        if (nextLen < bestLen) {
            best = next;
            bestLen = nextLen;
            bestSeg = i;
        }
    }
    return length(best.xy - pos);
}

float outsize(uint to, vec2 pos) {
    float dir = 1;
    float bestPos = 10000;
    for (uint i = 1; i <= to; i++) {
        SEGMENT sg = segments.segments[i];
        vec2 roots = rootsY(sg, pos.y);
        if (roots.x >= 0 && roots.x <= 1) {
            vec2 at1 = at(sg, roots.x);
            if (at1.x > pos.x && at1.x < bestPos) {
                bestPos = at1.x;
                vec2 at2 = at(sg,roots.x + 0.01);
                dir = at2.y < at1.y ? 1 : -1;
            }
        }
        if (roots.y >= 0 && roots.y <= 1) {
            vec2 at1 = at(sg, roots.y);
            if (at1.x > pos.x && at1.x < bestPos) {
                bestPos = at1.x;
                vec2 at2 = at(sg,roots.y + 0.01);
                dir = at2.y < at1.y ? 1 : -1;
            }
        }
    }
    return dir;
}

void main() {
    SEGMENT s0 = segments.segments[0];
    int w = int(s0.mid.x);
    int h = int(s0.mid.y);
    int y = int(gl_GlobalInvocationID.y);
    int x = int(gl_GlobalInvocationID.x);
    if (y > h || x > w) {
        return;
    }
    vec2 pt = s0.from + vec2(float(x) / float(w), float(y) / float(h)) * s0.to;
    float md = minDistance(s0.id, pt) * outsize(s0.id, pt);
    vec4 c = vec4(clamp(md / 2.0, -1, 1));
    imageStore(outputImage, ivec2(x,y),c);
}