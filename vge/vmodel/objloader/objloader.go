package objloader

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vasset"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type ObjLoader struct {
	Builder   *vmodel.ModelBuilder
	Loader    vasset.Loader
	Parent    *vmodel.NodeBuilder
	materials map[string]vmodel.MaterialIndex
	images    map[string]vmodel.ImageIndex
	mat       vmodel.MaterialIndex
	objName   string
	position  []mgl32.Vec3
	normal    []mgl32.Vec3
	uv        []mgl32.Vec2
	faces     []vertex
}

type vertex struct {
	position int
	uv       int
	normal   int
	fourth   bool
}

func (ol *ObjLoader) LoadFile(filename string) error {
	if ol.Loader == nil {
		return errors.New("Set loader")
	}
	rd, err := ol.Loader.Open(filename)
	if err != nil {
		return err
	}
	defer rd.Close()
	return ol.Load(rd)
}

func (ol *ObjLoader) Load(reader io.Reader) (err error) {

	if ol.Builder == nil {
		ol.Builder = &vmodel.ModelBuilder{}
	}
	ol.materials = make(map[string]vmodel.MaterialIndex)
	ol.mat = vmodel.MaterialIndex(-1)
	ol.images = make(map[string]vmodel.ImageIndex)
	sc := bufio.NewScanner(reader)

	for sc.Scan() {
		parts := ol.splitLine(sc.Text())
		if len(parts) > 0 {
			err := ol.processLine(parts)
			if err != nil {
				return fmt.Errorf("At line %s: %v", sc.Text(), err)
			}
		}
	}
	err = ol.buildObject("")
	return err
}

func (ol *ObjLoader) splitLine(text string) (result []string) {
	sb := &strings.Builder{}
	for _, r := range text {
		if r == '#' {
			if sb.Len() > 0 {
				result = append(result, sb.String())
			}
			return
		}
		if r == ' ' || r == '\t' {
			if sb.Len() > 0 {
				result = append(result, sb.String())
			}
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}
	if sb.Len() > 0 {
		result = append(result, sb.String())
	}
	return
}

func (ol *ObjLoader) processLine(parts []string) (err error) {
	switch parts[0] {
	case "v":
		ol.position, err = ol.pushv3(ol.position, parts)
	case "vt":
		ol.uv, err = ol.pushv2(ol.uv, parts)
	case "vn":
		ol.normal, err = ol.pushv3(ol.normal, parts)
	case "f":
		ol.faces, err = ol.pushFaces(ol.faces, parts)
	case "o":
		if len(parts) < 2 {
			return errors.New("Object without name")
		}
		err = ol.buildObject(parts[1])
	case "mtllib":
		if len(parts) < 2 {
			return errors.New("Matlib without file")
		}
		err = ol.loadMatLib(parts[1])
	case "usemtl":
		if len(parts) < 2 {
			return errors.New("Usemat without material")
		}
		m, ok := ol.materials[parts[1]]
		if !ok {
			return fmt.Errorf("Undefined material %s", parts[1])
		}
		ol.mat = m
	case "s":
	default:
		return fmt.Errorf("Unknown command %s", parts[0])
	}
	return
}

func (ol *ObjLoader) pushv3(ar []mgl32.Vec3, parts []string) ([]mgl32.Vec3, error) {
	if len(parts) != 4 {
		return nil, fmt.Errorf("Line should have 3 floats")
	}
	var v mgl32.Vec3
	for idx := 1; idx <= 3; idx++ {
		f, err := strconv.ParseFloat(parts[idx], 32)
		if err != nil {
			return nil, err
		}
		v[idx-1] = float32(f)
	}
	return append(ar, v), nil
}

func (ol *ObjLoader) pushv2(ar []mgl32.Vec2, parts []string) ([]mgl32.Vec2, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("Line should have 2 floats")
	}
	var v mgl32.Vec2
	for idx := 1; idx <= 2; idx++ {
		f, err := strconv.ParseFloat(parts[idx], 32)
		if err != nil {
			return nil, err
		}
		v[idx-1] = float32(f)
	}
	return append(ar, v), nil
}

func (ol *ObjLoader) pushFaces(vertices []vertex, parts []string) ([]vertex, error) {
	if len(parts) != 4 && len(parts) != 5 {
		return nil, errors.New("Vertex count per face must be 3 or 4")
	}
	for idx := 1; idx < len(parts); idx++ {
		v, err := ol.parseFace(parts[idx])
		if err != nil {
			return nil, err
		}
		if idx == 4 {
			v.fourth = true
		}
		vertices = append(vertices, v)

	}
	return vertices, nil
}

func (ol *ObjLoader) parseFace(s string) (vertex, error) {
	v := vertex{}
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return v, errors.New("Vertex must have 3 values separated with /")
	}
	var err error
	v.position, err = strconv.Atoi(parts[0])
	v.position--
	v.uv, _ = strconv.Atoi(parts[1])
	v.uv--
	v.normal, _ = strconv.Atoi(parts[2])
	v.normal--
	return v, err
}

