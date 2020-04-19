package vasset

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Default loader for application
var DefaultLoader Loader = DirectoryLoader{Directory: "."}

type Loader interface {
	Open(filename string) (io.ReadCloser, error)
}

type DirectoryLoader struct {
	Directory string
}

func (d DirectoryLoader) Open(filename string) (io.ReadCloser, error) {
	fullName := filepath.Join(d.Directory, filename)
	return os.Open(fullName)
}

type MultiDirectorLoader struct {
	Directories []string
}

func NewMultiDirectorLoader(directories string) MultiDirectorLoader {
	return MultiDirectorLoader{Directories: strings.Split(directories, ";")}
}

func (d MultiDirectorLoader) Open(filename string) (io.ReadCloser, error) {
	for _, dr := range d.Directories {
		fullName := filepath.Join(dr, filename)
		_, err := os.Stat(fullName)
		if os.IsNotExist(err) {
			continue
		}
		return os.Open(fullName)
	}
	return nil, fmt.Errorf("Can't locate file %s from any of directories %s", filename, strings.Join(d.Directories, ";"))
}

// Make load request relative to main component path.
type SubDirLoader struct {
	L         Loader
	DirPrefix string
}

func (s SubDirLoader) Open(filename string) (io.ReadCloser, error) {
	return s.L.Open(filepath.Join(s.DirPrefix, filename))
}

// Load content using given loader. If loader is nil, DefaultLoader is used
func Load(path string, l Loader) (content []byte, err error) {
	if l == nil {
		l = DefaultLoader
	}
	rd, err := l.Open(path)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	return ioutil.ReadAll(rd)
}
