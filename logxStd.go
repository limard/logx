package logx

import (
	"io"
)

var logStd = NewLogger("", "")

func init() {
	logStd.callSkip = 4
}

// Trace output a [DEBUG] trace string
func Trace() {
	logStd.Trace()
}

// Debug output a [DEBUG] string
func Debug(v ...interface{}) {
	logStd.Debug(v...)
}

// Debugf output a [DEBUG] string with format
func Debugf(format string, v ...interface{}) {
	logStd.Debugf(format, v...)
}

func DebugToJson(v ...interface{}) {
	logStd.DebugToJson(v...)
}

// Info output a [INFO ] string
func Info(v ...interface{}) {
	logStd.Info(v...)
}

// Infof output a [INFO ] string with format
func Infof(format string, v ...interface{}) {
	logStd.Infof(format, v...)
}

// Warn output a [WARN ] string
func Warn(v ...interface{}) {
	logStd.Warn(v...)
}

// Warnf output a [WARN ] string with format
func Warnf(format string, v ...interface{}) {
	logStd.Warnf(format, v...)
}

// Error output a [ERROR] string
func Error(v ...interface{}) {
	logStd.Error(v...)
}

// Errorf output a [ERROR] string with format
func Errorf(format string, v ...interface{}) {
	logStd.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	logStd.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	logStd.Fatalf(format, v...)
}

// SetLogPath set path of output log
func SetLogPath(s string) {
	logStd.LogPath = s
}

func SetExeName(s string) {
	logStd.LogName = s
}

// SetOutputFlag set output purpose(OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView)
func SetOutputFlag(flag int) {
	logStd.OutputFlag = flag
}

// SetOutputLevel set output level.
// OutputLevel_Debug
// OutputLevel_Info
// OutputLevel_Warn
// OutputLevel_Error
// OutputLevel_Fatal
func SetOutputLevel(level int) {
	logStd.OutputLevel = level
}

// SetPrefixFlag set time format(PrefixFlag_Shortfile | PrefixFlag_Date | PrefixFlag_Time)
func SetPrefixFlag(flag int) {
	logStd.PrefixFlag = flag
}

// SetConsoleOut set a writer instead of console
func SetConsoleOut(out io.Writer) {
	logStd.ConsoleOutWriter = out
}

func DefaultLog() *Logger {
	return logStd
}
