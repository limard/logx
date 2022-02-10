package logx

import (
	"testing"
)

func TestTrace(t *testing.T) {
	Trace()
	Debugf("Debug %v %v", "123", "456")
	Info("6789")
	Warn("5tyui")
	Errorf("%d", 2355)

	l := New("", "")
	l.PrefixFlag = PrefixFlag_level | PrefixFlag_Time
	l.Trace()
	l.Debugf("Debug %v %v", "123", "456")
	l.Error("ERROROOOOO")
}

func TestSort(t *testing.T) {
	files := []string{"bipkg.exe.181015_111646.log", "bipkg.exe.181015_111600.log", "bipkg.exe.181015_111518.log",
		"bipkg.exe.181015_111510.log", "bipkg.exe.181015_111349.log",
		"bipkg.exe.181015_111504.log", "bipkg.exe.181015_111438.log"}
	files = logStd.getNeedDeleteLogfile(files)
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
	l.OutputFlag = OutputFlag_File
	l.LineMaxLength = 1024
	for i := 0; i < 100000; i++ {
		l.Debug("1234567890qwertyuiopasdfghjklzxcvbnm,./[pljugftrdr4sdrtygfdsssssssssssssssssssddddddddddddddddddddddfasdlqamdlmkwlqmkdwmqklmdkwlqmlkdmwkmdklwqmdklqmwkdwqmdklwmkldqmkwmdkqlwmlkdqmlkwdmqlkmdlkqmwlkdmkmlkmkhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddvhjkjvbbnghk")
	}
}

//BenchmarkName-8   	  730000	      1554 ns/op
func BenchmarkSpeed(b *testing.B) {
	SetOutputFlag(0)
	for i := 0; i < b.N; i++ {
		Debug("123")
	}
}
