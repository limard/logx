package logxParse

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestReadFile2(t *testing.T) {
	f, e := os.Open(`C:\ProgramData\PrintSystem\Log\FileTransfer.190626_104258.log`)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()

	p := NewParser2(f)
	e = p.Read(func(data *LogxData) (e error) {
		t.Logf("%s", data.Content)
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestReadStream2(t *testing.T) {
	f, e := os.Open(`C:\ProgramData\PrintSystem\Log\FileTransfer.190626_104258.log`)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()

	buf := bytes.Buffer{}
	go func() {
		b := make([]byte, 128)
		for {
			n, e := f.Read(b)
			if e != nil {
				return
			}
			buf.Write(b[:n])
			time.Sleep(time.Second)
		}
	}()

	p := NewParser2(&buf)
	e = p.Read(func(data *LogxData) (e error) {
		t.Logf("%s", data.Content)
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestNewParser2(t *testing.T) {
	str := "2019/06/25 11:13:31 Server.go:144: [INFO ][Server]port 0.0.0.0:9090 ...\n2019/06/25 11:13:48 httpServer.go:33: [DEBUG][ServeHTTP]1561432428731122600 PUT /file/test/tmp0 HTTP/1.1  Go-http-client/1.1"
	//str := "2019/06/25 11:13:31 Server.go:144: [INFO ][Server]port 0.0.0.0:9090 ..."

	p := NewParser2(bytes.NewBufferString(str))
	e := p.Read(func(data *LogxData) (e error) {
		t.Logf("%+v", data)
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

//Benchmark_parseLine-8    	 1000000	      1536 ns/op
func Benchmark_parseLine(b *testing.B) {
	str := "2019/06/25 11:13:31 Server.go:144: [INFO ][Server]port 0.0.0.0:9090 ..."
	for i := 0; i < b.N; i++ {
		parseLine([]byte(str))
	}
}

//Benchmark_parseLine2-8   	 2000000	       683 ns/op
func Benchmark_parseLine2(b *testing.B) {
	str := "2019/06/25 11:13:31 Server.go:144: [INFO ][Server]port 0.0.0.0:9090 ..."
	for i := 0; i < b.N; i++ {
		parseLine2([]byte(str))
	}
}
