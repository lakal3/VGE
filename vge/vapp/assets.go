package vapp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lakal3/vge/vge/materials/pbr"
	"github.com/lakal3/vge/vge/materials/std"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"github.com/lakal3/vge/vge/vmodel/gltf2loader"
	"github.com/lakal3/vge/vge/vmodel/objloader"
	"github.com/lakal3/vge/vge/vscene"
)

var AM *vasset.AssetManager
var kAssetManager = vk.NewKey()

func MustLoadAsset(path string, construct func(content []byte) (asset interface{}, err error)) (asset interface{}) {
	asset, err := AM.Load(path, construct)
	if err != nil {
		Ctx.SetError(err)
		return nil
	}
	return asset
}

func LoadModel(path string) (model *vmodel.Model, err error) {
	// mb.ShaderFactory = unlit.UnlitFactory
	ext := strings.ToLower(filepath.Ext(path))
	dl := vasset.SubDirLoader{DirPrefix: filepath.Dir(path), L: vasset.DefaultLoader}
	rawModel, err := AM.Get(path, func(path string) (asset interface{}, err error) {
		mb := vmodel.ModelBuilder{}
		mb.ShaderFactory = pbr.PbrFactory
		if vscene.FrameMaxDynamicSamplers > 0 {
			mb.ShaderFactory = std.Factory
		}
		switch ext {
		case ".obj":
			ol := objloader.ObjLoader{Builder: &mb, Loader: dl}
			err := ol.LoadFile(filepath.Base(path))
			if err != nil {
				return nil, err
			}
		case ".gltf":
			ol := gltf2loader.GLTF2Loader{Builder: &mb, Loader: dl}
			err := ol.LoadGltf(filepath.Base(path))
			if err != nil {
				return nil, err
			}
			err = ol.Convert(0)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknown model type %s", ext)
		}
		return mb.ToModel(Ctx, Dev), nil
	})
	if err != nil {
		return nil, err
	}
	return rawModel.(*vmodel.Model), nil
}
