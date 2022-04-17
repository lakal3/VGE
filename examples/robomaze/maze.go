package main

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/materials/shadow"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vscene"
)

type Dir int

const (
	Left = Dir(1 << iota)
	Down
	Right
	Up
)

func (d Dir) prevDir() Dir {
	if d == Left {
		return Up
	}
	return d >> 1
}

func (d Dir) nextDir() Dir {
	if d == Up {
		return Left
	}
	return d << 1
}

type mazeSquare struct {
	open Dir
}

type maze struct {
	size        int
	squares     []*mazeSquare
	nFloor      *vscene.Node
	nFence      *vscene.Node
	nRoot       *vscene.Node
	nFloorRoot  *vscene.Node
	nCorner     *vscene.Node
	nLightPos   *vscene.Node
	nLampLight  *vscene.Node
	nRobots     *vscene.Node
	nTower      *vscene.Node
	nWall_1     *vscene.Node
	nWall_2     *vscene.Node
	robos       []*robo
	stains      []*stain
	walkControl *vapp.WalkControl
}

func (m *maze) getAt(x, y int) *mazeSquare {
	return m.squares[y*(m.size+2)+x]
}

func (m *maze) buildSquares() {
	m.squares = make([]*mazeSquare, (m.size+2)*(m.size+2))
	for idx, _ := range m.squares {
		m.squares[idx] = &mazeSquare{}
	}
	for idx := 0; idx < m.size+2; idx++ {
		m.getAt(0, idx).open = Left | Up | Down
		m.getAt(m.size+1, idx).open = Right | Up | Down
		m.getAt(idx, 0).open |= Right | Down | Left
		m.getAt(idx, m.size+1).open |= Right | Up | Left
	}
	m.randomize()
}

func (m *maze) randomize() {
	x := rand.Intn(m.size-2) + 2
	m.getAt(x, 1).open |= Up
	m.getAt(x, 2).open |= Down
	y := 2
	left := m.size*m.size - 2
	for left > 0 {
		s := m.getAt(x, y)
		if s.open == 0 {
			x = x + 1
			if x > m.size {
				x = 1
				y = y + 1
				if y > m.size {
					y = 1
				}
			}
			continue
		}
		switch rand.Intn(4) {
		case 0:
			if x > 1 && m.getAt(x-1, y).open == 0 {
				s.open |= Left
				x--
				left--
				m.getAt(x, y).open |= Right
				continue
			}
		case 1:
			if x < m.size && m.getAt(x+1, y).open == 0 {
				s.open |= Right
				x++
				left--
				m.getAt(x, y).open |= Left
				continue
			}
		case 2:
			if y > 1 && m.getAt(x, y-1).open == 0 {
				s.open |= Down
				y--
				left--
				m.getAt(x, y).open |= Up
				continue
			}
		case 3:
			if y < m.size && m.getAt(x, y+1).open == 0 {
				s.open |= Up
				y++
				left--
				m.getAt(x, y).open |= Down
				continue
			}
		}
		x, y = rand.Intn(m.size)+1, rand.Intn(m.size)+1
	}
	// Make last opening out
	m.getAt(x, m.size).open |= Up
	m.getAt(x, m.size+1).open |= Down
}

func buildMaze(sa *shrinkAmim, size int) {
	m := &maze{size: size}
	m.buildSquares()
	m.prepareScene()
	sa.m = m
}

func (m *maze) switchScene() {
	c := vscene.NewPerspectiveCamera(100)
	c.Position = mgl32.Vec3{float32(m.size)/2 + 1, 1, -0.5}
	c.Target = mgl32.Vec3{float32(m.size)/2 + 1, 1, -float32(m.size)/2 + 1}
	app.mainWnd.Scene.Update(func() {
		app.mainWnd.Env.Children = []*vscene.Node{vscene.NewNode(app.envMaze)}
		app.mainWnd.Model.Ctrl = app.probe
		app.mainWnd.Model.Children = []*vscene.Node{m.nRoot}
		app.mainWnd.Camera = c
		if app.orbitCamera {
			oc := vapp.OrbitControlFrom(c)
			oc.RegisterHandler(-100, app.mainWnd)
		} else {
			m.walkControl = vapp.WalkControlFrom(-100, app.mainWnd, c)
			m.walkControl.Adjust = func(pc *vscene.PerspectiveCamera) {
				if pc.Position[1] < 0.1 {
					pc.Position[1] = 0.1
				}
				pc.Target[1] = 1 + (pc.Position[1]-1)*0.5
			}
			app.mainWnd.Ui.Children = append(app.mainWnd.Ui.Children, vscene.NewNode(m.walkControl))
		}
	})
}

