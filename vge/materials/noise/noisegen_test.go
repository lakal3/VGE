package noise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
)

func TestNewPerlinNoise(t *testing.T) {
	pn := NewPerlinNoise(256)
	pn.Add(1, 35.6)
	pn.Add(0.4, 15.7)
	// pn.Add(0.5, 17)
	bTmp := pn.ToBytes()
	testDir := os.Getenv("VGE_TEST_DIR")
	if len(testDir) == 0 {
		t.Log("Unable to save test image, missing environment variable VGE_TEST_DIR")
		return
	}
	fPath := filepath.Join(testDir, "perlin1.dds")
	fOut, err := os.Create(fPath)
	if err != nil {
		t.Fatal("Write to ", fPath, " failed ", err)
	}
	defer fOut.Close()
	desc := vk.ImageDescription{Width: 256, Depth: 1, Height: 256, Format: vk.FORMATR8Unorm, Layers: 1, MipLevels: 1}
	err = vasset.WriteDDS(fOut, desc, bTmp)
	if err != nil {
		t.Error("Error writing perlin1.dds ", err)
	}
}
