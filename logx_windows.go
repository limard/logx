// +build windows

package logx

import (
	"syscall"
	"time"
	"unsafe"
)

// Bis path

func getDefaultLogPath() string {
	s, e := getCommonAppDataDirectory()
	if e != nil {
		s = `C:\log`
	}

	s += `\PrintSystem\Log\`
	return s
}

// LogSaveTime ...
var LogSaveTime = 6 * 24 * time.Hour

var (
	dShell32                = syscall.NewLazyDLL("Shell32.dll")
	pSHGetSpecialFolderPath = dShell32.NewProc("SHGetSpecialFolderPathW")
)

func getCommonAppDataDirectory() (string, error) {
	const CSIDL_COMMON_APPDATA = 0x23
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

func addNewLine(s string) string {
	l := len(s)
	if l == 0 {
		return "\r\n"
	}
	if l == 1 {
		return s + "\r\n"
	}
	if s[l-2] == '\r' && s[l-1] == '\n' {
		return s
	}
	if s[l-1] == '\r' || s[l-1] == '\n' {
		return s[:l-1] + "\r\n"
	}
	return s + "\r\n"
}
