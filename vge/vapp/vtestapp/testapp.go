package vtestapp

import (
	"testing"

	"github.com/lakal3/vge/vge/vk"
)

var TestApp struct {
	App          *vk.Application
	Dev          *vk.Device
	NoValidation bool
	owner        vk.Owner
	options      []TestOption
	pdIndex      int32
}

type TestContext struct {
	T *testing.T
}

func (t TestContext) SetError(err error) {
	t.T.Fatal("API call failed: ", err)
}

func (t TestContext) IsValid() bool {
	return true
}

func (t TestContext) Begin(callName string) (atEnd func()) {
	return nil
}

type TestOption interface {
	InitOption()
}

func Init(name string, options ...TestOption) (err error) {
	TestApp.options = options
	TestApp.App, err = vk.NewApplication(name)
	if !TestApp.NoValidation {
		TestApp.App.AddValidation()
	}
	TestApp.owner.AddChild(TestApp.App)
	for _, opt := range TestApp.options {
		opt.InitOption()
	}
	TestApp.App.Init()
	TestApp.Dev = TestApp.App.NewDevice(TestApp.pdIndex)
	TestApp.owner.AddChild(TestApp.Dev)
	return nil
}

func Terminate() {
	TestApp.owner.Dispose()
}

func AddChild(child vk.Disposable) {
	TestApp.owner.AddChild(child)
}
