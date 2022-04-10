package main

import (
	"syscall"
)

func fillLogicalDriveLetters() error {

	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")

	if ret, _, callErr := syscall.Syscall(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		return callErr
	} else {
		availableDrives := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

		for i := range availableDrives {
			if ret&1 == 1 {
				app.drives = append(app.drives, availableDrives[i]+":")
			}
			ret >>= 1
		}
	}
	return nil
}
