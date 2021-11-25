package vk

import (
	"os"
	"path/filepath"
	"testing"
)

const testFrag = `
#version 450

layout(location = 0) out vec4 outColor;
layout(location = 1) out vec4 outGray;
layout(location = 0) in vec2 i_uv;

layout(set = 0, binding = 0) uniform sampler2D tx_color[];

void main() {
    outColor = texture(tx_color[1], i_uv);
    float c = (outColor.r + outColor.g + outColor.b) / 3;
    outGray = vec4(c,c,c,1);
}
`

func TestNewCompiler(t *testing.T) {
	a, err := NewApplication("Test")
	if err != nil {
		t.Fatal("New application ", err)
	}
	a.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}
	a.AddValidation()
	a.Init()
	if a.hInst == 0 {
		t.Error("No instance for initialize app")
	}
	d := NewDevice(a, 0)
	if d == nil {
		t.Error("Failed to initialize device")
	}
	d.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}
	comp := NewCompiler(d)
	defer comp.Dispose()
	spirv, info, err := comp.Compile(SHADERStageFragmentBit, testFrag)
	if err != nil {
		t.Error("Compile failed ", err.Error())
	} else {
		t.Log("Compile info ", info)
		t.Log("Spirv length ", len(spirv))
	}
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		t.Log("No VGE_TEST_DIR defined")
	} else {
		fn := filepath.Join(testDir, "compiler_text.frag.spv")
		err = os.WriteFile(fn, spirv, 0660)
		if err != nil {
			t.Error("Error writing compiler_text.frag.spv", err)
		} else {
			t.Log("Shader written to ", fn)
		}
	}
}
