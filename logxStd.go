package logx

import (
	"fmt"
	"io"
)

var logxSTD = New("", "")

// Trace output a [DEBUG] trace string
func Trace() {
	logxSTD.trace()
}

// Debug output a [DEBUG] string
func Debug(v ...interface{}) {
	logxSTD.Debug(v)
}

// Debugf output a [DEBUG] string with format
func Debugf(format string, v ...interface{}) {
	logxSTD.Debugf(format, v)
}

// Info output a [INFO ] string
func Info(v ...interface{}) {
	logxSTD.Info(v)
}

// Infof output a [INFO ] string with format
func Infof(format string, v ...interface{}) {
	logxSTD.Infof(format, v)
}

// Warn output a [WARN ] string
func Warn(v ...interface{}) {
	logxSTD.Warn(v)
}

// Warnf output a [WARN ] string with format
func Warnf(format string, v ...interface{}) {
	logxSTD.Warnf(format, v)
}

// Error output a [ERROR] string
func Error(v ...interface{}) {
	logxSTD.Error(v)
}

// Errorf output a [ERROR] string with format
func Errorf(format string, v ...interface{}) {
	logxSTD.Errorf(format, v)
}

// SetLogPath set path of output log
func SetLogPath(s string) {
	logxSTD.SetLogPath(s)
}

// SetOutputFlag set output purpose(OutputFlag_File | OutputFlag_Console | OutputFlag_DbgView)
func SetOutputFlag(flag int) {
	logxSTD.SetOutputFlag(flag)
}

// SetOutputLevel set output level.
// OutputLevel_Debug
// OutputLevel_Info
// OutputLevel_Warn
// OutputLevel_Error
// OutputLevel_Unexpected
func SetOutputLevel(level int) {
	logxSTD.output(fmt.Sprintf("Log Level: %v Flag: %v", level, logxSTD.outputFlag))
	logxSTD.SetOutputLevel(level)
}

// SetTimeFlag set time format(Lshortfile | Ldate | Ltime)
func SetTimeFlag(flag int) {
	logxSTD.SetTimeFlag(flag)
}

// SetConsoleOut set a writer instead of console
func SetConsoleOut(out io.Writer) {
	logxSTD.SetConsoleOut(out)
}

// SetConsoleOutPrefix set prefix for console output
func SetConsoleOutPrefix(prefix []byte) {
	logxSTD.SetConsoleOutPrefix(prefix)
}
