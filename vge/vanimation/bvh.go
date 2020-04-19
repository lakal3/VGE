package vanimation

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vmodel"
	"math"
	"strconv"
	"strings"
	"text/scanner"
)

type BVHAnimation struct {
	err        error
	Root       *BvhJoint
	Joints     []*BvhJoint
	frames     [][]float32
	Values     map[string]float32
	noChannels int
}

type BvhJoint struct {
	Offset   mgl32.Vec3
	Name     string
	Children []*BvhJoint
	Len      mgl32.Vec3
	// Joint value Offset within a frame
	framePos int
	channels []string
	parent   int
	anim     *BVHAnimation
}

func (j *BvhJoint) GetRotation(frame int) mgl32.Mat4 {
	return j.anim.getJointRotation(j, frame)
}

func (j *BvhJoint) GetTranslation(frame int) mgl32.Mat4 {
	return j.anim.getJointTranslation(j, frame)
}

func (j *BvhJoint) GetPoseRotation() mgl32.Mat4 {
	if j.Len.Len() < 0.1 {
		return mgl32.Ident4()
	}
	jn := j.Len.Normalize()
	d := jn.Dot(mgl32.Vec3{0, 1, 0})
	t := jn.Cross(mgl32.Vec3{0, 1, 0})
	if t.Len() < 0.001 {
		if d < 0 { // Inverse
			return mgl32.HomogRotate3DX(math.Pi)
		}
		return mgl32.Ident4()
	}
	t = t.Normalize()
	angle := -float32(math.Acos(float64(d)))
	return mgl32.HomogRotate3D(angle, t)
}

func LoadBVH(l vasset.Loader, filename string) (*BVHAnimation, error) {
	rd, err := l.Open(filename)
	if err != nil {
		return nil, err
	}
	bvh := &BVHAnimation{Values: make(map[string]float32)}
	defer rd.Close()
	sc := &scanner.Scanner{Error: func(s *scanner.Scanner, msg string) {
		if bvh.err == nil {
			bvh.err = fmt.Errorf("Error %s at %d:%d", msg, s.Line, s.Column)
		}
	}}
	sc.Init(rd)
	err = bvh.read(sc)
	if err != nil {
		return nil, err
	}
	return bvh, bvh.err
}

// BuildAnimation tries to match skin bones orientation with BVH model orientation in world space for each recorded frame.
// BuildAnimation will create appropriate rotations to match target BVH bone direction. Therefore BVH will work even if original pose of BVH model and
// current model are different.
//
// However, currently this will only works if original skin don't have any rotation applied in bone direction
// and root bone (typically Hips) should have no rotation in pose position. If this not true you will see odd twists at some bones in animated model!
func (bvh *BVHAnimation) BuildAnimation(sk *vmodel.Skin, mj MapJoint) vmodel.Animation {
	ba := &buildAnimation{sk: sk}
	ft, ok := bvh.Values["FrameTime"]
	if !ok {
		ft = float32(1.0 / 60.0)
	}
	ba.joints = make(map[*BvhJoint]buildChannel)
	ba.inp = make([]float32, len(bvh.frames))
	for idx := 0; idx < len(ba.inp); idx++ {
		ba.inp[idx] = ft * float32(idx)
	}

	var rootJoint *BvhJoint
	for _, j := range bvh.Joints {
		jIdx := mj(sk, j.Name)
		if jIdx >= 0 {
			if jIdx == 0 {
				rootJoint = j
			}
			chanIndex := len(ba.an.Channels)
			ba.joints[j] = buildChannel{chanIndex: chanIndex, poseDif: sk.Joints[jIdx].Rotate.Mat4()}
			ba.an.Channels = append(ba.an.Channels, vmodel.Channel{Joint: jIdx, Input: ba.inp, Target: vmodel.TRotation})
		}
	}

	for frame := 0; frame < len(bvh.frames); frame++ {
		ba.buildChTime(frame, bvh.Root, mgl32.Ident4(), mgl32.Ident4())
	}

	if rootJoint != nil && rootJoint.Offset.Y() != 0 {
		// Add translation channel to root
		chMove := vmodel.Channel{Joint: 0, Input: ba.inp, Target: vmodel.TTranslation}
		scale := sk.Joints[0].Translate.Y() / rootJoint.Offset.Y()
		mScale := mgl32.Scale3D(scale, scale, scale)
		for frame := 0; frame < len(bvh.frames); frame++ {
			pos := mScale.Mul4(rootJoint.GetTranslation(frame)).Mul4x1(mgl32.Vec4{0, 0, 0, 1})
			chMove.Output = append(chMove.Output, pos[0], pos[1], pos[2])
		}
		ba.an.Channels = append(ba.an.Channels, chMove)
	}
	return ba.an
}

