package vdraw

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
)

// CompileShapes will compile predefined shapes in vdraw package
func CompileShapes(dev *vk.Device) (err error) {
	rectShape, err = CompileShape(dev, rectSDF)
	if err != nil {
		return err
	}
	rrectShape, err = CompileShape(dev, calcRR+rrectSDF)
	if err != nil {
		return err
	}
	borderShape, err = CompileShape(dev, borderSDF)
	if err != nil {
		return err
	}
	rborderShape, err = CompileShape(dev, calcRR+rborderSDF)
	if err != nil {
		return err
	}
	lineShape, err = CompileShape(dev, lineSDF)
	if err != nil {
		return err
	}

	return nil
}

// Shape is compiled signed distance field function that can be used as a drawable
type Shape struct {
	spirv []byte
	key   vk.Key
}

var rectShape Shape

const rectSDF = `
float calcSDF() {
	vec4 rect = getSegment(0);
	float dx = abs(rect.x - i_pos.x) - rect.z;
	float dy = abs(rect.y - i_pos.y) - rect.w;
	return max(dx, dy);
}
`

// Rect is rectangle shape with given area
type Rect struct {
	Area Area
}

func (r Rect) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	center := r.Area.From.Add(r.Area.To).Mul(0.5)
	edges := r.Area.Size().Mul(0.5)
	da := DrawArea{Min: r.Area.From.Sub(mgl32.Vec2{2, 2}), Max: r.Area.To.Add(mgl32.Vec2{2, 2})}
	da.To = 1
	ds := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	return []DrawArea{da}, []DrawSegment{ds}
}

func (r Rect) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (r Rect) GetShape() (shape *Shape) {
	return &rectShape
}

var rrectShape Shape

const calcRR = `
float calcRR(vec4 rect, vec4 corners) {
	vec2 tl = vec2(rect.x - rect.z + corners.x, rect.y - rect.w + corners.x);
	if (i_pos.x < tl.x && i_pos.y < tl.y) {
		return length(i_pos - tl) - corners.x;
	}
	vec2 tr = vec2(rect.x + rect.z - corners.y, rect.y - rect.w + corners.y);
	if (i_pos.x > tr.x && i_pos.y < tr.y) {
		return length(i_pos - tr) - corners.y;
	}
	vec2 bl = vec2(rect.x - rect.z + corners.z, rect.y + rect.w - corners.z);
	if (i_pos.x < bl.x && i_pos.y > bl.y) {
		return length(i_pos - bl) - corners.z;
	}
	vec2 br = vec2(rect.x + rect.z - corners.w, rect.y + rect.w - corners.w);
	if (i_pos.x > br.x && i_pos.y > br.y) {
		return length(i_pos - br) - corners.w;
	}
	float dx = abs(rect.x - i_pos.x) - rect.z;
	float dy = abs(rect.y - i_pos.y) - rect.w;
	return max(dx, dy);
}
`

const rrectSDF = `
float calcSDF() {
	vec4 rect = getSegment(0);
	vec4 corners = getSegment(1);
	return calcRR(rect, corners);
}
`

// RoudedRect is rectangle shape with given area and given rounded corners
type RoudedRect struct {
	Area    Area
	Corners Corners
}

func (rr RoudedRect) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	center := rr.Area.From.Add(rr.Area.To).Mul(0.5)
	edges := rr.Area.Size().Mul(0.5)
	da := DrawArea{Min: rr.Area.From.Sub(mgl32.Vec2{2, 2}), Max: rr.Area.To.Add(mgl32.Vec2{2, 2})}
	da.To = 2
	ds := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	ds2 := DrawSegment{V1: rr.Corners.TopLeft, V2: rr.Corners.TopRight, V3: rr.Corners.BottomLeft, V4: rr.Corners.BottomRight}
	return []DrawArea{da}, []DrawSegment{ds, ds2}
}

func (rr RoudedRect) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (rr RoudedRect) GetShape() (shape *Shape) {
	return &rrectShape
}

var borderShape Shape

const borderSDF = `
float calcSDF() {
	vec4 rect = getSegment(1);
	float dx = abs(rect.x - i_pos.x) - rect.z;
	float dy = abs(rect.y - i_pos.y) - rect.w;
	float r = max(dx,dy);
	if (r < 0) {
		return -r;
	}	
	rect = getSegment(0);
	dx = abs(rect.x - i_pos.x) - rect.z;
	dy = abs(rect.y - i_pos.y) - rect.w;
	return max(dx,dy);
}
`

// Border is border shape with given area and size of edges. Some edges may be 0
type Border struct {
	Area  Area
	Edges Edges
}

func (r Border) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	center := r.Area.From.Add(r.Area.To).Mul(0.5)
	edges := r.Area.Size().Mul(0.5)
	da := DrawArea{Min: r.Area.From.Sub(mgl32.Vec2{2, 2}), Max: r.Area.To.Add(mgl32.Vec2{2, 2})}
	da.To = 2
	ds1 := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	inside := r.Edges.Shrink(r.Area, 0)
	center = inside.From.Add(inside.To).Mul(0.5)
	edges = inside.Size().Mul(0.5)
	ds2 := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	return []DrawArea{da}, []DrawSegment{ds1, ds2}
}

func (r Border) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (r Border) GetShape() (shape *Shape) {
	return &borderShape
}

