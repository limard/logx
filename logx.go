package logx

// version: 2022/3/23

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

var logLevelStr = []string{"DEBUG", "INFO ", "WARN ", "ERROR", "FATAL"}
var log = NewLogger("", "")

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
	LstdFlags = Lshortfile | Ldate | Ltime | LfuncName | Llevel
)

//Logger struct
type Logger struct {
	OutFile          *os.File
	LastError        error
	FilePerm         os.FileMode
	LineMaxLength    int    // 一行最大的长度
	LogPath          string // log的保存目录
	LogName          string // log的文件名，默认为程序名
	OutputFlag       int    // 输出Flag
	OutputLevel      int    // 输出级别
	PrefixFlag       int    // properties
	MaxLogNumber     int    // 最多log文件个数
	ContinuousLog    bool   // 连续在上一个文件中输出，适用于经常被调用启动的程序日志
	LogSaveTime      time.Duration
	ConsoleOutWriter io.Writer // 可重定向到父进程中
	ConsoleColor     bool

	mu       sync.Mutex //log mutex
	writeCnt int        // 记录写入次数
	Prefix   []byte     // Prefix to write at beginning of each line
	muFile   sync.Mutex
	callSkip int
}

func NewLogger(path, name string) *Logger {
	l := &Logger{
		FilePerm:         os.FileMode(0666),
		LineMaxLength:    1024,
		LogPath:          path,
		LogName:          name,
		OutputFlag:       OutputFlag_File | OutputFlag_Console,
		OutputLevel:      OutputLevel_Debug,
		PrefixFlag:       LstdFlags,
		MaxLogNumber:     3,
		ContinuousLog:    true,
		LogSaveTime:      6 * 24 * time.Hour,
		ConsoleOutWriter: os.Stdout,
		ConsoleColor:     true,
		writeCnt:         0,
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
	buf, e := ioutil.ReadFile("log.json")
	if e == nil {
		c1 := struct {
			OutputLevel string
			OutputFlag  []string
		}{}
		json.Unmarshal(buf, &c1)

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
		switch sub.(type) {
		case string:
			ss = append(ss, sub.(string))
		default:
			buf, _ := json.Marshal(sub)
			ss = append(ss, string(buf))
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

func (t *Logger) getFileHandle() error {
	e := os.MkdirAll(t.LogPath, 0777)
	if e != nil {
		t.LastError = e
		return e
	}

	files := make([]string, 0)
	filepath.Walk(t.LogPath, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fInfo.IsDir() || !strings.HasPrefix(fInfo.Name(), t.LogName+`.`) || !strings.HasSuffix(fInfo.Name(), ".log") {
			return nil
		}
		if time.Now().Sub(fInfo.ModTime()) > t.LogSaveTime {
			os.Remove(fPath)
			return nil
		}
		files = append(files, fInfo.Name())
		return nil
	})
	for _, value := range t.getNeedDeleteLogfile(files) {
		os.Remove(t.LogPath + value)
	}

	if t.ContinuousLog {
		f := t.getNewestLogfile(files)
		if len(f) > 0 {
			filename := filepath.Join(t.LogPath, f)
			fi, e := os.Stat(filename)
			if e == nil && fi.Size() < 1024*1024*3 {
				t.OutFile, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, t.FilePerm)
				if e != nil {
					fmt.Println("logx:", e)
				} else {
					t.OutFile.Write([]byte("\r\n==================================================\r\n"))
				}
			} else if e != nil {
				fmt.Println("logx:", e)
			}
		}
	}
	if t.OutFile == nil {
		filename := filepath.Join(t.LogPath, t.LogName+`.`+time.Now().Format(`060102_150405`)+`.log`)
		t.OutFile, e = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, t.FilePerm)
	}
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
	if t.OutFile != nil && t.writeCnt < 200 {
		t.writeCnt++
		return nil
	}

	t.muFile.Lock()
	defer t.muFile.Unlock()

	t.writeCnt = 0

	// 检查当前文件的大小
	if t.OutFile != nil {
		fi, e := t.OutFile.Stat()
		if e == nil && fi.Size() < 1024*1024*3 {
			return nil
		}

		t.OutFile.Close()
	}

	// 新文件
	e = t.getFileHandle()
	if e != nil {
		return e
	}

	if t.OutFile == nil {
		return fmt.Errorf("OutFile is nil")
	}
	return nil
}

func (t *Logger) output(level int, format string, v ...interface{}) {
	buf := t.makeStr(level, format, v...)

	if t.OutputFlag&OutputFlag_File != 0 {
		e := t.renewLogFile()
		if e != nil {
			t.ConsoleOutWriter.Write([]byte(e.Error()))
			t.ConsoleOutWriter.Write([]byte("\n"))

			if strings.Contains(e.Error(), "permission denied") {
				t.OutputFlag &= ^OutputFlag_File
			}
		} else {
			t.muFile.Lock()
			t.OutFile.Write(buf)
			t.muFile.Unlock()
		}
	}

	if t.OutputFlag&OutputFlag_Console != 0 {
		t.mu.Lock()
		if t.ConsoleColor {
			switch level {
			case OutputLevel_Debug:
				t.ConsoleOutWriter.Write([]byte("\033[0;39;49m"))
			case OutputLevel_Info:
				t.ConsoleOutWriter.Write([]byte("\033[0;34;49m"))
			case OutputLevel_Warn:
				t.ConsoleOutWriter.Write([]byte("\033[1;33;49m"))
			case OutputLevel_Error:
				t.ConsoleOutWriter.Write([]byte("\033[1;31;49m"))
			}
			t.ConsoleOutWriter.Write(buf)
			t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
		} else {
			t.ConsoleOutWriter.Write(buf)
		}
		t.mu.Unlock()
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func (t *Logger) itoa(buf *[]byte, i int, wid int) {
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

func (t *Logger) makeStr(level int, format string, v ...interface{}) (buf []byte) {
	// level.. [DEBUG]
	if t.PrefixFlag&Llevel != 0 {
		buf = append(buf, '[')
		buf = append(buf, logLevelStr[level]...)
		buf = append(buf, ']', ' ')
	}

	// time.. 2022/02/10 15:00:22
	if t.PrefixFlag&(Ldate|Ltime|Lmicroseconds) != 0 {
		tm := time.Now()
		if t.PrefixFlag&LUTC != 0 {
			tm = tm.UTC()
		}
		if t.PrefixFlag&Ldate != 0 {
			year, month, day := tm.Date()
			t.itoa(&buf, year%100, 2)
			buf = append(buf, '/')
			t.itoa(&buf, int(month), 2)
			buf = append(buf, '/')
			t.itoa(&buf, day, 2)
			buf = append(buf, ' ')
		}
		if t.PrefixFlag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := tm.Clock()
			t.itoa(&buf, hour, 2)
			buf = append(buf, ':')
			t.itoa(&buf, min, 2)
			buf = append(buf, ':')
			t.itoa(&buf, sec, 2)
			if t.PrefixFlag&Lmicroseconds != 0 {
				buf = append(buf, '.')
				t.itoa(&buf, tm.Nanosecond()/1e3, 6)
			}
			buf = append(buf, ' ')
		}
	}

	// logx_test.go:9 (funcName):
	if t.PrefixFlag&(Lshortfile|Llongfile|LfuncName) != 0 {
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
				buf = append(buf, file...)
				buf = append(buf, ':')
				t.itoa(&buf, line, -1)
			}

			if t.PrefixFlag&LfuncName != 0 {
				funcName := runtime.FuncForPC(pc).Name()
				s := strings.Split(funcName, ".")
				funcName = s[len(s)-1]
				buf = append(buf, ' ')
				buf = append(buf, funcName...)
				//buf = append(buf, ')')
			}
			buf = append(buf, ':', ' ')
		}
	}

	// content
	if format == "" {
		buf = append(buf, fmt.Sprint(v...)...)
	} else {
		buf = append(buf, fmt.Sprintf(format, v...)...)
	}

	// limit max length
	if len(buf) > t.LineMaxLength {
		buf = append(buf, buf[:t.LineMaxLength]...)
		buf = append(buf, ' ', '.', '.', '.')
	}

	if len(buf) < 2 || buf[len(buf)-2] != '\r' || buf[len(buf)-1] != '\n' {
		buf = append(buf, '\r', '\n')
	}
	return buf
}
