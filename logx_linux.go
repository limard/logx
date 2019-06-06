// +build linux

package logx

import "time"

func getDefaultLogPath() string {
	return `/var/log/rundebug/bis/`
}

var LogSaveTime = 6 * 24 * time.Hour

func outputToDebugView(buf []byte) {
}

func addNewLine(s string) string {
	l := len(s)
	if l == 0 {
		return "\n"
	}
	if s[l-1] != '\n' {
		return s + "\n"
	}
	return s
}
