package vdebug

import (
	"fmt"
	"github.com/lakal3/vge/vge/forward"
	"image"
	"time"

	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vscene"
	"github.com/lakal3/vge/vge/vui"
)

// NewFPSTimer creates a simple UI to display FPS.
// FPS is calculated from start of scene to end of submit (CPU time).
// FPS times is attached to end of current scenes UI nodes.
func NewFPSTimer(rw *vapp.RenderWindow, theme vui.Theme) *FPSTimer {
	fp := &FPSTimer{rw: rw}
	fr, ok := rw.GetRenderer().(*forward.Renderer)
	if ok {
		fr.RenderDone = fp.renderDone
		fp.UIView = vui.NewUIView(theme, fp.getArea(), rw)
		fp.UIView.MainCtrl = vui.NewVStack(2, vui.NewLabel("FPS").AssignTo(&fp.lFPS),
			vui.NewLabel("").AssignTo(&fp.lGPU))
		rw.Scene.Update(func() {
			rw.Ui.Children = append(rw.Ui.Children, vscene.NewNode(fp.UIView))
		})
	}
	return fp
}

// AddGPUTiming adds support for timing from start of first submit in render phase to end of main command submit.
// GPU time should include all time used in rendering but not time used to manage swap chain
func (fp *FPSTimer) AddGPUTiming() {
	fr, ok := fp.rw.GetRenderer().(*forward.Renderer)
	if ok {
		fr.GPUTiming = func(times []float64) {
			var delta = (times[2] - times[0]) / 1e9
			fp.gpuCount++
			fp.gpuTotal += delta
		}
	}
}

type FPSTimer struct {
	rw       *vapp.RenderWindow
	count    int
	total    float64
	gpuCount int
	gpuTotal float64
	visible  bool
	lFPS     *vui.Label
	UIView   *vui.UIView
	lGPU     *vui.Label
}

func (t *FPSTimer) renderDone(started time.Time) {
	totalSec := time.Now().Sub(started).Seconds()
	t.total += totalSec
	t.count++
	if t.count > 50 {
		t.rw.Scene.Update(func() {
			if !t.visible {
				t.UIView.ShowInactive()
				t.visible = true
			}
			ft := t.total / float64(t.count) * 1000.0
			t.lFPS.Text = fmt.Sprintf("Render time %.2f ms (FPS %.1f)", ft, 1000/ft)
			if t.gpuCount > 0 {
				ft = t.gpuTotal / float64(t.gpuCount) * 1000.0
				t.lGPU.Text = fmt.Sprintf("GPU time %.2fms (FPS %.1f)", ft, 1000/ft)
			}
			t.count, t.gpuCount, t.total, t.gpuTotal = 0, 0, 0, 0
			t.UIView.Area = t.getArea()
		})
	}
}

func (t *FPSTimer) getArea() image.Rectangle {
	ws := t.rw.WindowSize
	return image.Rect(ws.X-350, 10, ws.X-10, 40)
}
