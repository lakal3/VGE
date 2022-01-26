package vdraw

import (
	"github.com/lakal3/vge/vge/vapp/vtestapp"
	"testing"
)

func TestCompileShapes(t *testing.T) {
	err := vtestapp.Init("compileshapes", opt{})
	if err != nil {
		t.Fatal("Init app: ", err)
	}
	err = CompileShapes(vtestapp.TestApp.Dev)
	if err != nil {
		t.Error("Compile shapes ", err)
	}
	t.Log("Len spirv rect ", len(rectShape.spirv))

	vtestapp.Terminate()
}
