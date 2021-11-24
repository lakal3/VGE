//

package env

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vscene"
)

// Render equirectangular projection image as background
type EquiRectBGNode struct {
	pool   *vk.MemoryPool
	imBg   *vk.Image
	dsPool *vk.DescriptorPool
	ds     *vk.DescriptorSet
}

func (eq *EquiRectBGNode) Process(pi *vscene.ProcessInfo) {
	db, ok := pi.Phase.(vscene.DrawPhase)
	if ok {
		dc := db.GetDC(vscene.LAYERBackground)
		if dc != nil {
			eq.Draw(dc)
		}
	}
}

func (e *EquiRectBGNode) Dispose() {
	if e.pool != nil {
		e.pool.Dispose()
		e.pool, e.imBg = nil, nil
	}
	if e.dsPool != nil {
		e.dsPool.Dispose()
		e.dsPool, e.ds = nil, nil
	}
}

func (eq *EquiRectBGNode) Draw(dc *vmodel.DrawContext) {
	sfc := vscene.GetSimpleFrame(dc.Frame)
	if sfc == nil {
		return // Not supported
	}
	cache := sfc.GetCache()
	pl := dc.Pass.Get(kEqPipeline, func() interface{} {
		return eq.newPipeline(dc)
	}).(vk.Pipeline)
	dsFrame := sfc.BindFrame()
	cb := getCube(cache.Device)
	dc.Draw(pl, 0, 36).AddInputs(cb.bVtx).AddDescriptors(dsFrame, eq.ds)
}

func (eq *EquiRectBGNode) newPipeline(dc *vmodel.DrawContext) *vk.GraphicsPipeline {
	cache := dc.Frame.GetCache()
	gp := vk.NewGraphicsPipeline(cache.Device)
	gp.AddVextexInput(vk.VERTEXInputRateVertex, vk.FORMATR32g32b32Sfloat)
	la := getLayout(cache.Device)
	laFrame := vscene.GetUniformLayout(cache.Device)
	gp.AddLayout(laFrame)
	gp.AddLayout(la)
	gp.AddShader(vk.SHADERStageVertexBit, eqrect_vert_spv)
	gp.AddShader(vk.SHADERStageFragmentBit, eqrect_frag_spv)
	gp.Create(dc.Pass)
	return gp

}

func NewEquiRectBGNode(dev *vk.Device, far float32, bgKind string, bgContent []byte) *EquiRectBGNode {
	la := getLayout(dev)
	sampler := getEnvSampler(dev)
	eq := &EquiRectBGNode{}
	eq.dsPool = vk.NewDescriptorPool(la, 1)
	eq.ds = eq.dsPool.Alloc()
	eq.pool = vk.NewMemoryPool(dev)
	var desc vk.ImageDescription
	vasset.DescribeImage(bgKind, &desc, bgContent)
	eq.imBg = eq.pool.ReserveImage(desc, vk.IMAGEUsageTransferDstBit|vk.IMAGEUsageSampledBit)
	eq.pool.Allocate()
	cp := vmodel.NewCopier(dev)
	defer cp.Dispose()
	cp.CopyToImage(eq.imBg, bgKind, bgContent, desc.FullRange(), vk.IMAGELayoutShaderReadOnlyOptimal)
	eq.ds.WriteImage(0, 0, eq.imBg.DefaultView(), sampler)
	return eq
}

var kEqLayout = vk.NewKey()
var kEnvSampler = vk.NewKey()
var kEqPipeline = vk.NewKey()
var kCube = vk.NewKey()

type cubeVertex struct {
	pool *vk.MemoryPool
	bVtx *vk.Buffer
}

func (cv *cubeVertex) Dispose() {
	if cv.pool != nil {
		cv.pool.Dispose()
		cv.pool, cv.bVtx = nil, nil
	}
}

func (cv *cubeVertex) addCube(tr mgl32.Mat4) (aPos []float32) {
	aPos = cv.addPlane(aPos, tr)
	aPos = cv.addPlane(aPos, tr.Mul4(mgl32.HomogRotate3DX(math.Pi/2)))
	aPos = cv.addPlane(aPos, tr.Mul4(mgl32.HomogRotate3DX(math.Pi/-2)))
	aPos = cv.addPlane(aPos, tr.Mul4(mgl32.HomogRotate3DY(math.Pi/2)))
	aPos = cv.addPlane(aPos, tr.Mul4(mgl32.HomogRotate3DY(math.Pi)))
	aPos = cv.addPlane(aPos, tr.Mul4(mgl32.HomogRotate3DY(-math.Pi/2)))
	return
}

func (cv *cubeVertex) addPlane(aPos []float32, tr mgl32.Mat4) []float32 {
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{-1, -1, 1})
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{1, -1, 1})
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{1, 1, 1})
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{-1, -1, 1})
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{1, 1, 1})
	aPos = cv.addPos(aPos, tr, mgl32.Vec3{-1, 1, 1})
	return aPos
}

func (cv *cubeVertex) addPos(aPos []float32, tr mgl32.Mat4, pos mgl32.Vec3) []float32 {
	pos2 := tr.Mul4x1(pos.Vec4(1))
	aPos = append(aPos, pos2[:3]...)
	return aPos
}

func getCube(dev *vk.Device) *cubeVertex {
	return dev.Get(kCube, func() interface{} {
		cv := &cubeVertex{}
		cv.pool = vk.NewMemoryPool(dev)
		cv.bVtx = cv.pool.ReserveBuffer(36*4*4, false, vk.BUFFERUsageTransferDstBit|vk.BUFFERUsageVertexBufferBit)
		cv.pool.Allocate()

		aPos := cv.addCube(mgl32.Ident4())
		cp := vmodel.NewCopier(dev)
		defer cp.Dispose()
		cp.CopyToBuffer(cv.bVtx, vk.Float32ToBytes(aPos))
		return cv
	}).(*cubeVertex)
}

func getLayout(dev *vk.Device) *vk.DescriptorLayout {
	return dev.Get(kEqLayout, func() interface{} {
		return vk.NewDescriptorLayout(dev, vk.DESCRIPTORTypeCombinedImageSampler, vk.SHADERStageFragmentBit, 1)
	}).(*vk.DescriptorLayout)
}

func getEnvSampler(dev *vk.Device) *vk.Sampler {
	sampler := dev.Get(kEnvSampler, func() interface{} {
		return vk.NewSampler(dev, vk.SAMPLERAddressModeClampToEdge)
	}).(*vk.Sampler)
	return sampler
}
