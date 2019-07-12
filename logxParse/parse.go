package logxParse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
)

type parser struct {
	r *bufio.Reader
}

func NewParser(reader io.Reader) (p *parser) {
	p = &parser{
		r: bufio.NewReader(reader),
	}
	return p
}

func (t *parser) Read(f func(data *LogxData) (e error)) (e error) {
	var data *LogxData
	for {
		line, _, e := t.r.ReadLine()
		if e != nil {
			if data != nil {
				if e1 := f(data); e1 != nil {
					return e1
				}
				data = nil
			}
			//if e == io.EOF {
			//	e = nil
			//}
			return e
		}
		parts := bytes.Split(line, []byte(" "))
		if len(parts) < 4 {
			if data != nil {
				data.Content = fmt.Sprintf("%s\r\n%s", data.Content, string(line))
			}
			continue
		}

		date, e := time.Parse("2006/01/02 15:04:05", fmt.Sprintf("%s %s", parts[0], parts[1]))
		parts2 := bytes.Split(parts[2], []byte(":"))
		if len(parts2) < 3 {
			if data != nil {
				data.Content = fmt.Sprintf("%s\r\n%s", data.Content, string(line))
			}
			continue
		}

		lineNo, _ := strconv.Atoi(string(parts2[1]))

		parts[3] = bytes.Join(parts[3:], []byte(" "))
		parts3 := bytes.Split(parts[3], []byte("]"))
		if len(parts3) < 3 {
			if data != nil {
				data.Content = fmt.Sprintf("%s\r\n%s", data.Content, string(line))
			}
			continue
		}

		if data != nil {
			if e = f(data); e != nil {
				return e
			}
			data = nil
		}

		data = &LogxData{
			Date:         date,
			FileName:     string(parts2[0]),
			LineNo:       lineNo,
			Level:        string(bytes.TrimSpace(parts3[0][1:])),
			FunctionName: string(bytes.TrimSpace(parts3[1][1:])),
			Content:      string(parts3[2]),
		}
	}
}
