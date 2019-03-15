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
	"sort"
	"strings"
	"time"
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

//type Loggerx struct
type Loggerx struct {
	hFile            *os.File
	hConsoleOut      io.Writer
	consoleOutPrefix []byte
	logPath          string // log的保存目录
	logName          string // log的文件名，默认为程序名
	logFile          *log.Logger
	logCounter       int
	outputFlag       int
	outputLevel      int
}

func New(path, name string) *Loggerx {
	l := new(Loggerx)
	l.logPath = path
	l.logName = name
	l.logFile = log.New(nil, "", log.Lshortfile|log.Ldate|log.Ltime)
	l.logCounter = 0
	l.outputFlag = OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView
	l.outputLevel = OutputLevel_Debug

	l.hConsoleOut = os.Stdout

	if len(l.logPath) == 0 {
		l.logPath = getDefaultLogPath()
	}

	if len(l.logName) == 0 {
		n, _ := exec.LookPath(os.Args[0])
		l.logName = filepath.Base(n)
	}

	//
	type configFile struct {
		OutputLevel string
		OutputFlag  []string
	}
	buf, e := ioutil.ReadFile("log.json")
	if e == nil {
		var c1 configFile
		json.Unmarshal(buf, &c1)

		if len(c1.OutputFlag) != 0 {
			l.outputFlag = 0
			for _, f := range c1.OutputFlag {
				switch strings.ToLower(f) {
				case "file":
					l.outputFlag |= OutputFlag_File
				case "console":
					l.outputFlag |= OutputFlag_Console
				case "dbgview":
					l.outputFlag |= OutputFlag_DbgView
				}
			}
		}

		if c1.OutputLevel != "" {
			switch strings.ToLower(c1.OutputLevel) {
			case "debug", "dbg":
				l.outputLevel = OutputLevel_Debug
			case "info":
				l.outputLevel = OutputLevel_Info
			case "warn", "warning":
				l.outputLevel = OutputLevel_Warn
			case "error", "err":
				l.outputLevel = OutputLevel_Error
			case "unexpected":
				l.outputLevel = OutputLevel_Unexpected
			}
		}
	}

	return l
}

func funcName() string {
	funcName := ""
	pc, _, _, ok := runtime.Caller(3)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
		s := strings.Split(funcName, ".")
		funcName = s[len(s)-1]
	}
	return funcName
}

func (t *Loggerx) Trace() {
	t.trace()
}

func (t *Loggerx) trace() {
	if t.outputLevel > OutputLevel_Debug {
		return
	}

	t.output(fmt.Sprintf("[TRACE][%s]", funcName()))
}

// Debug output a [DEBUG] string
func (t *Loggerx) Debug(v ...interface{}) {
	t.debug(v...)
}

func (t *Loggerx) debug(v ...interface{}) {
	if t.outputLevel > OutputLevel_Debug {
		return
	}
	t.output(fmt.Sprintf(`[DEBUG][%s]%s`, funcName(), fmt.Sprint(v...)))
}

// Debugf output a [DEBUG] string with format
func (t *Loggerx) Debugf(format string, v ...interface{}) {
	t.debugf(format, v...)
}

func (t *Loggerx) debugf(format string, v ...interface{}) {
	if t.outputLevel > OutputLevel_Debug {
		return
	}
	t.output(fmt.Sprintf(fmt.Sprintf(`[DEBUG][%s]%s`, funcName(), format), v...))
}

func (t *Loggerx) DebugToJson(v ...interface{}) {
	t.debugToJson(v...)
}

func (t *Loggerx) debugToJson(v ...interface{}) {
	if t.outputLevel > OutputLevel_Debug {
		return
	}
	ss := []string{`[DEBUG]`, `[` + funcName() + `]`}
	for _, sub := range v {
		switch sub.(type) {
		case string:
			ss = append(ss, sub.(string))
		default:
			buf, _ := json.Marshal(sub)
			ss = append(ss, string(buf))
		}
	}
	t.output(strings.Join(ss, ""))
}

// Info output a [INFO ] string
func (t *Loggerx) Info(v ...interface{}) {
	t.info(v...)
}

func (t *Loggerx) info(v ...interface{}) {
	if t.outputLevel > OutputLevel_Info {
		return
	}
	t.output(fmt.Sprintf(`[INFO ][%s]%s`, funcName(), fmt.Sprint(v...)))
}

// Infof output a [INFO ] string with format
func (t *Loggerx) Infof(format string, v ...interface{}) {
	t.infof(format, v...)
}

func (t *Loggerx) infof(format string, v ...interface{}) {
	if t.outputLevel > OutputLevel_Info {
		return
	}
	t.output(fmt.Sprintf(fmt.Sprintf(`[INFO ][%s]%s`, funcName(), format), v...))
}

// Warn output a [WARN ] string
func (t *Loggerx) Warn(v ...interface{}) {
	t.warn(v...)
}

