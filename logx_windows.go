package logx

import (
	"bytes"
)

func ensureLineEndingf(format string, buf *bytes.Buffer) {
	if len(format) < 2 {
		buf.Write([]byte{'\r', '\n'})
		return
	}

	// 如果已经以\r\n结尾，不需要修改
	length := len(format)
	if length >= 2 && format[length-2] == '\r' && format[length-1] == '\n' {
		return
	}

	// 如果不以换行符结尾，添加\r\n
	buf.Write([]byte{'\r', '\n'})
}

func ensureLineEnding(buf *bytes.Buffer) {
	if buf.Len() >= 2 && buf.Bytes()[buf.Len()-2] == '\r' && buf.Bytes()[buf.Len()-1] == '\n' {
		return
	}

	buf.Write([]byte{'\r', '\n'})
}
