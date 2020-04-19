package vanimation

import (
	"github.com/lakal3/vge/vge/vasset"
	"testing"
)

func TestLoadBVH(t *testing.T) {
	l := vasset.DirectoryLoader{Directory: "../../assets/bvh/tests"}
	bvh, err := LoadBVH(l, "test1.bvh")
	if err != nil {
		t.Error("Load BVH failed, ", err)
		return
	}
	if len(bvh.frames) != 1413 {
		t.Error("Invalid number of frames, assumed 1413 got ", len(bvh.frames))
	}
}
