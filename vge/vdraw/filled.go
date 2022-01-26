package vdraw

import (
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"math"
	"sort"
)

const epsilon = 0.00001

// Filled is Drawable shape. Is is create using Fill method of Path
type Filled struct {
	err error
	ds  []DrawSegment
	da  []DrawArea
}

func (f *Filled) GetShape() (shape *Shape) {
	return nil
}

func (f *Filled) GetGlyph() (view *vk.ImageView, layer uint32) {
	return nil, 0
}

func (f *Filled) GetDrawData() (areas []DrawArea, segments []DrawSegment) {
	return f.da, f.ds
}

type filler struct {
	lines   []line
	err     error
	areas   []*fillArea
	lastSeg uint32
	lastDir int
	cutId   uint32
}

type line struct {
	cutId uint32
	segId uint32
	from  mgl32.Vec2
	to    mgl32.Vec2
}

type fillArea struct {
	fromY  float32
	toY    float32
	lefts  []int
	rights []int
}

func (s *Filled) build(p *Path) *filler {
	f := filler{lines: make([]line, 0, len(p.segments)*2)}
	for _, sg := range p.segments {
		f.addSegment(sg)
	}

	f.buildAreas()
	f.emitDraw(s)
	s.err = f.err
	return &f
}

// Copy segments. Split Bézier curves that have derivative over t = 0. One curve can only go upwards or downwards
func (f *filler) addSegment(sg segment) {
	if sg.deg == 2 {
		// t0 = (y0 - y1) / (y0 - 2 y1 + y2), derivative of curve is zero
		d := sg.from[1] - 2*sg.mid[1] + sg.to[1]
		if d > -epsilon && d < epsilon {
			f.addCurve(sg)
			return
		}
		t0 := (sg.from[1] - 2*sg.mid[1]) / d
		if t0 <= 0 || t0 >= 1 {
			f.addCurve(sg)
			return
		}
		sg2, sg3 := sg.split(t0)
		f.addCurve(sg2)
		f.addCurve(sg3)
	} else {
		f.addLine(sg.from, sg.to, sg.id)
	}
}

const steps = 0.25

func (f *filler) addCurve(sg segment) {
	prev := sg.at(0)
	for step := float32(steps); step <= 1.0; step += steps {
		next := sg.at(step)
		f.addLine(prev, next, sg.id)
		prev = next
	}
}

func (f *filler) buildAreas() {
	type ySeg struct {
		y   float32
		idx int
	}
	var ys []ySeg
	for idx, sg := range f.lines {
		ys = append(ys, ySeg{y: sg.from[1], idx: idx}, ySeg{y: sg.to[1], idx: -idx - 1})
	}

	sort.Slice(ys, func(i, j int) bool {
		return ys[i].y < ys[j].y
	})
	prev := float32(0)
	for idx, sl := range ys {
		if idx == 0 {
			prev = sl.y
			continue
		}
		if idx > 0 && (sl.y-prev) < epsilon {
			if sl.idx < 0 {
				f.lines[-sl.idx-1].to[1] = prev
			} else {
				f.lines[sl.idx].from[1] = prev
			}
			continue
		}
		f.addSlices(prev, sl.y)
		prev = sl.y
	}
}

type segCenter struct {
	index  int
	center float32
}

func (f *filler) addSlices(fromY float32, toY float32) {
	mid := (fromY + toY) / 2
	var lefts []segCenter
	var rights []segCenter
	for idx, ln := range f.lines {
		tMid, ok := ln.solveY(mid)
		if !ok {
			continue
		}
		switch ln.dir() {
		case -1:
			// up
			lefts = append(lefts, segCenter{index: idx, center: ln.at(tMid)[0]})
		case 1:
			rights = append(rights, segCenter{index: idx, center: ln.at(tMid)[0]})
		}
	}
	if len(lefts) != len(rights) {
		f.setError(fmt.Errorf("No matching lefts and rights from %f to %f", fromY, toY))
		return
	}
	sort.Slice(lefts, func(i, j int) bool {
		return lefts[i].center < lefts[j].center
	})
	sort.Slice(rights, func(i, j int) bool {
		return rights[i].center < rights[j].center
	})
	for idx := 0; idx < len(lefts); idx++ {
		if lefts[idx].center > rights[idx].center {
			f.setError(fmt.Errorf("Right before left from %f to %f", fromY, toY))
		} else {
			f.updateArea(lefts[idx].index, rights[idx].index, fromY, toY)
		}
	}
}

func (f *filler) setError(err error) {
	if f.err == nil {
		f.err = err
	}
}

