package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vscene"
	"math"
	"math/rand"
)

type stain struct {
	dec       *decal.Decal
	pos       mgl32.Vec2
	rot       float32
	createdAt float64
	updateAt  float64
	removed   bool
	m         *maze
}

func (s *stain) Process(pi *vscene.ProcessInfo) {
	if pi.Time-s.updateAt > 0.001 {
		s.updateAt = pi.Time
		d := pi.Time - s.createdAt
		var scale float32
		if d < 2 {
			scale = float32(d) / 2
		} else {
			scale = 1 - (float32(d)-2)/10
		}
		if scale < 0 {
			s.removed = true
			var lst []*stain
			for _, st := range s.m.stains {
				if !st.removed {
					lst = append(lst, st)
				}
			}
			s.m.stains = lst

		} else {
			s.dec.At = mgl32.Translate3D(s.pos[0], 0, s.pos[1]).Mul4(
				mgl32.HomogRotate3DY(s.rot)).Mul4(mgl32.Scale3D(scale*0.2, 1, scale*0.2))
		}
	}
	if s.removed {
		return
	}
	s.dec.Process(pi)
}

var minStainDelta = 1.0

func (r *robo) checkNewStain(time float64) {
	if (time-r.lastDrop)*rand.Float64() < minStainDelta {
		return
	}
	r.lastDrop = time
	wd := time - r.walkStart
	// Add stain if possible
	if len(r.m.stains) >= 20 {
		// Failed, adjust minStainDelta
		minStainDelta *= 1.5
		return
	}

	st := &stain{
		m: r.m, pos: mgl32.Vec2{lerpFloat(wd, r.current.x, r.next.x), -lerpFloat(wd, r.current.y, r.next.y)},
		rot: rand.Float32() * math.Pi * 2, createdAt: time, dec: app.stainSet.NewInstance("oil_stain", mgl32.Ident4()),
	}
	r.m.stains = append(r.m.stains, st)
	updateStains(r.m)
}

func updateStains(m *maze) {
	app.mainWnd.Scene.Update(func() {
		var ncs []vscene.NodeControl
		for _, st := range m.stains {
			ncs = append(ncs, st)
		}
		m.nFloorRoot.Ctrl = vscene.NewMultiControl(ncs...)
	})
}
