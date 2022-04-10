package shaders

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lakal3/vge/vge/vk"
	"io"
	"os"
)

var fileMagic = []byte("VGESpiv\x00")

type SpirvCode struct {
	Vertex   []byte
	Fragment []byte
	Compute  []byte
	Geometry []byte
}

type Pack struct {
	programs map[string]*SpirvCode
}

// AddFrom adds shader from other pack to this
func (sp *Pack) AddFrom(spFrom *Pack) {
	for k, v := range spFrom.programs {
		sp.Add(k, v)
	}
}

func (sp *Pack) Add(name string, code *SpirvCode) {
	if sp.programs == nil {
		sp.programs = make(map[string]*SpirvCode)
	}
	sp.programs[name] = code
}

func (sp *Pack) Load(reader io.Reader) error {
	if sp.programs == nil {
		sp.programs = make(map[string]*SpirvCode)
	}
	var magic [8]byte
	_, err := io.ReadFull(reader, magic[:])
	if err != nil {
		return err
	}
	if bytes.Compare(magic[:], fileMagic) != 0 {
		return errors.New("file content is not VGE spirv pack")
	}
	next := true
	for next {
		next, err = sp.readOne(reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sp *Pack) LoadFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return sp.Load(f)
}

func (sp *Pack) Get(programName string) *SpirvCode {
	code, _ := sp.programs[programName]
	return code
}

func (sp *Pack) MustGet(dev *vk.Device, programName string) *SpirvCode {
	code, ok := sp.programs[programName]
	if !ok {
		dev.FatalError(fmt.Errorf("No program %s", programName))
		return nil
	}
	return code
}

func (sp *Pack) Save(write io.Writer) error {
	bw := bufio.NewWriter(write)
	_, _ = bw.Write(fileMagic)
	for k, v := range sp.programs {
		sp.writeOne(bw, k, v)
	}
	sp.writeUint32(bw, 0)
	return bw.Flush()
}

func (sp *Pack) readOne(reader io.Reader) (bool, error) {
	n, err := sp.readUInt32(reader)
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	name := make([]byte, n)
	_, err = io.ReadFull(reader, name)
	if err != nil {
		return false, err
	}
	kind := uint32(1)
	var spirLen uint32
	codes := &SpirvCode{}
	for kind != 0 {
		kind, err = sp.readUInt32(reader)
		if err != nil {
			return false, err
		}
		if kind == 0 {
			break
		}
		spirLen, err = sp.readUInt32(reader)
		spir := make([]byte, spirLen)
		_, err = io.ReadFull(reader, spir)
		if err != nil {
			return false, err
		}
		switch vk.ShaderStageFlags(kind) {
		case vk.SHADERStageFragmentBit:
			codes.Fragment = spir
		case vk.SHADERStageVertexBit:
			codes.Vertex = spir
		case vk.SHADERStageGeometryBit:
			codes.Geometry = spir
		case vk.SHADERStageComputeBit:
			codes.Compute = spir
		default:
			return false, fmt.Errorf("Unknown shader stage %x", kind)
		}
	}
	sp.programs[string(name)] = codes
	return true, nil
}

func (sp *Pack) readUInt32(reader io.Reader) (n uint32, err error) {
	var content [4]byte
	_, err = io.ReadFull(reader, content[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(content[:]), nil
}

func (sp *Pack) writeOne(bw *bufio.Writer, name string, code *SpirvCode) {
	sp.writeUint32(bw, uint32(len(name)))
	_, _ = bw.Write([]byte(name))
	if code.Vertex != nil {
		sp.writeCode(bw, vk.SHADERStageVertexBit, code.Vertex)
	}
	if code.Fragment != nil {
		sp.writeCode(bw, vk.SHADERStageFragmentBit, code.Fragment)
	}
	if code.Geometry != nil {
		sp.writeCode(bw, vk.SHADERStageGeometryBit, code.Geometry)
	}
	if code.Compute != nil {
		sp.writeCode(bw, vk.SHADERStageComputeBit, code.Compute)
	}
	sp.writeUint32(bw, 0)
}

func (sp *Pack) writeUint32(bw *bufio.Writer, u uint32) {
	var content [4]byte
	binary.LittleEndian.PutUint32(content[:], u)
	_, _ = bw.Write(content[:])

}

func (sp *Pack) writeCode(bw *bufio.Writer, kind vk.ShaderStageFlags, spirv []byte) {
	sp.writeUint32(bw, uint32(kind))
	sp.writeUint32(bw, uint32(len(spirv)))
	_, _ = bw.Write(spirv)
}
