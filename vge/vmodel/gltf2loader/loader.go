package gltf2loader

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func (l *GLTF2Loader) LoadGLB(name string) (err error) {
	glb, err := l.loadContent(name)
	if err != nil {
		return err
	}
	if len(glb) < 12 {
		return errors.New("Not a GLB content")
	}
	if binary.LittleEndian.Uint32(glb[0:4]) != 0x46546C67 {
		return errors.New("Not a GLB content")
	}
	if binary.LittleEndian.Uint32(glb[4:8]) != 2 {
		return errors.New("Version must be 2")
	}
	l.Model = &GLTF2{}
	offset := 12
	for offset < len(glb) {
		chLen := int(binary.LittleEndian.Uint32(glb[offset : offset+4]))
		chType := binary.LittleEndian.Uint32(glb[offset+4 : offset+8])
		switch chType {
		case 0x4E4F534A:
			err = json.Unmarshal(glb[offset+8:offset+chLen+8], l.Model)
			if err != nil {
				return err
			}
		case 0x004E4942:
			l.Model.Buffers[0].SetContent(glb[offset+8 : offset+chLen+8])
		default:

		}
		offset += chLen + 8
	}
	l.Model.Bind()
	return nil
}

func (l *GLTF2Loader) LoadGltf(mainUri string) (err error) {

	l.Model = &GLTF2{}
	core, err := l.loadContent(mainUri)
	if err != nil {
		return err
	}
	err = json.Unmarshal(core, l.Model)
	if err != nil {
		return err
	}

	for _, buf := range l.Model.Buffers {
		if len(buf.URI) > 0 {
			buf.data, err = l.loadContent(buf.URI)
			if err != nil {
				return err
			}
		}
	}
	for _, img := range l.Model.Images {
		if len(img.URI) > 0 {

			rd, err := l.Loader.Open(img.URI)
			if err != nil {
				return err
			}
			defer rd.Close()
			content, err := ioutil.ReadAll(rd)
			if err != nil {
				return err
			}
			img.kind = filepath.Ext(img.URI)[1:]
			img.content = content
		}
	}
	l.Model.Bind()
	return nil
}

func (l *GLTF2Loader) loadContent(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "data:") {
		idx := strings.Index(uri, "base64,")
		if idx < 0 {
			return nil, errors.New("Can't locate base64, from data uri")
		}
		return base64.StdEncoding.DecodeString(uri[idx+7:])
	}
	if l.Loader == nil {
		return nil, errors.New("Set loader")
	}
	r, err := l.Loader.Open(uri)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	glb, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return glb, nil
}

func DefaultLoader(dir string) func(uri string) (content []byte, err error) {
	return func(uri string) (content []byte, err error) {
		if strings.HasPrefix(uri, "data:") {
			idx := strings.Index(uri, "base64,")
			if idx < 0 {
				return nil, errors.New("Can't locate base64, from data uri")
			}
			return base64.StdEncoding.DecodeString(uri[idx+7:])
		}
		return ioutil.ReadFile(filepath.Join(dir, uri))
	}
}