func (bvh *BVHAnimation) read(sc *scanner.Scanner) error {
	err := bvh.assume(sc, "HIERARCHY")
	if err != nil {
		return err
	}
	err = bvh.assume(sc, "ROOT")
	if err != nil {
		return err
	}
	bvh.Root, err = bvh.readJoint(sc, -1)
	if err != nil {
		return err
	}
	err = bvh.assume(sc, "MOTION")
	if err != nil {
		return err
	}
	found := true
	for found {
		found, err = bvh.nextOpt(sc)
		if err != nil {
			return err
		}
	}
	bvh.calcChannels()
	more := true
	for more {
		more, err = bvh.readFrame(sc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bvh *BVHAnimation) assume(sc *scanner.Scanner, token string) error {
	r := sc.Scan()
	if r != scanner.Ident {
		return fmt.Errorf("Assume ident not %s", sc.TokenText())
	}
	if sc.TokenText() != token {
		return fmt.Errorf("Assume %s not %s", token, sc.TokenText())
	}
	return nil
}

func (bvh *BVHAnimation) readJoint(sc *scanner.Scanner, parent int) (*BvhJoint, error) {
	if r := sc.Scan(); r != scanner.Ident {
		return nil, fmt.Errorf("Assumed Name not  %s", sc.TokenText())
	}
	j := &BvhJoint{parent: parent, anim: bvh}
	j.Name = sc.TokenText()
	if r := sc.Scan(); r != '{' {
		return nil, fmt.Errorf("Assumed { not  %s", sc.TokenText())
	}
	idx := len(bvh.Joints)
	bvh.Joints = append(bvh.Joints, j)
	rn := sc.Scan()
	var err error
	if rn == scanner.Ident && sc.TokenText() == "OFFSET" {
		v := mgl32.Vec3{}
		v[0], err = bvh.readValue(sc)
		if err != nil {
			return nil, err
		}
		v[1], err = bvh.readValue(sc)
		if err != nil {
			return nil, err
		}
		v[2], err = bvh.readValue(sc)
		if err != nil {
			return nil, err
		}
		j.Offset = bvh.getOffset(v)
		rn = sc.Scan()
	}
	if rn == scanner.Ident && sc.TokenText() == "CHANNELS" {
		f, err := bvh.readValue(sc)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(f); i++ {
			r := sc.Scan()
			if r != scanner.Ident {
				return nil, fmt.Errorf("Expected channel Name not %s", sc.TokenText())
			}
			j.channels = append(j.channels, sc.TokenText())
		}
		rn = sc.Scan()
	} else {
		return nil, fmt.Errorf("Missing CHANNELS, got %s", sc.TokenText())
	}
	for rn == scanner.Ident && sc.TokenText() == "JOINT" {
		ch, err := bvh.readJoint(sc, idx)
		if err != nil {
			return nil, err
		}
		if len(j.Children) == 0 {
			j.Len = ch.Offset
		}
		j.Children = append(j.Children, ch)
		rn = sc.Scan()
	}
	if rn == '}' {
		return j, nil
	}
	if rn != scanner.Ident || sc.TokenText() != "End" {
		return nil, fmt.Errorf("Assumed End site, not %s", sc.TokenText())
	}
	offset, err := bvh.parseEnd(sc)
	if err != nil {
		return j, err
	}
	j.Len = bvh.getOffset(offset)
	return j, nil
}

func (bvh *BVHAnimation) readValue(sc *scanner.Scanner) (float32, error) {
	r := sc.Scan()
	mult := float32(1)
	if r == '-' {
		mult = -1
		r = sc.Scan()
	}
	switch r {
	case scanner.Int:
		f, err := strconv.ParseFloat(sc.TokenText(), 32)
		return float32(f) * mult, err
	case scanner.Float:
		f, err := strconv.ParseFloat(sc.TokenText(), 32)
		return float32(f) * mult, err
	}
	return 0, fmt.Errorf("Assume number not %s", sc.TokenText())
}

func (bvh *BVHAnimation) parseEnd(sc *scanner.Scanner) (offset mgl32.Vec3, err error) {
	err = bvh.assume(sc, "Site")
	if err != nil {
		return
	}
	r := sc.Scan()
	if r != '{' {
		err = fmt.Errorf("Expected { not %s", sc.TokenText())
		return
	}
	err = bvh.assume(sc, "OFFSET")
	if err != nil {
		return
	}
	offset[0], err = bvh.readValue(sc)
	if err != nil {
		return
	}
	offset[1], err = bvh.readValue(sc)
	if err != nil {
		return
	}
	offset[2], err = bvh.readValue(sc)
	if err != nil {
		return
	}
	for idx := 1; idx <= 2; idx++ {
		// Two times }
		r = sc.Scan()
		if r != '}' {
			err = fmt.Errorf("Expected { not %s", sc.TokenText())
			return
		}
	}
	return
}

func (bvh *BVHAnimation) nextOpt(sc *scanner.Scanner) (bool, error) {
	rc := sc.Scan()
	if rc != scanner.Ident {
		return false, nil
	}
	sb := strings.Builder{}
	sb.WriteString(sc.TokenText())
	rc = sc.Scan()
	for rc != ':' {
		if rc != scanner.Ident {
			return false, fmt.Errorf("Assumed key part, not %s", sc.TokenText())
		}
		sb.WriteString(sc.TokenText())
		rc = sc.Scan()
	}
	v, err := bvh.readValue(sc)
	if err != nil {
		return false, err
	}
	bvh.Values[sb.String()] = v
	return true, nil
}

func (bvh *BVHAnimation) calcChannels() {
	for _, j := range bvh.Joints {
		j.framePos = bvh.noChannels
		bvh.noChannels += len(j.channels)
	}
}

func (bvh *BVHAnimation) readFrame(sc *scanner.Scanner) (bool, error) {
	frame := make([]float32, bvh.noChannels)
	var rn rune
	for i := 0; i < bvh.noChannels; i++ {
		mult := float32(1)
		if sc.TokenText() == "-" {
			mult = float32(-1)
			rn = sc.Scan()
		}
		v, err := strconv.ParseFloat(sc.TokenText(), 32)
		if err != nil {
			return false, err
		}
		frame[i] = float32(v) * mult
		rn = sc.Scan()
	}
	bvh.frames = append(bvh.frames, frame)
	return rn != scanner.EOF, nil
}

func (bvh *BVHAnimation) getJointTranslation(j *BvhJoint, frame int) mgl32.Mat4 {
	var x, y, z float32
	for cIdx, ch := range j.channels {
		idx := cIdx + j.framePos
		if ch == "Xposition" {
			x = bvh.frames[frame][idx]
		}
		if ch == "Yposition" {
			y = bvh.frames[frame][idx]
		}
		if ch == "Zposition" {
			z = bvh.frames[frame][idx]
		}
	}
	return mgl32.Translate3D(x, y, z)
}

func (bvh *BVHAnimation) getJointRotation(j *BvhJoint, frame int) mgl32.Mat4 {
	m := mgl32.Ident4()
	for cIdx, ch := range j.channels {
		idx := cIdx + j.framePos
		if ch == "Xrotation" {
			m = m.Mul4(mgl32.HomogRotate3DX(toRad(bvh.frames[frame][idx])))
		}
		if ch == "Yrotation" {
			m = m.Mul4(mgl32.HomogRotate3DY(toRad(bvh.frames[frame][idx])))
		}
		if ch == "Zrotation" {
			m = m.Mul4(mgl32.HomogRotate3DZ(toRad(bvh.frames[frame][idx])))
		}
	}
	return m
}

func (bvh *BVHAnimation) getOffset(offset mgl32.Vec3) mgl32.Vec3 {

	return offset
}

func (ba *buildAnimation) buildChTime(frame int, j *BvhJoint, parentPos mgl32.Mat4, parentFin mgl32.Mat4) {
	chanInfo, ok := ba.joints[j]

	rOff := j.GetRotation(frame)

	tr := parentPos.Mul4(mgl32.Translate3D(j.Offset[0], j.Offset[1], j.Offset[2])).Mul4(rOff)
	bRot := j.GetPoseRotation()
	rFin := tr.Mul4(bRot)
	if ok {
		rLoc := parentFin.Inv().Mul4(rFin)
		/*
			vd := chanInfo.poseDif.Mul4x1(mgl32.Vec4{1,0,0,0})
			a := math.Atan2(float64(vd[1]), float64(vd[0]))
			if (a > 0.3 || a < -0.3 || j.Name == "LeftUpLeg") && frame == 60 {
				fmt.Println(j.Name, "  ", j.framePos)
				fmt.Println(a, a * 180 / math.Pi)
			}
			if (a > 0.3 || a < -0.3) {
				boneRoll := mgl32.HomogRotate3DY(-float32(a))
				rLoc = rLoc.Mul4(boneRoll)
				// tr = tr.Mul4(boneRoll)
				rFin = rFin.Mul4(boneRoll)
			}

		*/
		q := mgl32.Mat4ToQuat(rLoc)

		ba.an.Channels[chanInfo.chanIndex].Output = append(ba.an.Channels[chanInfo.chanIndex].Output,
			q.V[0], q.V[1], q.V[2], q.W)
	}
	for _, ch := range j.Children {
		ba.buildChTime(frame, ch, tr, rFin)
	}
}

func toRad(f float32) float32 {
	return f * math.Pi / 180.0
}

type buildAnimation struct {
	an     vmodel.Animation
	joints map[*BvhJoint]buildChannel
	inp    []float32
	sk     *vmodel.Skin
}

type buildChannel struct {
	chanIndex int
	poseDif   mgl32.Mat4
}
