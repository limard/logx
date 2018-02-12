// +build linux

package logx

import "time"

func getBisPath() string {
	//s := `/opt/PrintSystem/Log/`
	return `/var/log/rundebug/bis/`
}

var LogSaveTime = 6*24*time.Hour

func outputToDebugView(buf []byte) {
}

func output(s string) {
	l := len(s)
	if l > 1 {
		if s[l-1:] != "\n" {
			s += "\n"
		}
	}

	if outputFlag&OutputFlag_File != 0 {
		renewLogFile()
		logFile.Output(3, s)
	}

	if outputFlag&OutputFlag_Console != 0 {
		if len(consoleOutPrefix) != 0 {
			hConsoleOut.Write(consoleOutPrefix)
		}
		hConsoleOut.Write([]byte(s))
	}

	if outputFlag&OutputFlag_DbgView != 0 {
		outputToDebugView([]byte("[BIS]" + s))
	}
}