package logxParse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

type parser2 struct {
	r io.Reader
}

func NewParser2(reader io.Reader) (p *parser2) {
	p = &parser2{
		r: reader,
	}
	return p
}

func (t *parser2) Read(f func(data *LogxData) (e error)) error {
	var buf []byte
	tmp := make([]byte, 4096)
	var dataCache *LogxData
	for {
		n, eRead := t.r.Read(tmp)
		buf = append(buf, tmp[:n]...)
		for {
			lastPos := bytes.Index(buf, []byte("\n"))
			if lastPos == -1 {
				if eRead != io.EOF {
					break
				}
				lastPos = len(buf)
				if lastPos == 0 {
					fmt.Println("err:", eRead)
					return nil
				}
			}

			data, ep := parseLine2(buf[:lastPos])
			if ep == nil {
				if dataCache != nil {
					if ep = f(dataCache); ep != nil {
						return ep
					}
				}
				dataCache = data
			} else {
				dataCache.Content += string(buf[:lastPos])
			}

			if eRead != nil && eRead == io.EOF {
				if dataCache != nil {
					if ep = f(dataCache); ep != nil {
						return ep
					}
				}
				fmt.Println("err:", eRead)
				return nil
			}
			buf = buf[lastPos+1:]
		}
	}
}

func parseLine(line []byte) (data *LogxData, e error) {
	parts := bytes.Split(line, []byte(" "))
	if len(parts) < 4 {
		e = fmt.Errorf("error format (parts)")
		return
	}
	parts2 := bytes.Split(parts[2], []byte(":"))
	if len(parts2) < 3 {
		e = fmt.Errorf("error format (parts2)")
		return
	}
	parts[3] = bytes.Join(parts[3:], []byte(" "))
	parts3 := bytes.Split(parts[3], []byte("]"))
	if len(parts3) < 3 {
		e = fmt.Errorf("error format (parts3)")
		return
	}
	date, e := time.Parse("2006/01/02 15:04:05", fmt.Sprintf("%s %s", parts[0], parts[1]))
	if len(parts3) < 3 {
		e = fmt.Errorf("error format (%v)", e)
		return
	}
	lineNo, e := strconv.Atoi(string(parts2[1]))
	if len(parts3) < 3 {
		e = fmt.Errorf("error format (%v)", e)
		return
	}
	data = &LogxData{
		Date:         date,
		FileName:     string(parts2[0]),
		LineNo:       lineNo,
		Level:        string(bytes.TrimSpace(parts3[0][1:])),
		FunctionName: string(bytes.TrimSpace(parts3[1][1:])),
		Content:      string(parts3[2]),
	}
	return
}

func parseLine2(line []byte) (data *LogxData, e error) {
	data = &LogxData{}

	if len(line) < 19 {
		e = errors.New("error format")
		return
	}

	data.Date, e = time.Parse("2006/01/02 15:04:05", string(line[:19]))
	if e != nil {
		return
	}
	line = line[19+1:]

	n := bytes.Index(line, []byte(":"))
	if n == -1 {
		e = errors.New("error format")
		return
	}
	data.FileName = string(bytes.TrimSpace(line[:n]))
	line = line[n+1:]

	n = bytes.Index(line, []byte(":"))
	if n == -1 {
		e = errors.New("error format")
		return
	}
	data.LineNo, _ = strconv.Atoi(string(bytes.TrimSpace(line[:n])))
	line = line[n+1:]

	n1 := bytes.Index(line, []byte("["))
	n2 := bytes.Index(line, []byte("]"))
	if n1 == -1 || n2 == -1 {
		e = errors.New("error format")
		return
	}
	data.Level = string(line[n1+1 : n2-n1])
	line = line[n2+1:]

	n1 = bytes.Index(line, []byte("["))
	n2 = bytes.Index(line, []byte("]"))
	if n1 == -1 || n2 == -1 {
		e = errors.New("error format")
		return
	}
	data.FunctionName = string(line[n1+1 : n2-n1])
	line = line[n2+1:]

	data.Content = string(line)
	return
}
