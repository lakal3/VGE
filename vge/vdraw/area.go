package vdraw

import "github.com/go-gl/mathgl/mgl32"

// Area is rectangular area with given minimum (From) and maximum (To) points
type Area struct {
	From mgl32.Vec2
	To   mgl32.Vec2
}

func (a Area) ToVec4() mgl32.Vec4 {
	return mgl32.Vec4{a.From[0], a.From[1], a.To[0], a.To[1]}
}

// Intersect updates area a so that it only contains parts that are also part of area a2
func (a *Area) Intersect(a2 Area) {
	if a.From[0] < a2.From[0] {
		a.From[0] = a2.From[0]
	}
	if a.From[1] < a2.From[1] {
		a.From[1] = a2.From[1]
	}
	if a.To[0] > a2.To[0] {
		a.To[0] = a2.To[0]
	}
	if a.To[1] > a2.To[1] {
		a.To[1] = a2.To[1]
	}
}

// Include extends area a so that it will contain given point pt
func (a *Area) Include(pt mgl32.Vec2) {
	if a.From[0] > pt[0] {
		a.From[0] = pt[0]
	}
	if a.From[1] > pt[1] {
		a.From[1] = pt[1]
	}
	if a.To[0] < pt[0] {
		a.To[0] = pt[0]
	}
	if a.To[1] < pt[1] {
		a.To[1] = pt[1]
	}
}

// Contains check if area a contains point pt
func (a Area) Contains(pt mgl32.Vec2) bool {
	if a.From[0] > pt[0] || a.From[1] > pt[1] {
		return false
	}
	if a.To[0] < pt[0] || a.To[1] < pt[1] {
		return false
	}
	return true
}

func (a Area) Width() float32 {
	return a.To[0] - a.From[0]
}

func (a Area) Height() float32 {
	return a.To[1] - a.From[1]
}

func (a Area) Size() mgl32.Vec2 {
	return a.To.Sub(a.From)
}

func (a Area) IsNil() bool {
	s := a.Size()
	return s[0] <= epsilon || s[1] <= epsilon
}

// Corners contains dimensions for four corners
type Corners struct {
	TopLeft     float32
	TopRight    float32
	BottomLeft  float32
	BottomRight float32
}

func (c Corners) IsEmpty() bool {
	return c.TopLeft <= 0 && c.TopRight <= 0 && c.BottomLeft <= 0 && c.BottomRight <= 0
}

// UniformCorners creates Corner struct where all corners have same value (corner)
func UniformCorners(corner float32) Corners {
	return Corners{TopLeft: corner, TopRight: corner, BottomLeft: corner, BottomRight: corner}
}

// Edges contains dimensions for four edges
type Edges struct {
	Left   float32
	Top    float32
	Right  float32
	Bottom float32
}

// UniformCorners creates Edges struct where all edges have same value (u)
func UniformEdge(u float32) Edges {
	return Edges{Left: u, Right: u, Top: u, Bottom: u}
}

func (e Edges) IsEmpty() bool {
	return e == Edges{}
}

func (e Edges) ToVec4(min float32) mgl32.Vec4 {
	return mgl32.Vec4{max32(min, e.Left), max32(min, e.Top), max32(min, e.Right), max32(min, e.Bottom)}
}

// Shrink shrinks area with given edge values. Min ensures that size of each edge is at least min.
func (e Edges) Shrink(area Area, min float32) Area {
	area.From = area.From.Add(mgl32.Vec2{max32(e.Left, min), max32(e.Top, min)})
	area.To = area.To.Sub(mgl32.Vec2{max32(e.Right, min), max32(e.Bottom, min)})
	return area
}
