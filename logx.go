package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

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
	Ldate         = 1 << iota // the date in the local time zone: 09/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes PrefixFlag_Time.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides PrefixFlag_Longfile
	LUTC                      // if PrefixFlag_Date or PrefixFlag_Time is set, use UTC rather than the local time zone
	LfuncName
	Llevel
	LstdFlags = Lshortfile | Ldate | Ltime | LfuncName | Llevel | Lmicroseconds
)

// Logger struct
type Logger struct {
	LastError        error       // LOGX运行错误
	FilePerm         os.FileMode // 日志文件的文件权限 默认666
	LineMaxLength    int         // 一行最大的长度，-1为不限制
	LogPath          string      // 日志的保存目录
	LogName          string      // 日志的文件名，默认为程序名
	OutputFlag       int         // 日志输出位置 OutputFlag_File Console
	OutputLevel      int         // 输出级别 OutputLevel_Debug Info Warn Error Fatal
	PrefixFlag       int         // 日志的前缀信息 Ldate L...
	MaxLogNumber     int         // 日志文件保存个数
	MaxFileSize      int64       // 日志文件最大大小（字节）
	ContinuousLog    bool        // 是否连续在上一个文件中输出，适用于经常被调用启动的程序日志 默认是
	ConsoleOutWriter io.Writer   // 可重定向到父进程中
	ConsoleColor     bool        // 在控制台输出时，Warn和Error是否加重颜色标识

	mutexConsoleOutWriter sync.Mutex // 保护向控制台写数据时不乱序
	logCounter            int        // 记录写入次数
	mutexFile             sync.Mutex // 保护outFile
	outFile               *os.File   // 当前输出的文件
	callSkip              int        //
	bufferPool            sync.Pool  // 数据缓冲池
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
		ConsoleOutWriter: os.Stdout,
		ConsoleColor:     true,
		logCounter:       0,
		callSkip:         3,
		bufferPool: sync.Pool{
			New: func() interface{} { return &bytes.Buffer{} },
		},
	}

	executable, _ := os.Executable()
	if len(l.LogPath) == 0 {
		if runtime.GOOS == "linux" {
			l.LogPath = `/var/log/`
		} else {
			l.LogPath = filepath.Dir(executable)
		}
	}

	if len(l.LogName) == 0 {
		l.LogName = filepath.Base(executable)
	}

	// read json configuration
	buf, e := ioutil.ReadFile(filepath.Join(l.LogPath, "log.json"))
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

// 兼容io.Writer
func (t *Logger) Write(b []byte) (n int, err error) {
	if t.OutputLevel > OutputLevel_Debug {
		return
	}
	t.inner_output(OutputLevel_Debug, bytes.NewBuffer(b))
	return len(b), nil
}

// go官方自带的log
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

///////////

// 判断所给路径文件/文件夹是否存在
func (t *Logger) isFileExists(path string) bool {
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
		t.outFile = f
		return nil
	}

	// 删除最老的文件
	lastFileName := makeFileName(t.MaxLogNumber)
	if t.isFileExists(lastFileName) {
		e = os.Remove(lastFileName)
		if e != nil {
			fmt.Fprintf(os.Stderr, "logx: delete old log file %s failed\n", e.Error())
			t.LastError = e
		}
	}

	// 文件名逐个后移
	for i := t.MaxLogNumber; i >= 1; i-- {
		from := makeFileName(i - 1)
		to := makeFileName(i)
		e := os.Rename(from, to)
		if e != nil {
			fmt.Fprintf(os.Stderr, "logx: rename old log file %s to %s failed: %s\n", from, to, e.Error())
			t.LastError = e
		}
	}

	newFileName := makeFileName(0)
	t.outFile, e = os.OpenFile(newFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, t.FilePerm)
	if e != nil {
		fmt.Println("logx:", e)
		t.LastError = e
		return e
	}
	return nil
}

func (t *Logger) renewLogFile() (e error) {
	if t.outFile != nil && t.logCounter < 200 {
		t.logCounter++
		return nil
	}
	t.logCounter = 1

	t.mutexFile.Lock()
	defer t.mutexFile.Unlock()

	if t.outFile == nil {
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	fi, _ := t.outFile.Stat()
	if fi.Size() > t.MaxFileSize {
		t.outFile.Close()
		t.outFile = nil
		e = t.getFileHandle()
		if e != nil {
			return e
		}
	}

	if t.outFile == nil {
		return fmt.Errorf("OutFile is nil")
	}
	return nil
}

func (t *Logger) output(level int, format string, v ...interface{}) {
	buf := t.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		t.bufferPool.Put(buf)
	}()

	buf.Reset()
	t.makeStr(buf, level, format, v...)

	t.inner_output(level, buf)
}
func (t *Logger) inner_output(level int, buf *bytes.Buffer) {
	if t.OutputFlag&OutputFlag_File != 0 {
		e := t.renewLogFile()
		if e != nil {
			_, _ = t.ConsoleOutWriter.Write([]byte(e.Error()))
			_, _ = t.ConsoleOutWriter.Write([]byte("\n"))

			if strings.Contains(e.Error(), "permission denied") {
				t.OutputFlag &= ^OutputFlag_File
			}
		} else {
			t.mutexFile.Lock()
			_, _ = t.outFile.Write(buf.Bytes())
			t.mutexFile.Unlock()
		}
	}

	if t.OutputFlag&OutputFlag_Console != 0 {
		t.mutexConsoleOutWriter.Lock()
		if t.ConsoleColor {
			switch level {
			case OutputLevel_Debug:
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
			case OutputLevel_Info:
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
			case OutputLevel_Warn:
				_, _ = t.ConsoleOutWriter.Write([]byte("\033[1;33;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				_, _ = t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			case OutputLevel_Error:
				_, _ = t.ConsoleOutWriter.Write([]byte("\033[1;31;49m"))
				_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
				_, _ = t.ConsoleOutWriter.Write([]byte("\u001B[0m"))
			}
		} else {
			_, _ = t.ConsoleOutWriter.Write(buf.Bytes())
		}
		t.mutexConsoleOutWriter.Unlock()
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

	// logx_test.go:9 funcName:
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

	ensureLineEnding(buf)
}
