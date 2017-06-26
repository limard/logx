// +build windows

package logx

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

func getLogFile() *os.File {
	commonpath, _ := getCommmonAppDataDirectory()
	if commonpath == "" {
		commonpath = `C:\log`
	}

	file, _ := exec.LookPath(os.Args[0])
	filename := filepath.Base(file)
	os.MkdirAll(commonpath+`\PrintSystem\Log`, 0666)

	// for i := 0; i < 10; i++ {
	// 	_, fname, _, ok := runtime.Caller(i)
	// 	if ok == false {
	// 		break
	// 	}
	// 	log.Println(i, fname)
	// }

	filename = commonpath + `\PrintSystem\Log\` + filename + `.log`

	// is size > 2mb then clear
	fileflag := os.O_CREATE | os.O_RDWR
	if fi, err := os.Stat(filename); err == nil || (err != nil && os.IsExist(err)) {
		if fi.Size() > 1024*1024*2 {
			fileflag |= os.O_TRUNC
		} else {
			fileflag |= os.O_APPEND
		}
	}

	logfile, _ := os.OpenFile(filename, fileflag, 0666)

	return logfile
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

var (
	dllKernel             = syscall.NewLazyDLL("Kernel32.dll")
	procOutputDebugString = dllKernel.NewProc("OutputDebugStringW")
)

func outputToDebugView(buf []byte) {
	p, _ := syscall.UTF16PtrFromString(string(buf))
	procOutputDebugString.Call(uintptr(unsafe.Pointer(p)))
}