func (ol *ObjLoader) buildObject(newName string) (err error) {
	if len(ol.faces) == 0 {
		ol.objName = newName
		return nil
	}

	mesh := &vmodel.MeshBuilder{}
	for _, f := range ol.faces {
		vb := mesh.AddVertex(ol.position[f.position])
		// mesh.AABB.Add(mgl32.Vec3{ol.position[f.position*3], ol.position[f.position*3+1], ol.position[f.position*3+2]})
		if f.normal >= 0 {
			vb.AddNormal(ol.normal[f.normal])
		}
		if f.uv >= 0 {
			vb.AddUV(ol.uv[f.uv])
		}
		if f.fourth {
			// Position 0, 2 and 3 form second triangle
			mesh.AddIndex(vb.Index-3, vb.Index-1, vb.Index)
		} else {
			mesh.AddIndex(vb.Index)
		}
	}

	meshIndex := ol.Builder.AddMesh(mesh)
	if ol.mat < 0 {
		ol.mat = ol.Builder.AddMaterial("_default", vmodel.NewMaterialProperties())
	}
	n := ol.Builder.AddNode(newName, ol.Parent, mgl32.Ident4())
	n.SetMesh(meshIndex, ol.mat)
	ol.faces = nil
	ol.objName = newName
	return nil
}

func (ol *ObjLoader) loadMatLib(lib string) error {
	if ol.Loader == nil {
		return errors.New("Set Loader for material library")
	}
	r, err := ol.Loader.Open(lib)
	if err != nil {
		return err
	}
	defer r.Close()
	ml := &matLoader{parent: ol}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		parts := ol.splitLine(sc.Text())
		if len(parts) > 0 {
			err = ml.processMatLine(parts)
		}
		if err != nil {
			return fmt.Errorf("Error in material file (%s) line %s: %v", lib, sc.Text(), err)
		}
	}
	return ml.addMat()
}

func (ol *ObjLoader) LoadImage(path string) (vmodel.ImageIndex, error) {
	im, ok := ol.images[path]
	if ok {
		return im, nil
	}
	rd, err := ol.Loader.Open(path)
	if err != nil {
		return 0, err
	}
	defer rd.Close()
	content, err := ioutil.ReadAll(rd)
	if err != nil {
		return 0, err
	}
	kind := filepath.Ext(path)[1:]
	return ol.Builder.AddImage(kind, content, vk.IMAGEUsageSampledBit|vk.IMAGEUsageTransferDstBit), nil
}

type matLoader struct {
	parent  *ObjLoader
	matName string
	props   vmodel.MaterialProperties
}

func (ml *matLoader) processMatLine(parts []string) (err error) {
	var v mgl32.Vec3
	var f float32
	var img vmodel.ImageIndex
	switch parts[0] {
	case "newmtl":
		if len(parts) < 2 {
			return errors.New("Newmat without name")
		}
		err = ml.addMat()
		if err != nil {
			return err
		}
		ml.matName = parts[1]
		ml.props = vmodel.NewMaterialProperties()
	case "Ns":
		f, err = ml.parseFloat(parts)
		if err != nil {
			return err
		}
		_ = f
		// ml.mat.SetFactor(vscene.FSpeculaPower, f)
	case "Ka":
	case "Kd":
		v, err = ml.parseVec3(parts)
		if err != nil {
			return err
		}
		ml.props.SetColor(vmodel.CAlbedo, v.Vec4(1))
	case "Ks":
		v, err = ml.parseVec3(parts)
		if err != nil {
			return err
		}
		ml.props.SetColor(vmodel.CSpecular, v.Vec4(1))
	case "Ke":
		v, err = ml.parseVec3(parts)
		if err != nil {
			return err
		}
		ml.props.SetColor(vmodel.CEmissive, v.Vec4(1))
	case "Ni":
	case "d":
	case "illum":
	case "map_Bump":
		img, err = ml.loadImage(parts)
		if err != nil {
			return err
		}
		ml.props.SetImage(vmodel.TxBump, img)
	case "map_Kd":
		img, err = ml.loadImage(parts)
		if err != nil {
			return err
		}
		ml.props.SetImage(vmodel.TxAlbedo, img)
	case "map_Ks":
		img, err = ml.loadImage(parts)
		if err != nil {
			return err
		}
		ml.props.SetImage(vmodel.TxSpecular, img)
	case "map_Ke":
		img, err = ml.loadImage(parts)
		if err != nil {
			return err
		}
		ml.props.SetImage(vmodel.TxEmissive, img)
	}
	return
}

func (ml *matLoader) addMat() (err error) {
	if ml.props != nil {
		matIndex := ml.parent.Builder.AddMaterial(ml.matName, ml.props)
		ml.parent.materials[ml.matName] = matIndex
		ml.props = nil
	}
	return nil
}

func (ml *matLoader) parseFloat(parts []string) (float32, error) {
	if len(parts) < 2 {
		return 0, errors.New("Missing value")
	}
	f, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func (ml *matLoader) parseVec3(parts []string) (mgl32.Vec3, error) {
	r := mgl32.Vec3{}
	if len(parts) < 4 {
		return r, errors.New("Missing value(s)")
	}
	for idx := 0; idx < 3; idx++ {
		f, err := strconv.ParseFloat(parts[1+idx], 32)
		if err != nil {
			return r, err
		}
		r[idx] = float32(f)
	}
	return r, nil
}

func (ml *matLoader) loadImage(parts []string) (vmodel.ImageIndex, error) {
	if len(parts) < 2 {
		return 0, errors.New("Missing file name")
	}
	return ml.parent.LoadImage(parts[1])
}
