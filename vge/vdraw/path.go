package vdraw

import "github.com/go-gl/mathgl/mgl32"

// Path records line and quadratic Bézier curves that can be later filled or added to GlyphSet
type Path struct {
	lastID   uint32
	lastPos  mgl32.Vec2
	segments []segment
	beginPos mgl32.Vec2
}

type segment struct {
	deg  uint32
	id   uint32
	from mgl32.Vec2
	to   mgl32.Vec2
	mid  mgl32.Vec2
}

// MoveTo moves current point at given location p1. This should always be first call for a path. Calling move will start a new path segment
func (p *Path) MoveTo(p1 mgl32.Vec2) *Path {
	p.lastID++
	p.lastPos = p1
	p.beginPos = p1
	return p
}

// LineTo draws line from current point to new point p1
func (p *Path) LineTo(p1 mgl32.Vec2) *Path {
	l := segment{deg: 1, from: p.lastPos, to: p1, id: p.lastID}
	p.lastPos = p1
	p.segments = append(p.segments, l)
	return p
}

// ClosePath draws line from current point to start point of path
func (p *Path) ClosePath() *Path {
	if p.lastPos != p.beginPos {
		p.LineTo(p.beginPos)
	}
	return p
}

// BezierTo draws Bézier curve from current point to point p2. P1 is control point of curve.
func (p *Path) BezierTo(p1, p2 mgl32.Vec2) *Path {
	sg := segment{deg: 2, from: p.lastPos, mid: p1, to: p2, id: p.lastID}
	p.segments = append(p.segments, sg)
	p.lastPos = p2
	return p
}

// Clear path. One path may be used to Fill several shapes.
// Note! Do not clear path added to GlyphSet before GlyphSet is built
func (p *Path) Clear() {
	p.segments = nil
	p.beginPos = mgl32.Vec2{}
	p.lastPos = p.beginPos
}

// AddRect adds outside or inside rectangle to Path
func (p *Path) AddRect(outside bool, left mgl32.Vec2, size mgl32.Vec2) *Path {
	rt := left.Add(mgl32.Vec2{size[0], 0})
	rb := left.Add(mgl32.Vec2{size[0], size[1]})
	lb := left.Add(mgl32.Vec2{0, size[1]})
	if outside {
		p.MoveTo(left)
		p.LineTo(rt)
		p.LineTo(rb)
		p.LineTo(lb)
		p.LineTo(left)
	} else {
		p.MoveTo(left)
		p.LineTo(lb)
		p.LineTo(rb)
		p.LineTo(rt)
		p.LineTo(left)
	}
	return p
}

// AddCornerRect adds outside or inside rectangle to Path. Rectangle has cut corners
func (p *Path) AddCornerRect(outside bool, left mgl32.Vec2, size mgl32.Vec2, corners Corners) *Path {
	from := len(p.segments)
	pp := left.Add(mgl32.Vec2{corners.TopLeft, 0})
	pn := left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners.TopRight, 0})
	p.MoveTo(pp)
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners.TopRight})
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners.BottomRight})
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{corners.BottomRight, 0})
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners.BottomLeft, 0})
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners.BottomLeft})
	p.LineTo(pn)
	pn = left.Add(mgl32.Vec2{0, 0}).Add(mgl32.Vec2{0, corners.TopLeft})
	p.LineTo(pn)
	p.LineTo(pp)
	if !outside {
		to := len(p.segments)
		p.invert(from, to)
	}
	return p
}

