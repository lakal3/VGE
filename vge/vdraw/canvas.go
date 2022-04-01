package vdraw

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vscene"
	font2 "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"math"
	"unsafe"
)

// Canvas is used to create CanvasPainter that can fill drawable shapes to output of current render pass
type Canvas struct {
	dev            *vk.Device
	vbSize, sbSize uint32
	key            vk.Key
	la             *vk.DescriptorLayout
	glyphSampler   *vk.Sampler
}

// CanvasPainter draws Drawables to output of current render pass
type CanvasPainter struct {
	Clip Area

	ca *Canvas
	rp *vk.GeneralRenderPass
	fi *vk.FrameInstance

	imageCount   uint32
	vextexCount  uint32
	segmentCount uint32
	maxVertex    uint32
	maxSegment   uint32
	vertexBuf    *vk.ASlice
	segmentBuf   *vk.ASlice
	// instanceCount uint32
	drawList   *vk.DrawList
	offsets    map[Drawable]uint32
	dsWorld    *vk.DescriptorSet
	dsSegments *vk.DescriptorSet
	images     map[vk.VImageView]uint32
}

func NewCanvas(dev *vk.Device) *Canvas {
	c := &Canvas{dev: dev, vbSize: 16384, sbSize: 16384, key: vk.NewKey()}
	c.buildDsLayout(dev)
	c.buildSampler(dev)
	return c
}

func (c *Canvas) ReportError(err error) {
	c.dev.ReportError(err)
}

// Projection is helper to calculate world projection for canvas that starts at point at. worldSize should be size of current output
// Both dimensions should be in pixels
func (c *Canvas) Projection(at mgl32.Vec2, worldSize mgl32.Vec2) mgl32.Mat4 {
	return mgl32.Translate3D(-1, -1, 0.1).
		Mul4(mgl32.Scale3D(2/worldSize[0], 2/worldSize[1], 1)).
		Mul4(mgl32.Translate3D(at[0], at[1], 0))
}

// Reserve is called during reserve phase of rendering. Canvas will reserve enough buffer / uniform / vertex storage to emit output
// If output changes a lot, canvas may reserve too little space for rendering. In that cases canvas will omit some output and adjust
// buffers for next frame
func (c *Canvas) Reserve(fi *vk.FrameInstance) {
	cf := &CanvasPainter{maxVertex: c.vbSize, maxSegment: c.sbSize, offsets: make(map[Drawable]uint32),
		images: make(map[vk.VImageView]uint32)}
	fi.ReserveSlice(vk.BUFFERUsageStorageBufferBit, uint64(cf.maxSegment)*uint64(unsafe.Sizeof(mgl32.Vec4{})))
	fi.ReserveSlice(vk.BUFFERUsageVertexBufferBit, uint64(cf.maxVertex)*uint64(unsafe.Sizeof(vertexData{})))
	fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, 1024)
	fi.ReserveDescriptor(c.la)
	fi.ReserveDescriptor(vscene.GetUniformLayout(c.dev))
	fi.Set(c.key, cf)
}

// BeginDraw start actual recording of drawing instructions to drawing list. Called of beginDraw must ensure that drawList (dl) is
// recorded to actual command buffer
func (c *Canvas) BeginDraw(fi *vk.FrameInstance, rp *vk.GeneralRenderPass, dl *vk.DrawList, projection mgl32.Mat4) *CanvasPainter {
	cp := fi.Get(c.key, nil).(*CanvasPainter)
	cp.ca, cp.rp = c, rp
	cp.drawList = dl
	cp.imageCount = 1
	sl := fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, 1024)
	copy(sl.Bytes(), vk.Float32ToBytes(projection[:]))
	cp.dsWorld = fi.AllocDescriptor(vscene.GetUniformLayout(c.dev))
	cp.dsWorld.WriteSlice(0, 0, sl)
	cp.dsSegments = fi.AllocDescriptor(c.la)
	cp.segmentBuf = fi.AllocSlice(vk.BUFFERUsageStorageBufferBit, uint64(cp.maxSegment)*uint64(unsafe.Sizeof(mgl32.Vec4{})))
	cp.vertexBuf = fi.AllocSlice(vk.BUFFERUsageVertexBufferBit, uint64(cp.maxVertex)*uint64(unsafe.Sizeof(vertexData{})))
	cp.dsSegments.WriteSlice(0, 0, cp.segmentBuf)
	return cp
}

