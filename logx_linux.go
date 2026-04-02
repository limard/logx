package logx

import "bytes"

func ensureLineEndingf(format string, buf *bytes.Buffer) {
	if format == "" {
		buf.WriteByte('\n')
		return
	}

	// 如果已经以\n结尾
	length := len(format)
	if format[length-1] == '\n' {
		return
	}

	// 如果不以换行符结尾，添加\n
	buf.WriteByte('\n')
}

func ensureLineEnding(buf *bytes.Buffer) {
	if buf.Len() >= 1 && buf.Bytes()[buf.Len()-1] == '\n' {
		return
	}
	buf.Write([]byte{'\n'})
}
