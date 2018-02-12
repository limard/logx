// +build windows

package logx

import (
	"syscall"
	"unsafe"
	"time"
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

var LogSaveTime = 15*24*time.Hour

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

func output(s string) {
	l := len(s)
	if l > 1 {
		if s[l-2:] == "\r\n" {
			// OK
		} else if s[l-1:] == "\r" || s[l-1:] == "\n" {
			s = s[0:l-1] + "\r\n"
		} else {
			s += "\r\n"
		}
	}

	if outputFlag&OutputFlag_File != 0 {
		renewLogFile()
		logFile.Output(3, s)
	}

	if outputFlag&OutputFlag_Console != 0 {
		if len(consoleOutPrefix) != 0 {
			hConsoleOut.Write(consoleOutPrefix)
		}
		hConsoleOut.Write([]byte(s))
	}

	if outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}
