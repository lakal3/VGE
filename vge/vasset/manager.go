package vasset

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/lakal3/vge/vge/vk"
)

type AssetManager struct {
	Loader Loader
	mx     *sync.Mutex
	assets map[string]interface{}
}

func NewAssetManager(l Loader) *AssetManager {
	return &AssetManager{Loader: l, mx: &sync.Mutex{}, assets: make(map[string]interface{})}
}

func (a *AssetManager) Dispose() {
	a.mx.Lock()
	defer a.mx.Unlock()
	for _, v := range a.assets {
		disp, ok := v.(vk.Disposable)
		if ok {
			disp.Dispose()
		}
	}
	a.assets = make(map[string]interface{})
}

// Get retrieves asset from cache. If asset is not found, contructor is called to retrieve asset
// Added path parameter to constructor for generic functions in 0.20.1
func (a *AssetManager) Get(path string, construct func(path string) (asset interface{}, err error)) (asset interface{}, err error) {
	a.mx.Lock()
	defer a.mx.Unlock()
	asset, ok := a.assets[path]
	if ok {
		return asset, nil
	}
	asset, err = construct(path)
	if err != nil {
		return nil, err
	}
	a.assets[path] = asset
	return asset, nil
}

func (a *AssetManager) Load(path string, construct func(content []byte) (asset interface{}, err error)) (asset interface{}, err error) {
	var rd io.ReadCloser
	var content []byte
	return a.Get(path, func(path string) (asset interface{}, err error) {
		rd, err = a.Loader.Open(path)
		if err != nil {
			return nil, err
		}
		defer rd.Close()
		content, err = ioutil.ReadAll(rd)
		if err != nil {
			return nil, err
		}
		return construct(content)
	})
}
