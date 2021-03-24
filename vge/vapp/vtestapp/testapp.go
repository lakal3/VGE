package vtestapp

import (
	"testing"

	"github.com/lakal3/vge/vge/vk"
)

var TestApp struct {
	Ctx          TestContext
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

func Init(ctx TestContext, name string, options ...TestOption) {
	TestApp.Ctx = ctx
	TestApp.options = append(TestApp.options, options...)
	TestApp.App = vk.NewApplication(ctx, name)
	if !TestApp.NoValidation {
		TestApp.App.AddValidation(ctx)
	}
	TestApp.owner.AddChild(TestApp.App)
	for _, opt := range TestApp.options {
		opt.InitOption()
	}
	TestApp.App.Init(ctx)
	TestApp.Dev = TestApp.App.NewDevice(ctx, TestApp.pdIndex)
	TestApp.owner.AddChild(TestApp.Dev)
}

func Terminate() {
	TestApp.owner.Dispose()
}

func AddChild(child vk.Disposable) {
	TestApp.owner.AddChild(child)
}
