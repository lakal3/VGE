package vglyph

import (
	"image"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

type Position struct {
	// Clip left, top, right, bottom
	Clip      image.Rectangle
	ImageSize image.Point
	GlyphArea image.Rectangle
	Rotate    float32
}

func (pos Position) AddClip(clip image.Rectangle) Position {
	if pos.Clip.Min.X < clip.Min.X {
		pos.Clip.Min.X = clip.Min.X
	}
	if pos.Clip.Min.Y < clip.Min.Y {
		pos.Clip.Min.Y = clip.Min.Y
	}
	if pos.Clip.Max.X > clip.Max.X {
		pos.Clip.Max.X = clip.Max.X
	}
	if pos.Clip.Max.Y > clip.Max.Y {
		pos.Clip.Max.Y = clip.Max.Y
	}
	return pos
}

func (pos Position) MouseArea() image.Rectangle {
	r := pos.GlyphArea
	if pos.Clip.Min.X > r.Min.X {
		r.Min.X = pos.Clip.Min.X
	}
	if pos.Clip.Min.Y > r.Min.Y {
		r.Min.Y = pos.Clip.Min.Y
	}
	if pos.Clip.Max.X < r.Max.X {
		r.Max.X = pos.Clip.Max.X
	}
	if pos.Clip.Max.Y < r.Max.Y {
		r.Max.Y = pos.Clip.Max.Y
	}
	return r
}

func (pos Position) Inset(min image.Point, max image.Point) Position {
	pos.GlyphArea.Min = pos.GlyphArea.Min.Add(min)
	pos.GlyphArea.Max = pos.GlyphArea.Max.Sub(max)
	return pos
}

type Appearance struct {
	GlyphSet  GlyphSetIndex
	GlyphName string
	ForeColor mgl32.Vec4
	BackColor mgl32.Vec4
	FgMask    MaskIndex
	BgMask    MaskIndex
	Edges     image.Rectangle
}

func (pl *Palette) Draw(dc *vmodel.DrawContext, position Position, appearance Appearance) bool {
	gs := pl.GetSet(appearance.GlyphSet)
	if gs == nil {
		return false
	}
	gl := gs.Get(appearance.GlyphName)
	if len(gl.Name) == 0 {
		return false
	}
	gp := dc.Pass.Get(dc.Cache.Ctx, kGlyphPipeline, func(ctx vk.APIContext) interface{} {
		return newPipeline(ctx, dc)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(dc.Cache)

	scx := 2 / float32(position.ImageSize.X)
	scy := 2 / float32(position.ImageSize.Y)
	scMat := mgl32.Translate2D(-1, -1).Mul3(mgl32.Scale2D(scx, scy))
	clMin := scMat.Mul3x1(point2Vec3(position.Clip.Min))
	clMax := scMat.Mul3x1(point2Vec3(position.Clip.Max))
	clip := mgl32.Vec4{clMin[0], clMin[1], clMax[0], clMax[1]}
	for edge := 0; edge < 9; edge++ {
		gi := glyphInstance{
			forecolor: appearance.ForeColor,
			backcolor: appearance.BackColor,
			clip:      clip,
			// position_1: di.Position.Row(0).Vec4(0),
			// position_2: di.Position.Row(1).Vec4(0),
			// uvGlyph_1:  di.UVGlyph.Row(0).Vec4(float32(di.GS)),
			// uvGlyph_2:  di.UVGlyph.Row(1).Vec4(0),
			// uvMask_1:   di.UVMask.Row(0).Vec4(float32(di.FgMask)),
			// uvMask_2:   di.UVMask.Row(0).Vec4(float32(di.BgMask)),
		}
		ok := pl.drawEdge(edge, &gi, gs, gl, scMat, position.GlyphArea, appearance.Edges, position.Rotate)
		if !ok {
			continue
		}
		gi.uvGlyph_1[3] = float32(appearance.GlyphSet)
		gi.uvGlyph_2[3] = float32(gs.kind)
		gi.uvMask_1[3] = float32(appearance.FgMask)
		gi.uvMask_2[3] = float32(appearance.BgMask)
		pl.drawInstance(dc, uc, gp, gi)
	}
	return true
}

func (pl *Palette) drawInstance(dc *vmodel.DrawContext, uc *vscene.UniformCache, gp *vk.GraphicsPipeline, gi glyphInstance) {
	gis := dc.Cache.GetPerFrame(kGlyphInstances, func(ctx vk.APIContext) interface{} {
		ds, sl := uc.Alloc(ctx)
		item := dc.Draw(gp, 0, 6).AddDescriptors(ds, pl.ds)
		return &glyphInstances{ds: ds, sl: sl, di: item}
	}).(*glyphInstances)
	lInst := uint32(unsafe.Sizeof(glyphInstance{}))
	b := *(*[unsafe.Sizeof(glyphInstance{})]byte)(unsafe.Pointer(&gi))
	copy(gis.sl.Content[gis.count*lInst:(gis.count+1)*lInst], b[:])
	gis.count++
	gis.di.SetInstances(0, gis.count)
	if gis.count >= maxInstances {
		dc.Cache.SetPerFrame(kGlyphInstances, nil)
	}
}

const NOMINALFontSize = 32

func (pl *Palette) DrawString(dc *vmodel.DrawContext, fontSize int, text string,
	position Position, appearance Appearance) bool {
	gs := pl.GetSet(appearance.GlyphSet)
	if gs == nil {
		return false
	}

	gp := dc.Pass.Get(dc.Cache.Ctx, kGlyphPipeline+vk.Key(gs.kind), func(ctx vk.APIContext) interface{} {
		return newPipeline(ctx, dc)
	}).(*vk.GraphicsPipeline)
	uc := vscene.GetUniformCache(dc.Cache)
	scx := 2 / float32(position.ImageSize.X)
	scy := 2 / float32(position.ImageSize.Y)
	scMat := mgl32.Translate2D(-1, -1).Mul3(mgl32.Scale2D(scx, scy))
	clMin := scMat.Mul3x1(point2Vec3(position.Clip.Min))
	clMax := scMat.Mul3x1(point2Vec3(position.Clip.Max))
	clip := mgl32.Vec4{clMin[0], clMin[1], clMax[0], clMax[1]}

	pos := float32(position.GlyphArea.Min.X)
	baseLine := float32(position.GlyphArea.Min.Y + fontSize*7/8)
	w := float32(0)
	prevChar := rune(0)
	for _, ch := range text {
		gl := gs.Get(string(ch))
		if prevChar != 0 {
			pos += w + gs.Advance(fontSize, prevChar, ch)
		}
		if len(gl.Name) > 0 {
			sz := gl.Location.Size()
			w = float32(fontSize*sz.X) / NOMINALFontSize
			h := float32(fontSize*sz.Y) / NOMINALFontSize
			uvLoc := gl.Location

			min := point2Uv(uvLoc.Min, gs.Desc.Width, gs.Desc.Height)
			max := point2Uv(uvLoc.Max, gs.Desc.Width, gs.Desc.Height)
			d := max.Sub(min)
			mUVGlyph := mgl32.Translate2D(min.X(), min.Y()).Mul3(mgl32.Scale2D(d.X(), d.Y()))

			pos += float32(fontSize*gl.CharOffset.X) / NOMINALFontSize
			yTop := baseLine + float32(fontSize*gl.CharOffset.Y)/NOMINALFontSize

			min = scMat.Mul3x1(mgl32.Vec3{pos, yTop, 1}).Vec2()
			max = scMat.Mul3x1(mgl32.Vec3{pos + w + 1, yTop + h + 0, 1}).Vec2()
			d = max.Sub(min)
			mPos := mgl32.Scale2D(d.X(), d.Y())
			if position.Rotate != 0 {
				mPos = mgl32.Rotate3DZ(position.Rotate)
			}
			mPos = mgl32.Translate2D(min.X(), min.Y()).Mul3(mPos)
			appearance.GlyphName = string(ch)
			gi := glyphInstance{
				forecolor:  appearance.ForeColor,
				backcolor:  appearance.BackColor,
				clip:       clip,
				position_1: mPos.Row(0).Vec4(0),
				position_2: mPos.Row(1).Vec4(0),
				uvGlyph_1:  mUVGlyph.Row(0).Vec4(float32(appearance.GlyphSet)),
				uvGlyph_2:  mUVGlyph.Row(1).Vec4(float32(gs.kind)),
				uvMask_1:   mgl32.Vec3{1, 0, 0}.Vec4(float32(appearance.FgMask)),
				uvMask_2:   mgl32.Vec3{0, 1, 0}.Vec4(float32(appearance.BgMask)),
			}
			pl.drawInstance(dc, uc, gp, gi)
		}
		prevChar = ch

	}
	return true
}

func (pl *Palette) MeasureString(gsIndex GlyphSetIndex, text string, fontHeight int) int {
	gs := pl.GetSet(gsIndex)
	if gs == nil {
		return 0
	}
	return gs.MeasureString(text, fontHeight)
}

func (th *Palette) drawEdge(edge int, gi *glyphInstance, gs *GlyphSet, gl Glyph, scMat mgl32.Mat3,
	area image.Rectangle, edges image.Rectangle, angle float32) bool {
	uvLoc := gl.Location
	bArea := area
	p0 := area.Min
	switch edge {
	case 0: // topleft
		if gl.Edges.Min.X > 0 && gl.Edges.Min.Y > 0 {
			if edges.Min.X == 0 || edges.Min.Y == 0 {
				return false
			}
			area.Max = area.Min.Add(edges.Min)
			uvLoc.Max = uvLoc.Min.Add(gl.Edges.Min)
		} else {
			return false
		}
	case 1: // top
		if gl.Edges.Min.Y > 0 {
			if edges.Min.Y == 0 {
				return false
			}
			area.Max.Y = area.Min.Y + edges.Min.Y
			uvLoc.Max.Y = uvLoc.Min.Y + gl.Edges.Min.Y
			if edges.Min.X > 0 && gl.Edges.Min.X > 0 {
				area.Min.X += edges.Min.X
				uvLoc.Min.X += gl.Edges.Min.X
			}
			if edges.Max.X > 0 && gl.Edges.Max.X > 0 {
				area.Max.X -= edges.Max.X
				uvLoc.Max.X -= gl.Edges.Max.X
			}
		} else {
			return false
		}
	case 2: // topright
		if gl.Edges.Max.X > 0 && gl.Edges.Min.Y > 0 {
			if edges.Max.X == 0 || edges.Min.Y == 0 {
				return false
			}
			area.Min.X = area.Max.X - edges.Max.X
			area.Max.Y = area.Min.Y + edges.Min.Y
			uvLoc.Min.X = uvLoc.Max.X - gl.Edges.Max.X
			uvLoc.Max.Y = uvLoc.Min.Y + gl.Edges.Min.Y
		} else {
			return false
		}
	case 3: // Left
		if gl.Edges.Min.X > 0 {
			if edges.Min.X == 0 {
				return false
			}
			area.Max.X = area.Min.X + edges.Min.X
			uvLoc.Max.X = uvLoc.Min.X + gl.Edges.Min.X
			if edges.Min.Y > 0 && gl.Edges.Min.Y > 0 {
				area.Min.Y += edges.Min.Y
				uvLoc.Min.Y += gl.Edges.Min.Y
			}
			if edges.Max.Y > 0 && gl.Edges.Max.Y > 0 {
				area.Max.Y -= edges.Max.Y
				uvLoc.Max.Y -= gl.Edges.Max.Y
			}
		} else {
			return false
		}
	case 4:
		if gl.Edges.Min.X > 0 {
			area.Min.X += edges.Min.X
		}
		if gl.Edges.Min.Y > 0 {
			area.Min.Y += edges.Min.Y
		}
		uvLoc.Min = uvLoc.Min.Add(gl.Edges.Min)
		if gl.Edges.Max.X > 0 {
			area.Max.X -= edges.Max.X
		}
		if gl.Edges.Max.Y > 0 {
			area.Max.Y -= edges.Max.Y
		}
		uvLoc.Max = uvLoc.Max.Sub(gl.Edges.Max)

	case 5: // right
		if gl.Edges.Max.X > 0 {
			if edges.Max.X == 0 {
				return false
			}
			area.Min.X = area.Max.X - edges.Max.X
			uvLoc.Min.X = uvLoc.Max.X - gl.Edges.Max.X
			if edges.Min.Y > 0 && gl.Edges.Min.Y > 0 {
				area.Min.Y += edges.Min.Y
				uvLoc.Min.Y += gl.Edges.Min.Y
			}
			if edges.Max.Y > 0 && gl.Edges.Max.Y > 0 {
				area.Max.Y -= edges.Max.Y
				uvLoc.Max.Y -= gl.Edges.Max.Y
			}
		} else {
			return false
		}
	case 6: // bottomleft
		if gl.Edges.Min.X > 0 && gl.Edges.Max.Y > 0 {
			if edges.Min.X == 0 || edges.Max.Y == 0 {
				return false
			}
			area.Min.Y = area.Max.Y - edges.Max.Y
			area.Max.X = area.Min.X + edges.Min.Y
			uvLoc.Min.Y = uvLoc.Max.Y - gl.Edges.Max.Y
			uvLoc.Max.X = uvLoc.Min.X + gl.Edges.Min.X
		} else {
			return false
		}
	case 7: // bottom
		if gl.Edges.Max.Y > 0 {
			if edges.Max.Y == 0 {
				return false
			}
			area.Min.Y = area.Max.Y - edges.Max.Y
			uvLoc.Min.Y = uvLoc.Max.Y - gl.Edges.Max.Y
			if edges.Min.X > 0 && gl.Edges.Min.X > 0 {
				area.Min.X += edges.Min.X
				uvLoc.Min.X += gl.Edges.Min.X
			}
			if edges.Max.X > 0 && gl.Edges.Max.X > 0 {
				area.Max.X -= edges.Max.X
				uvLoc.Max.X -= gl.Edges.Max.X
			}
		} else {
			return false
		}
	case 8: // topleft
		if gl.Edges.Max.X > 0 && gl.Edges.Max.Y > 0 {
			if edges.Max.X == 0 || edges.Max.Y == 0 {
				return false
			}
			area.Min = area.Max.Sub(edges.Max)
			uvLoc.Min = uvLoc.Max.Sub(gl.Edges.Max)
		} else {
			return false
		}
	default:
		return false
	}
	min := point2Uv(uvLoc.Min, gs.Desc.Width, gs.Desc.Height)
	max := point2Uv(uvLoc.Max, gs.Desc.Width, gs.Desc.Height)
	d := max.Sub(min)
	mUVGlyph := mgl32.Translate2D(min.X(), min.Y()).Mul3(mgl32.Scale2D(d.X(), d.Y()))
	gi.uvGlyph_1 = mUVGlyph.Row(0).Vec4(0)
	gi.uvGlyph_2 = mUVGlyph.Row(1).Vec4(0)
	min = scMat.Mul3x1(point2Vec3(area.Min)).Vec2()
	max = scMat.Mul3x1(point2Vec3(area.Max)).Vec2()
	p0min := scMat.Mul3x1(point2Vec3(p0)).Vec2()
	d = max.Sub(min)
	mPos := mgl32.Scale2D(d.X(), d.Y())
	if angle != 0 {
		pd := min.Sub(p0min)
		rot := mgl32.Rotate3DZ(angle).Mul3(mgl32.Translate2D(pd.X(), pd.Y())).Mul3(mPos)
		mPos = mgl32.Translate2D(-pd.X(), -pd.Y()).Mul3(rot)
	}
	mPos = mgl32.Translate2D(min.X(), min.Y()).Mul3(mPos)
	gi.position_1 = mPos.Row(0).Vec4(0)
	gi.position_2 = mPos.Row(1).Vec4(0)
	uPos := mgl32.Translate2D(float32(area.Min.X-bArea.Min.X)/float32(bArea.Size().X),
		float32(area.Min.Y-bArea.Min.Y)/float32(bArea.Size().Y))
	uPos = uPos.Mul3(mgl32.Scale2D(float32(area.Max.X-area.Min.X)/float32(bArea.Max.X-bArea.Min.X),
		float32(area.Max.Y-area.Min.Y)/float32(bArea.Max.Y-bArea.Min.Y)))
	gi.uvMask_1 = uPos.Row(0).Vec4(0)
	gi.uvMask_2 = uPos.Row(1).Vec4(0)
	return true
}

const maxInstances = 100

var kGlyphPipeline = vk.NewKey()
var kGlyphInstances = vk.NewKey()

type glyphInstances struct {
	sl    *vk.Slice
	ds    *vk.DescriptorSet
	di    *vk.DrawItem
	count uint32
}

type glyphInstance struct {
	forecolor  mgl32.Vec4
	backcolor  mgl32.Vec4
	clip       mgl32.Vec4 // l, t, r, b
	position_1 mgl32.Vec4 //
	position_2 mgl32.Vec4
	uvGlyph_1  mgl32.Vec4 // w = glyph index
	uvGlyph_2  mgl32.Vec4
	uvMask_1   mgl32.Vec4 // w = foreground mask
	uvMask_2   mgl32.Vec4 // w = bgMask
}

func newPipeline(ctx vk.APIContext, dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	rc := dc.Cache
	laTheme := getThemeLayout(ctx, rc.Device)
	la := vscene.GetUniformLayout(ctx, rc.Device)
	gp := vk.NewGraphicsPipeline(ctx, rc.Device)
	gp.AddLayout(ctx, la)
	gp.AddLayout(ctx, laTheme)
	gp.AddShader(ctx, vk.SHADERStageVertexBit, glyph_vert_spv)
	gp.AddShader(ctx, vk.SHADERStageFragmentBit, glyph_frag_spv)
	gp.AddAlphaBlend(ctx)
	gp.Create(ctx, dc.Pass)
	return gp
}

func point2Uv(point image.Point, width uint32, height uint32) mgl32.Vec2 {
	return mgl32.Vec2{
		float32(point.X)/float32(width) + 0.5/float32(width),
		float32(point.Y)/float32(height) + 0.5/float32(height),
	}
}

func point2Vec3(pt image.Point) mgl32.Vec3 {
	return mgl32.Vec3{float32(pt.X), float32(pt.Y), 1}
}