func (m *maze) prepareScene() {
	m.buildNodes()
	tf := app.mainWnd.GetSceneTime()
	endTime := tf + 2

	for y := 0; y <= m.size+1; y++ {
		for x := 0; x <= m.size+1; x++ {
			m.nFloorRoot.Children = append(m.nFloorRoot.Children, vscene.NewNode(
				&tileControl{m: m, endTime: endTime, x: x, y: y}, m.nFloor))
		}
	}
	if app.oil {
		m.nFloorRoot.Ctrl = &stainPainter{m: m}
	}
	// Add towers
	for idx := 0; idx < 4; idx++ {
		angle := float32(idx) * math.Pi / 2
		x, y := 0, 0
		if idx%2 == 1 {
			x = m.size + 2
		}
		if idx >= 2 {
			y = m.size + 2
		}
		m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
			&tileControl{m: m, endTime: endTime, x: x, y: y, angle: angle}, m.nTower))
	}
	// And walls
	for x := 0; x <= m.size+1; x++ {
		nw := m.nWall_1
		if x%3 == 2 {
			nw = m.nWall_2
		}
		m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
			&tileControl{m: m, endTime: endTime, x: x, y: 0}, nw))
		m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
			&tileControl{m: m, endTime: endTime, x: x, y: m.size + 2}, nw))
	}
	for y := 0; y <= m.size+1; y++ {
		nw := m.nWall_1
		if y%3 == 2 {
			nw = m.nWall_2
		}
		m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
			&tileControl{m: m, endTime: endTime, x: 0, y: y, angle: math.Pi / 2}, nw))
		m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
			&tileControl{m: m, endTime: endTime, x: m.size + 2, y: y, angle: math.Pi / 2}, nw))
	}

	for y := 1; y <= m.size+1; y++ {
		for x := 1; x <= m.size+1; x++ {
			m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
				&tileControl{m: m, x: x, y: y, endTime: endTime}, m.nCorner))
		}
	}
	for y := 1; y <= m.size+1; y++ {
		for x := 1; x <= m.size+1; x++ {
			if (m.getAt(x, y).open & Down) == 0 {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x, y: y, endTime: endTime}, m.nFence))
			}
			if (m.getAt(x, y).open & Left) == 0 {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x, y: y, endTime: endTime, angle: math.Pi / 2}, m.nFence))
			}
		}
	}
	for y := 1; y <= m.size+1; y += 4 {
		for x := 1; x <= m.size+1; x += 4 {
			if x <= m.size && y < m.size {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x, y: y, endTime: endTime},
					m.nLightPos, newLightNode(m, endTime)))
			}
			if x > 1 && y < m.size {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x - 1, y: y, endTime: endTime, angle: math.Pi / 2},
					m.nLightPos, newLightNode(m, endTime+rand.Float64()*10)))
			}
			if x < m.size && y > 1 {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x, y: y - 1, endTime: endTime, angle: -math.Pi / 2},
					m.nLightPos, newLightNode(m, endTime+rand.Float64()*10)))
			}
			if x > 1 && y > 1 {
				m.nRoot.Children = append(m.nRoot.Children, vscene.NewNode(
					&tileControl{m: m, x: x - 1, y: y - 1, endTime: endTime, angle: math.Pi},
					m.nLightPos, newLightNode(m, endTime+rand.Float64()*10)))
			}
		}
	}
	m.nRobots = vscene.NewNode(&newRoboCtrl{m: m, prevAdded: endTime - 2})
	m.nRoot.Children = append(m.nRoot.Children, m.nRobots)
}

func newLightNode(m *maze, onTime float64) *vscene.Node {
	onOff := &lightOnOff{onTime: onTime}
	tr := &vscene.TransformControl{Transform: mgl32.Translate3D(0.81, 1.85, -0.81)}
	lp := shadow.NewPointLight(vscene.PointLight{Intensity: mgl32.Vec3{1, 1, 0.7}, Attenuation: mgl32.Vec3{0, 0, 0.3}}, 512)
	if app.slowShadows {
		lp.MaxDistance, lp.UpdateDelay = 7, 15
	} else {
		lp.MaxDistance = 4
	}

	onOff.light = lp
	// We must add NoShadow control to lamp light element, otherwise it will shadow all light coming from point light
	return vscene.NewNode(vscene.NewMultiControl(onOff, tr, lp), vscene.NewNode(shadow.NoShadow{}, m.nLampLight))
}

func (m *maze) buildNodes() {
	m.nFloorRoot = vscene.NewNode(nil)
	m.nRoot = vscene.NewNode(nil, m.nFloorRoot)
	m.nFloor = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Floor"), true)
	m.nFence = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Fence"), true)
	m.nCorner = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Corner"), true)
	m.nTower = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Tower"), true)
	m.nWall_1 = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Wall_1"), true)
	m.nWall_2 = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("Wall_2"), true)
	m.nLightPos = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("LightPost"), true)
	m.nLampLight = vscene.NodeFromModel(app.fenceModel, app.fenceModel.FindNode("LampLight"), true)
}

type tileControl struct {
	m       *maze
	x, y    int
	endTime float64
	angle   float32
}

func (c *tileControl) Process(pi *vscene.ProcessInfo) {
	td := float32(c.endTime - pi.Time)
	pi.World = pi.World.Mul4(mgl32.Translate3D(float32(c.x), 0, float32(-c.y)))
	if c.angle != 0 {
		pi.World = pi.World.Mul4(mgl32.HomogRotate3DY(c.angle))
	}
	if td > 0 {
		pi.World = pi.World.Mul4(mgl32.Scale3D(1, 1/(1+td), 1))
	}
}

type lightOnOff struct {
	onTime float64
	light  *shadow.PointLight
}

func (l *lightOnOff) Process(pi *vscene.ProcessInfo) {
	if pi.Time < l.onTime {
		pi.Visible = false
		return
	}
	f := float32((pi.Time - l.onTime) * 2)
	if rand.Float32()*200 < f-5 {
		pi.Visible = false
		l.onTime = rand.Float64()*10 + pi.Time
		return
	}
	if f > 1.4 {
		f = 1.4
	}
	l.light.Intensity = mgl32.Vec3{f, f, f}
}