func (f *filler) updateArea(li int, ri int, fromY float32, toY float32) {
	for _, fa := range f.areas {
		if fa.toY == fromY && f.lines[fa.lefts[0]].cutId == f.lines[li].cutId &&
			f.lines[fa.rights[0]].cutId == f.lines[ri].cutId {
			fa.toY = toY
			fa.lefts = f.addUnique(fa.lefts, li)
			fa.rights = f.addUnique(fa.rights, ri)
			return
		}
	}
	f.areas = append(f.areas, &fillArea{fromY: fromY, toY: toY, lefts: []int{li}, rights: []int{ri}})
}

func (f *filler) emitDraw(s *Filled) {

	for _, fa := range f.areas {
		f.emitDrawArea(fa, s)
	}
}

func (f *filler) addUnique(vecs []int, idx int) []int {
	for _, v := range vecs {
		if v == idx {
			return vecs
		}
	}
	return append(vecs, idx)
}

func (f *filler) emitDrawArea(fa *fillArea, s *Filled) {
	ya := make([]float32, 0, len(fa.lefts)+len(fa.rights))
	ya = append(ya, fa.fromY, fa.toY)
	for _, lIdx := range fa.lefts {
		l := f.lines[lIdx]
		if l.from[1] > fa.fromY && l.from[1] < fa.toY {
			ya = append(ya, l.from[1])
		}
		if l.to[1] > fa.fromY && l.to[1] < fa.toY {
			ya = append(ya, l.to[1])
		}
	}
	for _, lIdx := range fa.rights {
		l := f.lines[lIdx]
		if l.from[1] > fa.fromY && l.from[1] < fa.toY {
			ya = append(ya, l.from[1])
		}
		if l.to[1] > fa.fromY && l.to[1] < fa.toY {
			ya = append(ya, l.to[1])
		}
	}
	sort.Slice(ya, func(i, j int) bool {
		return ya[i] < ya[j]
	})

	da := DrawArea{From: uint32(len(s.ds))}
	isOpen := float32(0)
	if f.lines[fa.lefts[0]].cutId != f.lines[fa.rights[0]].cutId {
		isOpen = 1
	}
	var minX, maxX, prev float32
	for idx, y := range ya {
		if idx > 0 && y-prev < epsilon {
			continue
		}

		ds := DrawSegment{V3: y, V1: f.at(fa.lefts, y), V2: f.at(fa.rights, y), V4: isOpen}
		s.ds = append(s.ds, ds)
		if idx == 0 {
			minX, maxX = ds.V1, ds.V2
		} else {
			minX = min32(ds.V1, minX)
			maxX = max32(ds.V2, maxX)
		}
		prev = y
	}
	da.To = uint32(len(s.ds))
	da.Min = mgl32.Vec2{minX - 1, fa.fromY}
	da.Max = mgl32.Vec2{maxX + 1, fa.toY}
	s.da = append(s.da, da)
}

func (f *filler) at(lines []int, y float32) float32 {
	for _, ln := range lines {
		l := f.lines[ln]
		t, ok := l.solveY(y)
		if ok {
			return l.at(t)[0]
		}
	}
	f.setError(errors.New("At failed"))
	return -1
}

func (f *filler) addLine(from mgl32.Vec2, to mgl32.Vec2, id uint32) {
	l := line{from: from, to: to}
	d := l.dir()
	if d == 0 {
		f.lastDir = 0
		return
	}

	if f.lastSeg != id || f.lastDir != d {
		f.cutId++
		f.lastSeg, f.lastDir = id, d
	}
	l.cutId = f.cutId
	f.lines = append(f.lines, l)

}

func (sg segment) solve1(axis int, c0 float32) (root float32, found bool) {
	r1, r2 := sg.solve(axis, c0)
	if is_0_1(r1) {
		if !is_0_1(r2) {
			return r1, true
		}
		return r1, r2-r1 > -epsilon && r1-r2 > -epsilon
	}
	if is_0_1(r2) {
		return r2, true
	}
	return -1, false
}

func is_0_1(d float32) bool {
	return d > -epsilon && d < 1+epsilon
}

func (sg segment) solve(axis int, c0 float32) (root1, root2 float32) {
	if sg.deg == 1 {
		d := sg.to[axis] - sg.from[axis]
		if d > -epsilon && d < epsilon {
			return -1, -1
		}
		// p0 + t * d = c0 -> (p0 - c0) /d = -t
		t := (c0 - sg.from[axis]) / d
		return t, -1
	} else {
		// to * u ^ 2 + 2 * u * (1 - u) * mid + (1 - u) ^ 2 * from = result
		// u ^ 2 * (to + 2 * mid + from) + u * (-2 * mid + -2 * from) + from
		a := sg.to[axis] - 2*sg.mid[axis] + sg.from[axis]
		b := 2*sg.mid[axis] - 2*sg.from[axis]
		c := sg.from[axis] - c0
		// -b +/- sqrt(b^2 - 4 * a * c) / 2a
		s := b*b - 4*a*c
		if s < 0 {
			return -1, -1
		}
		s = float32(math.Sqrt(float64(s)))
		u1 := (-b - s) / (2 * a)
		u2 := (-b + s) / (2 * a)
		return u1, u2
	}
}