// Draw a single drawable to drawing list given in BeginDraw. All shapes can be moved and scaled from their initial position
func (cp *CanvasPainter) Draw(dr Drawable, at mgl32.Vec2, scale mgl32.Vec2, br *Brush) {
	segOffset, ok := cp.offsets[dr]
	areas, segments := dr.GetDrawData()
	vertexCount := uint32(len(areas) * 2)
	segCount := uint32(len(segments))
	vertexOffset := cp.vextexCount
	cp.vextexCount += vertexCount
	if !ok {
		segOffset = cp.segmentCount
		cp.segmentCount += segCount
		cp.offsets[dr] = segOffset
	}
	if cp.segmentCount > cp.maxSegment || cp.vextexCount > cp.maxVertex {
		return // Buffer too short
	}
	if !ok && segCount > 0 {
		segSlice := unsafe.Slice((*DrawSegment)(unsafe.Pointer(&cp.segmentBuf.Bytes()[0])), cp.maxSegment)
		copy(segSlice[segOffset:], segments)
	}
	ptr, instOffset := cp.drawList.AllocPushConstants(szInstance)
	ci := (*canvasInstance)(ptr)
	// TODO: Get texture index
	txIndex := uint32(0) // c.addTexture(br)
	if br.Image != nil {
		txIndex = cp.addImage(br.Image, br.Sampler)
	}
	ci.uv1 = br.UVTransform.Row(0).Vec4(float32(txIndex))
	ci.uv2 = br.UVTransform.Row(1).Vec4(0)
	ci.scale = scale
	ci.color = br.Color
	ci.color2 = br.Color2
	ci.clip = cp.Clip.ToVec4()
	glImage, glLayer := dr.GetGlyph()
	if glImage != nil {
		imIdex := cp.addImage(glImage, cp.ca.glyphSampler)
		ci.glyph = mgl32.Vec2{float32(imIdex), float32(glLayer)}
	} else {
		ci.glyph = mgl32.Vec2{0, 0}
	}
	var pl *vk.GraphicsPipeline
	shape := dr.GetShape()
	if shape == nil {
		pl = cp.buildPipeline(glImage != nil)
	} else {
		pl = cp.buildShapePipeline(shape)
	}
	verSlice := unsafe.Slice((*vertexData)(unsafe.Pointer(&cp.vertexBuf.Bytes()[0])), cp.maxVertex)
	m3 := mgl32.Translate2D(at[0], at[1]).Mul3(mgl32.Scale2D(scale[0], scale[1]))
	for idx, area := range areas {
		vd := &verSlice[vertexOffset+uint32(2*idx)]
		vd.origCorner = area.Min.Sub(mgl32.Vec2{2 / scale[0], 0})
		vd.corner = m3.Mul3x1(vd.origCorner.Vec3(1)).Vec2()
		vd.segments = mgl32.Vec2{float32(area.From + segOffset), float32(area.To + segOffset)}
		vd = &verSlice[vertexOffset+uint32(2*idx)+1]
		vd.origCorner = area.Max.Add(mgl32.Vec2{2 / scale[0], 0})
		vd.corner = m3.Mul3x1(vd.origCorner.Vec3(1)).Vec2()
		vd.segments = mgl32.Vec2{float32(area.From + segOffset), float32(area.To + segOffset)}
	}
	cp.drawList.Draw(pl, vertexOffset, vertexCount).AddInput(0, cp.vertexBuf).
		AddDescriptors(cp.dsWorld, cp.dsSegments).AddPushConstants(szInstance, instOffset)
}

