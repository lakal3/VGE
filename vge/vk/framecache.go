package vk

import "errors"

type FrameCache struct {
	shared    Owner
	dev       *Device
	Instances []*FrameInstance
	ac        *Allocator
}

func (f *FrameCache) Dispose() {
	if f.ac == nil {
		return
	}
	for _, fi := range f.Instances {
		fi.dispose()
	}
	f.ac.Dispose()
	f.shared.Dispose()
	f.Instances, f.ac = nil, nil
}

type fMemory struct {
	mem        *AMemory
	neededSize uint64
	usedSize   uint64
	realloc    bool
	hostMem    bool
}

type fDescriptorPool struct {
	dp     *DescriptorPool
	layout *DescriptorLayout
	needed int
	used   int
	ds     []*DescriptorSet
}

type fSlice struct {
	buffer     *ABuffer
	neededSize uint64
	usedSize   uint64
	alignment  uint32
}

type FrameInstance struct {
	index       int
	perFrame    Owner
	fc          *FrameCache
	mx          *SpinLock
	state       int
	memories    map[uint32]fMemory
	descriptors map[*DescriptorLayout]fDescriptorPool
	slices      map[BufferUsageFlags]fSlice
}

func NewFrameCache(dev *Device, instances int) *FrameCache {
	fc := &FrameCache{shared: NewOwner(true), dev: dev}
	fc.ac = NewAllocator(dev)
	for idx := 0; idx < instances; idx++ {
		fc.Instances = append(fc.Instances, newFrameInstance(fc, idx))
	}
	return fc
}

func (fi *FrameInstance) Index() (current int, total int) {
	return fi.index, len(fi.fc.Instances)
}

func (fi *FrameInstance) Get(key Key, ctor Constructor) interface{} {
	return fi.perFrame.Get(key, ctor)
}

func (fi *FrameInstance) GetShared(key Key, ctor Constructor) interface{} {
	return fi.fc.shared.Get(key, ctor)
}

