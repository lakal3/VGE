package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/noise"
	"github.com/lakal3/vge/vge/vscene"
)

func openFireTest() {
	nModelRool := vscene.NewNode(nil)
	nModelRool.Children = append(nModelRool.Children,
		vscene.NodeFromModel(app.tools, app.tools.FindNode("Dullbg"), true),
		vscene.NewNode(&vscene.TransformControl{Transform: mgl32.Translate3D(0, 0.2, 0.5)},
			vscene.NewNode(noise.NewFire(0.3, 0.5))),
	)

	app.rw.Scene.Update(func() {
		app.rw.Model.Children = []*vscene.Node{nModelRool}
	})
}
