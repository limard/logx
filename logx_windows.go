// +build windows

package logx

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

var (
	defaultFilePerm = os.FileMode(0666)
)

func getDefaultLogPath() string {
	return filepath.Dir(os.Args[0])
}

// LogSaveTime ...
var LogSaveTime = 6 * 24 * time.Hour

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