// AddRoundedRect adds outside or inside rectangle to Path. Rectangle has rounded corners
func (p *Path) AddRoundedRect(outside bool, left mgl32.Vec2, size mgl32.Vec2, corners Corners) *Path {
	from := len(p.segments)
	pp := left.Add(mgl32.Vec2{corners.TopLeft, 0})
	pn := left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners.TopRight, 0})
	p.MoveTo(pp)
	p.LineTo(pn)
	if corners.TopRight > 0 {
		pn = left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners.TopRight})
		p.addRoundedCorner(left.Add(mgl32.Vec2{size[0], 0}), pn)
	}
	pn = left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners.BottomRight})
	p.LineTo(pn)
	if corners.BottomRight > 0 {
		pn = left.Add(size).Sub(mgl32.Vec2{corners.BottomRight, 0})
		p.addRoundedCorner(left.Add(size), pn)
	}
	pn = left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners.BottomLeft, 0})
	p.LineTo(pn)
	if corners.BottomLeft > 0 {
		pn = left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners.BottomLeft})
		p.addRoundedCorner(left.Add(mgl32.Vec2{0, size[1]}), pn)
	}
	pn = left.Add(mgl32.Vec2{0, 0}).Add(mgl32.Vec2{0, corners.TopLeft})
	p.LineTo(pn)
	if corners.TopLeft > 0 {
		pn = left.Add(mgl32.Vec2{corners.TopLeft, 0})
		p.addRoundedCorner(left, pn)
	}
	p.LineTo(pp)
	if !outside {
		to := len(p.segments)
		p.invert(from, to)
	}

	return p
}

// Bounds get bounds of path (left, top, right, bottom)
func (p *Path) Bounds() (a Area) {
	if len(p.segments) == 0 {
		return Area{}
	}
	first := p.segments[0].from
	a.From, a.To = first, first

	for _, sg := range p.segments {
		if sg.deg == 2 {
			for u := float32(0); u <= 1.0; u += steps {
				p := sg.at(u)
				a.Include(p)
			}
		} else {
			a.Include(sg.from)
			a.Include(sg.to)
		}
	}
	return a
}

// Fill converts path to Filled shape. All individual path segments must be
// - closed
// - non overlapping
// - clockwise paths are inside of shape, counterclockwise are outside
// If these rules are not follow, generated path will be invalid!
func (p *Path) Fill() *Filled {
	st := &Filled{}
	st.build(p)
	return st
}

func (p *Path) IsEmpty() bool {
	return len(p.segments) == 0
}

// Translate all points of path with given transformation matrix tr
func (p *Path) Translate(tr mgl32.Mat3) {
	for idx, s := range p.segments {
		p.segments[idx] = segment{deg: s.deg,
			from: tr.Mul3x1(s.from.Vec3(1)).Vec2(),
			to:   tr.Mul3x1(s.to.Vec3(1)).Vec2(),
			mid:  tr.Mul3x1(s.mid.Vec3(1)).Vec2(),
		}
	}
}

const sq2_1 = 0.414213562373 // 2 / sqrt(2) - 1

func (p *Path) addRoundedCorner(mid mgl32.Vec2, to mgl32.Vec2) *Path {
	if to.Sub(p.lastPos).LenSqr() < epsilon {
		return p
	}
	mid1 := p.lastPos.Mul(1 - sq2_1).Add(mid.Mul(sq2_1))
	mid2 := to.Mul(1 - sq2_1).Add(mid.Mul(sq2_1))
	mid3 := mid1.Add(mid2).Mul(0.5)
	sg1 := segment{deg: 2, from: p.lastPos, mid: mid1, to: mid3}
	sg2 := segment{deg: 2, from: mid3, mid: mid2, to: to}
	p.segments = append(p.segments, sg1, sg2)
	p.lastPos = to
	return p
}

func (p *Path) invert(from int, to int) {
	d := to - 1
	e := (to - from + 1) / 2
	for idx := 0; idx < e; idx++ {
		sg1, sg2 := p.segments[from+idx], p.segments[d-idx]
		sg1.from, sg1.to = sg1.to, sg1.from
		if from+idx == d-idx {
			p.segments[from+idx] = sg1
			continue
		}
		sg2.from, sg2.to = sg2.to, sg2.from
		p.segments[from+idx], p.segments[d-idx] = sg2, sg1
	}
}
