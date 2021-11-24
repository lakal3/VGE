package vk

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

var ErrDisposed = errors.New("Item disposed")

func byteArrayToUintptr(arr []byte) uintptr {
	if len(arr) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&arr[0]))
}

func boolToUintptr(val bool) uintptr {
	if val {
		return 1
	}
	return 0
}

func sliceToUintptr(views interface{}) uintptr {
	if reflect.TypeOf(views).Kind() != reflect.Slice {
		panic("Slice to uintptr called without pointer slice")
	}
	v := reflect.ValueOf(views)
	if v.Len() == 0 {
		return 0
	}
	return v.Pointer()
}

func handleError(ctx apicontext, rc uintptr) {
	if rc != 0 {
		var msg [4096]byte
		var lMsg int32
		call_Exception_GetError(hException(rc), msg[:], &lMsg)
		err := errors.New(string(msg[:lMsg]))
		call_Disposable_Dispose(hDisposable(rc))
		ctx.setError(err)
	}
}

func Float32ToBytes(src []float32) []byte {
	p := unsafe.Pointer(&src[0])
	return unsafe.Slice((*byte)(p), len(src)*4)
}

func BytesToFloat32(src []byte) []float32 {
	p := unsafe.Pointer(&src[0])
	return unsafe.Slice((*float32)(p), len(src)/4)
}

func UInt32ToBytes(src []uint32) []byte {
	p := unsafe.Pointer(&src[0])
	return unsafe.Slice((*byte)(p), len(src)*4)
}

func UInt16ToBytes(src []uint16) []byte {
	p := unsafe.Pointer(&src[0])
	return unsafe.Slice((*byte)(p), len(src)*2)
}

// Disposable interface will be implemented on object that have allocate external, typically Vulkan devices.
// You must ensure that all disposable object gets disposed. If object is owned you must just dispose owner.
// You can always call dispose on already disposed items without any errors. Send call will be just silently ignored.
// Follow this pattern if you implement Disposable interface
type Disposable interface {
	Dispose()
}

// Owner is helper structure used by multiple entities like application or device to keep list of child disposable objects that
// are bound to object lifetime.
// Typically owner instance is private and owner struct will offer similar methods (Get and sometimes Set, AddChild) to access private owner collection
type Owner struct {
	children []Disposable
	mx       *sync.Mutex
	keyMap   map[Key]interface{}
}

// Dispose all owner members in LIFO order
func (o *Owner) Dispose() {
	if o.mx != nil {
		o.mx.Lock()
		defer o.mx.Unlock()
	}
	for idx := len(o.children) - 1; idx >= 0; idx-- {
		o.children[idx].Dispose()
	}
	o.children, o.keyMap = nil, nil
}

// Retrieve object by key. You can create unique key with NewKey. If key does not exists, constructor will be called to create one.
// If constructed item is Disposable, it will be disposed when Owner collection is disposed
// It is possible to contruct non Disposable objects also. They will still be remembered by key
func (o *Owner) Get(key Key, cons Constructor) interface{} {
	if o.mx == nil {
		if o.keyMap == nil {
			o.keyMap = make(map[Key]interface{})
		}
		v, ok := o.keyMap[key]
		if !ok {
			v = cons()
			o.keyMap[key] = v
			disp, okDisp := v.(Disposable)
			if okDisp {
				o.children = append(o.children, disp)
			}
		}
		return v
	} else {
		o.mx.Lock()
		if o.keyMap == nil {
			o.keyMap = make(map[Key]interface{})
		}
		v, ok := o.keyMap[key]
		if !ok {
			o.mx.Unlock()
			vNew := cons()
			o.mx.Lock()
			defer o.mx.Unlock()
			v, ok = o.keyMap[key]
			disp, okDisp := vNew.(Disposable)
			if ok {
				// Get 2 copies. Dispose later
				if okDisp {
					disp.Dispose()
				}
				return v
			}
			o.keyMap[key] = vNew
			if okDisp {
				o.children = append(o.children, disp)
			}
			return vNew
		} else {
			o.mx.Unlock()
			return v
		}
	}
}

// Allocate new Owner collection. Non multithreaded version can be constructed by default structure constructor
func NewOwner(multithreaded bool) Owner {
	o := Owner{}
	if multithreaded {
		o.mx = &sync.Mutex{}
	}
	return o
}

// AddChild adds disposable member to this collection
func (o *Owner) AddChild(child Disposable) {
	if o.mx != nil {
		o.mx.Lock()
		defer o.mx.Unlock()
	}
	o.children = append(o.children, child)
}

// Set updates key value in owner collection.
// Nil value will clear entry and if there was previous value with same key that value is disposed
func (o *Owner) Set(key Key, val interface{}) {
	if o.mx != nil {
		o.mx.Lock()
		defer o.mx.Unlock()
	}
	if o.keyMap == nil {
		return
	}
	if val == nil {
		delete(o.keyMap, key)
	} else {
		disp, ok := val.(Disposable)
		if ok {
			o.children = append(o.children, disp)
		}
		o.keyMap[key] = val
	}
}

// Simple APIContext to panic on all errors
type PanicContext struct {
}

func (p PanicContext) SetError(err error) {
	panic(err)
}

func (p PanicContext) IsValid() bool {
	return true
}

func (p PanicContext) Begin(callName string) (atEnd func()) {
	return nil
}
