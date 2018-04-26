package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	hFile            *os.File
	hConsoleOut      io.Writer
	consoleOutPrefix []byte
	logPath          string
	logFile          = log.New(nil, "", log.Lshortfile|log.Ldate|log.Ltime)
	logCounter       = 0
	outputFlag       = OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView
	outputLevel      = OutputLevel_Debug
)

// const value
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
	OutputFlag  []string
}

func init() {
	hConsoleOut = os.Stdout

	buf, e := ioutil.ReadFile("log.json")
	if e == nil {
		var c1 configFile
		json.Unmarshal(buf, &c1)

		if len(c1.OutputFlag) != 0 {
			outputFlag = 0
			for _, f := range c1.OutputFlag {
				switch strings.ToLower(f) {
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
			case "debug", "dbg":
				outputLevel = OutputLevel_Debug
			case "info":
				outputLevel = OutputLevel_Info
			case "warn", "warning":
				outputLevel = OutputLevel_Warn
			case "error", "err":
				outputLevel = OutputLevel_Error
			case "unexpected":
				outputLevel = OutputLevel_Unexpected
			}
		}
	}
}

func getLogFile(fDir string) (*os.File, error) {
	e := os.MkdirAll(fDir, 0666)
	if e != nil {
		return nil, e
	}

	file, _ := exec.LookPath(os.Args[0])

	filepath.Walk(fDir, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fInfo.IsDir() {
			return nil
		}
		if strings.Contains(filepath.Base(fPath), filepath.Base(file)) {
			if time.Now().Sub(fInfo.ModTime()) > LogSaveTime {
				os.Remove(fPath)
			}
		}
		return nil
	})

	filename := fDir + filepath.Base(file) + `.` + time.Now().Format(`060102_150405`) + `.log`
	// linux: 该目录所有模块可写、创建、删除、不能读（只保留6天），用户只读
	logfile, e := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0442)
	if e != nil {
		return nil, e
	}
	return logfile, nil
}

func renewLogFile() (e error) {
	if hFile != nil && logCounter < 100 {
		logCounter++
		return nil
	}
	logCounter = 1

	if logPath == "" {
		logPath = getDefaultLogPath()
	}

	if hFile == nil {
		hFile, e = getLogFile(logPath)
		if e != nil {
			return e
		}
	}

	fi, _ := hFile.Stat()
	if fi.Size() > 1024*1024*5 {
		hFile.Close()
		hFile, e = getLogFile(logPath)
		if e != nil {
			return e
		}
	}

	if hFile == nil {
		return fmt.Errorf("hFile is nil")
	}

	logFile.SetOutput(hFile)
	return nil
}

func output(s string) {
	s = addNewLine(s)

	if outputFlag&OutputFlag_File != 0 {
		e := renewLogFile()
		if e != nil {
			es := addNewLine(e.Error())
			hConsoleOut.Write([]byte(es))
			outputToDebugView([]byte(es))
			if strings.Contains(e.Error(), "permission denied") {
				outputFlag &= ^OutputFlag_File
			}
		} else {
			logFile.Output(3, s)
		}
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

// Trace output a [DEBUG] trace string
func Trace() {
	if outputLevel > OutputLevel_Debug {
		return
	}

	funcName := ""
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
		s := strings.Split(funcName, ".")
		funcName = s[len(s)-1]
	}
	output(fmt.Sprintf("[TRACE]%v", funcName))
}

// Debug output a [DEBUG] string
func Debug(v ...interface{}) {
	if outputLevel > OutputLevel_Debug {
		return
	}
	output(fmt.Sprintf(`[DEBUG]%s`, fmt.Sprint(v...)))
}

// Debugf output a [DEBUG] string with format
func Debugf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Debug {
		return
	}
	output(fmt.Sprintf(`[DEBUG]`+format, v...))
}

// Info output a [INFO ] string
func Info(v ...interface{}) {
	if outputLevel > OutputLevel_Info {
		return
	}
	output(fmt.Sprintf(`[INFO ]%s`, fmt.Sprint(v...)))
}

// Infof output a [INFO ] string with format
func Infof(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Info {
		return
	}
	output(fmt.Sprintf(`[INFO ]`+format, v...))
}

// Warn output a [WARN ] string
func Warn(v ...interface{}) {
	if outputLevel > OutputLevel_Warn {
		return
	}
	output(fmt.Sprintf(`[WARN ]%s`, fmt.Sprint(v...)))
}

// Warnf output a [WARN ] string with format
func Warnf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Warn {
		return
	}
	output(fmt.Sprintf(`[WARN ]`+format, v...))
}

// Error output a [ERROR] string
func Error(v ...interface{}) {
	if outputLevel > OutputLevel_Error {
		return
	}
	output(fmt.Sprintf(`[ERROR]%s`, fmt.Sprint(v...)))
}

// Errorf output a [ERROR] string with format
func Errorf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Error {
		return
	}
	output(fmt.Sprintf(`[ERROR]`+format, v...))
}

// Unexpected output a [UNEXP] string
func Unexpected(v ...interface{}) {
	if outputLevel > OutputLevel_Unexpected {
		return
	}
	output(fmt.Sprintf(`[UNEXP]%s`, fmt.Sprint(v...)))
}

// Unexpectedf output a [UNEXP] string with format
func Unexpectedf(format string, v ...interface{}) {
	if outputLevel > OutputLevel_Unexpected {
		return
	}
	output(fmt.Sprintf(`[UNEXP]`+format, v...))
}

// SetLogPath set path of output log
func SetLogPath(s string) {
	logPath = s
}

// SetOutputFlag set output purpose(OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView)
func SetOutputFlag(flag int) {
	output(fmt.Sprintf("Log Level: %v Flag: %v", outputLevel, flag))
	outputFlag = flag
}

// SetOutputLevel set output level.
// OutputLevel_Debug
// OutputLevel_Info
// OutputLevel_Warn
// OutputLevel_Error
// OutputLevel_Unexpected
func SetOutputLevel(level int) {
	output(fmt.Sprintf("Log Level: %v Flag: %v", level, outputFlag))
	outputLevel = level
}

// SetTimeFlag set time format(Lshortfile | Ldate | Ltime)
func SetTimeFlag(flag int) {
	logFile.SetFlags(flag)
}

// SetConsoleOut set a writer instead of console
func SetConsoleOut(out io.Writer) {
	hConsoleOut = out
}

// SetConsoleOutPrefix set prefix for console output
func SetConsoleOutPrefix(prefix []byte) {
	consoleOutPrefix = prefix
}
