package logx

import "fmt"

func Debug(v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[DEBUG]%s`, fmt.Sprint(v...)))
}

func Debugf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[DEBUG]`+format, v...))
}

func Info(v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[INFO ]%s`, fmt.Sprint(v...)))
}

func Infof(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[INFO ]`+format, v...))
}

func Warn(v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[WARN ]%s`, fmt.Sprint(v...)))
}

func Warnf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[WARN ]`+format, v...))
}

func Error(v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[ERROR]%s`, fmt.Sprint(v...)))
}

func Errorf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[ERROR]`+format, v...))
}

func Unexpected(v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[UNEXP]%s`, fmt.Sprint(v...)))
}

func Unexpectedf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(`[UNEXP]`+format, v...))
}

