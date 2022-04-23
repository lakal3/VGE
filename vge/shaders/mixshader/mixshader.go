package mixshader

import (
	"bytes"
	_ "embed"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vdraw3d"
)

func LoadPack() (*shaders.Pack, error) {
	sp, err := vdraw3d.LoadDefault()
	if err != nil {
		return nil, err
	}
	return sp, AddPack(sp)
}

func AddPack(base *shaders.Pack) error {
	return base.Load(bytes.NewReader(mixshader_bin))
}

//go:embed mixshader.bin
var mixshader_bin []byte
