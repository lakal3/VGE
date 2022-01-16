package vk

import "testing"

func BenchmarkAllocator_AllocMemory(b *testing.B) {
	a, err := NewApplication("Test")
	if err != nil {
		b.Fatal("NewApplication ", err)
	}
	a.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	a.Init()
	if a.hInst == 0 {
		b.Error("No instance for initialize app")
	}
	d := NewDevice(a, 0)
	if d == nil {
		b.Error("Failed to initialize application")
	}
	d.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	b.ResetTimer()
	al := NewAllocator(d)
	bTmp := al.AllocBuffer(BUFFERUsageStorageBufferBit|BUFFERUsageTransferDstBit|BUFFERUsageTransferSrcBit, 1024*1024, true)
	for idx := 0; idx < b.N; idx++ {
		mem := al.AllocMemory(1024*1024, bTmp.MemoryType(), true)
		mem.Dispose()
	}
	bTmp.Dispose()
	al.Dispose()
	d.Dispose()
	a.Dispose()
}

func BenchmarkAllocator_AllocBuffer(b *testing.B) {
	a, err := NewApplication("Test")
	if err != nil {
		b.Fatal("NewApplication ", err)
	}
	a.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	a.Init()
	if a.hInst == 0 {
		b.Error("No instance for initialize app")
	}
	d := NewDevice(a, 0)
	if d == nil {
		b.Error("Failed to initialize application")
	}
	d.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	b.ResetTimer()
	al := NewAllocator(d)
	bTmp := al.AllocBuffer(BUFFERUsageStorageBufferBit|BUFFERUsageTransferDstBit|BUFFERUsageTransferSrcBit, 1024*1024, true)
	mem := al.AllocMemory(1024*1024, bTmp.MemoryType(), true)
	bTmp.Dispose()

	for idx := 0; idx < b.N; idx++ {
		buf := al.AllocBuffer(BUFFERUsageStorageBufferBit|BUFFERUsageTransferDstBit|BUFFERUsageTransferSrcBit, 1024, true)
		buf.Bind(mem, 1024)
		buf.Dispose()
	}
	mem.Dispose()
	al.Dispose()
	d.Dispose()
	a.Dispose()
}

func BenchmarkAllocator_AllocImage(b *testing.B) {
	a, err := NewApplication("Test")
	if err != nil {
		b.Fatal("NewApplication ", err)
	}
	a.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	// a.AddValidation()
	a.Init()
	if a.hInst == 0 {
		b.Error("No instance for initialize app")
	}
	d := NewDevice(a, 0)
	if d == nil {
		b.Error("Failed to initialize application")
	}
	d.OnFatalError = func(fatalError error) {
		b.Fatal(fatalError)
	}
	b.ResetTimer()
	al := NewAllocator(d)
	desc := ImageDescription{Format: FORMATR8g8b8a8Unorm, Width: 1024, Height: 768, Depth: 1, Layers: 1, MipLevels: 1}
	iTmp := al.AllocImage(IMAGEUsageSampledBit|IMAGEUsageTransferSrcBit|IMAGEUsageTransferDstBit, desc)
	sz, alignment := iTmp.Size()
	mem := al.AllocMemory(sz+uint64(alignment), iTmp.MemoryType(), false)
	iTmp.Dispose()

	for idx := 0; idx < b.N; idx++ {
		img := al.AllocImage(IMAGEUsageSampledBit|IMAGEUsageTransferSrcBit|IMAGEUsageTransferDstBit, desc)
		img.Bind(mem, uint64(alignment))
		ir := ImageRange{LayerCount: 1, LevelCount: 1}
		v := img.AllocView(ir, false)
		v.Dispose()
		img.Dispose()
	}
	mem.Dispose()
	al.Dispose()
	d.Dispose()
	a.Dispose()
}
