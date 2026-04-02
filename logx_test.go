package logx

import (
	"testing"
)

func TestTrace(t *testing.T) {
	log := NewLogger("log", "text")
	log.Trace()
	log.Debugf("Debug %v %v", "123", "456")
	log.Info("6789")
	log.Warn("5tyui")
	log.Errorf("%d", 2355)

	l := NewLogger("", "")
	l.PrefixFlag = Llevel | Ltime
	l.Trace()
	l.Debugf("Debug %v %v", "123", "456")
	l.Error("ERROROOOOO")
}

func TestCleanFile(t *testing.T) {
	l := NewLogger("", "testCleanFile")
	l.OutputFlag = OutputFlag_File
	l.LineMaxLength = 1024
	for i := 0; i < 100000; i++ {
		l.Debug("1234567890qwertyuiopasdfghjklzxcvbnm,./[pljugftrdr4sdrtygfdsssssssssssssssssssddddddddddddddddddddddfasdlqamdlmkwlqmkdwmqklmdkwlqmlkdmwkmdklwqmdklqmwkdwqmdklwmkldqmkwmdkqlwmlkdqmlkwdmqlkmdlkqmwlkdmkmlkmkhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddvhjkjvbbnghk")
	}
}

// 1445377               883.2 ns/op           248 B/op          2 allocs/op
func BenchmarkSpeed(b *testing.B) {
	log := NewLogger("log", "temp_test_log")
	log.OutputFlag = 0
	for i := 0; i < b.N; i++ {
		log.Debug("1234567890")
	}
}
