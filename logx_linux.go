package logx

import "bytes"

func ensureLineEnding(buf *bytes.Buffer) {
	if buf.Len() == 0 {
		return
	}

	data := buf.Bytes()
	length := len(data)

	// 如果已经以\n结尾
	if data[length-1] == '\n' {
		// 如果是\r\n结尾，替换为\n
		if length >= 2 && data[length-2] == '\r' {
			buf.Truncate(length - 2)
			buf.WriteByte('\n')
		}
		// 如果只是\n结尾，不需要修改
		return
	}

	// 如果不以换行符结尾，添加\n
	buf.WriteByte('\n')
}
