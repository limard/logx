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
	logxSTD.debug(v)
}

// Debugf output a [DEBUG] string with format
func Debugf(format string, v ...interface{}) {
	logxSTD.debugf(format, v)
}

// Info output a [INFO ] string
func Info(v ...interface{}) {
	logxSTD.info(v)
}

// Infof output a [INFO ] string with format
func Infof(format string, v ...interface{}) {
	logxSTD.infof(format, v)
}

// Warn output a [WARN ] string
func Warn(v ...interface{}) {
	logxSTD.warn(v)
}

// Warnf output a [WARN ] string with format
func Warnf(format string, v ...interface{}) {
	logxSTD.warnf(format, v)
}

// Error output a [ERROR] string
func Error(v ...interface{}) {
	logxSTD.error(v)
}

// Errorf output a [ERROR] string with format
func Errorf(format string, v ...interface{}) {
	logxSTD.errorf(format, v)
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
