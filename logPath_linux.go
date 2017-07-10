package logx

import (
	"os"
	"os/exec"
	"path/filepath"
)

func getLogFile() *os.File {
	logPath := `/opt/PrintSystem/Log/`
	if _, err := os.Stat(logPath); err != nil && os.IsNotExist(err) {
		os.MkdirAll(logPath, 0666)
	}

	filename, _ := exec.LookPath(os.Args[0])
	filename = filepath.Base(filename)
	filename = logPath + filename + `.log`

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

func outputToDebugView(buf []byte) {
}
