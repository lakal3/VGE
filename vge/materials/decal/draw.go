package decal

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/forward"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

func (d *Decal) Process(pi *vscene.ProcessInfo) {
	pre, ok := pi.Phase.(*vscene.PredrawPhase)
	if ok {
		f := forward.MustGetForwardFrame(pre.Cache.Ctx, pre.Frame)
		d.set.addImage(f, pre.Cache, d.txAlbedo)
		d.set.addImage(f, pre.Cache, d.txMetalRoughness)
		d.set.addImage(f, pre.Cache, d.txNormal)
		return
	}
	dp, ok := pi.Phase.(vscene.DrawPhase)
	if !ok {
		return
	}
	dc := dp.GetDC(vscene.LAYER3D)
	if dc == nil {
		return
	}
	raw := pi.Get(kDecalProcessInfo)
	var dpi *decalProcessInfo
	if raw == nil {
		dpi = &decalProcessInfo{}
		pi.Set(kDecalProcessInfo, dpi)
	} else {
		dpi = raw.(*decalProcessInfo)
	}
	at := pi.World.Mul4(d.At)
	rAt := at.Inv()
	txAlbedo := d.set.getImage(dc.Cache, d.txAlbedo)
	txMetalRoughness := d.set.getImage(dc.Cache, d.txMetalRoughness)
	txNormal := d.set.getImage(dc.Cache, d.txNormal)
	if txNormal < 0 {
		// not enough space
		return
	}
	hasNormals := float32(0)
	if txNormal != 0 {
		hasNormals = 1
	}
	dpi.instances = append(dpi.instances, rAt)
	var rRest mgl32.Mat4
	rRest.SetCol(0, d.AlbedoFactor)
	rRest.SetCol(2, mgl32.Vec4{d.MetalnessFactor, d.RoughnessFactor, hasNormals, d.NormalAttenuation})
	rRest.SetCol(3, mgl32.Vec4{float32(txAlbedo), float32(txNormal), float32(txMetalRoughness), 0})
	dpi.instances = append(dpi.instances, rRest)
}

// GetDeacls will return active decals in matrix format. Decals will be encoded to 64 float (2 x matrix). See decalGPUInstance
// for actual layout
// Matrix format is used so that decals can be packed to same uniform set with skin matrixes
func GetDecals(extra vmodel.ShaderExtra) []mgl32.Mat4 {
	raw := extra.Get(kDecalProcessInfo)
	if raw == nil {
		return nil
	}
	return raw.(*decalProcessInfo).instances
}

var kDecalProcessInfo = vk.NewKey()

type decalProcessInfo struct {
	instances []mgl32.Mat4
}

var kDecals = vk.NewKey()

type decalGPUInstance struct {
	toDecalSpace     mgl32.Mat4
	albedoFactor     mgl32.Vec4
	filler2          mgl32.Vec4
	metalnessFactor  float32
	roughnessFactor  float32
	hasNormalMap     float32
	filler           float32
	txAlbedo         float32
	txNormal         float32
	txMetalRoughness float32
	filler1          float32
}
