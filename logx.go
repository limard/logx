package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
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

	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

//type Loggerx struct
type Loggerx struct {
	OutFile   *os.File
	LastError error
	FilePerm  os.FileMode

	hConsoleOut io.Writer
	logPath     string // log的保存目录
	logName     string // log的文件名，默认为程序名
	logCounter  int
	outputFlag  int
	outputLevel int
	mu          sync.Mutex // ensures atomic writes; protects the following fields
	prefix      []byte     // prefix to write at beginning of each line
	flag        int        // properties
	buf         []byte     // for accumulating text to write
	muFile      sync.Mutex
}

func New(path, name string) *Loggerx {
	l := new(Loggerx)
	l.logPath = path
	l.logName = name
	l.logCounter = 0
	l.outputFlag = OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView
	l.outputLevel = OutputLevel_Debug
	l.FilePerm = defaultFilePerm
	l.flag = Lshortfile | Ldate | Ltime

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
	t.flag = flag
}

// SetConsoleOut set a writer instead of console
func (t *Loggerx) SetConsoleOut(out io.Writer) {
	t.hConsoleOut = out
}

// SetConsoleOutPrefix set prefix for console output
func (t *Loggerx) SetConsoleOutPrefix(prefix []byte) {
	t.prefix = prefix
}

func (t *Loggerx) getFileHandle() error {
	e := os.MkdirAll(t.logPath, 0666)
	if e != nil {
		t.LastError = e
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
	t.OutFile, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, t.FilePerm)
	if e != nil {
		t.LastError = e
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
	if t.OutFile != nil && t.logCounter < 100 {
		t.logCounter++
		return nil
	}
	t.logCounter = 1

	t.muFile.Lock()
	defer t.muFile.Unlock()

	if t.OutFile == nil {
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	fi, _ := t.OutFile.Stat()
	if fi.Size() > 1024*1024*5 {
		t.OutFile.Close()
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	if t.OutFile == nil {
		return fmt.Errorf("OutFile is nil")
	}
	return nil
}

func (t *Loggerx) output(s string) {
	s = addNewLine(s)

	buf := t.makeStr(4, s)

	if t.outputFlag&OutputFlag_File != 0 {
		e := t.renewLogFile()
		if e != nil {
			es := addNewLine(e.Error())
			if t.hConsoleOut != nil {
				t.hConsoleOut.Write([]byte(es))
			}
			outputToDebugView([]byte(es))
			if strings.Contains(e.Error(), "permission denied") {
				t.outputFlag &= ^OutputFlag_File
			}
		} else {
			t.OutFile.Write(buf)
		}
	}

	if t.outputFlag&OutputFlag_Console != 0 && t.hConsoleOut != nil {
		t.hConsoleOut.Write(buf)
	}

	if t.outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (t *Loggerx) formatHeader(buf *[]byte, tm time.Time, file string, line int) {
	*buf = append(*buf, t.prefix...)
	if t.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if t.flag&LUTC != 0 {
			tm = tm.UTC()
		}
		if t.flag&Ldate != 0 {
			year, month, day := tm.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if t.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := tm.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if t.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, tm.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if t.flag&(Lshortfile|Llongfile) != 0 {
		if t.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (t *Loggerx) makeStr(calldepth int, s string) []byte {
	now := time.Now() // get this early.
	var file string
	var line int
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		t.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		t.mu.Lock()
	}
	t.buf = t.buf[:0]
	t.formatHeader(&t.buf, now, file, line)
	t.buf = append(t.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		t.buf = append(t.buf, '\n')
	}
	return t.buf
}
