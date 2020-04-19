package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/debugmat"
	"github.com/lakal3/vge/vge/materials/decal"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vscene"
	"log"
	"path/filepath"
)

func loadDecals() (set *decal.Set) {
	// Add decals
	rs, err := vapp.AM.Get("asset/decal/stain", func() (asset interface{}, err error) {
		abContent, err := vasset.Load("assets/decals/stain_albedo.png", vapp.AM.Loader)
		if err != nil {
			return nil, err
		}
		b := &decal.Builder{}
		stAlbedo := b.AddImage("png", abContent, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
		props := vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, stAlbedo)
		b.AddDecal("stain", props)
		props = vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, stAlbedo).SetColor(vmodel.CAlbedo, mgl32.Vec4{0.7, 0.7, 0.7, 0.5}).SetFactor(vmodel.FNormalAttenuation, 0)
		b.AddDecal("stain2", props)
		abContent, err = vasset.Load("assets/decals/stone_albedo.png", vapp.AM.Loader)
		if err != nil {
			return nil, err
		}
		stoneAlbedo := b.AddImage("png", abContent, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
		normalContent, err := vasset.Load("assets/decals/stone_normal.png", vapp.AM.Loader)
		if err != nil {
			return nil, err
		}
		stoneNormal := b.AddImage("png", normalContent, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
		props = vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, stoneAlbedo).SetImage(vmodel.TxBump, stoneNormal)
		b.AddDecal("stone", props)
		abContent, err = vasset.Load("assets/decals/uc_albedo.png", vapp.AM.Loader)
		if err != nil {
			return nil, err
		}
		ucAlbedo := b.AddImage("png", abContent, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
		normalContent, err = vasset.Load("assets/decals/uc_normal.png", vapp.AM.Loader)
		if err != nil {
			return nil, err
		}
		ucNormal := b.AddImage("png", normalContent, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
		props = vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, ucAlbedo).SetImage(vmodel.TxBump, ucNormal)
		b.AddDecal("underconstruction", props)
		return b.Build(vapp.Ctx, vapp.Dev), nil
	})
	if err != nil {
		log.Fatal("Load decals failed ", err)
	}
	return rs.(*decal.Set)
}

func openDecalTest() {
	testModel := loadTestModel()
	nModelRool := vscene.NewNode(nil)
	nModelRool.Children = append(nModelRool.Children,
		vscene.NewNode(&vscene.DirectionalLight{Intensity: mgl32.Vec3{0.7, 0.7, 0.7}, Direction: mgl32.Vec3{-0.1, -1, -0.2}.Normalize()}),
		vscene.NodeFromModel(testModel, 0, true),
	)

	dSet := loadDecals()

	nModelRool.Ctrl = vscene.NewMultiControl(
		dSet.NewInstance("stone", mgl32.Ident4()),
		dSet.NewInstance("underconstruction", mgl32.Translate3D(2, 0, 2).Mul4(mgl32.Scale3D(1.5, 5, 1.5))),
		dSet.NewInstance("stain", mgl32.Translate3D(2, 0, -2).Mul4(mgl32.HomogRotate3DX(-1))),
		dSet.NewInstance("stain2", mgl32.Translate3D(-2, 0, -2).Mul4(mgl32.Scale3D(1.5, 1.5, 1.5))),
	)

	app.rw.Scene.Update(func() {
		app.cam.Position = mgl32.Vec3{1, 2, 5}
		app.rw.Model.Children = []*vscene.Node{nModelRool}
	})
}

func loadTestModel() *vmodel.Model {
	dl := vasset.SubDirLoader{DirPrefix: "assets/gltf/testparts", L: vasset.DefaultLoader}
	mb := vmodel.ModelBuilder{}
	mb.ShaderFactory = debugmat.DebugMatFactory
	debugmat.DebugModes = mgl32.Vec4{1, 1, 1, 0}
	ol := gltf2loader.GLTF2Loader{Builder: &mb, Loader: dl}
	err := ol.LoadGltf(filepath.Base("testparts.gltf"))
	if err != nil {
		log.Fatal("Load testparts model failed ", err)
	}
	err = ol.Convert(0)
	if err != nil {
		log.Fatal("Load testparts model failed ", err)
	}
	return mb.ToModel(vapp.Ctx, vapp.Dev)
}
