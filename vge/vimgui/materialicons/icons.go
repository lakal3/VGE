package materialicons

import (
	"bufio"
	"bytes"
	_ "embed"
	"github.com/lakal3/vge/vge/vdraw"
	"strconv"
	"strings"
)

var Icons *vdraw.Font
var CodePoints map[string]rune

func LoadIcons() (err error) {
	if Icons != nil {
		return nil
	}
	Icons, err = vdraw.LoadFont(material_icons)
	if err != nil {
		return err
	}
	parseCodePoints()
	return nil
}

func GetRunes(iconNames ...string) (runes []rune) {
	for _, in := range iconNames {
		runes = append(runes, CodePoints[in])
	}
	return
}

func parseCodePoints() {
	CodePoints = make(map[string]rune, 3000)
	r := bytes.NewReader(code_points)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		t := sc.Text()
		idx := strings.IndexRune(t, ' ')
		if idx > 0 {
			r, _ := strconv.ParseInt(t[idx+1:], 16, 32)
			if r > 0 {
				CodePoints[t[:idx]] = rune(r)
			}
		}
	}
}

//go:embed MaterialIcons-Regular.ttf
var material_icons []byte

//go:embed MaterialIcons-Regular.codepoints
var code_points []byte
