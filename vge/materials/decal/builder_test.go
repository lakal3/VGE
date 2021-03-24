package decal

import (
	"io/ioutil"
	"testing"

	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
)

func TestBuilder_Build(t *testing.T) {
	ctx := vtestapp.TestContext{T: t}
	vtestapp.Init(ctx, "decal_test")
	vasset.RegisterNativeImageLoader(ctx, vtestapp.TestApp.App)
	sb := &Builder{}
	am := loadImage(ctx, sb, "../../../assets/decals/stone_albedo.png")
	n := loadImage(ctx, sb, "../../../assets/decals/stone_normal.png")
	props := vmodel.NewMaterialProperties().SetImage(vmodel.TxAlbedo, am).SetImage(vmodel.TxBump, n)
	sb.AddDecal("oilStain", props)
	s := sb.Build(ctx, vtestapp.TestApp.Dev)
	s.Dispose()
	vtestapp.Terminate()
}

func loadImage(ctx vtestapp.TestContext, sb *Builder, imageName string) vmodel.ImageIndex {
	content, err := ioutil.ReadFile(imageName)
	if err != nil {
		ctx.SetError(err)
		return 0
	}
	return sb.AddImage("png", content, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit)
}
