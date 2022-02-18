package vk

import "errors"

type FrameCache struct {
	shared    Owner
	dev       *Device
	Instances []*FrameInstance
	ac        *Allocator
	rings     map[Key]*frameRing
}

type frameRing struct {
	current   int
	recording bool
	instances [2]*FrameInstance
}

const hostMemOffset uint32 = 1024

func (f *FrameCache) Dispose() {
	if f.ac == nil {
		return
	}
	for _, fi := range f.Instances {
		fi.dispose()
	}
	for _, r := range f.rings {
		r.dispose()
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

type fImage struct {
	image      *AImage
	usage      ImageUsageFlags
	neededSize uint64
	alignment  uint32
	ranges     []ImageRange
	views      []*AImageView
}

type FrameInstance struct {
	Output VImage

	index       int
	perFrame    Owner
	fc          *FrameCache
	mx          *SpinLock
	state       int
	memories    map[uint32]fMemory
	descriptors map[*DescriptorLayout]fDescriptorPool
	slices      map[BufferUsageFlags]fSlice
	images      map[Key]fImage
	ring        Key
	rRings      []Key
}

func NewFrameCache(dev *Device, instances int) *FrameCache {
	fc := &FrameCache{shared: NewOwner(true), dev: dev}
	fc.ac = NewAllocator(dev)
	for idx := 0; idx < instances; idx++ {
		fc.Instances = append(fc.Instances, newFrameInstance(fc, idx, 0))
	}
	fc.rings = make(map[Key]*frameRing)
	return fc
}

func (fi *FrameInstance) Index() (current int, total int) {
	return fi.index, len(fi.fc.Instances)
}

func (fi *FrameInstance) GetRing(ring Key) (ringInstance *FrameInstance, recording bool) {
	r := fi.fc.rings[ring]
	if r == nil || r.current < 0 {
		return nil, false
	}
	return r.instances[r.current], r.recording
}

func (fi *FrameInstance) Get(key Key, ctor Constructor) interface{} {
	return fi.perFrame.Get(key, ctor)
}

func (fi *FrameInstance) Set(key Key, value interface{}) {
	fi.perFrame.Set(key, value)
}

func (fi *FrameInstance) GetShared(key Key, ctor Constructor) interface{} {
	return fi.fc.shared.Get(key, ctor)
}

func (fi *FrameInstance) Device() *Device {
	return fi.fc.dev
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

type FrameView struct {
	Layout ImageLayout
	Range  ImageRange
}

func (fi *FrameInstance) ReserveImage(key Key, usage ImageUsageFlags, description ImageDescription, views ...ImageRange) {
	if !fi.checkReserve() {
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	im := fi.images[key]
	if im.image == nil {
		im.image = fi.fc.ac.AllocImage(usage, description)
		im.usage = usage
		im.neededSize, im.alignment = im.image.Size()
		im.ranges = views
	} else {
		im.neededSize, _ = im.image.Size()
	}
	rem := im.neededSize % uint64(im.alignment)
	if rem > 0 {
		im.neededSize += uint64(im.alignment) - rem
	}
	fi.images[key] = im
}

func (fi *FrameInstance) AllocImage(key Key) (image *AImage, views []*AImageView) {
	if !fi.checkAlloc() {
		return nil, nil
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	im := fi.images[key]
	return im.image, im.views
}

func (fi *FrameInstance) AllocCommand(queue QueueFlags) *Command {
	cmd := NewCommand(fi.Device(), queue, true)
	fi.perFrame.AddChild(cmd)
	return cmd
}

func (fi *FrameInstance) Commit() {
	if !fi.checkReserve() {
		return
	}
	if fi.ring != 0 {
		fi.fc.dev.FatalError(errors.New("FrameInstance is not main instance"))
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	for _, ring := range fi.rRings {
		r := fi.fc.rings[ring]
		r.instances[r.current].commitOne()
	}
	fi.commitOne()
}

func (fi *FrameInstance) commitOne() {
	fi.state = 2
	fi.commitDescriptors()
	fi.reserveMemory()
	fi.commitBuffers()
	fi.commitImages()
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
	if fi.ring != 0 {
		fi.fc.dev.FatalError(errors.New("FrameInstance is not main instance"))
		return
	}
	fi.rRings = nil

	fi.state = 1
}

func (fi *FrameInstance) BeginRing(ring Key) {
	if fi.state != 1 {
		fi.fc.dev.FatalError(errors.New("FrameInstance not in Reserve state"))
		return
	}
	if fi.ring != 0 {
		fi.fc.dev.FatalError(errors.New("FrameInstance is not main instance"))
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()

	r := fi.fc.rings[ring]
	if r == nil {
		r = newRing(ring, fi.fc)
		fi.fc.rings[ring] = r
	} else {
		if r.recording {
			return
		}
	}
	fi.rRings = append(fi.rRings, ring)
	r.current++
	if r.current > 1 {
		r.current = 0
	}
	r.recording = true
	if r.instances[r.current].state == 4 {
		r.instances[r.current].cleanUp()
	}
	r.instances[r.current].state = 1
}

func (fi *FrameInstance) Freeze() {
	fi.checkAlloc()
	if fi.ring != 0 {
		fi.fc.dev.FatalError(errors.New("FrameInstance is not main instance"))
		return
	}
	fi.mx.Lock()
	defer fi.mx.Unlock()
	for _, ring := range fi.rRings {
		r := fi.fc.rings[ring]
		r.instances[r.current].state = 4
		r.recording = false
	}
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
	for idx, im := range fi.images {
		im.neededSize = 0
		fi.images[idx] = im
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
	for _, img := range fi.images {
		if img.image != nil {
			for _, v := range img.views {
				v.Dispose()
			}
			img.image.Dispose()
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
	fi.slices, fi.memories, fi.descriptors, fi.images = nil, nil, nil, nil
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
		mt := sl.buffer.MemoryType() + hostMemOffset
		m := fi.memories[mt]
		if m.realloc {
			sl.buffer.Dispose()
			sl.buffer = fi.fc.ac.AllocBuffer(usage, sl.neededSize, true)
			sl.buffer.Bind(m.mem, m.usedSize)
			m.usedSize += sl.neededSize
			fi.memories[mt] = m
		}
		sl.usedSize = 0
		fi.slices[usage] = sl
	}
}

func (fi *FrameInstance) commitImages() {
	for key, img := range fi.images {
		mt := img.image.MemoryType()
		m := fi.memories[mt]
		if m.realloc {
			if img.views != nil {
				for _, v := range img.views {
					v.Dispose()
				}
				desc := img.image.Description
				img.image.Dispose()
				img.image = fi.fc.ac.AllocImage(img.usage, desc)
			}
			img.views = make([]*AImageView, len(img.ranges))
			img.image.Bind(m.mem, m.usedSize)
			m.usedSize += img.neededSize
			for idx := range img.views {
				rg := img.ranges[idx]
				img.views[idx] = img.image.AllocView(rg)
			}
			fi.memories[mt] = m
			fi.images[key] = img
		}
	}
}

func (fi *FrameInstance) reserveMemory() {
	for _, sl := range fi.slices {
		if sl.neededSize != 0 {
			mt := sl.buffer.MemoryType() + hostMemOffset
			m := fi.memories[mt]
			m.neededSize += sl.neededSize
			sz, _ := sl.buffer.Size()
			if sz < sl.neededSize {
				m.realloc = true
			}
			fi.memories[mt] = m
		}
	}
	for _, img := range fi.images {
		if img.neededSize != 0 {
			mt := img.image.MemoryType()
			m := fi.memories[mt]
			m.neededSize += img.neededSize
			if img.views == nil {
				m.realloc = true
			}
			fi.memories[mt] = m
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
			if mt >= hostMemOffset {
				mem.mem = fi.fc.ac.AllocMemory(mem.neededSize, mt-hostMemOffset, true)
			} else {
				mem.mem = fi.fc.ac.AllocMemory(mem.neededSize, mt, false)
			}
		}
		fi.memories[mt] = mem
	}
}

func newFrameInstance(fc *FrameCache, idx int, ring Key) *FrameInstance {
	fi := &FrameInstance{fc: fc, index: idx}
	fi.descriptors = make(map[*DescriptorLayout]fDescriptorPool)
	fi.slices = make(map[BufferUsageFlags]fSlice)
	fi.memories = make(map[uint32]fMemory)
	fi.images = make(map[Key]fImage)
	fi.perFrame = NewOwner(true)
	fi.mx = &SpinLock{}
	fi.ring = ring
	return fi
}

func (r *frameRing) dispose() {
	if r.instances[0] != nil {
		r.instances[0].dispose()
	}
	if r.instances[1] != nil {
		r.instances[1].dispose()
	}
	r.instances = [2]*FrameInstance{nil, nil}
	r.current = -1
}

func newRing(key Key, fc *FrameCache) *frameRing {
	r := &frameRing{current: -1}
	r.instances[0] = newFrameInstance(fc, 0, key)
	r.instances[1] = newFrameInstance(fc, 1, key)
	return r
}