func (fi *FrameInstance) ReserveSlice(usage BufferUsageFlags, size uint64) {
	if !fi.checkReserve() {
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	b := fi.slices[usage]
	if b.alignment == 0 {
		// Need to allocate one buffer to check alignment
		b.buffer = fi.fc.ac.AllocBuffer(usage, size, true)
		_, b.alignment = b.buffer.Size()
	}
	rem := size % uint64(b.alignment)
	if rem > 0 {
		size += uint64(b.alignment) - rem
	}
	b.neededSize += size
	fi.slices[usage] = b
}

func (fi *FrameInstance) AllocSlice(usage BufferUsageFlags, size uint64) *ASlice {
	if !fi.checkAlloc() {
		return nil
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	sl := fi.slices[usage]
	a := sl.buffer.Slice(sl.usedSize, size)
	rem := size % uint64(sl.alignment)
	if rem > 0 {
		size += uint64(sl.alignment) - rem
	}
	sl.usedSize += size
	fi.slices[usage] = sl
	return a
}

func (fi *FrameInstance) ReserveDescriptor(layout *DescriptorLayout) {
	if !fi.checkReserve() {
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	b := fi.descriptors[layout]
	b.needed++
	fi.descriptors[layout] = b
}

func (fi *FrameInstance) AllocDescriptor(layout *DescriptorLayout) *DescriptorSet {
	if !fi.checkAlloc() {
		return nil
	}
	if !layout.isValid() {
		return nil
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	des := fi.descriptors[layout]
	ds := des.ds[des.used]
	des.used++
	fi.descriptors[layout] = des
	return ds
}

func (fi *FrameInstance) Commit() {
	if !fi.checkReserve() {
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	fi.state = 2
	fi.commitDescriptors()
	fi.reserveMemory()
	fi.commitBuffers()
	fi.state = 3
}

func (fi *FrameInstance) BeginFrame() {
	fi.mx.Lock()
	defer fi.mx.Unlock()
	if fi.state == 4 {
		fi.cleanUp()
		fi.state = 0
	}
	if fi.state != 0 {
		fi.fc.dev.FatalError(errors.New("FrameInstance not in Initial state"))
		return
	}
	fi.state = 1
}

func (fi *FrameInstance) Freeze() {
	fi.checkAlloc()
	fi.mx.Lock()
	defer fi.mx.Unlock()
	fi.state = 4
}

func (fi *FrameInstance) checkReserve() bool {
	if fi.state != 1 {
		fi.fc.dev.FatalError(errors.New("FrameInstance not in Reservation state"))
		return false
	}
	return true
}

func (fi *FrameInstance) checkAlloc() bool {
	if fi.state != 3 {
		fi.fc.dev.FatalError(errors.New("FrameInstance not in Allocation state"))
		return false
	}
	return true
}

func (fi *FrameInstance) cleanUp() {
	fi.perFrame.Dispose()
	for idx, sl := range fi.slices {
		sl.neededSize = 0
		fi.slices[idx] = sl
	}
	for idx, des := range fi.descriptors {
		des.needed = 0
		fi.descriptors[idx] = des
	}
	for idx, mem := range fi.memories {
		mem.neededSize, mem.realloc = 0, false
		fi.memories[idx] = mem
	}
}

func (fi *FrameInstance) dispose() {
	fi.perFrame.Dispose()
	for _, sl := range fi.slices {
		if sl.buffer != nil {
			sl.buffer.Dispose()
		}
	}
	for _, des := range fi.descriptors {
		if des.dp != nil {
			des.dp.Dispose()
		}
	}
	for _, mem := range fi.memories {
		if mem.mem != nil {
			mem.mem.Dispose()
		}
	}
	fi.slices, fi.memories, fi.descriptors = nil, nil, nil
}

func (fi *FrameInstance) commitDescriptors() {
	for la, des := range fi.descriptors {
		if len(des.ds) < des.needed {
			if des.dp != nil {
				des.dp.Dispose()
			}
			des.dp = NewDescriptorPool(la, des.needed)
			des.ds = make([]*DescriptorSet, des.needed)
			for idx := 0; idx < des.needed; idx++ {
				des.ds[idx] = des.dp.Alloc()
			}
		}
		des.used = 0
		fi.descriptors[la] = des
	}
}

func (fi *FrameInstance) commitBuffers() {
	for usage, sl := range fi.slices {
		m := fi.memories[sl.buffer.MemoryType()]
		if m.realloc {
			sl.buffer.Dispose()
			sl.buffer = fi.fc.ac.AllocBuffer(usage, sl.neededSize, true)
			sl.buffer.Bind(m.mem, m.usedSize)
			m.usedSize += sl.neededSize
			fi.memories[sl.buffer.MemoryType()] = m
		}
		sl.usedSize = 0
		fi.slices[usage] = sl
	}
}

func (fi *FrameInstance) reserveMemory() {
	for _, sl := range fi.slices {
		if sl.neededSize != 0 {
			m := fi.memories[sl.buffer.MemoryType()]
			m.neededSize += sl.neededSize
			m.hostMem = true
			sz, _ := sl.buffer.Size()
			if sz < sl.neededSize {
				m.realloc = true
			}
			fi.memories[sl.buffer.MemoryType()] = m
		}
	}
	for mt, mem := range fi.memories {
		mem.usedSize = 0
		if mem.mem != nil {
			oldSize := mem.mem.Size()
			if oldSize < mem.neededSize {
				mem.mem.Dispose()
				mem.mem = nil
			}
		}
		if mem.mem == nil {
			mem.realloc = true
			mem.mem = fi.fc.ac.AllocMemory(mem.neededSize, mt, mem.hostMem)
		}
		fi.memories[mt] = mem
	}
}

func newFrameInstance(fc *FrameCache, idx int) *FrameInstance {
	fi := &FrameInstance{fc: fc, index: idx}
	fi.descriptors = make(map[*DescriptorLayout]fDescriptorPool)
	fi.slices = make(map[BufferUsageFlags]fSlice)
	fi.memories = make(map[uint32]fMemory)
	fi.perFrame = NewOwner(true)
	fi.mx = &SpinLock{}
	return fi
}