func (t *Loggerx) warn(v ...interface{}) {
	if t.outputLevel > OutputLevel_Warn {
		return
	}
	t.output(fmt.Sprintf(`[WARN ][%s]%s`, funcName(), fmt.Sprint(v...)))
}

// Warnf output a [WARN ] string with format
func (t *Loggerx) Warnf(format string, v ...interface{}) {
	t.warnf(format, v...)
}

func (t *Loggerx) warnf(format string, v ...interface{}) {
	if t.outputLevel > OutputLevel_Warn {
		return
	}
	t.output(fmt.Sprintf(fmt.Sprintf(`[WARN ][%s]%s`, funcName(), format), v...))
}

// Error output a [ERROR] string
func (t *Loggerx) Error(v ...interface{}) {
	t.error(v...)
}

func (t *Loggerx) error(v ...interface{}) {
	if t.outputLevel > OutputLevel_Error {
		return
	}
	t.output(fmt.Sprintf(`[ERROR][%s]%s`, funcName(), fmt.Sprint(v...)))
}

// Errorf output a [ERROR] string with format
func (t *Loggerx) Errorf(format string, v ...interface{}) {
	t.errorf(format, v...)
}

func (t *Loggerx) errorf(format string, v ...interface{}) {
	if t.outputLevel > OutputLevel_Error {
		return
	}
	t.output(fmt.Sprintf(fmt.Sprintf(`[ERROR][%s]%s`, funcName(), format), v...))
}

// SetLogPath set path of output log
func (t *Loggerx) SetLogPath(s string) {
	t.logPath = s
}

// SetOutputFlag set output purpose(OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView)
func (t *Loggerx) SetOutputFlag(flag int) {
	t.outputFlag = flag
}

// SetOutputLevel set output level.
// OutputLevel_Debug
// OutputLevel_Info
// OutputLevel_Warn
// OutputLevel_Error
// OutputLevel_Unexpected
func (t *Loggerx) SetOutputLevel(level int) {
	t.outputLevel = level
}

// SetTimeFlag set time format(Lshortfile | Ldate | Ltime)
func (t *Loggerx) SetTimeFlag(flag int) {
	t.logFile.SetFlags(flag)
}

// SetConsoleOut set a writer instead of console
func (t *Loggerx) SetConsoleOut(out io.Writer) {
	t.hConsoleOut = out
}

// SetConsoleOutPrefix set prefix for console output
func (t *Loggerx) SetConsoleOutPrefix(prefix []byte) {
	t.consoleOutPrefix = prefix
}

func (t *Loggerx) getFileHandle() error {
	e := os.MkdirAll(t.logPath, 0666)
	if e != nil {
		return e
	}

	files := make([]string, 0)
	filepath.Walk(t.logPath, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fInfo.IsDir() || strings.HasPrefix(filepath.Base(fPath), t.logName) == false {
			return nil
		}
		if time.Now().Sub(fInfo.ModTime()) > LogSaveTime {
			os.Remove(fPath)
			return nil
		}
		files = append(files, fInfo.Name())
		return nil
	})
	for _, value := range t.getNeedDeleteLogfile(files) {
		fmt.Println("delete log file:", value)
		os.Remove(t.logPath + `\` + value)
	}

	filename := t.logPath + t.logName + `.` + time.Now().Format(`060102_150405`) + `.log`
	// linux: 该目录所有模块可写、创建、删除、不能读（只保留6天），用户只读
	t.hFile, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0442)
	if e != nil {
		return e
	}
	return nil
}

func (t *Loggerx) getNeedDeleteLogfile(filesName []string) []string {
	if len(filesName) < 6 {
		return nil
	}
	sort.Strings(filesName)

	result := make([]string, 0)
	for i := 0; i < len(filesName)-6; i++ {
		result = append(result, filesName[i])
	}
	return result
}

func (t *Loggerx) renewLogFile() (e error) {
	if t.hFile != nil && t.logCounter < 100 {
		t.logCounter++
		return nil
	}
	t.logCounter = 1

	if t.hFile == nil {
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	fi, _ := t.hFile.Stat()
	if fi.Size() > 1024*1024*5 {
		t.hFile.Close()
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	if t.hFile == nil {
		return fmt.Errorf("hFile is nil")
	}

	t.logFile.SetOutput(t.hFile)
	return nil
}

func (t *Loggerx) output(s string) {
	s = addNewLine(s)

	if t.outputFlag&OutputFlag_File != 0 {
		e := t.renewLogFile()
		if e != nil {
			es := addNewLine(e.Error())
			t.hConsoleOut.Write([]byte(es))
			outputToDebugView([]byte(es))
			if strings.Contains(e.Error(), "permission denied") {
				t.outputFlag &= ^OutputFlag_File
			}
		} else {
			t.logFile.Output(4, s)
		}
	}

	if t.outputFlag&OutputFlag_Console != 0 {
		if len(t.consoleOutPrefix) != 0 {
			t.hConsoleOut.Write(t.consoleOutPrefix)
		}
		t.hConsoleOut.Write([]byte(s))
	}

	if t.outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}
