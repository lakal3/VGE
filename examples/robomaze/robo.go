package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vscene"
	"math"
	"math/rand"
)

type roboPos struct {
	x   int
	y   int
	dir Dir
}

type robo struct {
	n         *vscene.Node
	m         *maze
	visited   map[int]int
	current   roboPos
	next      roboPos
	walkStart float64
	lastDrop  float64
	done      bool
}

func (rp roboPos) nextPoint() roboPos {
	switch rp.dir {
	case Left:
		rp.x--
	case Right:
		rp.x++
	case Up:
		rp.y++
	case Down:
		rp.y--
	}
	return rp
}

func (rp roboPos) canMoveTo(m *maze, dir Dir) bool {
	switch dir {
	case Up:
		if rp.x%2 == 0 {
			return false
		}
		if rp.y%2 == 0 {
			return true
		}
		return rp.getAt(m).open&Up != 0
	case Down:
		if rp.x%2 == 1 {
			return false
		}
		if rp.y%2 == 1 {
			return true
		}
		return rp.getAt(m).open&Down != 0
	case Right:
		if rp.y%2 == 1 {
			return false
		}
		if rp.x%2 == 0 {
			return true
		}
		return rp.getAt(m).open&Right != 0
	case Left:
		if rp.y%2 == 0 {
			return false
		}
		if rp.x%2 == 1 {
			return true
		}
		return rp.getAt(m).open&Left != 0
	}
	panic("Invalid direction")
}

func (rp roboPos) getRot() float32 {
	switch rp.dir {
	case Left:
		return math.Pi / 2
	case Right:
		return -math.Pi / 2
	case Up:
		return math.Pi
	}
	return 0
}
func (rp roboPos) getAt(m *maze) *mazeSquare {
	return m.getAt(rp.x/2, rp.y/2)
}

func (rp roboPos) vp() int {
	return (rp.y/2)*100 + (rp.x / 2)
}

var roboScale = mgl32.HomogRotate3DX(math.Pi / 2).Mul4(mgl32.Scale3D(0.002, 0.002, 0.002))

func (r *robo) Process(pi *vscene.ProcessInfo) {
	if r.done {
		pi.Visible = false
		return
	}
	if app.oil {
		r.checkNewStain(pi.Time)
	}
	wd := pi.Time - r.walkStart
	if wd > 1 {
		r.nextDir()
		wd = 0
	}
	ag := lerpAngle(wd, r.current.getRot(), r.next.getRot())
	pi.World = pi.World.Mul4(mgl32.Translate3D(lerpFloat(wd, r.current.x, r.next.x), 0, -lerpFloat(wd, r.current.y, r.next.y))).
		Mul4(mgl32.HomogRotate3DY(ag)).Mul4(roboScale)
}

func lerpAngle(wd float64, a1 float32, a2 float32) float32 {
	ag := a1*(1-float32(wd)) + a2*float32(wd)
	if ag > math.Pi {
		return ag - math.Pi*2
	}
	return ag
}

func lerpFloat(wd float64, s1 int, s2 int) float32 {
	f1 := float32(s1)*0.5 + 0.25
	f2 := float32(s2)*0.5 + 0.25
	return f1*(1-float32(wd)) + f2*float32(wd)
}

func (rb *robo) nextDir() {
	mustTurn := true
	turning := rb.current.dir != rb.next.dir
	rb.current = rb.next
	if rb.current.y > 2*rb.m.size+1 {
		// Done
		rb.done = true
		return
	}
	rb.walkStart = app.mainWnd.GetSceneTime()
	if rb.current.canMoveTo(rb.m, rb.current.dir) {
		mustTurn = false
	}

	if mustTurn || !turning {
		if rb.current.canMoveTo(rb.m, rb.current.dir.prevDir()) {
			rb.next = rb.current
			rb.next.dir = rb.current.dir.prevDir()
			np := rb.next.nextPoint()
			if mustTurn || rb.shouldGoTo(rb.current, np) {
				return
			}
		}
		if rb.current.canMoveTo(rb.m, rb.current.dir.nextDir()) {
			rb.next = rb.current
			rb.next.dir = rb.current.dir.nextDir()
			np := rb.next.nextPoint()
			if mustTurn || rb.shouldGoTo(rb.current, np) {
				return
			}
		}
	}
	rb.next = rb.current.nextPoint()
	rb.visited[rb.next.vp()]++
}

func (r *robo) shouldGoTo(current roboPos, np roboPos) bool {
	v0 := r.visited[current.vp()]
	v1 := r.visited[np.vp()]
	if float64(v1+1)/float64(v0+v1+1) > rand.Float64() {
		return false
	}
	return true
}

func addRobo(m *maze, at float64) {
	if len(m.robos) > m.size*2 {
		return
	}
	rb := &robo{next: roboPos{y: 2, dir: Up}, m: m, visited: make(map[int]int), lastDrop: at}
	rb.next.x = rand.Intn(m.size)*2 + 3
	rb.nextDir()

	m.robos = append(m.robos, rb)
	rb.n = vscene.NodeFromModel(app.robotModel, app.robotModel.FindNode("droid"), true)
	app.mainWnd.Scene.Update(func() {
		m.nRobots.Children = append(m.nRobots.Children, vscene.NewNode(rb, rb.n))
	})
}

type newRoboCtrl struct {
	prevAdded float64
	m         *maze
}

func (n *newRoboCtrl) Process(pi *vscene.ProcessInfo) {
	if pi.Time-n.prevAdded > 3.7 {
		n.prevAdded = pi.Time
		go addRobo(n.m, pi.Time)
	}
}