const rborderSDF = `
float calcSDF() {
	vec4 rectIn = getSegment(1);
	vec4 rectOut = getSegment(0);
	vec4 corners = getSegment(2);
	float r = calcRR(rectIn, corners);
	if (r < 0) {
		return -r;
	}
	return calcRR(rectOut, corners); 
}
`

var rborderShape Shape

// RoundedBorder is border shape with given area and size of edges. You can set rounding size for each Corner. Some edges or corners may be 0
// Corner size may not be > width/2 or height/2
type RoundedBorder struct {
	Area    Area
	Edges   Edges
	Corners Corners
}

func (rb RoundedBorder) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	center := rb.Area.From.Add(rb.Area.To).Mul(0.5)
	edges := rb.Area.Size().Mul(0.5)
	da := DrawArea{Min: rb.Area.From.Sub(mgl32.Vec2{2, 2}), Max: rb.Area.To.Add(mgl32.Vec2{2, 2})}
	da.To = 2
	ds1 := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	inside := rb.Edges.Shrink(rb.Area, 0)
	center = inside.From.Add(inside.To).Mul(0.5)
	edges = inside.Size().Mul(0.5)
	ds2 := DrawSegment{V1: center[0], V2: center[1], V3: edges[0], V4: edges[1]}
	ds3 := DrawSegment{V1: rb.Corners.TopLeft, V2: rb.Corners.TopRight, V3: rb.Corners.BottomLeft, V4: rb.Corners.BottomRight}

	return []DrawArea{da}, []DrawSegment{ds1, ds2, ds3}
}

func (rb RoundedBorder) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (rb RoundedBorder) GetShape() (shape *Shape) {
	return &rborderShape
}

var lineShape Shape

const lineSDF = `
float len2(vec2 v) {
    return v.x * v.x + v.y * v.y;
}

float ptToVector(vec2 from, vec2 to, vec2 pos) {
    vec2 v = to - from;
    return dot(pos - from, v) / len2(v);
}

float calcSDF() {
	vec4 line = getSegment(0);
	vec4 thickness = getSegment(1);
	float u = ptToVector(line.xy, line.zw, i_pos);
	if (u < 0 || u > 1) {
		return 1;
	}
	vec2 pos = line.xy + u * (line.zw - line.xy);
	return length(pos - i_pos) - thickness.x / 2;
}

`

// Line draw straight line from point From to To with given Thickness
type Line struct {
	From      mgl32.Vec2
	To        mgl32.Vec2
	Thickness float32
}

func (l Line) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	a := l.area()
	return []DrawArea{{Min: a.From, Max: a.To, To: 2}}, []DrawSegment{
		{V1: l.From[0], V2: l.From[1], V3: l.To[0], V4: l.To[1]},
		{V1: l.Thickness},
	}
}

func (l Line) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (l Line) GetShape() (shape *Shape) {
	return &lineShape
}

func (l Line) area() Area {
	return Area{
		From: mgl32.Vec2{min32(l.From[0], l.To[0]), min32(l.From[1], l.To[1])}.Sub(mgl32.Vec2{2, 2}),
		To:   mgl32.Vec2{max32(l.From[0], l.To[0]), max32(l.From[1], l.To[1])}.Add(mgl32.Vec2{2, 2}),
	}
}

// CompileShape convert SDFfunction written is glsl to a shape program. See example functions from shape.go
func CompileShape(dev *vk.Device, SDFfunction string) (shape Shape, err error) {
	c := vk.NewCompiler(dev)
	defer c.Dispose()
	source := shapeShaderStart + SDFfunction + shapeShaderEnd
	shape.spirv, _, err = c.Compile(vk.SHADERStageFragmentBit, source)
	if err != nil {
		return
	}
	shape.key = vk.NewKey()
	return shape, nil
}

const shapeShaderStart = `
#version 450
#extension GL_EXT_nonuniform_qualifier: require

layout(set=1, binding=0) buffer SEGMENTS {
    vec4 segments[];
} segments;

layout(push_constant) uniform INSTANCE {
    vec4 clip;  // Clip in world coordinates
    vec4 color; // Color
    vec4 color2; // Color
    vec4 uv1; // first row of uv matrix from pos, w - image index
    vec4 uv2;
    vec2 scale;
    vec2 glyph;// Glyph image if x > 0, image layer in y
} inst;


layout(set=1, binding=1) uniform sampler2D tx_masks[];
layout(set=1, binding=1) uniform sampler2DArray tx_glyphs[];

layout(location = 0) in vec2 i_pos;
layout(location = 1) in vec2 i_wpos;
layout(location = 2) in vec2 i_uv0;
layout(location = 3) in vec2 i_uvGlyph;
layout(location = 4) in flat uvec2 i_segments;

layout(location = 0) out vec4 o_color;

vec4 getSegment(int relIndex) {
	return segments.segments[i_segments.x + relIndex];
}


`

const shapeShaderEnd = `


void main() {
    if (i_wpos.x < inst.clip.x || i_wpos.x > inst.clip.z || i_wpos.y < inst.clip.y || i_wpos.y > inst.clip.w) {
        discard; // Clipped
    }
    float a = 0;
    float r = calcSDF();
    a = clamp((0.4-r) * 0.75 , 0, 1);
    vec4 c = inst.color + fract(i_uv0.x) * inst.color2;
    int textImage = int(inst.uv1.w);
    if (textImage > 0) {
        c = c * texture(tx_masks[textImage], i_uv0);
    }

    o_color = vec4(c.rgb, c.a * a);
}

`
