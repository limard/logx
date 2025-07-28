package logx

import "bytes"

func ensureLineEnding(buf *bytes.Buffer) {
	if buf.Len() == 0 {
		return
	}

	data := buf.Bytes()
	length := len(data)

	// 如果已经以\r\n结尾，不需要修改
	if length >= 2 && data[length-2] == '\r' && data[length-1] == '\n' {
		return
	}

	// 如果以\n结尾但没有\r，替换为\r\n
	if data[length-1] == '\n' {
		// 将最后的\n替换为\r\n
		buf.Truncate(length - 1)
		buf.WriteString("\r\n")
		return
	}

	// 如果不以换行符结尾，添加\r\n
	buf.WriteString("\r\n")
}
