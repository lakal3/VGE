package vk

import (
	"errors"
	"unsafe"
)

type GlslCompiler struct {
	dev  *Device
	comp hGlslCompiler
}

func (g *GlslCompiler) Dispose() {
	if g.comp != 0 {
		call_Disposable_Dispose(hDisposable(g.comp))
		g.comp = 0
	}
}

func NewCompiler(dev *Device) *GlslCompiler {
	gc := &GlslCompiler{dev: dev}
	call_Device_NewGlslCompiler(dev, dev.hDev, &gc.comp)
	return gc
}

func (gc *GlslCompiler) Compile(stage ShaderStageFlags, src string) (spirv []byte, info string, err error) {
	var inst uintptr
	call_GlslCompiler_Compile(gc.dev, gc.comp, stage, []byte(src), &inst)
	var msg uintptr
	var msgLen uint64
	var result uint32
	call_GlslCompiler_GetOutput(gc.dev, gc.comp, inst, &msg, &msgLen, &result)
	s := string(unsafe.Slice((*byte)(unsafe.Pointer(msg)), msgLen))
	if result >= 10 {
		call_GlslCompiler_Free(gc.dev, gc.comp, inst)
		return nil, "", errors.New(s)
	}
	var rawSpirv uintptr
	var spirLen uint64
	call_GlslCompiler_GetSpirv(gc.dev, gc.comp, inst, &rawSpirv, &spirLen)
	spirv = make([]byte, spirLen)
	copy(spirv, unsafe.Slice((*byte)(unsafe.Pointer(rawSpirv)), spirLen))
	call_GlslCompiler_Free(gc.dev, gc.comp, inst)
	return spirv, s, nil
}