func (sg segment) at(u float32) mgl32.Vec2 {
	if sg.deg == 1 {
		d := sg.to.Sub(sg.from)
		return sg.from.Add(d.Mul(u))
	} else {
		u1 := 1 - u
		return sg.to.Mul(u * u).Add(sg.mid.Mul(2 * u * u1)).Add(sg.from.Mul(u1 * u1))
	}
}

func (sg segment) split(t float32) (s1 segment, s2 segment) {
	if sg.deg == 1 {
		mid := sg.to.Sub(sg.from).Mul(t).Add(sg.from)
		return segment{deg: 1, from: sg.from, to: mid, id: sg.id},
			segment{deg: 1, from: mid, to: sg.to, id: sg.id}
	}
	// https://math.stackexchange.com/questions/1408478/subdividing-a-b%C3%A9zier-curve-into-n-curves
	sg2 := segment{
		// P0,2 (1−z)^2P0+2(1−z)zP1+z2P2 P2,1=(1−z)P1+zP2 P2,2=P2
		deg: 2, id: sg.id,
		from: sg.from.Mul((1 - t) * (1 - t)).Add(sg.mid.Mul(2 * (1 - t) * t)).Add(sg.to.Mul(t * t)),
		mid:  sg.mid.Mul((1 - t)).Add(sg.to.Mul(t)), to: sg.to,
	}
	sg3 := segment{deg: 2, id: sg.id,
		// P1,0=P0 P1,1=(1−z)P0+zP1 P1,2=(1−z)^2P0+2(1−z)zP1+z^2P2
		from: sg.from, mid: sg.from.Mul((1 - t)).Add(sg.mid.Mul(t)),
		to: sg.from.Mul((1 - t) * (1 - t)).Add(sg.mid.Mul(2 * (1 - t) * t)).Add(sg.to.Mul(t * t)),
	}
	return sg3, sg2
}

func (sg segment) split2(r1 float32, r2 float32) segment {
	_, sg2 := sg.split(r1)
	sg2, _ = sg2.split(r2 / (1 - r1))
	return sg2
}

// dir -1 up, 0 = parallel, 1 down
func (sg segment) dir() int {
	d := sg.to[1] - sg.from[1]
	if d < -epsilon {
		return -1
	}
	if d > epsilon {
		return 1
	}
	return 0
}

func (sg segment) outbound() int {
	if sg.deg == 1 {
		return 0 // Line
	}
	l1 := sg.to.Sub(sg.from).Vec3(0)
	l2 := sg.mid.Sub(sg.from).Vec3(0)
	cr := l2.Cross(l1)
	if cr[2] > epsilon {
		return 1 // Outbound. First midline clockwise to center line
	}
	if cr[2] < -epsilon {
		return -1 // Inbound. First midline counterclockwise to center line
	}
	return 0
}

// Convert position from midlines to center line (for Outbound lines)
func (sg segment) atCenter(y float32) mgl32.Vec2 {
	delta := sg.to.Sub(sg.from)
	t := (y - sg.from[1]) / (delta[1])
	return sg.from.Add(delta.Mul(t))
}

func (ln line) solveY(y float32) (float32, bool) {
	d := ln.to[1] - ln.from[1]
	if d > -epsilon && d < epsilon {
		return 0, false
	}
	// p0 + t * d = c0 -> (p0 - c0) /d = -t
	t := (y - ln.from[1]) / d
	if t > -epsilon && t < epsilon+1 {
		return t, true
	}
	return t, false
}

func (ln line) at(t float32) mgl32.Vec2 {
	return ln.from.Mul(1 - t).Add(ln.to.Mul(t))
}

// dir -1 up, 0 = parallel, 1 down
func (l line) dir() int {
	d := l.to[1] - l.from[1]
	if d < -epsilon {
		return -1
	}
	if d > epsilon {
		return 1
	}
	return 0
}

func min32(m1, m2 float32) float32 {
	if m1 < m2 {
		return m1
	}
	return m2
}

func max32(m1, m2 float32) float32 {
	if m1 < m2 {
		return m2
	}
	return m1
}
