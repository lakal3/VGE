package vk

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

type Application struct {
	OnFatalError    func(fatalError error)
	owner           Owner
	hApp            hApplication
	hInst           hInstance
	dynamicIndexing bool
	fatalError      error
}

func (a *Application) setError(err error) {
	a.fatalError = err
	if a.OnFatalError != nil {
		a.OnFatalError(err)
	} else {
		panic("Fatal application error: " + err.Error())
	}
}

func (a *Application) isValid() bool {
	return a.fatalError == nil
}

func (a *Application) begin(callName string) (atEnd func()) {
	return nil
}

type Device struct {
	Props        DeviceInfo
	OnFatalError func(fatalError error)
	OnError      func(err error)
	hDev         hDevice
	owner        Owner
	keyMap       map[Key]interface{}
	mxQueue      *sync.Mutex
	mxMap        *sync.Mutex
	app          *Application
	fatalError   error
}

func (d *Device) setError(err error) {
	d.fatalError = err
	if d.OnFatalError != nil {
		d.OnFatalError(err)
	} else {
		panic("Fatal device error: " + err.Error())
	}
}

func (d *Device) isValid() bool {
	return d.hDev != 0 && d.fatalError == nil
}

func (d *Device) begin(callName string) (atEnd func()) {
	return nil
}

func DebugPoint(point string) {
	call_DebugPoint([]byte(point))
}

// VGEDllPath sets name of default vgelib path.
var VGEDllPath string = "vgelib.dll"

// GetDllPath gets full path of vgelib.dll (.so). You can override this function to match your preferences / OS.
// By default in Windows file name is kept as is. In linux -> vgelib will be converted to libvgelib and .dll -> .so
var GetDllPath = func() string {
	if runtime.GOOS == "linux" {
		p := strings.ReplaceAll(VGEDllPath, "vgelib", "libvgelib")
		return strings.ReplaceAll(p, ".dll", ".so")
	}
	return VGEDllPath
}

// AddValidationException register validation exception to ignore. Normally validation errors cause Vulkan API call to fail if
// validation is enabled for application and validation layer reports an error.
// Some errors are not always valid and call to AddValidationException can put validation message to ignore list
//
// In VGE validation ignore list is global, not per application instance
func AddValidationException(msgId int32) {
	call_AddValidationException(&errContext{}, msgId)
}

func NewApplication(name string) (*Application, error) {
	err := loadLib()
	if err != nil {
		return nil, err
	}
	AddValidationException(0x9cacd67a - 0x100000000) // UNASSIGNED-CoreValidation-DrawState-QueryNotReset
	// VGE bind cube and normal images on same descriptor slot using slot override in glsl
	AddValidationException(0xa44449d4 - 0x100000000) // VUID-vkCmdDrawIndexed-None-02699
	a := &Application{}
	var ec errContext
	call_NewApplication(&ec, []byte(name), &a.hApp)
	return a, ec.err
}

// AddValidation will load Vulkan Validations layers and register them when application is initialize.
// This call must be before actual application is initialized
// You must have Vulkan SDK installed on your machine because validation layers are part of Vulkan SDK, not driver API:s.
// See https://vulkan-tutorial.com/Drawing_a_triangle/Setup/Validation_layers for more info
//
// Validation layer can be also configured with Vulkan Configurator from Vulkan SDK. If you use Vulkan Configurator to validate application,
// you should not add validation layer to application.
func (a *Application) AddValidation() {
	if a.hInst != 0 {
		a.setError(errors.New("Already initialized"))
		return
	}
	call_AddValidation(a, a.hApp)
}

// AddDynamicDescriptors adds dynamics descriptor support to device.
// VGE 0.20 now sets update after bind to flag to dynamics descriptors that seems to allow at least ~500000 samples image per stage!
// See maxDescriptorSetUpdateAfterBindSamplers from Vulkan database for actual recorder limits
//
// This call must be done before any device is created.
// Note that device creating will fail if dynamic descriptor are not supported or request maxSize is too high
func (a *Application) AddDynamicDescriptors() {
	if a.hInst != 0 {
		a.setError(errors.New("Already initialized"))
		return
	}
	call_AddDynamicDescriptors(a, a.hApp)
	a.dynamicIndexing = true
}

// Initialize Vulkan application. Create Vulkan application and Vulkan instance.
// See https://gpuopen.com/understanding-vulkan-objects/ about Vulkan object and their dependencies
func (a *Application) Init() {
	if a.hInst != 0 {
		a.setError(errors.New("Already initialized"))
	}
	call_Application_Init(a, a.hApp, &a.hInst)
}

