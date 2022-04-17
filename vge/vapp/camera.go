package vapp

import (
	"image"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vscene"
)

type OrbitControl struct {
	PC          *vscene.PerspectiveCamera
	Yaw         float64
	Pitch       float64
	PanMode     Mods
	RotateMode  Mods
	ZoomMode    Mods
	Active      bool
	Sensitivity float64
	Clamp       func(point mgl32.Vec3, target bool) mgl32.Vec3

	prevPoint image.Point
}

func (c *OrbitControl) CameraProjection(size image.Point) (projection, view mgl32.Mat4) {
	return c.PC.CameraProjection(size)
}

func (c *OrbitControl) Handle(ev Event) {
	if !c.Active {
		return
	}
	mm, ok := ev.(*MouseMoveEvent)
	if ok {
		// fmt.Println(mm.CurrentMods)
		if c.RotateMode != 0 && mm.HasMods(c.RotateMode) {
			if c.prevPoint.X != 0 {
				c.Yaw += float64(mm.MousePos.X-c.prevPoint.X) * c.Sensitivity / 100
				c.Pitch += float64(mm.MousePos.Y-c.prevPoint.Y) * c.Sensitivity / 100
				c.Pitch = clamp(c.Pitch, -math.Pi/2+0.1, math.Pi/2-0.1)
				l := c.PC.Target.Sub(c.PC.Position).Len()
				position := c.PC.Target.Sub(c.GetDirection().Mul(-l))
				c.apply(c.PC.Target, position)
			}
			c.prevPoint = mm.MousePos
		} else if c.PanMode != 0 && mm.HasMods(c.PanMode) {
			if c.prevPoint.X != 0 {
				l := c.PC.Target.Sub(c.PC.Position).Len()
				dy := l * float32(mm.MousePos.Y-c.prevPoint.Y) * float32(c.Sensitivity) / 100
				dx := -float64(l) * float64(mm.MousePos.X-c.prevPoint.X) * c.Sensitivity / 100
				target := c.PC.Target.Add(mgl32.Vec3{float32(math.Cos(-c.Yaw) * dx), dy, float32(math.Sin(-c.Yaw) * dx)})
				position := c.PC.Target.Sub(c.GetDirection().Mul(-l))
				c.apply(target, position)
			}
			c.prevPoint = mm.MousePos
		} else {
			c.prevPoint = image.Point{}
		}
	}
	ms, ok := ev.(*ScrollEvent)
	if ok && ms.HasMods(c.ZoomMode) {
		l := c.PC.Target.Sub(c.PC.Position).Len()
		if ms.Range.Y < 0 {
			l = l * 0.95
		} else {
			l = l / 0.95
		}
		position := c.PC.Target.Sub(c.GetDirection().Mul(-l))
		c.apply(c.PC.Target, position)
	}
}

func clamp(a float64, min float64, max float64) float64 {
	if a > max {
		return max
	}
	if a < min {
		return min
	}
	return a
}

func (oc *OrbitControl) RegisterHandler(priority float64, win *RenderWindow) {
	RegisterHandler(priority, func(ev Event) (unregister bool) {
		if win.Closed() {
			return true
		}
		es, ok := ev.(SourcedEvent)
		if ok && es.IsSource(win) {
			oc.Handle(ev)
		}
		return false
	})
}

func NewOrbitControl(priority float64, win *RenderWindow) *OrbitControl {
	oc := &OrbitControl{RotateMode: MODMouseButton1, Yaw: 0.1, Pitch: 0.2, Sensitivity: 1,
		Active: true, PanMode: MODMouseButton2}
	oc.PC = vscene.NewPerspectiveCamera(1000)
	position := oc.PC.Target.Sub(oc.GetDirection().Mul(-1))
	if win != nil {
		win.Camera = oc

	}
	oc.apply(oc.PC.Target, position)
	return oc
}

func OrbitControlFrom(pc *vscene.PerspectiveCamera) *OrbitControl {
	oc := &OrbitControl{RotateMode: MODMouseButton1, Sensitivity: 1,
		Active: true, PanMode: MODMouseButton2}
	oc.PC = pc
	position, target := pc.Position, pc.Target
	dir := oc.PC.Position.Sub(target).Normalize()
	oc.Yaw = math.Atan2(float64(dir.X()), float64(dir.Z()))
	oc.Pitch = math.Asin(float64(dir.Y()))
	oc.apply(target, position)
	return oc
}

func (oc *OrbitControl) Zoom(sc *vscene.Scene) {
	bb := &vscene.BoudingBox{}
	sc.Process(0, vscene.NullFrame{}, bb)
	aabb, empty := bb.Get()
	if empty {
		oc.ZoomTo(mgl32.Vec3{}, 1)
	} else {
		oc.ZoomTo(aabb.Center(), aabb.Len())
	}
}

func (oc *OrbitControl) ZoomTo(target mgl32.Vec3, distance float32) {
	position := target.Sub(oc.GetDirection().Mul(-distance))
	oc.apply(target, position)
}

