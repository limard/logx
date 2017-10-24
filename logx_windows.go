// +build windows

package logx

import (
	"syscall"
	"unsafe"
)

// Bis path

func getBisPath() string {
	s, e := getCommmonAppDataDirectory()
	if e != nil {
		s = `C:\log`
	}

	s += `\PrintSystem\Log\`
	return s
}

var (
	dshell32                = syscall.NewLazyDLL("Shell32.dll")
	pSHGetSpecialFolderPath = dshell32.NewProc("SHGetSpecialFolderPathW")
	CSIDL_COMMON_APPDATA    = 0x23
)

func getCommmonAppDataDirectory() (string, error) {
	return shGetSpecialFolderPath(CSIDL_COMMON_APPDATA)
}

func shGetSpecialFolderPath(nFolder int) (string, error) {
	if err := pSHGetSpecialFolderPath.Find(); err != nil {
		return "", err
	}
	pt := make([]uint16, syscall.MAX_PATH)
	ret, _, err := pSHGetSpecialFolderPath.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(&pt[0])),
		uintptr(nFolder),
		uintptr(1))
	if ret != 0 {
		err = nil
	}

	return syscall.UTF16ToString(pt), err
}

// OutputDebugStringW

var (
	dllKernel             = syscall.NewLazyDLL("Kernel32.dll")
	procOutputDebugString = dllKernel.NewProc("OutputDebugStringW")
)

func outputToDebugView(buf []byte) {
	p, _ := syscall.UTF16PtrFromString(string(buf))
	procOutputDebugString.Call(uintptr(unsafe.Pointer(p)))
}
