package shaders

import (
	"errors"
	"fmt"
	"github.com/lakal3/vge/vge/vk"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

type Flags map[string]string

func NewFlags() Flags {
	c := make(Flags)

	return c
}

func (c Flags) Set(cond string, val string) Flags {
	c[cond] = val
	return c
}

func (c Flags) Clear(cond string) Flags {
	delete(c, cond)
	return c
}

func (c Flags) AddAll(all Flags) {
	for k, v := range all {
		c[k] = v
	}
}

type ShaderTool struct {
	ErrOutput io.Writer
	Loader    Loader
	fr        map[string]string
}

var ErrNotFound = errors.New("File not found")

type Loader func(include string) (content []byte, err error)

func NewShaderTool() *ShaderTool {
	return &ShaderTool{fr: make(map[string]string), ErrOutput: os.Stderr}
}

func (st *ShaderTool) AddFragment(name string, source string) {
	st.fr[name] = source
}

func (st *ShaderTool) CompileConfig(dev *vk.Device, sp *Pack, config Config) (err error) {
	var code *SpirvCode
	if st.Loader == nil {
		st.Loader = NewDirectoryLoader(config.RootDir, config.Include)
	}
	for k, prog := range config.Programs {
		code, err = st.CompileProg(dev, prog)
		if err != nil {
			return err
		}
		sp.Add(k, code)
	}
	return nil
}

func NewDirectoryLoader(rootDir string, includes []string) Loader {
	return func(include string) (content []byte, err error) {
		for _, dir := range includes {
			fp := filepath.Join(rootDir, dir, include+".glsl")
			content, err = os.ReadFile(fp)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			return content, nil
		}
		return nil, ErrNotFound
	}

}

func (st *ShaderTool) CompileProg(dev *vk.Device, prog Program) (code *SpirvCode, err error) {
	fl := NewFlags()
	for _, flString := range prog.Flags {
		idx := strings.IndexRune(flString, '=')
		if idx > 0 {
			fl.Set(flString[:idx], flString[idx+1:])
		} else {
			fl.Set(flString, "1")
		}
	}
	code = &SpirvCode{}
	code.Vertex, err = st.compileOne(dev, vk.SHADERStageVertexBit, prog.Vertex, fl)
	if err != nil {
		return nil, err
	}
	code.Fragment, err = st.compileOne(dev, vk.SHADERStageFragmentBit, prog.Fragment, fl)
	if err != nil {
		return nil, err
	}
	code.Geometry, err = st.compileOne(dev, vk.SHADERStageGeometryBit, prog.Geometry, fl)
	if err != nil {
		return nil, err
	}
	code.Compute, err = st.compileOne(dev, vk.SHADERStageComputeBit, prog.Compute, fl)
	if err != nil {
		return nil, err
	}
	return code, nil
}

func (st *ShaderTool) compileOne(dev *vk.Device, bit vk.ShaderStageFlags, fragment string, flags Flags) ([]byte, error) {
	if len(fragment) == 0 {
		return nil, nil
	}
	spirv, code, err := st.Build(dev, bit, fragment, flags)
	if err != nil {
		_, _ = fmt.Fprintf(st.ErrOutput, "// at fragment %s\n ", fragment)
		for k, v := range flags {
			_, _ = fmt.Fprintf(st.ErrOutput, "// set %s = %s\n ", k, v)
		}
		st.formatCode(code)
		return nil, err
	}
	return spirv, nil
}

func (st *ShaderTool) formatCode(code string) {
	for idx, line := range strings.Split(code, "\n") {
		_, _ = fmt.Fprintf(st.ErrOutput, "%d. %s\n", idx+1, line)
	}

}

var kCompiler = vk.NewKey()

func (st *ShaderTool) Build(dev *vk.Device, stage vk.ShaderStageFlags, fragment string, flags Flags) (spirv []byte, builtCode string, err error) {
	var code strings.Builder
	err = st.parse(&code, fragment, flags)
	if err != nil {
		return nil, code.String(), err
	}
	compiler := dev.Get(kCompiler, func() interface{} {
		return vk.NewCompiler(dev)
	}).(*vk.GlslCompiler)
	spirv, _, err = compiler.Compile(stage, code.String())
	return spirv, code.String(), err
}

func (st *ShaderTool) parse(code *strings.Builder, fragment string, flags Flags) error {
	fr, err := st.GetFragment(fragment)
	if err != nil {
		return err
	}
	if len(fr) == 0 {
		return fmt.Errorf("No fragment %s", fragment)
	}
	rd := strings.NewReader(fr)
	var end bool
	end, err = st.buildFragment(code, rd, flags)
	if err != nil {
		return err
	}
	if end {
		return errors.New("Extra #end is fragment")
	}
	return nil
}

func (st *ShaderTool) GetFragment(name string) (string, error) {
	fr := st.fr[name]
	if len(fr) > 0 {
		return fr, nil
	}
	code, err := st.Loader(name)
	if err == ErrNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	st.fr[name] = string(code)
	return st.fr[name], nil
}

func (st *ShaderTool) buildFragment(sb *strings.Builder, rd *strings.Reader, flags Flags) (end bool, err error) {
	var ch rune
	for true {
		ch, _, err = rd.ReadRune()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if ch == '#' {
			end, err = st.buildCommand(sb, rd, flags)
			if end || err != nil {
				return end, err
			}
		} else {
			sb.WriteRune(ch)
		}
	}
	return false, nil
}

func (st *ShaderTool) buildCommand(sb *strings.Builder, rd *strings.Reader, flags Flags) (end bool, err error) {
	var sbCommand strings.Builder
	var ch rune
	for ch != '\n' && ch != '#' {
		ch, _, err = rd.ReadRune()
		if err != nil {
			return false, err
		}
		if ch == '#' {
			continue
		}
		if ch >= 32 || ch == '\t' {
			sbCommand.WriteRune(ch)
		}
	}
	sc := scanner.Scanner{}
	sc.Init(strings.NewReader(sbCommand.String()))
	if sc.Scan() != scanner.Ident {
		return false, fmt.Errorf("assumed command not %s", sc.TokenText())
	}
	mustFind := true
	var val string
	switch sc.TokenText() {
	case "end":
		fallthrough
	case "endif":
		return true, nil
	case "r":
		fallthrough
	case "replace":
		val, err = st.eval(&sc, flags)
		if err != nil {
			return false, err
		}
		sb.WriteString(val)
		return false, nil
	case "define":
		fallthrough
	case "set":
		if sc.Scan() != scanner.Ident {
			return false, fmt.Errorf("assumed variable name %s", sc.TokenText())
		}
		vn := sc.TokenText()
		val, err = st.eval(&sc, flags)
		if err != nil {
			return false, err
		}
		flags.Set(vn, val)
		return false, nil
	case "if":
		val, err = st.eval(&sc, flags)
		if err != nil {
			return false, err
		}
		if st.asBool(val) {
			end, err = st.buildFragment(sb, rd, flags)
			if err != nil {
				return false, err
			}
		} else {
			var sbTemp strings.Builder
			flTemp := NewFlags()
			flTemp.AddAll(flags)
			flTemp.Set("_skipped", "1")
			end, err = st.buildFragment(&sbTemp, rd, flTemp)
			if err != nil {
				return false, err
			}
		}

		if !end {
			return false, errors.New("Missign $end")
		}
		return false, nil
	case "undef":
		if sc.Scan() != scanner.Ident {
			return false, fmt.Errorf("assumed variable name %s", sc.TokenText())
		}
		vn := sc.TokenText()
		flags.Clear(vn)
	case "cinclude":
		mustFind = false
		fallthrough
	case "include":
		{
			if sc.Scan() != scanner.Ident {
				return false, fmt.Errorf("assumed fragment name %s", sc.TokenText())
			}
			fn := sc.TokenText()
			var f2 string
			f2, err = st.GetFragment(fn)
			if err != nil {
				if err == ErrNotFound {
					if mustFind {
						return false, fmt.Errorf("Can't find include %s", fn)
					}
					return false, nil
				}
				return false, err
			}
			rd2 := strings.NewReader(f2)
			return st.buildFragment(sb, rd2, flags)
		}
	case "version":
		fallthrough
	case "extension":
		sb.WriteRune('#')
		sb.WriteString(sbCommand.String())
		sb.WriteRune('\n')
		return false, nil
	}
	return false, fmt.Errorf("unknown command %s", sc.TokenText())
}

func (st *ShaderTool) addSubstitute(sb *strings.Builder, rd *strings.Reader, flags Flags) (err error) {
	var sbCond strings.Builder
	var ch rune
	for true {
		ch, _, err = rd.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			sbCond.WriteRune(ch)
		} else {
			flName := sbCond.String()
			if len(flName) == 0 {
				if ch == '#' {
					sb.WriteRune(ch)
					return nil
				}
				return errors.New("empty #")
			}
			sb.WriteString(flags[flName])
			return nil
		}
	}
	return nil
}

