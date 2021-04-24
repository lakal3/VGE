package main

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vscene"
)

type stain struct {
	pos       mgl32.Vec2
	rot       float32
	createdAt float64
	updateAt  float64
	removed   bool
	location  mgl32.Mat4
}

type stainPainter struct {
	painter decal.LocalPainter
	m       *maze
}

func (s *stainPainter) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(*vscene.AnimatePhase)
	if ok {
		var lst []*stain
		s.painter.Decals = nil
		for _, st := range s.m.stains {
			st.update(pi.Time)
			if !st.removed {
				lst = append(lst, st)
				s.painter.AddDecal(app.stainSet, 0, st.location, mgl32.Vec4{1, 1, 1, 1})
			}
		}
		s.m.stains = lst
	} else {
		s.painter.Process(pi)
	}
}

func (s *stain) update(time float64) {
	if time-s.updateAt > 0.001 {
		s.updateAt = time
		d := time - s.createdAt
		var scale float32
		if d < 2 {
			scale = float32(d) / 2
		} else {
			scale = 1 - (float32(d)-2)/10
		}
		if scale < 0 {
			s.removed = true

		} else {
			s.location = mgl32.Translate3D(s.pos[0], 0, s.pos[1]).Mul4(
				mgl32.HomogRotate3DY(s.rot)).Mul4(mgl32.Scale3D(scale*0.2, 1, scale*0.2))
		}
	}
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
		pos: mgl32.Vec2{lerpFloat(wd, r.current.x, r.next.x), -lerpFloat(wd, r.current.y, r.next.y)},
		rot: rand.Float32() * math.Pi * 2, createdAt: time,
	}
	r.m.stains = append(r.m.stains, st)
}
