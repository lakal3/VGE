package main

import (
	"errors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/mintheme"
	"github.com/lakal3/vge/vge/vk"
	"image"
	"io/fs"
	"os"
	"path/filepath"
)

var kGlyphs = vk.NewKey()

func addFileTree(dir string) error {
	err := mintheme.BuildMinTheme()
	if err != nil {
		return err
	}
	if settings.glyphFont {
		vapp.Dev.Get(kGlyphs, func() interface{} {
			gs := &vdraw.GlyphSet{}
			// Convert characters from range 33 to 255 into glyph font
			err = gs.AddRunes(mintheme.PrimaryFont, 33, 255)
			if err != nil {
				return nil
			}
			gs.Build(vapp.Dev, image.Pt(32, 32))
			return gs
		})
	}
	tf := &fileTree{update: make(chan currentDir, 10)}
	tf.fillFiles(dir)
	fv := vimgui.NewView(vapp.Dev, vimgui.VMNormal, mintheme.Theme, tf.draw)
	fv.OnSize = func(fullArea vdraw.Area) vdraw.Area {
		return vdraw.Area{To: mgl32.Vec2{fullArea.Width()/4 - 10, fullArea.Height()}}
	}
	app.rw.AddView(fv)
	return nil
}

type currentDir struct {
	dir     string
	files   []os.DirEntry
	fileErr error
}

type fileTree struct {
	currentDir
	offset   mgl32.Vec2
	update   chan currentDir
	selected int
	mode     int
}

var kFile = vk.NewKey()

func (ft *fileTree) draw(fr *vimgui.UIFrame) {
	select {
	case newFiles := <-ft.update:
		ft.files, ft.fileErr, ft.dir = newFiles.files, newFiles.fileErr, newFiles.dir
		ft.offset = mgl32.Vec2{}
		ft.selected, ft.mode = -1, 0
	default:
	}

	fr.NewLine(-45, 28, 2)
	vimgui.TabButton(fr, kFile, "Files", 0, &ft.mode)
	if len(app.drives) > 0 {
		fr.NewColumn(-45, 5)
		vimgui.TabButton(fr, kFile, "Drives", 1, &ft.mode)
	}

	fr.NewLine(0, 0, 2)
	if ft.mode == 1 {
		for idx, di := range app.drives {
			fr.NewLine(0, 20, 2)
			fr.NewColumn(-100, 3)
			tmp := -1
			if vimgui.ToggleButton(fr, kFile+vk.Key(idx)+2, "tab", di, idx, &tmp) {
				ft.fillFiles(di + "/")
			}
		}
		return
	}
	fr.DrawArea.From = fr.ControlArea.From
	vimgui.ScrollArea(fr, mgl32.Vec2{0, 22 * float32(len(ft.files))}, &ft.offset, func(uf *vimgui.UIFrame) {
		for idx, di := range ft.files {

			uf.NewLine(0, 20, 2)
			uf.NewColumn(-100, 3)
			n := di.Name()
			if di.IsDir() {
				n = "[" + n + "]"
			}
			if vimgui.ToggleButton(uf, kFile+vk.Key(idx)+2, "tab", n, idx, &ft.selected) {
				ft.selectFile()
			}
		}

	})
}

type tempEntry struct {
	n string
}

func (t tempEntry) Name() string {
	return t.n
}

func (t tempEntry) IsDir() bool {
	return true
}

func (t tempEntry) Type() fs.FileMode {
	return 0
}

func (t tempEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("Not implemented")
}

func (ft *fileTree) fillFiles(dir string) {
	var files []os.DirEntry

	if !isRoot(dir) {
		files = []os.DirEntry{tempEntry{n: ".."}}
	}
	actualFiles, err := os.ReadDir(dir)
	if err != nil {
		ft.update <- currentDir{fileErr: err}
		return
	}

	for _, f := range actualFiles {
		if f.IsDir() {
			files = append(files, f)
		}
	}
	for _, f := range actualFiles {
		if !f.IsDir() {
			files = append(files, f)
		}
	}
	ft.update <- currentDir{files: files, dir: dir}
}

func isRoot(dir string) bool {
	return filepath.Dir(dir) == dir
}

func (ft *fileTree) selectFile() {
	e := ft.files[ft.selected]
	if e.IsDir() {
		var dir string
		if e.Name() == ".." {
			dir = filepath.Dir(ft.dir)
		} else if isRoot(e.Name()) {
			dir = e.Name()
		} else {
			dir = filepath.Join(ft.dir, e.Name())
		}
		go ft.fillFiles(dir)
	} else {
		viewFile(filepath.Join(ft.dir, e.Name()))
	}
}