func (oc *OrbitControl) GetDirection() mgl32.Vec3 {
	return mgl32.Vec3{float32(math.Sin(oc.Yaw) * math.Cos(oc.Pitch)),
		float32(math.Sin(oc.Pitch)),
		float32(math.Cos(oc.Yaw) * math.Cos(oc.Pitch))}
}

func (c *OrbitControl) apply(target mgl32.Vec3, position mgl32.Vec3) {
	if c.Clamp != nil {
		c.PC.Target = c.Clamp(target, true)
		c.PC.Position = c.Clamp(position, false)
		return
	}
	c.PC.Target, c.PC.Position = target, position
}

type WalkControl struct {
	Active      bool
	Yaw         float64
	Speed       float64
	WalkSpeed   float64
	Adjust      func(pc *vscene.PerspectiveCamera)
	Keys        [6]GLFWKeyCode
	pc          *vscene.PerspectiveCamera
	win         *RenderWindow
	prevUpdate  float64
	To          [6]bool
	prevX       int
	updateMouse bool
}

func (c *WalkControl) Process(pi *vscene.ProcessInfo) {
	dif := pi.Time - c.prevUpdate
	dir := c.pc.Target.Sub(c.pc.Position)
	needAdjust := false
	if c.updateMouse {
		c.pc.Target = c.pc.Position.Add(mgl32.Vec3{float32(math.Sin(c.Yaw)), 0, float32(math.Cos(c.Yaw))}.Mul(dir.Len() * -1))
		needAdjust = true
		c.updateMouse = false
	}

	dir = dir.Normalize().Mul(float32(dif * c.WalkSpeed))
	if c.To[2] {
		dir = mgl32.Vec3{float32(math.Sin(c.Yaw - math.Pi/2)), 0, float32(math.Cos(c.Yaw - math.Pi/2))}.
			Mul(float32(dif * c.WalkSpeed))
		c.pc.Target = c.pc.Target.Add(dir)
		c.pc.Position = c.pc.Position.Add(dir)
		needAdjust = true
	}
	if c.To[3] {
		dir = mgl32.Vec3{float32(math.Sin(c.Yaw + math.Pi/2)), 0, float32(math.Cos(c.Yaw + math.Pi/2))}.
			Mul(float32(dif * c.WalkSpeed))
		c.pc.Target = c.pc.Target.Add(dir)
		c.pc.Position = c.pc.Position.Add(dir)
		needAdjust = true
	}
	if c.To[0] {
		c.pc.Target = c.pc.Target.Add(dir)
		c.pc.Position = c.pc.Position.Add(dir)
		needAdjust = true
	}
	if c.To[1] {
		c.pc.Target = c.pc.Target.Sub(dir)
		c.pc.Position = c.pc.Position.Sub(dir)
		needAdjust = true
	}
	if c.To[4] {
		dir = mgl32.Vec3{0, 1, 0}.Mul(float32(dif * c.WalkSpeed))
		c.pc.Target = c.pc.Target.Add(dir)
		c.pc.Position = c.pc.Position.Add(dir)
		needAdjust = true
	}
	if c.To[5] {
		dir = mgl32.Vec3{0, 1, 0}.Mul(float32(dif * c.WalkSpeed))
		c.pc.Target = c.pc.Target.Sub(dir)
		c.pc.Position = c.pc.Position.Sub(dir)
		needAdjust = true
	}
	c.prevUpdate = pi.Time
	if needAdjust && c.Adjust != nil {
		c.Adjust(c.pc)
	}
}

func (c *WalkControl) eventHandler(ev Event) (unregister bool) {
	kd, ok := ev.(*KeyDownEvent)
	if ok {
		for idx := 0; idx < 6; idx++ {
			if kd.KeyCode == c.Keys[idx] {
				c.To[idx] = true
			}
		}
	}
	ku, ok := ev.(*KeyUpEvent)
	if ok {
		for idx := 0; idx < 6; idx++ {
			if ku.KeyCode == c.Keys[idx] {
				c.To[idx] = false
			}
		}
	}
	mm, ok := ev.(*MouseMoveEvent)
	if ok && mm.IsWin(c.win) {
		// fmt.Println(mm.CurrentMods)
		if mm.HasMods(MODMouseButton2) {
			if c.prevX != 0 {
				c.Yaw += float64(mm.MousePos.X-c.prevX) * c.Speed / 100
				c.updateMouse = true
			}
			c.prevX = mm.MousePos.X
		} else {
			c.prevX = 0
		}
	}
	return !c.Active
}

func WalkControlFrom(priority float64, win *RenderWindow, pc *vscene.PerspectiveCamera) *WalkControl {
	oc := &WalkControl{win: win, Active: true, WalkSpeed: 3, Speed: 3, Keys: [6]GLFWKeyCode{'W', 'S', 'A', 'D', 'E', 'C'}}
	oc.pc = pc
	dir := pc.Position.Sub(pc.Target).Normalize()
	oc.Yaw = math.Atan2(float64(dir.X()), float64(dir.Z()))
	oc.prevUpdate = win.GetSceneTime()
	RegisterHandler(priority, oc.eventHandler)
	return oc
}
