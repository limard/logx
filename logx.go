package logx

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"strings"
	"runtime"
	"gopkg.in/yaml.v2"
)

var (
	hFile       *os.File
	logFile     = log.New(nil, "", log.Lshortfile|log.Ldate|log.Ltime)
	logPath     string
	logCounter  = 0
	outputFlag  = OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView
	outputLevel = OutputLevel_Debug
)

const (
	OutputFlag_File = 1 << iota
	OutputFlag_Console
	OutputFlag_DbgView

	OutputLevel_Debug      = 100
	OutputLevel_Info       = 200
	OutputLevel_Warn       = 300
	OutputLevel_Error      = 400
	OutputLevel_Unexpected = 500
)

type configFile struct {
	OutputLevel string
	OutputFlag []string
}

func init() {
	buf, e := ioutil.ReadFile("log.yaml")
	if e == nil {
		var c1 configFile
		yaml.Unmarshal(buf, &c1)

		if len(c1.OutputFlag) != 0 {
			outputFlag = 0
			for _, f := range c1.OutputFlag {
				switch strings.ToLower(f)  {
				case "file":
					outputFlag |= OutputFlag_File
				case "console":
					outputFlag |= OutputFlag_Console
				case "dbgview":
					outputFlag |= OutputFlag_DbgView
				}
			}
		}

		if c1.OutputLevel != "" {
			switch strings.ToLower(c1.OutputLevel) {
			case "debug":
				outputLevel = OutputLevel_Debug
			case "info":
				outputLevel = OutputLevel_Info
			case "warn":
				outputLevel = OutputLevel_Warn
			case "error":
				outputLevel = OutputLevel_Error
			case "unexpected":
				outputLevel = OutputLevel_Unexpected
			}
		}
	}
}

func getLogFile(fDir string) *os.File {
	os.MkdirAll(fDir, 0666)

	file, _ := exec.LookPath(os.Args[0])
	filename := fDir + filepath.Base(file) + `.log`

	os.Remove(filename + ".old")
	os.Rename(filename, filename+".old")

	logfile, _ := os.OpenFile(filename, os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0666)
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
		fmt.Print(s)
	}

	if outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}

func Trace() {
	if outputLevel > OutputLevel_Debug {
		return
	}

	funcName := ""
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	output(fmt.Sprintf("[TRACE]%v", funcName))
}

func Debug(v ...interface{}) {
	if outputLevel > OutputLevel_Debug {
		return
	}
	output(fmt.Sprintf(`[DEBUG]%s`, fmt.Sprint(v...)))
}

func Debugf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Debug {
		return
	}
	output(fmt.Sprintf(`[DEBUG]`+format, v...))
}

func Info(v ...interface{}) {
	if outputLevel > OutputLevel_Info {
		return
	}
	output(fmt.Sprintf(`[INFO ]%s`, fmt.Sprint(v...)))
}

func Infof(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Info {
		return
	}
	output(fmt.Sprintf(`[INFO ]`+format, v...))
}

func Warn(v ...interface{}) {
	if outputLevel > OutputLevel_Warn {
		return
	}
	output(fmt.Sprintf(`[WARN ]%s`, fmt.Sprint(v...)))
}

func Warnf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Warn {
		return
	}
	output(fmt.Sprintf(`[WARN ]`+format, v...))
}

func Error(v ...interface{}) {
	if outputLevel > OutputLevel_Error {
		return
	}
	output(fmt.Sprintf(`[ERROR]%s`, fmt.Sprint(v...)))
}

func Errorf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Error {
		return
	}
	output(fmt.Sprintf(`[ERROR]`+format, v...))
}

func Unexpected(v ...interface{}) {
	if outputLevel > OutputLevel_Unexpected {
		return
	}
	output(fmt.Sprintf(`[UNEXP]%s`, fmt.Sprint(v...)))
}

func Unexpectedf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Unexpected {
		return
	}
	output(fmt.Sprintf(`[UNEXP]`+format, v...))
}

func SetLogPath(s string) {
	logPath = s
}

func SetOutputFlag(flag int) {
	output(fmt.Sprintf("Log Level: %v Flag: %v", outputLevel, flag))
	outputFlag = flag
}

func SetOutputLevel(level int) {
	output(fmt.Sprintf("Log Level: %v Flag: %v", level, outputFlag))
	outputLevel = level
}
