package logx

import (
	"testing"
	"fmt"
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