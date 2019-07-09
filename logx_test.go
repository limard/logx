package logx

import (
	"fmt"
	"testing"
)

func TestTrace(t *testing.T) {
	Trace()
	Debugf("Debug %v %v", "123", "456")

	l := New("", "Logx.Obj")
	l.Trace()
	l.Debugf("Debug %v %v", "123", "456")
}

func Test111(t *testing.T) {
	t.Log(d("Debug %v %v", "123", "456"))
	t.Log(d2("Debug %v %v", "123", "456"))
}

func d2(f string, v ...interface{}) string {
	return d(f, v...)
}

func d(f string, v ...interface{}) string {
	return fmt.Sprintf(f, v...)
}

func TestSort(t *testing.T) {
	files := []string{"bipkg.exe.181015_111646.log", "bipkg.exe.181015_111600.log", "bipkg.exe.181015_111518.log",
		"bipkg.exe.181015_111510.log", "bipkg.exe.181015_111349.log",
		"bipkg.exe.181015_111504.log", "bipkg.exe.181015_111438.log"}
	files = logxSTD.getNeedDeleteLogfile(files)
	t.Log(files)
}

func TestLoggerx_DebugToJson(t *testing.T) {
	type WE struct {
		A string
	}
	we := WE{"123qwe"}
	DebugToJson("SQ", we)
}

func TestLogx(t *testing.T) {
	Debug("123", "456", "789")
	Debugf("123 %s %s", "456", "789")
}

func TestCleanFile(t *testing.T) {
	l := New("", "testCleanFile")
	l.SetOutputFlag(OutputFlag_File)
	for i := 0; i < 1000*1000; i++ {
		l.Debug("1234567890qwertyuiopasdfghjklzxcvbnm,./[pljugftrdr4sdrtygfvhjkjvbbnghk")
	}
}
