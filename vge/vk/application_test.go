package vk

import (
	"fmt"
	"testing"
)

func TestAddDynamicDescriptors(t *testing.T) {
	a, err := NewApplication("IndexTest")
	if err != nil {
		t.Fatal("New application ", err)
	}
	a.OnFatalError = func(fatalError error) {
		t.Fatal(fatalError)
	}
	a.AddDynamicDescriptors()
	a.Init()
	devices := a.GetDevices()
	for idx, dev := range devices {
		t.Logf("%d. %s", idx+1, string(dev.Name[:dev.NameLen]))
		if dev.ReasonLen > 0 {
			t.Logf("Incompatible %s", string(dev.Reason[:dev.ReasonLen]))
		}

	}
	a.Dispose()
}

func TestNewApplication(t *testing.T) {
	// VGEDllPath = "vgelibd.dll"
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
	err = testCopy(d)
	if err != nil {
		t.Error("Copy failed ", err)
	}
	d.Dispose()
	a.Dispose()
}

func testCopy(dev *Device) error {
	mp := NewMemoryPool(dev)
	defer mp.Dispose()
	b1 := mp.ReserveBuffer(999, true, BUFFERUsageTransferSrcBit)
	b2 := mp.ReserveBuffer(1234, false, BUFFERUsageTransferSrcBit|BUFFERUsageTransferDstBit)
	b3 := mp.ReserveBuffer(999, true, BUFFERUsageTransferDstBit)
	mp.Allocate()
	if len(b1.Bytes()) != 999 {
		return fmt.Errorf("Invalid buffer length %d", len(b1.Bytes()))
	}
	var testBuf [1024]byte
	for idx := 0; idx < len(testBuf); idx++ {
		testBuf[idx] = byte(idx)
	}
	copy(b1.Bytes(), testBuf[:])
	cmd := NewCommand(dev, QUEUETransferBit, false)
	defer cmd.Dispose()
	cmd.Begin()
	cmd.CopyBuffer(b2, b1)
	cmd.Submit()
	cmd.Wait()
	cmd.Begin()
	cmd.CopyBuffer(b3, b2)
	cmd.Submit()
	cmd.Wait()
	endBytes := b3.Bytes()
	for idx, b := range endBytes {
		if b != byte(idx) {
			return fmt.Errorf("Error at %d (%d)", idx, b)
		}
	}
	return nil
}
