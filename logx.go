package log

// version: 2023/11/29

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// var log = NewLogger("", "")

var logLevelStr = []string{"[DBG]", "[INF]", "[WAR]", "[ERR]", "[FAL]"}

// const value
const (
	OutputFlag_File = 1 << iota
	OutputFlag_Console
)

const (
	OutputLevel_Debug = iota
	OutputLevel_Info
	OutputLevel_Warn
	OutputLevel_Error
	OutputLevel_Fatal
)

const (
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes PrefixFlag_Time.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides PrefixFlag_Longfile
	LUTC                      // if PrefixFlag_Date or PrefixFlag_Time is set, use UTC rather than the local time zone
	Lmsgprefix
	LfuncName
	Llevel
	LstdFlags = Lshortfile | Ldate | Ltime | LfuncName | Llevel | Lmicroseconds
)

// Logger struct
type Logger struct {
	OutFile          *os.File
	LastError        error
	FilePerm         os.FileMode
	LineMaxLength    int    // 一行最大的长度
	LogPath          string // log的保存目录
	LogName          string // log的文件名，默认为程序名
	OutputFlag       int    // 输出Flag
	OutputLevel      int    // 输出级别
	PrefixFlag       int    // properties L...
	FileInfoLevel    int    // 输出级别，用于在debug等常见的级别上，不输出文件、函数信息
	MaxLogNumber     int    // 最多log文件个数
	MaxFileSize      int64  // 最大文件尺寸（字节）
	ContinuousLog    bool   // 连续在上一个文件中输出，适用于经常被调用启动的程序日志
	LogSaveTime      time.Duration
	ConsoleOutWriter io.Writer // 可重定向到父进程中
	ConsoleColor     bool

	mu         sync.Mutex //log mutex
	logCounter int        // 记录写入次数
	Prefix     []byte     // Prefix to write at beginning of each line
	muFile     sync.Mutex
	callSkip   int
}

func NewLogger(path, name string) *Logger {
	l := &Logger{
		FilePerm:         os.FileMode(0666),
		LineMaxLength:    -1,
		LogPath:          path,
		LogName:          name,
		OutputFlag:       OutputFlag_File | OutputFlag_Console,
		OutputLevel:      OutputLevel_Debug,
		PrefixFlag:       LstdFlags,
		MaxLogNumber:     3,
		MaxFileSize:      3 * 1024 * 1024,
		ContinuousLog:    true,
		LogSaveTime:      6 * 24 * time.Hour,
		ConsoleOutWriter: os.Stdout,
		ConsoleColor:     true,
		logCounter:       0,
		callSkip:         3,
	}

	if len(l.LogPath) == 0 {
		if runtime.GOOS == "linux" {
			l.LogPath = `/var/log/`
		} else {
			l.LogPath = filepath.Dir(os.Args[0])
		}
	}

	if len(l.LogName) == 0 {
		n, _ := exec.LookPath(os.Args[0])
		l.LogName = filepath.Base(n)
	}

	// read json configuration
	executable, _ := os.Executable()
	exeDir := filepath.Dir(executable)
	buf, e := os.ReadFile(filepath.Join(exeDir, "log.json"))
	if e == nil {
		// 切掉BOM
		if buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
			buf = buf[3:]
		}

		c1 := struct {
			OutputLevel  string
			OutputFlag   []string
			MaxFileSize  int64
			MaxLogNumber int
		}{}
		_ = json.Unmarshal(buf, &c1)

		if len(c1.OutputFlag) != 0 {
			l.OutputFlag = 0
			for _, f := range c1.OutputFlag {
				switch strings.ToLower(f) {
				case "file":
					l.OutputFlag |= OutputFlag_File
				case "console":
					l.OutputFlag |= OutputFlag_Console
				}
			}
		}

		if c1.OutputLevel != "" {
			switch strings.ToLower(c1.OutputLevel) {
			case "debug", "dbg":
				l.OutputLevel = OutputLevel_Debug
			case "info":
				l.OutputLevel = OutputLevel_Info
			case "warn", "warning":
				l.OutputLevel = OutputLevel_Warn
			case "error", "err":
				l.OutputLevel = OutputLevel_Error
			case "fatal":
				l.OutputLevel = OutputLevel_Fatal
			}
		}

		if c1.MaxLogNumber != 0 {
			l.MaxLogNumber = c1.MaxLogNumber
		}

		if c1.MaxFileSize != 0 {
			l.MaxFileSize = c1.MaxFileSize
		}
	}

	return l
}

func (t *Logger) Trace() {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, "TRACE")
}

// Debug output a [DEBUG] string
func (t *Logger) Debug(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, "", v...)
}

// Debugf output a [DEBUG] string with format
func (t *Logger) Debugf(format string, v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, format, v...)
}

