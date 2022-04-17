package vdraw3d

import (
	"bytes"
	_ "embed"
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"sort"
	"time"
	"unsafe"
)

type Camera interface {
	CameraProjection(size image.Point) (projection, view mgl32.Mat4)
}

type ActiveCamera interface {
	Camera
	Handle(ev vapp.Event)
}

type View struct {
	OnSize  func(fi *vk.FrameInstance) vdraw.Area
	OnEvent func(ev vapp.Event)
	Elapsed float64
	Camera  Camera

	debug    uint32
	prevDraw time.Time
	dev      *vk.Device
	// 0 - output image, 1 - depth buffer, 2 - frame buffer, 3 - command, 4 - blend render pass
	key          vk.Key
	shaders      *shaders.Pack
	nextRings    []vk.Key
	ambient      mgl32.Vec4
	ambienty     mgl32.Vec4
	staticScene  func(v *View, dl *FreezeList)
	dynamicScene func(v *View, dl *FreezeList)
	staticList   *FreezeList
	area         vdraw.Area
}

var StaticRing = vk.NewKey()

const (
	StaticBase      FrozenID = 1_000_000
	vkiImage2                = 1
	vkiDepth                 = 2
	vkDepthFrame             = 2
	vkColorFrame1            = 4
	vkColorFrame2            = 5
	vkBlendPipeline          = 6
	storageSize              = 16 * 4
)

func LoadDefault() (*shaders.Pack, error) {
	sp := &shaders.Pack{}
	err := sp.Load(bytes.NewReader(stdshaders_bin))
	if err != nil {
		return nil, err
	}
	return sp, nil
}

func NewView(dev *vk.Device, staticScene func(v *View, dl *FreezeList), dynamicScene func(v *View, dl *FreezeList)) *View {
	sp, err := LoadDefault()
	if err != nil {
		dev.FatalError(err)
	}
	return NewCustomView(dev, sp, staticScene, dynamicScene)
}

func NewCustomView(dev *vk.Device, sp *shaders.Pack, staticScene func(v *View, dl *FreezeList), dynamicScene func(v *View, dl *FreezeList)) *View {
	v := &View{dev: dev, staticScene: staticScene, dynamicScene: dynamicScene, shaders: sp,
		ambient: mgl32.Vec4{0.4, 0.4, 0.4}, ambienty: mgl32.Vec4{0.2, 0.2, 0.2}}
	v.Camera = vapp.NewOrbitControl(0, nil)
	v.key = vk.NewKeys(7)
	return v
}

func (v *View) SetDebug(debugMode uint32) {
	v.debug = debugMode
}

func (v *View) SetAmbient(ambient mgl32.Vec4, ambienty mgl32.Vec4) {
	v.ambient, v.ambienty = ambient, ambienty
}