// End indicates that we will not draw anything else using this CanvasPainter. This call may update buffer requirements
func (cp *CanvasPainter) End() {
	if cp.ca.vbSize < cp.vextexCount {
		cp.ca.vbSize = cp.vextexCount * 3 / 2
	}
	if cp.ca.sbSize < cp.segmentCount {
		cp.ca.sbSize = cp.segmentCount * 3 / 2
	}
}

// DrawText draws text with current color. Draw text only support horizontal left to right text. Yse DrawTextWith for
// more control of filling and movement between characters
func (cp *CanvasPainter) DrawText(font *Font, size float32, at mgl32.Vec2, br *Brush, text string) {
	scale := mgl32.Vec2{size / 64, size / 64}
	cp.DrawTextWith(font, size, at, text, func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2) {
		return at.Add(mgl32.Vec2{advance, 0})
	}, func(idx int, at mgl32.Vec2, ch Drawable) {
		cp.Draw(ch, at, scale, br)
	})
}

// DrawTextWith draws text one rune at a time. DrawTextWidth will call advance function to move position for next characters.
// Advance must also fill recorded path for character.
func (cp *CanvasPainter) DrawTextWith(font *Font, size float32, at mgl32.Vec2, text string,
	advance func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2),
	draw func(idx int, at mgl32.Vec2, ch Drawable)) {
	var iPrev sfnt.GlyphIndex

	for idx, r := range text {

		iGl, err := font.sf.GlyphIndex(font.buf, r)
		if err != nil {
			cp.ca.ReportError(err)
			return
		}
		if iGl == 0 {
			at = advance(idx, at, size/4, false)
			iPrev = 0
			continue
		}
		var defKern fixed.Int26_6
		if iPrev > 0 {
			defKern, err = font.sf.Kern(font.buf, iPrev, iGl, toFixed(size), font2.HintingFull)
			if err == nil {
				at = advance(idx, at, DefaultKern+float32(defKern)/FontStrokeSize, true)
			} else {
				at = advance(idx, at, DefaultKern, true)
			}
		}
		defAdvance, err := font.sf.GlyphAdvance(font.buf, iGl, toFixed(size), font2.HintingFull)
		if err != nil {
			cp.ca.ReportError(err)
			return
		}
		chGlyph := font.GetGlyph(r)
		if chGlyph != nil {
			draw(idx, at, chGlyph)
		} else {
			ch, err := font.GetFilled(r)
			if err != nil {
				cp.ca.ReportError(err)
				return
			}
			if ch != nil {
				draw(idx, at, ch)
			}
		}
		at = advance(idx, at, float32(math.Ceil(float64(defAdvance)/FontStrokeSize)), false)
		iPrev = iGl
	}
}

func (cp *CanvasPainter) addImage(image vk.VImageView, sampler *vk.Sampler) uint32 {

	imgIndex, ok := cp.images[image]
	if ok {
		return imgIndex
	}
	if cp.imageCount >= vscene.FrameMaxDynamicSamplers {
		return 0
	}
	imgIndex = cp.imageCount
	cp.imageCount++
	cp.images[image] = imgIndex
	cp.dsSegments.WriteView(1, imgIndex, image, vk.IMAGELayoutShaderReadOnlyOptimal, sampler)
	return imgIndex
}

