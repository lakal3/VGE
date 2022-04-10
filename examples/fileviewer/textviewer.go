package main

import (
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"os"
	"strings"
	"unicode/utf8"
)

type textViewer struct {
	path   string
	info   os.FileInfo
	lines  []string
	rm     func()
	view   *vimgui.View
	size   mgl32.Vec2
	fs     float32
	offset mgl32.Vec2
}

func (t *textViewer) Close() {
	t.rm()
}

func (t *textViewer) Stat(fr *vimgui.UIFrame) {
	DrawFileInfo(fr, t.path, t.info)
	fr.NewLine(-100, 20, 5)
	vimgui.Label(fr, fmt.Sprintf("%d lines", len(t.lines)))
}

func (t *textViewer) draw(fr *vimgui.UIFrame) {
	vimgui.ScrollArea(fr, t.size, &t.offset, func(uf *vimgui.UIFrame) {
		for _, l := range t.lines {
			uf.NewLine(-100, t.fs, 1)
			vimgui.Label(fr, l)
		}
	})
}

func parseText(path string, info os.FileInfo, content []byte) (isBinary bool) {
	s := mintheme.Theme.GetStyles("*label")
	fs := s.Get(vimgui.FontStyle{}).(vimgui.FontStyle)
	if fs.Font == nil {
		vapp.Dev.FatalError(errors.New("No font in style"))
	}

	tv := &textViewer{path: path, info: info, fs: fs.Size}
	var sb strings.Builder
	for idx := 0; idx < len(content); {
		b := content[idx]
		if b < 32 {
			switch b {
			case '\n':
				tv.lines = append(tv.lines, sb.String())
				sz, _ := fs.Font.MeasureText(fs.Size, sb.String())
				if sz[0] > tv.size[0] {
					tv.size[0] = sz[0]
				}
				sb.Reset()
			case '\r':
			case '\t':
				l := sb.Len()
				l = 4 - l%4
				for l > 0 {
					sb.WriteRune(' ')
					l--
				}
			default:
				return true
			}
			idx++
		} else {
			r, s := utf8.DecodeRune(content[idx:])
			sb.WriteRune(r)
			idx += s
		}
	}
	if sb.Len() > 0 {
		tv.lines = append(tv.lines, sb.String())
	}
	tv.view = vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, tv.draw)
	tv.view.OnSize = func(fi *vk.FrameInstance) vdraw.Area {
		desc := fi.Output.Describe()
		return vdraw.Area{From: mgl32.Vec2{float32(desc.Width)/4 + 1, StatHeight + 1}, To: mgl32.Vec2{float32(desc.Width), float32(desc.Height)}}
	}
	tv.size[1] = float32(len(tv.lines)) * (1 + fs.Size)
	tv.rm = SetViewer(tv, tv.view)
	return false
}
