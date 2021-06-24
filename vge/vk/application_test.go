package vk

import (
	"fmt"
	"testing"
)

type testContext struct {
	t *testing.T
}

func (t *testContext) IsValid() bool {
	return true
}

func (t *testContext) SetError(err error) {
	t.t.Fatal("Vulcan api error ", err)
}

func (t *testContext) Begin(callName string) (atEnd func()) {
	return nil
}

func TestAddDynamicDescriptors(t *testing.T) {
	tc := &testContext{t: t}
	a := NewApplication(tc, "IndexTest")
	a.AddDynamicDescriptors(tc)
	a.Init(tc)
	devices := a.GetDevices(tc)
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
	tc := &testContext{t: t}
	a := NewApplication(tc, "Test")
	a.AddValidation(tc)
	a.Init(tc)
	if a.hInst == 0 {
		t.Error("No instance for initialize app")
	}
	d := NewDevice(tc, a, 0)
	if d == nil {
		t.Error("Failed to initialize application")
	}
	err := testCopy(tc, d)
	if err != nil {
		t.Error("Copy failed ", err)
	}
	d.Dispose()
	a.Dispose()
}

func testCopy(ctx *testContext, dev *Device) error {
	mp := NewMemoryPool(dev)
	defer mp.Dispose()
	b1 := mp.ReserveBuffer(ctx, 999, true, BUFFERUsageTransferSrcBit)
	b2 := mp.ReserveBuffer(ctx, 1234, false, BUFFERUsageTransferSrcBit|BUFFERUsageTransferDstBit)
	b3 := mp.ReserveBuffer(ctx, 999, true, BUFFERUsageTransferDstBit)
	mp.Allocate(ctx)
	if len(b1.Bytes(ctx)) != 999 {
		return fmt.Errorf("Invalid buffer length %d", len(b1.Bytes(ctx)))
	}
	var testBuf [1024]byte
	for idx := 0; idx < len(testBuf); idx++ {
		testBuf[idx] = byte(idx)
	}
	copy(b1.Bytes(ctx), testBuf[:])
	cmd := NewCommand(ctx, dev, QUEUETransferBit, false)
	defer cmd.Dispose()
	cmd.Begin()
	cmd.CopyBuffer(b2, b1)
	cmd.Submit()
	cmd.Wait()
	cmd.Begin()
	cmd.CopyBuffer(b3, b2)
	cmd.Submit()
	cmd.Wait()
	endBytes := b3.Bytes(ctx)
	for idx, b := range endBytes {
		if b != byte(idx) {
			return fmt.Errorf("Error at %d (%d)", idx, b)
		}
	}
	return nil
}
