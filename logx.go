package logx

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	hFile      *os.File
	logFile    = log.New(nil, "", log.Lshortfile|log.Ldate|log.Ltime)
	logPath    string
	logCounter = 0
	outputFlag = OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView
)

const (
	OutputFlag_File            = 1 << iota
	OutputFlag_Console
	OutputFlag_DbgView
)

func getLogFile(fDir string) *os.File {
	os.MkdirAll(fDir, 0666)

	file, _ := exec.LookPath(os.Args[0])
	filename := fDir + filepath.Base(file) + `.log`

	os.Remove(filename + ".old")
	os.Rename(filename, filename+".old")

	fileflag := os.O_CREATE | os.O_RDWR | os.O_TRUNC
	logfile, _ := os.OpenFile(filename, fileflag, 0666)
	return logfile
}

func renewLogFile() {
	if logCounter != 0 && logCounter < 50 {
		logCounter++
		return
	}
	logCounter = 1

	if logPath == "" {
		logPath = getBisPath()
	}

	if hFile == nil {
		hFile = getLogFile(logPath)
	}

	fi, _ := hFile.Stat()
	if fi.Size() > 1024*1024*5 {
		hFile.Close()
		hFile = getLogFile(logPath)
	}
	logFile.SetOutput(hFile)
}

func output(s string) {
	if outputFlag&OutputFlag_File != 0 {
		renewLogFile()
		logFile.Output(2, s)
	}

	if outputFlag&OutputFlag_Console != 0 {
		fmt.Print(s)
	}

	if outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}

func Debug(v ...interface{}) {
	output(fmt.Sprintf(`[DEBUG]%s`, fmt.Sprint(v...)))
}

func Debugf(format string, v ...interface{}) {
	output(fmt.Sprintf(`[DEBUG]`+format, v...))
}

func Info(v ...interface{}) {
	output(fmt.Sprintf(`[INFO ]%s`, fmt.Sprint(v...)))
}

func Infof(format string, v ...interface{}) {
	output(fmt.Sprintf(`[INFO ]`+format, v...))
}

func Warn(v ...interface{}) {
	output(fmt.Sprintf(`[WARN ]%s`, fmt.Sprint(v...)))
}

func Warnf(format string, v ...interface{}) {
	output(fmt.Sprintf(`[WARN ]`+format, v...))
}

func Error(v ...interface{}) {
	output(fmt.Sprintf(`[ERROR]%s`, fmt.Sprint(v...)))
}

func Errorf(format string, v ...interface{}) {
	output(fmt.Sprintf(`[ERROR]`+format, v...))
}

func Unexpected(v ...interface{}) {
	output(fmt.Sprintf(`[UNEXP]%s`, fmt.Sprint(v...)))
}

func Unexpectedf(format string, v ...interface{}) {
	output(fmt.Sprintf(`[UNEXP]`+format, v...))
}

func SetLogPath(s string) {
	logPath = s
}

func SetOutputFlag(flag int) {
	outputFlag = flag
}