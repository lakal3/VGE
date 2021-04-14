package env

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vscene"
)

type AmbientLight struct {
	Intensity mgl32.Vec3
}

func (al *AmbientLight) Process(pi *vscene.ProcessInfo) {
	_, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		lf, ok := pi.Frame.(EnvFrame)
		if ok {
			var sph [9]mgl32.Vec4
			sph[0] = al.Intensity.Vec4(1)
			lf.AddEnvironment(sph, 0, pi)
		}

	}
}
