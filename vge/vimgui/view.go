package vimgui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vk"
	"sync"
	"time"
)

type ViewMode int

const (
	VMTransparent ViewMode = 0
	VMNormal      ViewMode = 1
	VMDialog      ViewMode = 2
	VMPopup       ViewMode = 3
)

type View struct {
	OnSize  func(fi *vk.FrameInstance) vdraw.Area
	OnClose func()

	dev       *vk.Device
	c         *vdraw.Canvas
	nextFrame UIFrame
	mx        *sync.Mutex
	painter   func(fr *UIFrame)
	started   time.Time
	mode      ViewMode
}

func NewView(dev *vk.Device, mode ViewMode, th *Theme, painter func(fr *UIFrame)) *View {
	f := &View{dev: dev, painter: painter, mx: &sync.Mutex{}, started: time.Now()}
	f.c = vdraw.NewCanvas(dev)
	f.mode = mode
	f.nextFrame.theme = th
	f.nextFrame.states = make(map[string]*vk.State)
	return f
}

func (v *View) HandleEvent(event vapp.Event) {
	se, ok := event.(vapp.SourcedEvent)
	if !ok {
		return
	}
	l := mgl32.Vec2{float32(se.Location().X), float32(se.Location().Y)}
	switch v.mode {
	case VMTransparent:
		if !v.nextFrame.winArea.Contains(l) {
			return
		}
	case VMNormal:
		if !v.nextFrame.winArea.Contains(l) {
			return
		}
		se.SetHandled()
	case VMDialog:
		se.SetHandled()
	case VMPopup:
		se.SetHandled()
		if !v.nextFrame.winArea.Contains(l) {
			_, ok = event.(*vapp.MouseDownEvent)
			if ok && v.OnClose != nil {
				v.OnClose()
			}
		}
	}
	v.mx.Lock()
	defer v.mx.Unlock()
	v.nextFrame.handleEvent(event)
}

func (v *View) Reserve(fi *vk.FrameInstance) {
	if v.nextFrame.TotalTime == 0 {
		v.nextFrame.TotalTime, v.nextFrame.DeltaTime = time.Now().Sub(v.started).Seconds(), 0
	} else {
		tn := time.Now().Sub(v.started).Seconds()
		v.nextFrame.TotalTime, v.nextFrame.DeltaTime = tn, tn-v.nextFrame.DeltaTime
	}
	v.c.Reserve(fi)
}

func (v *View) PreRender(fi *vk.FrameInstance) (done vapp.Completed) {
	return nil
}

func (v *View) Render(fi *vk.FrameInstance, cmd *vk.Command, rp *vk.GeneralRenderPass) {
	uif := v.beginDraw(fi)
	if uif == nil {
		return
	}
	dl := &vk.DrawList{}
	outDesc := fi.Output.Describe()
	cp := v.c.BeginDraw(fi, rp, dl, v.c.Projection(uif.winArea.From, mgl32.Vec2{float32(outDesc.Width), float32(outDesc.Height)}))
	defer v.endDraw(uif)
	cp.Clip = uif.DrawArea
	uif.cp = cp
	v.painter(uif)
	uif.cp.End()
	cmd.Draw(dl)
}

func (f *View) beginDraw(fi *vk.FrameInstance) *UIFrame {
	f.mx.Lock()
	defer f.mx.Unlock()
	oDesc := fi.Output.Describe()
	if f.OnSize != nil {
		f.nextFrame.winArea = f.OnSize(fi)
	} else {
		f.nextFrame.winArea = vdraw.Area{To: mgl32.Vec2{float32(oDesc.Width), float32(oDesc.Height)}}
	}
	f.nextFrame.dev = f.dev
	uif := f.nextFrame
	f.nextFrame.Ev = UIEvent{}
	uif.Offset = mgl32.Vec2{0, 0}
	uif.DrawArea.From = mgl32.Vec2{0, 0}
	uif.DrawArea.To = mgl32.Vec2{uif.winArea.Width(), uif.winArea.Height()}
	return &uif
}

func (f *View) endDraw(uif *UIFrame) {
	f.mx.Lock()
	defer f.mx.Unlock()
	f.nextFrame.states = uif.states
	f.nextFrame.focusCtrl = uif.focusCtrl
}
