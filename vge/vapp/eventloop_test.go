package vapp

import (
	"sync"
	"testing"
)

type testCtx struct {
	t *testing.T
}

func (t testCtx) SetError(err error) {
	t.t.Fatal(err)
}

func (t testCtx) IsValid() bool {
	return true
}

func (t testCtx) Begin(callName string) (atEnd func()) {
	return nil
}

type countUpEvent struct {
	count   int
	handled bool
	wg      *sync.WaitGroup
}

func (c *countUpEvent) Done() {
	c.wg.Done()
}

func (c *countUpEvent) Handled() bool {
	return c.handled
}

var results = []int{2, 4, 6, 7, 8, 9, 10, 11, 12}

func TestNewEventLoop(t *testing.T) {
	err := startEventLoop()
	if err != nil {
		t.Fatal("Start event loop ", err)
	}
	RegisterHandler(3, countUp(3, false))
	RegisterHandler(1, countUp(4, false))
	cue := &countUpEvent{wg: &sync.WaitGroup{}}
	for idx := 0; idx < 9; idx++ {
		cue.wg.Add(1)
		cue.handled = false
		Post(cue)
		cue.wg.Wait()
		if results[idx] != cue.count {
			t.Errorf("Excepted %d got %d on loop %d", results[idx], cue.count, idx)
		}
		if idx == 0 {
			RegisterHandler(2, countUp(5, true))
		}
	}
	if len(eventLoop.handlers) > 0 {
		t.Error("No all handler removed")
	}
	stopEventLoop()
}

func countUp(total int, handle bool) EventHandler {
	myCount := 0
	return func(ev Event) (unregister bool) {
		cce, ok := ev.(*countUpEvent)
		if ok {
			cce.count++
			myCount++
			cce.handled = handle
			return myCount >= total
		}
		return false
	}
}