func (t *Logger) DebugToJson(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	var ss []string
	for _, sub := range v {
		switch sub := sub.(type) {
		case string:
			ss = append(ss, sub)
		default:
			b := &bytes.Buffer{}
			en := json.NewEncoder(b)
			en.SetEscapeHTML(false)
			_ = en.Encode(sub)
			ss = append(ss, b.String())
		}
	}
	t.output(OutputLevel_Debug, strings.Join(ss, ""))
}

func (t *Logger) Print(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, "", v...)
}

func (t *Logger) Println(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, "", v...)
}

func (t *Logger) Printf(format string, v ...interface{}) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.output(OutputLevel_Debug, format, v...)
}

// Info output a [INFO ] string
func (t *Logger) Info(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Info {
		return
	}
	t.output(OutputLevel_Info, "", v...)
}

// Infof output a [INFO ] string with format
func (t *Logger) Infof(format string, v ...interface{}) {
	if t.OutputLevel > OutputLevel_Info {
		return
	}
	t.output(OutputLevel_Info, format, v...)
}

// Warn output a [WARN ] string
func (t *Logger) Warn(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Warn {
		return
	}
	t.output(OutputLevel_Warn, "", v...)
}

// Warnf output a [WARN ] string with format
func (t *Logger) Warnf(format string, v ...interface{}) {
	if t.OutputLevel > OutputLevel_Warn {
		return
	}
	t.output(OutputLevel_Warn, format, v...)
}

// Error output a [ERROR] string
func (t *Logger) Error(v ...interface{}) {
	if t.OutputLevel > OutputLevel_Error {
		return
	}
	t.output(OutputLevel_Error, "", v...)
}

// Errorf output a [ERROR] string with format
func (t *Logger) Errorf(format string, v ...interface{}) {
	if t.OutputLevel > OutputLevel_Error {
		return
	}
	t.output(OutputLevel_Error, format, v...)
}

func (t *Logger) Fatal(v ...interface{}) {
	t.output(OutputLevel_Fatal, "", v...)
	os.Exit(1)
}

func (t *Logger) Fatalf(format string, v ...interface{}) {
	t.output(OutputLevel_Fatal, format, v...)
	os.Exit(1)
}

func (t *Logger) SetFlags(flag int) {
	t.OutputFlag = flag
}

///////////

// 判断所给路径文件/文件夹是否存在
func isFileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (t *Logger) getFileHandle() error {
	e := os.MkdirAll(t.LogPath, 0777)
	if e != nil {
		t.LastError = e
		return e
	}

	makeFileName := func(index int) string {
		if index == 0 {
			return filepath.Join(t.LogPath, fmt.Sprintf(`%s.log`, t.LogName))
		} else {
			return filepath.Join(t.LogPath, fmt.Sprintf(`%s.log.%d`, t.LogName, index))
		}
	}

	firstName := makeFileName(0)

	f, e := os.OpenFile(firstName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, t.FilePerm)
	if e != nil {
		fmt.Println("logx:", e)
		t.LastError = e
		return e
	}

	fi, _ := f.Stat()

	if fi.Size() > t.MaxFileSize {
		f.Close()
	} else {
		t.OutFile = f
		return nil
	}

	lastFileName := makeFileName(t.MaxLogNumber)
	if isFileExists(lastFileName) {
		e = os.Remove(lastFileName)
		if e != nil {
			fmt.Fprintf(os.Stderr, "delete old log file %s failed\n", e.Error())
		}
	}

	for i := t.MaxLogNumber; i >= 1; i-- {
		from := makeFileName(i - 1)
		to := makeFileName(i)
		e := os.Rename(from, to)
		if e != nil {
			fmt.Fprintf(os.Stderr, "rename old log file %s to %s failed: %s\n", from, to, e.Error())
		}
	}

	newFileName := makeFileName(0)
	t.OutFile, e = os.OpenFile(newFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, t.FilePerm)
	if e != nil {
		fmt.Println("logx:", e)
		t.LastError = e
		return e
	}
	return nil
}

// 获取同名Log中最老的数个
func (t *Logger) getNeedDeleteLogfile(filesName []string) []string {
	if len(filesName) < t.MaxLogNumber {
		return nil
	}
	sort.Strings(filesName)
	return filesName[0 : len(filesName)-t.MaxLogNumber]
}

// 获取同名Log中最新的一个
func (t *Logger) getNewestLogfile(filesName []string) string {
	if len(filesName) == 0 {
		return ""
	}
	sort.Strings(filesName)
	return filesName[len(filesName)-1]
}

