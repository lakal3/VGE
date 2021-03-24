package vglyph

import (
	"errors"
	"image"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type segment struct {
	deg    int // 1 - for line, 2 for quadratic bezier
	points [3]mgl32.Vec2
}

type VectorBuilder struct {
	name       string
	segments   []segment
	min        mgl32.Vec2
	max        mgl32.Vec2
	margin     int
	charOffset image.Point
	edges      image.Rectangle

	size    image.Point
	offset  image.Point
	content []byte
	hasDim  bool
}

func (vb *VectorBuilder) AddLine(p1, p2 mgl32.Vec2) *VectorBuilder {
	vb.addLimit(p1)
	vb.segments = append(vb.segments, segment{deg: 1, points: [3]mgl32.Vec2{p1, p2, mgl32.Vec2{}}})
	vb.addLimit(p2)
	return vb
}

func (vb *VectorBuilder) AddPoint(p1 mgl32.Vec2) *VectorBuilder {
	vb.addLimit(p1)
	return vb
}

func (vb *VectorBuilder) AddQuadratic(p1, p2, p3 mgl32.Vec2) *VectorBuilder {
	vb.addLimit(p1)
	vb.segments = append(vb.segments, segment{deg: 2, points: [3]mgl32.Vec2{p1, p2, p3}})
	vb.addLimit(p2)
	vb.addLimit(p3)
	return vb
}

func (vb *VectorBuilder) AddRect(outside bool, left mgl32.Vec2, size mgl32.Vec2) *VectorBuilder {
	rt := left.Add(mgl32.Vec2{size[0], 0})
	rb := left.Add(mgl32.Vec2{size[0], size[1]})
	lb := left.Add(mgl32.Vec2{0, size[1]})
	if outside {
		vb.AddLine(left, rt)
		vb.AddLine(rt, rb)
		vb.AddLine(rb, lb)
		vb.AddLine(lb, left)
	} else {
		vb.AddLine(left, lb)
		vb.AddLine(lb, rb)
		vb.AddLine(rb, rt)
		vb.AddLine(rt, left)
	}
	return vb
}

func (vb *VectorBuilder) AddCornerRect(outside bool, left mgl32.Vec2, size mgl32.Vec2, corners mgl32.Vec4) *VectorBuilder {
	if outside {
		pp := left.Add(mgl32.Vec2{corners[0], 0})
		pn := left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
	} else {
		pp := left.Add(mgl32.Vec2{corners[0], 0})
		pn := left.Add(mgl32.Vec2{0, 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
	}
	return vb
}

func (vb *VectorBuilder) AddRoundedRect(outside bool, left mgl32.Vec2, size mgl32.Vec2, corners mgl32.Vec4) *VectorBuilder {
	if outside {
		pp := left.Add(mgl32.Vec2{corners[0], 0})
		pn := left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddCorner(pp, left.Add(mgl32.Vec2{size[0], 0}), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(size).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddCorner(pp, left.Add(size), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddCorner(pp, left.Add(mgl32.Vec2{0, size[1]}), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{corners[0], 0})
		vb.AddCorner(pp, left, pn)
	} else {
		pp := left.Add(mgl32.Vec2{corners[0], 0})
		pn := left.Add(mgl32.Vec2{0, corners[1]})
		vb.AddCorner(pp, left, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{0, size[1]}).Add(mgl32.Vec2{corners[0], 0})
		vb.AddCorner(pp, left.Add(mgl32.Vec2{0, size[1]}), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], size[1]}).Sub(mgl32.Vec2{0, corners[3]})
		vb.AddCorner(pp, left.Add(size), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Add(mgl32.Vec2{0, corners[1]})
		vb.AddLine(pp, pn)
		pp, pn = pn, left.Add(mgl32.Vec2{size[0], 0}).Sub(mgl32.Vec2{corners[2], 0})
		vb.AddCorner(pp, left.Add(mgl32.Vec2{size[0], 0}), pn)
		pp, pn = pn, left.Add(mgl32.Vec2{corners[0], 0})
		vb.AddLine(pp, pn)
	}
	return vb
}

func (vb *VectorBuilder) AddCorner(from mgl32.Vec2, mid mgl32.Vec2, to mgl32.Vec2) {
	vb.AddQuadratic(from, mid, to)
}

func (vb *VectorBuilder) addLimit(p mgl32.Vec2) {
	if !vb.hasDim {
		vb.min, vb.max, vb.hasDim = p, p, true
		return
	}
	if vb.min[0] > p[0] {
		vb.min[0] = p[0]
	}
	if vb.min[1] > p[1] {
		vb.min[1] = p[1]
	}
	if vb.max[0] < p[0] {
		vb.max[0] = p[0]
	}
	if vb.max[1] < p[1] {
		vb.max[1] = p[1]
	}
}

func (vb *VectorBuilder) calcSize() {
	limits := vb.max.Sub(vb.min).Add(mgl32.Vec2{1, 1})

	vb.size = image.Pt(int(limits[0]), int(limits[1])).Add(image.Pt(vb.margin*2, vb.margin*2))
	// sd.im = image.NewRGBA(image.Rectangle{Max: ptLimits})
	offset := vb.min.Sub(mgl32.Vec2{float32(vb.margin), float32(vb.margin)})
	for idx, edge := range vb.segments {
		vb.segments[idx].points[0] = edge.points[0].Sub(offset)
		vb.segments[idx].points[1] = edge.points[1].Sub(offset)
		if edge.deg > 1 {
			vb.segments[idx].points[2] = edge.points[2].Sub(offset)
		}
	}
}

func (vb *VectorBuilder) renderOne(maxDistance float32) {
	vb.content = make([]byte, vb.size.Y*vb.size.X)
	for y := 0; y < vb.size.Y; y++ {
		for x := 0; x < vb.size.X; x++ {
			vb.fillPoint(maxDistance, x, y)
		}
	}
}

const notSetDistance = float32(1e10)

// from http://geomalgorithms.com/a03-_inclusion.html
// isLeft(): tests if a point is Left|On|Right of an infinite line.
//    Input:  three points P0, P1, and P2
//    Return: >0 for P2 left of the line through P0 and P1
//            =0 for P2  on the line
//            <0 for P2  right of the line
//    See: Algorithm 1 "Area of Triangles and Polygons"

func isLeft(pos, p1, p2 mgl32.Vec2) float32 {
	return (p2[0]-p1[0])*(pos[1]-p1[1]) - (pos[0]-p1[0])*(p2[1]-p1[1])
}

func (vb *VectorBuilder) fillPoint(maxDistance float32, x int, y int) {
	pos := mgl32.Vec2{float32(x), float32(y)}
	dist := notSetDistance
	var vn int

	for _, sg := range vb.segments {
		edgeDist := maxDistance
		var vnEdge int
		switch sg.deg {
		case 1:
			edgeDist, vnEdge = vb.lineLen(pos, sg.points[0], sg.points[1])
		case 2:
			edgeDist, vnEdge = vb.quadLen(pos, sg.points[0], sg.points[1], sg.points[2])
		}
		if edgeDist < dist {
			dist = edgeDist
		}
		vn += vnEdge
	}
	if dist > maxDistance {
		dist = maxDistance
	}
	var f float32
	if vn <= 0 { // Outside
		f = 0.5 + 0.5*dist/maxDistance
	} else {
		f = 0.5 - 0.5*dist/maxDistance
	}
	vb.content[y*vb.size.X+x] = uint8(f * 255)
}

func (vb *VectorBuilder) lineLen(pos mgl32.Vec2, p1 mgl32.Vec2, p2 mgl32.Vec2) (float32, int) {
	a := pos.Sub(p1)
	v := p2.Sub(p1)

	l2 := v.LenSqr()
	if l2 == 0 {
		return a.Len(), 0
	}
	vn := 0
	if p1[1] <= pos[1] {
		if p2[1] > pos[1] { // Upwards
			if isLeft(pos, p1, p2) > 0 {
				vn = 1
			}
		}
	} else {
		if p2[1] <= pos[1] { // Downwards
			if isLeft(pos, p1, p2) < 0 {
				vn = -1
			}
		}
	}

	t := a.Dot(v) / l2
	if t < 0 {
		return a.Len(), vn
	}
	if t > 1 {
		return p2.Sub(pos).Len(), vn
	}
	return p1.Add(v.Mul(t)).Sub(pos).Len(), vn
}

func (vb *VectorBuilder) quadLen(pos mgl32.Vec2, p0 mgl32.Vec2, pMid mgl32.Vec2, p1 mgl32.Vec2) (quadDist float32, quadVn int) {
	quadDist = 1e10
	segPrev := p0
	for t := float32(0.125); t <= 1; t += 0.125 {
		segNext := p0.Mul((1 - t) * (1 - t)).Add(pMid.Mul(2 * t * (1 - t))).Add(p1.Mul(t * t))
		dist, vn := vb.lineLen(pos, segPrev, segNext)
		if dist < quadDist {
			quadDist = dist
		}
		segPrev = segNext
		quadVn += vn
	}
	return
}

func (vb *VectorBuilder) addSegments(fSegs []float32) []float32 {
	for _, sg := range vb.segments {
		fSegs = append(fSegs, float32(sg.deg), 0, sg.points[0][0], sg.points[0][1], sg.points[1][0], sg.points[1][1],
			sg.points[2][0], sg.points[2][1])
	}
	fSegs = append(fSegs, 0, 0, 0, 0, 0, 0, 0, 0)
	return fSegs
}

// VectorSetBuilder converts vector based images (fonts and drawn) to glyph set
type VectorSetBuilder struct {
	Ctx         vk.APIContext
	MaxDistance float32
	glyphs      []*VectorBuilder
	b           *sfnt.Buffer
}

// Add glyph using Vector builder. Margin will be added to final glyph. You must have few pixel around edges
// to separate glyph inside from outside
func (vsb *VectorSetBuilder) AddGlyph(name string, margin int) *VectorBuilder {
	vb := &VectorBuilder{name: name, margin: margin}
	vsb.glyphs = append(vsb.glyphs, vb)
	return vb
}

// Add glyph that have 3 sides: top, center, bottom or left, center, right.
// 9 side glyphs have left, top, right, bottom, center and all corners.
func (vsb *VectorSetBuilder) AddEdgedGlyph(name string, margin int, edges image.Rectangle) *VectorBuilder {
	vb := vsb.AddGlyph(name, margin)
	vb.edges = edges
	return vb
}

// Range of unicode points, ranges characters including
type Range struct {
	From rune
	To   rune
}

// Add font to vector set builder. Glyph names will be directly font character converted as string
// Currently only fonts containing lines and quadratics bezier lines are support (this should include all ttf fonts)
// Cubic bezier lines are not
func (vsb *VectorSetBuilder) AddFont(ctx vk.APIContext, fontContent []byte, ranges ...Range) {
	err := vsb.addFont(fontContent, ranges)
	if err != nil {
		ctx.SetError(err)
	}
}

func (vsb *VectorSetBuilder) addFont(content []byte, ranges []Range) error {
	font, err := sfnt.Parse(content)
	if err != nil {
		return err
	}
	for _, rg := range ranges {
		for ch := rg.From; ch <= rg.To; ch++ {
			vsb.AddChar(font, NOMINALFontSize, ch)
		}
	}
	return nil
}

// Add individual character from font to vector builder
func (vsb *VectorSetBuilder) AddChar(font *sfnt.Font, pixelSize int, r rune) *VectorBuilder {
	if vsb.b == nil {
		vsb.b = &sfnt.Buffer{}
	}
	idx, err := font.GlyphIndex(vsb.b, r)
	if err != nil || idx == 0 {
		return nil
	}
	vb := vsb.AddGlyph(string(r), 3)
	segments, err := font.LoadGlyph(vsb.b, idx, fixed.I(pixelSize), nil)
	if err != nil {
		return nil
	}
	var prevPos mgl32.Vec2
	for _, sg := range segments {
		switch sg.Op {
		case sfnt.SegmentOpMoveTo:
			prevPos = toVector(sg, 0)
		case sfnt.SegmentOpLineTo:
			pos := toVector(sg, 0)
			vb.AddLine(prevPos, pos)
			prevPos = pos
		case sfnt.SegmentOpQuadTo:
			mid := toVector(sg, 0)
			pos := toVector(sg, 1)
			vb.AddQuadratic(prevPos, mid, pos)
			prevPos = pos
		default:
			panic("Invalid OP")
		}
	}
	vb.charOffset = image.Pt(int(vb.min[0]), int(vb.min[1]))
	return vb
}

// Convert added vector sets to glyph set. This glyph set will be signed depth field
func (vsb *VectorSetBuilder) Build(ctx vk.APIContext, dev *vk.Device) *GlyphSet {
	for _, vb := range vsb.glyphs {
		vb.calcSize()
	}
	w := 256
	h := 0
	for h == 0 || h > w {
		w = 2 * w
		h = vsb.calcImageSize(w)
		if w > MAXImageWidth {
			ctx.SetError(errors.New("Glyph set too large"))
			return nil
		}
	}
	bi := &vectorBuildInfo{}
	defer bi.owner.Dispose()
	bi.gs = newGlyphSet(ctx, dev, len(vsb.glyphs), w, h, SETDepthField)
	bi.prepare(ctx, dev, vsb)
	bi.render(ctx, dev, vsb)
	vsb.addGlyphs(bi.gs)
	bi.gs.Advance = DefaultAdvance
	return bi.gs
}

func (vsb *VectorSetBuilder) addGlyphs(gs *GlyphSet) {
	for _, gb := range vsb.glyphs {
		max := gb.offset.Add(gb.size)
		gs.glyphs[gb.name] = Glyph{Name: gb.name, CharOffset: gb.charOffset,
			Location: image.Rectangle{Min: gb.offset.Add(image.Pt(1, 1)), Max: max.Sub(image.Pt(2, 2))}, Edges: gb.edges}
	}
}

func (vsb *VectorSetBuilder) calcImageSize(w int) int {
	h := 0
	var offset image.Point
	var lHeight int
	for _, im := range vsb.glyphs {
		if im.size.X > w {
			return 0 // Won't fit
		}
		if im.size.X+offset.X > w {
			h += lHeight
			offset = image.Point{0, h}
			lHeight = 0
		}
		im.offset = offset
		if lHeight < im.size.Y {
			lHeight = im.size.Y
		}
		offset.X = offset.X + im.size.X
	}
	return h + lHeight
}

func toVector(sg sfnt.Segment, pos int) mgl32.Vec2 {
	p := sg.Args[pos]
	return mgl32.Vec2{float32(p.X), float32(p.Y)}.Mul(1.0 / 64)
}

type vectorBuildInfo struct {
	gs        *GlyphSet
	owner     vk.Owner
	segments  *vk.Buffer
	sUniforms []*vk.Slice
	dsIn      []*vk.DescriptorSet
	dsOut     *vk.DescriptorSet
	pl        *vk.ComputePipeline
}

var kVDepthInputLayout = vk.NewKeys(2)
var kVDepthPipeline = vk.NewKey()

func (bi *vectorBuildInfo) prepare(ctx vk.APIContext, dev *vk.Device, vsb *VectorSetBuilder) {
	pool := vk.NewMemoryPool(dev)
	bi.owner.AddChild(pool)
	segments := 0
	for _, gl := range vsb.glyphs {
		segments += len(gl.segments) + 1
	}
	bi.segments = pool.ReserveBuffer(ctx, uint64(segments*4*8), true, vk.BUFFERUsageStorageBufferBit)
	bUniforms := pool.ReserveBuffer(ctx, vk.MinUniformBufferOffsetAlignment*uint64(len(vsb.glyphs)), true,
		vk.BUFFERUsageUniformBufferBit)
	pool.Allocate(ctx)
	laIn1 := dev.Get(ctx, kVDepthInputLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeUniformBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	laIn := dev.Get(ctx, kVDepthInputLayout+1, func(ctx vk.APIContext) interface{} {
		return laIn1.AddBinding(ctx, vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)
	laOut := dev.Get(ctx, kOutputLayout, func(ctx vk.APIContext) interface{} {
		return vk.NewDescriptorLayout(ctx, dev, vk.DESCRIPTORTypeStorageImage, vk.SHADERStageComputeBit, 1)
	}).(*vk.DescriptorLayout)

	bi.pl = dev.Get(ctx, kVDepthPipeline, func(ctx vk.APIContext) interface{} {
		cp := vk.NewComputePipeline(ctx, dev)
		cp.AddShader(ctx, vdepth_comp_spv)
		cp.AddLayout(ctx, laIn)
		cp.AddLayout(ctx, laOut)
		cp.Create(ctx)
		return cp
	}).(*vk.ComputePipeline)

	dpOut := vk.NewDescriptorPool(ctx, laOut, 1)
	bi.owner.AddChild(dpOut)
	bi.dsOut = dpOut.Alloc(ctx)
	dpIn := vk.NewDescriptorPool(ctx, laIn, len(vsb.glyphs))
	bi.owner.AddChild(dpIn)
	bi.dsIn = make([]*vk.DescriptorSet, len(vsb.glyphs))
	bi.sUniforms = make([]*vk.Slice, len(vsb.glyphs))

	for idx, _ := range vsb.glyphs {
		ds := dpIn.Alloc(ctx)
		sl := bUniforms.Slice(ctx, uint64(idx)*vk.MinUniformBufferOffsetAlignment, uint64(idx+1)*vk.MinUniformBufferOffsetAlignment)
		ds.WriteSlice(ctx, 0, 0, sl)
		ds.WriteBuffer(ctx, 1, 0, bi.segments)
		bi.sUniforms[idx] = sl
		bi.dsIn[idx] = ds
	}
	fr := bi.gs.image.FullRange()
	fr.Layout = vk.IMAGELayoutGeneral
	view := vk.NewImageView(ctx, bi.gs.image, &fr)
	bi.owner.AddChild(view)
	bi.dsOut.WriteImage(ctx, 0, 0, view, nil)
}

const wgSize = 16

func (bi *vectorBuildInfo) render(ctx vk.APIContext, dev *vk.Device, vsb *VectorSetBuilder) {
	cmd := vk.NewCommand(ctx, dev, vk.QUEUEComputeBit, true)
	defer cmd.Dispose()
	cmd.Begin()
	ir := bi.gs.image.FullRange()
	cmd.SetLayout(bi.gs.image, &ir, vk.IMAGELayoutGeneral)
	var fSegments []float32
	for idx, gl := range vsb.glyphs {
		ubfs := []float32{float32(gl.size.X), float32(gl.size.Y),
			float32(gl.offset.X), float32(gl.offset.Y), float32(len(fSegments) / 8)}
		fSegments = gl.addSegments(fSegments)
		copy(bi.sUniforms[idx].Content, vk.Float32ToBytes(ubfs))
		cmd.Compute(bi.pl, uint32(gl.size.X/wgSize)+1, uint32(gl.size.Y/wgSize)+1, 1, bi.dsIn[idx], bi.dsOut)

	}
	cmd.SetLayout(bi.gs.image, &ir, vk.IMAGELayoutShaderReadOnlyOptimal)
	copy(bi.segments.Bytes(ctx), vk.Float32ToBytes(fSegments))
	cmd.Submit()
	cmd.Wait()
}
