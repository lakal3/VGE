package phongshader

import (
	"bytes"
	_ "embed"
	"github.com/lakal3/vge/vge/shaders"
)

func LoadPack() (*shaders.Pack, error) {
	sp := &shaders.Pack{}
	err := sp.Load(bytes.NewReader(phongshader_bin))
	return sp, err
}

//go:embed phongshader.bin
var phongshader_bin []byte