func (v *View) Reserve(fi *vk.FrameInstance) {
	var nextRings []vk.Key
	nextRings, v.nextRings = v.nextRings, []vk.Key{}
	tn := time.Now()
	if !v.prevDraw.IsZero() {
		v.Elapsed += tn.Sub(v.prevDraw).Seconds()
	}
	v.prevDraw = tn

	if v.staticList == nil {
		nextRings = append(nextRings, StaticRing)
		v.staticList = &FreezeList{BaseID: StaticBase}
		v.staticScene(v, v.staticList)
	} else {
		v.staticList.Clone()
	}
	dl := &FreezeList{}
	v.dynamicScene(v, dl)
	desc := fi.Output.Describe()
	cv := fi.Get(v.key, func() interface{} {
		cv := &currentView{fl: dl, imageIndex: 2}
		if v.OnSize != nil {
			cv.area = v.OnSize(fi)
			desc.Width = uint32(cv.area.Size()[0])
			desc.Height = uint32(cv.area.Size()[0])
		} else {
			cv.area.To = mgl32.Vec2{float32(desc.Width), float32(desc.Height)}
		}
		cv.views = make(map[vk.VImageView]uint32)
		return cv
	}).(*currentView)
	v.area = cv.area
	desc.Depth, desc.MipLevels, desc.Layers = 1, 1, 1
	desc.Format = vk.FORMATR8g8b8a8Unorm
	dDepth := desc
	dDepth.Format = vk.FORMATD32Sfloat
	fi.ReserveImage(v.key, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit, desc,
		desc.FullRange())
	fi.ReserveImage(v.key+vkiImage2, vk.IMAGEUsageColorAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit, desc,
		desc.FullRange())
	fi.ReserveImage(v.key+vkiDepth, vk.IMAGEUsageDepthStencilAttachmentBit|vk.IMAGEUsageTransferSrcBit|vk.IMAGEUsageSampledBit, dDepth,
		desc.FullRange())
	fi.ReserveDescriptor(GetFrameLayout(fi.Device()))
	offset := uint32(1)
	for _, ring := range nextRings {
		fi.BeginRing(ring)
	}
	for _, fr := range v.staticList.items {
		offset = fr.Reserve(fi, offset)
	}
	for _, fr := range dl.items {
		offset = fr.Reserve(fi, offset)
	}
	cv.storageSize = offset
	cv.size = image.Pt(int(desc.Width), int(desc.Height))
	fi.ReserveSlice(vk.BUFFERUsageStorageBufferBit, uint64(offset)*storageSize)
	fi.ReserveSlice(vk.BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(frame{})))

}

func (v *View) PreRender(fi *vk.FrameInstance) (done vapp.Completed) {
	cv := fi.Get(v.key, func() interface{} {
		fi.Device().FatalError(errors.New("No view instance"))
		return nil
	}).(*currentView)
	cv.dsFrame = fi.AllocDescriptor(GetFrameLayout(fi.Device()))
	fStorage := v.updateFrame(fi, cv)
	rm := RenderMaps{Render: Render{Target: Main, Shaders: v.shaders, DSFrame: cv.dsFrame},
		Static: v.staticList, Dynamic: cv.fl}
	var preSubmit []vk.SubmitInfo
	rm.AtEnd = func(end func() []vk.SubmitInfo) {
		preSubmit = append(preSubmit, end()...)
	}
	rm.UpdateStorage = func(storagePosition uint32, index uint32, values ...float32) {
		copy(fStorage[storagePosition*16+index:storagePosition*16+16], values)
	}
	v.staticList.RenderAll(fi, rm)
	cv.fl.RenderAll(fi, rm)
	cmd := fi.AllocCommand(vk.QUEUEGraphicsBit)
	cmd.Begin()
	cv.cmd = cmd
	v.renderImage(fi, cv)
	si := cmd.SubmitForWait(1, vk.PIPELINEStageEarlyFragmentTestsBit, preSubmit...)
	return func() []vk.SubmitInfo {
		return []vk.SubmitInfo{si}
	}
}

func (v *View) Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass) {
	cv := fi.Get(v.key, func() interface{} {
		fi.Device().FatalError(errors.New("No view instance"))
		return nil
	}).(*currentView)
	sampler := vmodel.GetDefaultSampler(fi.Device())
	cv.dsFrame.WriteView(2, 0, cv.outputView, vk.IMAGELayoutShaderReadOnlyOptimal, sampler)
	pl := v.getBlendPipeline(fi.Device(), rp)
	dl := &vk.DrawList{}
	dl.Draw(pl, 0, 6).AddDescriptor(0, cv.dsFrame)
	cmd.Draw(dl)
}

func (v *View) HandleEvent(event vapp.Event) {
	if v.OnEvent != nil {
		v.OnEvent(event)
		if event.Handled() {
			return
		}
	}
	se, ok := event.(vapp.SourcedEvent)
	if !ok {
		return
	}
	l := mgl32.Vec2{float32(se.Location().X), float32(se.Location().Y)}
	if v.area.Contains(l) {
		ac, ok := v.Camera.(ActiveCamera)
		if ok {
			ac.Handle(event)
		}
	}
}