func (cp *CanvasPainter) buildPipeline(glyph bool) *vk.GraphicsPipeline {
	dev := cp.ca.dev

	ul := vscene.GetUniformLayout(dev)
	imLayout := cp.ca.la

	k := kPipeline
	if glyph {
		k++
	}
	return cp.rp.Get(k, func() interface{} {
		pl := vk.NewGraphicsPipeline(dev)
		pl.AddAlphaBlend()
		// position (2),  original position (2), segments
		pl.AddVextexInput(vk.VERTEXInputRateVertex, vk.FORMATR32g32Sfloat, vk.FORMATR32g32Sfloat, vk.FORMATR32g32Sfloat)
		pl.AddLayout(ul)
		pl.AddLayout(imLayout)
		pl.AddShader(vk.SHADERStageGeometryBit, draw_geom_spv)
		pl.AddShader(vk.SHADERStageVertexBit, draw_vert_spv)
		if glyph {
			pl.AddShader(vk.SHADERStageFragmentBit, draw_glyph_frag_spv)
		} else {
			pl.AddShader(vk.SHADERStageFragmentBit, draw_frag_spv)
		}
		pl.SetTopology(vk.PRIMITIVETopologyLineList)
		pl.AddPushConstants(vk.SHADERStageVertexBit|vk.SHADERStageGeometryBit|vk.SHADERStageFragmentBit, szInstance)
		pl.Create(cp.rp)
		return pl
	}).(*vk.GraphicsPipeline)
}

func (cp *CanvasPainter) buildShapePipeline(shape *Shape) *vk.GraphicsPipeline {
	dev := cp.ca.dev

	ul := vscene.GetUniformLayout(dev)
	imLayout := cp.ca.la

	if len(shape.spirv) == 0 {
		dev.FatalError(errors.New("Shapes not compiled"))
	}
	return cp.rp.Get(shape.key, func() interface{} {
		pl := vk.NewGraphicsPipeline(dev)
		pl.AddAlphaBlend()
		// position (2),  original position (2), segments
		pl.AddVextexInput(vk.VERTEXInputRateVertex, vk.FORMATR32g32Sfloat, vk.FORMATR32g32Sfloat, vk.FORMATR32g32Sfloat)
		pl.AddLayout(ul)
		pl.AddLayout(imLayout)
		pl.AddShader(vk.SHADERStageGeometryBit, draw_geom_spv)
		pl.AddShader(vk.SHADERStageVertexBit, draw_vert_spv)
		pl.AddShader(vk.SHADERStageFragmentBit, shape.spirv)
		pl.SetTopology(vk.PRIMITIVETopologyLineList)
		pl.AddPushConstants(vk.SHADERStageVertexBit|vk.SHADERStageGeometryBit|vk.SHADERStageFragmentBit, szInstance)
		pl.Create(cp.rp)
		return pl
	}).(*vk.GraphicsPipeline)
}

func (c *Canvas) buildDsLayout(dev *vk.Device) {
	la := dev.Get(kLayout, func() interface{} {
		la1 := dev.NewDescriptorLayout(vk.DESCRIPTORTypeStorageBuffer, vk.SHADERStageFragmentBit, 1)
		return la1.AddDynamicBinding(vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit,
			vscene.FrameMaxDynamicSamplers, vk.DESCRIPTORBindingPartiallyBoundBitExt|vk.DESCRIPTORBindingUpdateAfterBindBitExt)
	}).(*vk.DescriptorLayout)
	c.la = la
}

func (c *Canvas) buildSampler(dev *vk.Device) {
	sampler := dev.Get(kSampler, func() interface{} {
		return dev.NewSampler(vk.SAMPLERAddressModeClampToEdge)
	}).(*vk.Sampler)
	c.glyphSampler = sampler
}

var kLayout = vk.NewKey()
var kSampler = vk.NewKey()
var kPipeline = vk.NewKeys(2)

type vertexData struct {
	corner     mgl32.Vec2
	origCorner mgl32.Vec2
	segments   mgl32.Vec2
}

const szInstance = uint32(unsafe.Sizeof(canvasInstance{}))

type canvasInstance struct {
	clip   mgl32.Vec4 // Clip in world coordinates
	color  mgl32.Vec4 // Color
	color2 mgl32.Vec4
	uv1    mgl32.Vec4 // first row of uv matrix from pos, w - image index
	uv2    mgl32.Vec4
	scale  mgl32.Vec2
	glyph  mgl32.Vec2
}