// Dispose Vulkan application. This will dispose device, instance and all resources bound to them.
// Disposing application is typically last call to Vulkan API
func (a *Application) Dispose() {
	if a.hApp != 0 {
		a.owner.Dispose()
		call_Disposable_Dispose(hDisposable(a.hApp))
		a.hInst, a.hApp = 0, 0
	}
}

// NewDevice allocates actual device that will be used to execute Vulkan rendering commands.
// pdIndex is index of physical device on your machine. 0 is default
func (a *Application) NewDevice(pdIndex int32) *Device {
	d := NewDevice(a, pdIndex)
	a.owner.AddChild(d)
	return d
}

// IsValid checks that application is created and not disposed.
// Can be used to validate application before calling any api requiring Vulkan Application or Vulkan Instance.
func (a *Application) IsValid() bool {
	if a.hInst == 0 {
		return false
	}
	return true
}

// List all physical devices available
func (a *Application) GetDevices() (result []DeviceInfo) {
	if !a.IsValid() {
		a.setError(errors.New("Application not initialized"))
		return nil
	}
	idx := int32(0)
	var di DeviceInfo
	call_Instance_GetPhysicalDevice(a, a.hInst, idx, &di)
	for di.Valid < 2 {
		result = append(result, di)
		idx++
		call_Instance_GetPhysicalDevice(a, a.hInst, idx, &di)
	}
	return result
}

// NewDevice will create new device from valid application.
// Unlike with app.NewDevice, you are now responsible of disposing device before disposing application.
// It is possible to use multiple devices. However, there is currently no support to directly copy assets between devices using this library
func NewDevice(app *Application, pdIndex int32) *Device {
	if !app.IsValid() {
		app.setError(errors.New("Application not initialized"))
		return nil
	}
	pds := app.GetDevices()
	if pdIndex < 0 || pdIndex >= int32(len(pds)) {
		app.setError(errors.New("No such device"))
		return nil
	}
	pd := pds[pdIndex]
	if pd.ReasonLen > 0 {
		app.setError(errors.New("Device misses support for: " + string(pd.Reason[:pd.ReasonLen])))
		return nil
	}
	d := &Device{keyMap: make(map[Key]interface{}), mxMap: &sync.Mutex{}, mxQueue: &sync.Mutex{}, app: app, Props: pd}
	call_Instance_NewDevice(app, app.hInst, pdIndex, &d.hDev)
	d.owner = NewOwner(true)
	return d
}

// Dispose device and all resources allocated from device
func (d *Device) Dispose() {
	if d.hDev != 0 {
		d.owner.Dispose()
		call_Disposable_Dispose(hDisposable(d.hDev))
		d.hDev = 0
	}
}

// NewMemoryPool creates a new memory pool that will be disposed when device is disposed. Safe for concurrent access.
func (d *Device) NewMemoryPool() *MemoryPool {
	mp := NewMemoryPool(d)
	d.owner.AddChild(mp)
	return mp
}

// Create a new sampler that will be disposed when device is disposed. Safe for concurrent access.
func (d *Device) NewSampler(mode SamplerAddressMode) *Sampler {
	s := NewSampler(d, mode)
	d.owner.AddChild(s)
	return s
}

// IsValid check that device is created and not disposed. Used to validate device before calling some Vulkan API requiring active device
func (d *Device) IsValid() bool {
	if d.hDev == 0 {
		return false
	}
	if d.fatalError != nil {
		return false
	}
	return true
}

// NewDescriptorLayout will create new DescriptorLayout that will be disposed when device is disposed. Safe for concurrent access.
func (d *Device) NewDescriptorLayout(descriptorType DescriptorType, stages ShaderStageFlags, elements uint32) *DescriptorLayout {
	dl := NewDescriptorLayout(d, descriptorType, stages, elements)
	d.owner.AddChild(dl)
	return dl
}

// NewDynamicDescriptorLayout will create new DescriptorLayout that will be disposed when device is disposed. Safe for concurrent access.
func (d *Device) NewDynamicDescriptorLayout(descriptorType DescriptorType, stages ShaderStageFlags,
	elements uint32, flags DescriptorBindingFlagBitsEXT) *DescriptorLayout {
	dl := NewDynamicDescriptorLayout(d, descriptorType, stages, elements, flags)
	d.owner.AddChild(dl)
	return dl
}

// See Owner.Get. Safe for concurrent access.
func (d *Device) Get(key Key, cons Constructor) interface{} {
	return d.owner.Get(key, cons)
}

// Application returns application that device was created from.
func (d *Device) Application() *Application {
	return d.app
}

// ReportError report non fatal usage error that can be recovered but most likely should be fixed. Override OnError to handle errors
// Default implementation will write them to os.Stderr
func (d *Device) ReportError(err error) {
	if d.OnError != nil {
		d.OnError(err)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "VGE error: ", err)
	}
}
