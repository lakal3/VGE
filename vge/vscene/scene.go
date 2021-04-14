package vscene

import (
	"github.com/lakal3/vge/vge/vmodel"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
)

// Mat4ToBytes will convert array of matrixes to byte array using slice manipulation. Returned byte slice is not an copy
// instead if will point to same location that original slice!
// This helper method can be used to copy matrixies to Vulkan buffers
func Mat4ToBytes(src []mgl32.Mat4) []byte {
	dPtr := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d2Ptr := &reflect.SliceHeader{Len: dPtr.Len * 64, Cap: dPtr.Cap * 64, Data: dPtr.Data}
	d2 := (*[]byte)(unsafe.Pointer(d2Ptr))
	runtime.KeepAlive(src)
	return *d2
}

// Scene in container whole rendered "view". Scene will consist hierarchical set of nodes starting from Root node
// See Update and Process on how to safely changes nodes of scene
type Scene struct {
	Root      Node
	mx        *sync.Mutex
	lockCount int32
	pending   []func()
	Time      float64
}

// Init must be called before Process or Update
func (sc *Scene) Init() {
	if sc.mx == nil {
		sc.mx = &sync.Mutex{}
	}
}

// Update will allow safe update to live scenes. When render loop is rendering scene or sum subpart of if like shadow maps
// scene will remain in readonly state so that each phase will see consistent view of scene. Update will schedule changes
// to next possible moment when scene is not used in read only phase
func (sc *Scene) Update(action func()) {
	sc.mx.Lock()
	defer func() {
		atomic.AddInt32(&sc.lockCount, -1)
		sc.mx.Unlock()
	}()
	if atomic.AddInt32(&sc.lockCount, 1) > 0 {
		sc.pending = append(sc.pending, action)
	} else {
		action()
	}
}

// Process will change scene to read only state and process through all phases. Some phases like shadow map rendering
// might again recursively call Process to render shadow map. Process function keep consistent state of scene until all process
// invocations have exited
func (sc *Scene) Process(time float64, frame vmodel.Frame, phases ...Phase) {
	defer func() {
		atomic.AddInt32(&sc.lockCount, -1)
	}()
	sc.mx.Lock()
	if atomic.AddInt32(&sc.lockCount, 1) == 1 {
		sc.updatePending()
	} else {
		sc.mx.Unlock()
	}

	for _, ph := range phases {
		sc.processPhase(time, frame, mgl32.Ident4(), ph)
	}
}

// Add new child node to parent. This method should be called from scene.Update if there is any change that
// someone is already rendering scene
func (sc *Scene) AddNode(parent *Node, ctrl NodeControl, children ...*Node) *Node {
	n := &Node{Ctrl: ctrl, Children: children}
	if parent == nil {
		parent = &sc.Root
	}
	parent.Children = append(parent.Children, n)
	return n
}

// Check if scene is in readonly state. You should use Update method instead of relying on this to property update live scene
func (sc *Scene) Locked() bool {
	return atomic.LoadInt32(&sc.lockCount) > 0
}

func (sc *Scene) processPhase(time float64, frame vmodel.Frame, world mgl32.Mat4, ph Phase) {
	atEnd := ph.Begin()
	if atEnd != nil {
		defer atEnd()
	}
	pi := ProcessInfo{Time: time, World: world, Frame: frame, Phase: ph, Visible: true}
	sc.processNode(&sc.Root, &pi)
}

func (sc *Scene) processNode(n *Node, piParent *ProcessInfo) {
	if n.Ctrl != nil {
		pi := *piParent
		pi.parent = piParent
		n.Ctrl.Process(&pi)
		if !pi.Visible {
			return
		}
		for _, ch := range n.Children {
			sc.processNode(ch, &pi)
		}
	} else {
		for _, ch := range n.Children {
			sc.processNode(ch, piParent)
		}
	}
}

func (sc *Scene) updatePending() {
	defer sc.mx.Unlock()
	var pending []func()
	pending, sc.pending = sc.pending, nil
	for _, ac := range pending {
		ac()
	}
}
