package vdebug

import (
	"fmt"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
	"image"
	"time"
)

func NewFPSTimer(rw *vapp.RenderWindow, theme vui.Theme) *FPSTimer {
	fp := &FPSTimer{rw: rw}
	fr, ok := rw.GetRenderer().(*vapp.ForwardRenderer)
	if ok {
		fr.RenderDone = fp.renderDone
		fp.uiView = vui.NewUIView(theme, fp.getArea(), rw)
		fp.uiView.MainCtrl = vui.NewLabel("FPS").AssignTo(&fp.lFPS)
		rw.Scene.Update(func() {
			rw.Env.Children = append(rw.Env.Children, vscene.NewNode(fp.uiView))
		})
	}
	return fp
}

type FPSTimer struct {
	rw      *vapp.RenderWindow
	count   int
	total   float64
	visible bool
	lFPS    *vui.Label
	uiView  *vui.UIView
}

func (t *FPSTimer) renderDone(started time.Time) {
	totalSec := time.Now().Sub(started).Seconds()
	t.total += totalSec
	t.count++
	if t.count > 50 {
		t.rw.Scene.Update(func() {
			if !t.visible {
				t.uiView.ShowInactive()
				t.visible = true
			}
			ft := t.total / float64(t.count) * 1000.0
			t.lFPS.Text = fmt.Sprintf("Average render time %.2f ms (FPS %.1f)", ft, 1000/ft)
			t.uiView.Area = t.getArea()
		})
	}
}

func (t *FPSTimer) getArea() image.Rectangle {
	ws := t.rw.WindowSize
	return image.Rect(ws.X-350, 10, ws.X-10, 40)
}