func (t *Logger) renewLogFile() (e error) {
	if t.OutFile != nil && t.logCounter < 200 {
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
	if fi.Size() > t.MaxFileSize {
		t.OutFile.Close()
		t.OutFile = nil
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

var bufferPool = sync.Pool{
	New: func() any { return &bytes.Buffer{} },
}

func (t *Logger) output(level int, format string, v ...interface{}) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		bufferPool.Put(buf)
	}()

	buf.Reset()
	t.makeStr(buf, level, format, v...)

	if t.OutputFlag&OutputFlag_File != 0 {
		e := t.renewLogFile()
		if e != nil {
			_, _ = t.ConsoleOutWriter.Write([]byte(e.Error()))
			_, _ = t.ConsoleOutWriter.Write([]byte("\n"))

			if strings.Contains(e.Error(), "permission denied") {
				t.OutputFlag &= ^OutputFlag_File
			}
		} else {
			t.muFile.Lock()
			_, _ = t.OutFile.Write(buf.Bytes())
			t.muFile.Unlock()
		}
	}

	if t.OutputFlag&OutputFlag_Console != 0 {
		t.mu.Lock()
		if t.ConsoleColor {
			switch level {
			case OutputLevel_Debug:
				// t.ConsoleOutWriter.Write([]byte("\033[0;39;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				// t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			case OutputLevel_Info:
				// t.ConsoleOutWriter.Write([]byte("\033[0;34;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				// t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			case OutputLevel_Warn:
				// _, _ = t.ConsoleOutWriter.Write([]byte("\033[1;33;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				// _, _ = t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			case OutputLevel_Error:
				// _, _ = t.ConsoleOutWriter.Write([]byte("\033[1;31;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				// _, _ = t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			}
		} else {
			_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
		}
		t.mu.Unlock()
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func (t *Logger) itoa(buf *bytes.Buffer, i int, wid int) {
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
	buf.Write(b[bp:])
}

func (t *Logger) makeStr(buf *bytes.Buffer, level int, format string, v ...interface{}) {
	// level.. [DBG]
	if t.PrefixFlag&Llevel != 0 {
		buf.WriteString(logLevelStr[level])
	}

	// time.. 2022/02/10 15:00:22
	if t.PrefixFlag&(Ldate|Ltime|Lmicroseconds) != 0 {
		tm := time.Now()
		if t.PrefixFlag&LUTC != 0 {
			tm = tm.UTC()
		}
		if t.PrefixFlag&Ldate != 0 {
			year, month, day := tm.Date()
			t.itoa(buf, year%100, 2)
			buf.WriteByte('/')
			t.itoa(buf, int(month), 2)
			buf.WriteByte('/')
			t.itoa(buf, day, 2)
			buf.WriteByte(' ')
		}
		if t.PrefixFlag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := tm.Clock()
			t.itoa(buf, hour, 2)
			buf.WriteByte(':')
			t.itoa(buf, min, 2)
			buf.WriteByte(':')
			t.itoa(buf, sec, 2)
			if t.PrefixFlag&Lmicroseconds != 0 {
				buf.WriteByte('.')
				t.itoa(buf, tm.Nanosecond()/1e3, 6)
			}
			buf.WriteByte(' ')
		}
	}

	// logx_test.go:9 (funcName):
	if level >= t.FileInfoLevel && t.PrefixFlag&(Lshortfile|Llongfile|LfuncName) != 0 {
		pc, file, line, ok := runtime.Caller(t.callSkip)
		if ok {
			if t.PrefixFlag&(Lshortfile|Llongfile) != 0 {
				if t.PrefixFlag&Lshortfile != 0 {
					short := file
					for i := len(file) - 1; i > 0; i-- {
						if file[i] == '/' {
							short = file[i+1:]
							break
						}
					}
					file = short
				}
				buf.WriteString(file)
				buf.WriteByte(':')
				t.itoa(buf, line, -1)
			}

			if t.PrefixFlag&LfuncName != 0 {
				funcName := runtime.FuncForPC(pc).Name()
				s := strings.Split(funcName, ".")
				funcName = s[len(s)-1]
				buf.WriteByte(' ')
				buf.WriteString(funcName)
				//buf = append(buf, ')')
			}
			buf.WriteByte(':')
			buf.WriteByte(' ')
		}
	}

	// content
	if format == "" {
		buf.WriteString(fmt.Sprint(v...))
	} else {
		buf.WriteString(fmt.Sprintf(format, v...))
	}

	// limit max length
	if t.LineMaxLength > 0 && buf.Len() > t.LineMaxLength {
		buf.Truncate(t.LineMaxLength)
		buf.Write([]byte{' ', '.', '.', '.'})
	}

	if buf.Len() < 1 || buf.Bytes()[len(buf.Bytes())-1] != '\n' {
		buf.WriteByte('\n')
	}
	return
}