func (st *ShaderTool) eval(sc *scanner.Scanner, flags Flags) (val string, err error) {
	val, err = st.evalAtom(sc, flags)
	token := sc.Scan()
	tt := sc.TokenText()
	for token != scanner.EOF {
		var val2 string
		if token == ')' {
			return val, nil
		}
		val2, err = st.evalAtom(sc, flags)
		if err != nil {
			return "", err
		}
		switch token {
		case '+':
			val, err = st.numOp(val, val2, func(v1, v2 int) int {
				return v1 + v2
			})
			if err != nil {
				return "", err
			}
		case '&':
			val = st.boolOp(val, val2, func(v1, v2 bool) bool {
				return v1 && v2
			})
		case '|':
			val = st.boolOp(val, val2, func(v1, v2 bool) bool {
				return v1 && v2
			})
		default:
			return val, fmt.Errorf("unknwon token %s", tt)
		}
		token = sc.Scan()
		tt = sc.TokenText()
	}
	return val, nil
}

func (st *ShaderTool) evalAtom(sc *scanner.Scanner, flags Flags) (string, error) {
	token := sc.Scan()
	if token == '(' {
		return st.eval(sc, flags)
	}
	if token == '!' {
		val, err := st.eval(sc, flags)
		if err != nil {
			return "", err
		}
		if st.asBool(val) {
			return "", nil
		}
		return "1", nil
	}
	if token == scanner.Int {
		return sc.TokenText(), nil
	}
	if token == scanner.Ident {
		return flags[sc.TokenText()], nil
	}
	return "", fmt.Errorf("invalid atom %s", sc.TokenText())
}

func (st *ShaderTool) numOp(val string, val2 string, t func(v1, v2 int) int) (string, error) {
	v1, err := st.asNum(val)
	if err != nil {
		return "", err
	}
	v2, err := st.asNum(val2)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(t(v1, v2)), nil
}

func (st *ShaderTool) boolOp(val string, val2 string, t func(v1, v2 bool) bool) string {
	v1 := st.asBool(val)
	v2 := st.asBool(val2)
	if t(v1, v2) {
		return "1"
	}
	return ""
}

func (st *ShaderTool) asNum(val string) (int, error) {
	if len(val) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func (st *ShaderTool) asBool(val string) bool {
	if len(val) == 0 || val == "0" {
		return false
	}
	return true
}
