//+build linux

package dldyn

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>

void *invoke9(void *addr, void *arg1, void *arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7, void *arg8, void *arg9) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7, void *arg8, void *arg9) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9);
}
void *invoke8(void *addr, void *arg1, void *arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7, void *arg8) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7, void *arg8) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8);
}
void *invoke7(void *addr, void *arg1, void *arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4, void *arg5, void *arg6, void *arg7) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4, arg5, arg6, arg7);
}
void *invoke6(void *addr, void *arg1, void *arg2, void *arg3, void *arg4, void *arg5, void *arg6) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4, void *arg5, void *arg6) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4, arg5, arg6);
}
void *invoke5(void *addr, void *arg1, void *arg2, void *arg3, void *arg4, void *arg5) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4, void *arg5) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4, arg5);
}
void *invoke4(void *addr, void *arg1, void *arg2, void *arg3, void *arg4) {
void *(*ptr)(void *arg1, void * arg2, void *arg3, void *arg4) = addr;
	return (*ptr)(arg1, arg2, arg3, arg4);
}
void *invoke3(void *addr, void *arg1, void *arg2, void *arg3) {
void *(*ptr)(void *arg1, void * arg2, void *arg3) = addr;
	return (*ptr)(arg1, arg2, arg3);
}
void *invoke2(void *addr, void *arg1, void *arg2) {
void *(*ptr)(void *arg1, void * arg2) = addr;
	return (*ptr)(arg1, arg2);
}
void *invoke1(void *addr, void *arg1)  {
	unsigned long long (*ptr)(void *arg1) = addr;
	(*ptr)(arg1);
}
void *invoke0(void *addr)  {
	unsigned long long (*ptr)() = addr;
	(*ptr)();
}
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

type Handle uintptr

// DLOpen opens libVGElib.so.
func DLOpen(libraryPath string) (Handle, error) {
	if strings.HasSuffix(strings.ToLower(libraryPath), ".dll") {
		// Convert Windows dll name to Linux equivalent
		libraryPath = "lib" + libraryPath[:len(libraryPath)-4] + ".so"
	}
	rHandle := C.dlopen(C.CString(libraryPath), C.RTLD_NOW)
	if uintptr(rHandle) == 0 {
		return 0, fmt.Errorf("Load %s failed; %s", libraryPath, C.GoString(C.dlerror()))
	}
	libHandle := Handle(rHandle)
	return libHandle, nil
}

// Retrieve address to exported function
func GetProcAddress(libHandle Handle, exportName string) (trap uintptr, err error) {
	mh := C.dlsym(unsafe.Pointer(libHandle), C.CString(exportName))
	if uintptr(mh) == 0 {
		return 0, fmt.Errorf("Failed to load function: %s", exportName)
	}
	return uintptr(mh), nil
}

// DLClose closes libVGElib.so
func DLClose(libHandle Handle) {
	// FreeLibrary releases previously loaded library
	C.dlclose(unsafe.Pointer(libHandle))
}

// Invoke function retrieved with GetProcAddress
func Invoke(trap uintptr, nargs int, a1 uintptr, a2 uintptr, a3 uintptr) uintptr {
	switch nargs {
	case 0:
		return uintptr(C.invoke0(unsafe.Pointer(trap)))
	case 1:
		return uintptr(C.invoke1(unsafe.Pointer(trap), unsafe.Pointer(a1)))
	case 2:
		return uintptr(C.invoke2(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2)))
	case 3:
		return uintptr(C.invoke3(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3)))
	}
	panic("Nargs must be from 0 to 3")
}

// Invoke function retrieved with GetProcAddress
func Invoke6(trap uintptr, nargs int, a1 uintptr, a2 uintptr, a3 uintptr, a4 uintptr, a5 uintptr, a6 uintptr) uintptr {
	switch nargs {
	case 4:
		return uintptr(C.invoke4(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4)))
	case 5:
		return uintptr(C.invoke5(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4), unsafe.Pointer(a5)))
	case 6:
		return uintptr(C.invoke6(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4), unsafe.Pointer(a5), unsafe.Pointer(a6)))

	}
	panic("Nargs must be from 4 to 6")
}

func Invoke9(trap uintptr, nargs int, a1 uintptr, a2 uintptr, a3 uintptr, a4 uintptr, a5 uintptr, a6 uintptr,
	a7 uintptr, a8 uintptr, a9 uintptr) uintptr {
	switch nargs {

	case 7:
		return uintptr(C.invoke7(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4), unsafe.Pointer(a5), unsafe.Pointer(a6), unsafe.Pointer(a7)))
	case 8:
		return uintptr(C.invoke8(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4), unsafe.Pointer(a5), unsafe.Pointer(a6), unsafe.Pointer(a7), unsafe.Pointer(a8)))
	case 9:
		return uintptr(C.invoke9(unsafe.Pointer(trap), unsafe.Pointer(a1), unsafe.Pointer(a2), unsafe.Pointer(a3),
			unsafe.Pointer(a4), unsafe.Pointer(a5), unsafe.Pointer(a6), unsafe.Pointer(a7), unsafe.Pointer(a8),
			unsafe.Pointer(a9)))
	}
	panic("Nargs must be from 7 to 9")
}
