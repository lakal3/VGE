package vk

import (
	"errors"
	"reflect"
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
	mx       *SpinLock
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
		o.mx = &SpinLock{}
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
		o.keyMap = make(map[Key]interface{})
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

// State hold any number of different state values. State values keys are types them self
// State values should always be plain structs. If you need reference or pointer values put those in structure members
type State struct {
	properties map[reflect.Type]interface{}
	mx         *SpinLock
}

// NewState creates new state. Multithreaded flag makes state thread save using a spinlock.
func NewState(multithreaded bool) State {
	s := State{}
	if multithreaded {
		s.mx = &SpinLock{}
	}
	return s
}

// CloneProperty if implemented is called when State is Cloned to clone property value.
// Otherwise, value is just copied from state to new state
type CloneProperty interface {
	CloneProperty() interface{}
}

// Get retrieves property value or default value from state. Value should be struct of some type
func (st State) Get(defaultValue interface{}) interface{} {
	if st.mx != nil {
		st.mx.Lock()
		defer st.mx.Unlock()
	}
	if st.properties == nil {
		return defaultValue
	}
	pv, ok := st.properties[reflect.TypeOf(defaultValue)]
	if ok {
		return pv
	}
	return defaultValue
}

// GetExists retrieves property from state and boolean flag to indicate if value existed
func (st State) GetExists(valueType interface{}) (value interface{}, exists bool) {
	if st.mx != nil {
		st.mx.Lock()
		defer st.mx.Unlock()
	}
	if st.properties == nil {
		return valueType, false
	}
	pv, ok := st.properties[reflect.TypeOf(valueType)]
	if ok {
		return pv, true
	}
	return valueType, false
}

// Has check if value has been set
func (st State) Has(valueType interface{}) bool {
	if st.mx != nil {
		st.mx.Lock()
		defer st.mx.Unlock()
	}
	if st.properties == nil {
		return false
	}
	_, ok := st.properties[reflect.TypeOf(valueType)]
	return ok
}

// Set value
func (st *State) Set(values ...interface{}) {
	if st.mx != nil {
		st.mx.Lock()
		defer st.mx.Unlock()
	}
	if st.properties == nil {
		st.properties = make(map[reflect.Type]interface{})
	}
	for _, value := range values {
		st.properties[reflect.TypeOf(value)] = value
	}
}

// Clear removes existing value
func (st *State) Clear(valueType interface{}) {
	if st.mx != nil {
		st.mx.Lock()
		defer st.mx.Unlock()
	}
	if st.properties != nil {
		delete(st.properties, reflect.TypeOf(valueType))
	}
}

// Clone state
func (st State) Clone() State {
	sNew := State{properties: make(map[reflect.Type]interface{})}
	if st.mx != nil {
		st.mx.Lock()
		sNew.mx = &SpinLock{}
		defer st.mx.Unlock()
	}

	for t, v := range st.properties {
		cl, ok := v.(CloneProperty)
		if ok {
			c := cl.CloneProperty()
			if reflect.TypeOf(c) == t {
				sNew.properties[t] = c
				continue
			}
		}
		sNew.properties[t] = v
	}
	return sNew
}