func (v *View) updateFrame(fi *vk.FrameInstance, cv *currentView) []float32 {
	var fr frame
	fr.projection, fr.view = v.Camera.CameraProjection(cv.size)
	cv.view = fr.view
	fr.cameraPos = fr.view.Inv().Col(3)
	fr.ambient, fr.ambienty = v.ambient, v.ambienty
	desc := fi.Output.Describe()
	fr.viewPosition[0] = (cv.area.From[0]/float32(desc.Width))*2 - 1
	fr.viewPosition[1] = (cv.area.From[1]/float32(desc.Height))*2 - 1
	fr.viewPosition[2] = (cv.area.To[0]/float32(desc.Width))*2 - 1
	fr.viewPosition[3] = (cv.area.To[1]/float32(desc.Height))*2 - 1
	bf := buildFrame{cv: cv, fr: &fr, isStatic: true}
	bf.stBuf = fi.AllocSlice(vk.BUFFERUsageStorageBufferBit, uint64(cv.storageSize)*storageSize)
	bf.fBuf = unsafe.Slice((*float32)(unsafe.Pointer(&bf.stBuf.Bytes()[0])), cv.storageSize*16)
	v.staticList.RenderAll(fi, bf)
	bf.isStatic = false
	cv.fl.RenderAll(fi, bf)
	fr.lights = cv.lights
	fr.lightPos = cv.lightPos
	fr.decals, fr.decalPos = cv.decals, cv.decalPos
	fr.debug = v.debug
	ubBuf := fi.AllocSlice(vk.BUFFERUsageUniformBufferBit, uint64(unsafe.Sizeof(frame{})))
	*(*frame)(unsafe.Pointer(&ubBuf.Bytes()[0])) = fr
	cv.dsFrame.WriteSlice(0, 0, ubBuf)
	cv.dsFrame.WriteSlice(1, 0, bf.stBuf)
	return bf.fBuf
}

type trList struct {
	priority float32
	render   func(dl *vk.DrawList, pass *vk.GeneralRenderPass)
}

func (v *View) renderImage(fi *vk.FrameInstance, cv *currentView) {
	rpDepth := getDepthRenderPass(fi.Device())
	rp := getDrawRenderPass1(fi.Device())
	rp2 := getDrawRenderPass2(fi.Device())
	_, vOut := fi.AllocImage(v.key)
	_, vOut2 := fi.AllocImage(v.key + vkiImage2)
	_, vDepth := fi.AllocImage(v.key + vkiDepth)
	cv.outputView = vOut[0]
	cv.depthView = vDepth[0]
	fpDepth := fi.Get(v.key+vkDepthFrame, func() interface{} {
		return vk.NewFramebuffer2(rpDepth, vDepth[0])
	}).(*vk.Framebuffer)
	cmd := cv.cmd
	cmd.BeginRenderPass(rpDepth, fpDepth)
	renderCommon := Render{DSFrame: cv.dsFrame, Target: Main, Shaders: v.shaders}
	rd := RenderDepth{DL: &vk.DrawList{}, Pass: rpDepth, Render: renderCommon}
	v.staticList.RenderAll(fi, rd)
	cv.fl.RenderAll(fi, rd)
	cmd.Draw(rd.DL)
	cmd.EndRenderPass()
	fpColor1 := fi.Get(v.key+vkColorFrame1, func() interface{} {
		return vk.NewFramebuffer2(rp, vOut[0], vOut2[0], vDepth[0])
	}).(*vk.Framebuffer)

	cmd.BeginRenderPass(rp, fpColor1)
	var probe uint32
	var transparents []trList
	rc := RenderColor{DL: &vk.DrawList{}, Pass: rp, Render: renderCommon, Probe: &probe, ViewTransform: cv.view}
	rc.RenderTransparent = func(priority float32, render func(dl *vk.DrawList, pass *vk.GeneralRenderPass)) {
		transparents = append(transparents, trList{priority: priority, render: render})
	}
	v.staticList.RenderAll(fi, rc)
	cv.fl.RenderAll(fi, rc)
	cmd.Draw(rc.DL)
	cmd.EndRenderPass()
	if len(transparents) == 0 {
		return
	}
	sort.Slice(transparents, func(i, j int) bool {
		return transparents[i].priority > transparents[j].priority
	})
	sampler := vmodel.GetDefaultSampler(fi.Device())
	cv.dsFrame.WriteView(2, 0, vOut2[0], vk.IMAGELayoutShaderReadOnlyOptimal, sampler)
	cv.dsFrame.WriteView(2, 1, vDepth[0], vk.IMAGELayoutShaderReadOnlyOptimal, sampler)
	fpColor2 := fi.Get(v.key+vkColorFrame2, func() interface{} {
		return vk.NewFramebuffer2(rp2, vOut[0])
	}).(*vk.Framebuffer)
	dl2 := &vk.DrawList{}
	cmd.BeginRenderPass(rp2, fpColor2)
	for _, tr := range transparents {
		tr.render(dl2, rp2)
	}
	cmd.Draw(dl2)
	cmd.EndRenderPass()
}

