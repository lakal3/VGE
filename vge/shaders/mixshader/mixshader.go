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
	err = sp.Load(bytes.NewReader(mixshader_bin))
	return sp, err
}

//go:embed mixshader.bin
var mixshader_bin []byte
