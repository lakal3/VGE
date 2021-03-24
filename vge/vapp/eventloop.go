package vapp

import (
	"errors"
	"sort"
	"sync"

	"github.com/lakal3/vge/vge/vk"
)

type Event interface {
	Handled() bool
}

type WaitEvent interface {
	Event
	Done()
}

type EventHandler func(ctx vk.APIContext, ev Event) (unregister bool)

type ehItem struct {
	handler  EventHandler
	priority float64
}

var eventLoop struct {
	ctx        vk.APIContext
	handlers   []ehItem
	newHandler []ehItem
	mxHandle   *sync.Mutex
	chPost     chan Event
	wg         *sync.WaitGroup
	shutdown   bool
}

func startEventLoop() {
	if eventLoop.chPost != nil {
		Ctx.SetError(errors.New("Already running"))
		return
	}
	eventLoop.wg = &sync.WaitGroup{}
	eventLoop.wg.Add(1)
	eventLoop.chPost = make(chan Event, 100)
	go runEventLoop()
}

func stopEventLoop() {
	if eventLoop.chPost != nil {
		sh := ShutdownEvent{wg: &sync.WaitGroup{}}
		sh.wg.Add(1)
		Post(sh)
		sh.wg.Wait()
		close(eventLoop.chPost)
		eventLoop.wg.Wait()
		eventLoop.wg, eventLoop.chPost = nil, nil
	}
}

// RegisterHandler will add new EventHandler to event queue. Event are passed to handlers in decreasing priority order
// EventHandler can unregister it self by returning true
func RegisterHandler(priority float64, handler EventHandler) {
	eventLoop.mxHandle.Lock()
	eventLoop.newHandler = append(eventLoop.newHandler, ehItem{priority: priority, handler: handler})
	eventLoop.mxHandle.Unlock()
}

func runEventLoop() {
	for ev := range eventLoop.chPost {
		addHandlers()
		prevIdx := 0

		for idx, ehItem := range eventLoop.handlers {
			remove := false
			if !ev.Handled() {
				remove = ehItem.handler(Ctx, ev)
			}
			if !remove {
				if prevIdx != idx {
					eventLoop.handlers[prevIdx] = eventLoop.handlers[idx]
				}
				prevIdx++
			}
		}
		if prevIdx < len(eventLoop.handlers) {
			eventLoop.handlers = eventLoop.handlers[:prevIdx]
		}
		we, ok := ev.(WaitEvent)
		if ok {
			we.Done()
		}
	}
	eventLoop.wg.Done()
}

func addHandlers() {
	found := false
	eventLoop.mxHandle.Lock()
	if len(eventLoop.newHandler) > 0 {
		eventLoop.handlers = append(eventLoop.handlers, eventLoop.newHandler...)
		eventLoop.newHandler = nil
		found = true
	}
	eventLoop.mxHandle.Unlock()
	if found {
		sort.Slice(eventLoop.handlers, func(i, j int) bool {
			return eventLoop.handlers[i].priority > eventLoop.handlers[j].priority
		})
	}
}

// Post event to event queue
func Post(ev Event) {
	if eventLoop.chPost == nil {
		Ctx.SetError(errors.New("Event loop not started"))
	}
	eventLoop.chPost <- ev
}

// Wait until event queue is stopped. Usually used at end of main function to wait that application shuts down
func WaitForShutdown() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	RegisterHandler(PRILast, func(ctx vk.APIContext, ev Event) (unregister bool) {
		_, ok := ev.(ShutdownEvent)
		if ok {
			wg.Done()
			return true
		}
		return false
	})
	wg.Wait()
}

func init() {
	eventLoop.mxHandle = &sync.Mutex{}
}