func (v *View) getBlendPipeline(dev *vk.Device, rp *vk.GeneralRenderPass) *vk.GraphicsPipeline {
	return rp.Get(v.key+vkBlendPipeline, func() interface{} {
		pl := vk.NewGraphicsPipeline(dev)
		pl.AddLayout(GetFrameLayout(dev))
		code := v.shaders.MustGet(v.dev, "blend")
		pl.AddShader(vk.SHADERStageVertexBit, code.Vertex)
		pl.AddShader(vk.SHADERStageFragmentBit, code.Fragment)
		pl.Create(rp)
		return pl
	}).(*vk.GraphicsPipeline)
}

type currentView struct {
	area        vdraw.Area
	fl          *FreezeList
	imageIndex  uint32
	lightPos    uint32
	lights      uint32
	decalPos    uint32
	decals      uint32
	views       map[vk.VImageView]uint32
	dsFrame     *vk.DescriptorSet
	outputView  *vk.AImageView
	depthView   *vk.AImageView
	cmd         *vk.Command
	storageSize uint32
	usedStorage uint32
	size        image.Point
	view        mgl32.Mat4
}

type buildFrame struct {
	target   FrozenID
	cv       *currentView
	fr       *frame
	stBuf    *vk.ASlice
	fBuf     []float32
	isStatic bool
}

func (bf buildFrame) IsStatic() bool {
	return bf.isStatic
}

func (bf buildFrame) UpdateStorage(storagePosition uint32, index uint32, values ...float32) {
	copy(bf.fBuf[storagePosition*16+index:storagePosition*16+16], values)
}

func (bf buildFrame) AddLight(storagePosition uint32) (prev float32) {
	bf.cv.lights++
	prev, bf.cv.lightPos = float32(bf.cv.lightPos), storagePosition
	return
}

func (bf buildFrame) AddDecal(storagePosition uint32) (prev float32) {
	bf.cv.decals++
	prev, bf.cv.decalPos = float32(bf.cv.decalPos), storagePosition
	return
}

func (bf buildFrame) TargetID() FrozenID {
	return bf.target
}

func (bf buildFrame) AddView(view vk.VImageView, sampler *vk.Sampler) float32 {
	c := bf.cv
	val, ok := c.views[view]
	if ok {
		return float32(val)
	}
	if c.imageIndex < MaxImages {
		val = c.imageIndex
		c.views[view] = val
		bf.cv.dsFrame.WriteView(2, val, view, vk.IMAGELayoutShaderReadOnlyOptimal, sampler)
		c.imageIndex++
		return float32(val)
	}
	return 0
}

//go:embed stdshader.bin
var stdshaders_bin []byte